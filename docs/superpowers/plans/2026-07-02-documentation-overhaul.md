# PIGO8 Documentation Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the PIGO8 mdBook docs site into an exhaustive, accurate, visually-demonstrated reference: fix broken/dead content, add a complete API Reference generated from the real source, add FAQ/Testing pages, and build a headless-Chrome tool that captures real screenshots/GIFs of the example games and embeds them across the docs.

**Architecture:** Seven sequential phases, each producing a working, committable increment: (1) content audit/fixes, (2) generated API Reference under `docs/reference/api/`, (3) two new pages (FAQ, Testing), (4) a new `cmd/docshots` tool built on a refactored, shared `internal/webbuild` package (extracted from `cmd/webexport`) plus `chromedp` for headless capture, (5) generating and embedding the visuals, (6) retrofitting consistency (`See Also` footers, `SUMMARY.md`, light branding), (7) final validation.

**Tech Stack:** Go 1.25 (existing module), `go/doc` for API extraction (used manually via `go doc`, not scripted), `github.com/chromedp/chromedp` (new dependency) for headless browser capture, Go stdlib `image/gif` + `image/draw` for GIF assembly, mdBook (existing) for the site build.

---

## Phase 1: Content Audit & Fixes

### Task 1.1: Build `cmd/linkcheck` - a reusable internal doc link checker

**Files:**
- Create: `cmd/linkcheck/main.go`
- Create: `cmd/linkcheck/main_test.go`

This tool scans all `.md` files under `docs/`, extracts markdown links of the form `[text](target)`, and verifies that any relative (non-URL) target resolves to a real file relative to the linking file's directory. It ignores external links (`http://`, `https://`) and in-page anchors (`#...`).

- [ ] **Step 1: Write the failing test**

```go
// cmd/linkcheck/main_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	content := "See [Installation](01-installation.md) and [broken](nope.md) and [ext](https://example.com) and [anchor](#section)."
	links := extractLinks(content)
	want := []string{"01-installation.md", "nope.md", "https://example.com", "#section"}
	if len(links) != len(want) {
		t.Fatalf("got %d links, want %d: %v", len(links), len(want), links)
	}
	for i, w := range want {
		if links[i] != w {
			t.Errorf("link %d: got %q, want %q", i, links[i], w)
		}
	}
}

func TestCheckFileBrokenLinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "target.md"), []byte("# ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "source.md")
	body := "[good](target.md) and [bad](missing.md) and [ext](https://example.com)"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	broken, err := checkFile(src)
	if err != nil {
		t.Fatalf("checkFile: %v", err)
	}
	if len(broken) != 1 || broken[0] != "missing.md" {
		t.Fatalf("got broken links %v, want [missing.md]", broken)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/linkcheck/... -run TestExtractLinks -v`
Expected: FAIL with "undefined: extractLinks" (package doesn't have the functions yet).

- [ ] **Step 3: Write minimal implementation**

```go
// cmd/linkcheck/main.go

// Package main implements a documentation link checker for the PIGO8 mdBook site.
// It walks docs/ for markdown files, extracts relative links and image references,
// and reports any that don't resolve to a real file.
//
// Usage:
//
//	go run ./cmd/linkcheck -dir docs
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var linkPattern = regexp.MustCompile(`\]\(([^)]+)\)`)

// extractLinks returns every link target found inside markdown link/image
// syntax `[text](target)` or `![alt](target)`, in the order they appear.
func extractLinks(content string) []string {
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	links := make([]string, 0, len(matches))
	for _, m := range matches {
		links = append(links, strings.TrimSpace(m[1]))
	}
	return links
}

// isCheckable reports whether a link target should be verified against the
// filesystem (relative paths only; external URLs and in-page anchors are skipped).
func isCheckable(target string) bool {
	if target == "" {
		return false
	}
	if strings.HasPrefix(target, "#") {
		return false
	}
	if strings.Contains(target, "://") {
		return false
	}
	if strings.HasPrefix(target, "mailto:") {
		return false
	}
	return true
}

// checkFile reads the markdown file at path and returns any link targets that
// don't resolve to an existing file relative to the file's directory. Anchor
// fragments (e.g. "page.md#section") are stripped before resolution.
func checkFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	var broken []string
	for _, link := range extractLinks(string(data)) {
		if !isCheckable(link) {
			continue
		}
		target := link
		if idx := strings.Index(target, "#"); idx >= 0 {
			target = target[:idx]
		}
		if target == "" {
			continue
		}
		full := filepath.Join(dir, target)
		if _, err := os.Stat(full); err != nil {
			broken = append(broken, link)
		}
	}
	return broken, nil
}

func main() {
	dir := flag.String("dir", "docs", "Root directory of markdown files to check")
	flag.Parse()

	var totalBroken int
	err := filepath.WalkDir(*dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		broken, err := checkFile(path)
		if err != nil {
			return err
		}
		for _, b := range broken {
			fmt.Printf("%s: broken link -> %s\n", path, b)
			totalBroken++
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking %s: %v\n", *dir, err)
		os.Exit(1)
	}

	if totalBroken > 0 {
		fmt.Printf("\n%d broken link(s) found.\n", totalBroken)
		os.Exit(1)
	}
	fmt.Println("No broken links found.")
}
```

Note: `filepath.WalkDir` requires the `os.DirEntry` type (Go 1.16+), already satisfied by this module's Go 1.25.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/linkcheck/... -v`
Expected: PASS for both `TestExtractLinks` and `TestCheckFileBrokenLinks`.

- [ ] **Step 5: Run against the real docs to find current breakage**

Run: `go run ./cmd/linkcheck -dir docs`
Expected output includes (at minimum):
```
docs/editor.md: broken link -> installing_go.md
docs/editor.md: broken link -> embedding.md
```

- [ ] **Step 6: Commit**

```bash
git add cmd/linkcheck
git commit -m "feat: add docs link checker tool"
```

### Task 1.2: Fix the broken links found in `docs/editor.md`

**Files:**
- Modify: `docs/editor.md`

- [ ] **Step 1: Fix the two broken links**

In `docs/editor.md`, change:

```markdown
To install the PIGO8 editor, you need to have Go installed on your system. If you haven't installed Go yet, please refer to the [Installing Go](installing_go.md) guide.
```

to:

```markdown
To install the PIGO8 editor, you need to have Go installed on your system. If you haven't installed Go yet, please refer to the [Installation](01-installation.md) guide.
```

And change:

```markdown
After creating your sprites and maps with the editor, you can use them in your PIGO8 games. See the [Resource Embedding](embedding.md) guide for details on how to include these resources in your game.
```

to:

```markdown
After creating your sprites and maps with the editor, you can use them in your PIGO8 games. See the [Resource Embedding](90-advanced/01-resource-embedding.md) guide for details on how to include these resources in your game.
```

- [ ] **Step 2: Verify the fix**

Run: `go run ./cmd/linkcheck -dir docs`
Expected: `No broken links found.` (assuming no other breakage exists at this point - Task 1.1 Step 5 confirmed these were the only two).

- [ ] **Step 3: Commit**

```bash
git add docs/editor.md
git commit -m "fix: correct broken links in editor docs"
```

### Task 1.3: Confirm no docs reference the deleted profiling internals

The recent cleanup commit (`31d2d44`) deleted `frame_metrics.go` and `perf_profiler.go`, which contained now-removed symbols such as `GetPerformanceProfiler`, `PerformanceProfiler`, `ProfilerSection`, `ProfileSection`, `ProfileFrame`, `EnableProfiling`, `GetProfilingReport`, `PrintProfilingReport`. Frame-stats functions like `EnableFrameStats`, `GetCurrentFrameNumber`, and `GetPerformanceReport` still exist (moved into `metrics.go`/`engine_integration.go`), so they are not dead references.

- [ ] **Step 1: Search for references to the removed symbols**

Run:
```bash
grep -rniE "PerformanceProfiler|ProfilerSection|ProfileSection|ProfileFrame|EnableProfiling|GetProfilingReport|PrintProfilingReport" docs/ README.md
```

Expected: no output (these were internal implementation details never covered in the public docs).

- [ ] **Step 2: If any matches are found, remove or rewrite them**

For each match, replace the reference with the current equivalent public API (`GetPerformanceReport()` from `engine_integration.go`, `EnableFrameStats`/`IsFrameStatsEnabled` from the still-present frame-stats functions), or delete the mention if it describes removed functionality with no replacement.

- [ ] **Step 3: Commit (only if Step 2 made changes)**

```bash
git add -A
git commit -m "fix: remove references to deleted profiling internals"
```

---

## Phase 2: API Reference

Create a new `docs/reference/api/` section covering the curated public game-development API - the same functions already covered by the existing conceptual docs and cheatsheet - documented exhaustively with parameter tables and examples. Internal/advanced symbols not intended for typical game code (caching internals like `LRUCache`, `SpriteHashTable`, `MultiTierBufferPool`, `ResourceManager`, `PixelBatchSystem`, `SpriteImageCache`, `SpritePixelCache`, `MetricsCollector`, and low-level tuning knobs like `SetRenderConfig`/`ResourceLimits`) are deliberately excluded from the curated reference and instead called out briefly in the overview page, since dumping every exported symbol (there are ~90) would bury the actual game-making API that users need under implementation details.

### Task 2.1: Write the API Reference overview page

**Files:**
- Create: `docs/reference/api/00-overview.md`

- [ ] **Step 1: Create the file**

```markdown
# API Reference

This section documents every public function, type, and constant in the PIGO8 game-development
API - organized by domain, with full parameter descriptions and runnable examples.

If you want a fast one-line-per-function lookup instead, see the [Cheatsheet](../cheatsheet.md).
This reference is the deep version: full explanations, parameter tables, and examples for each
function.

## How to Read an Entry

Each function entry follows this format:

- **Signature** - the exact Go function signature, including generic type parameters where used.
- **Description** - what it does and how it behaves, including edge cases (out-of-range
  coordinates, invalid color indices, etc.).
- **Parameters** - a table of every parameter, its type, and its meaning.
- **Example** - a runnable snippet showing typical usage.
- **See also** - links to the matching conceptual guide page.

## Generic Coordinate Types

Many drawing functions (`Spr`, `Line`, `Rect`, `Circ`, `Mget`, `Sget`, and others) accept
coordinates as Go generics constrained to a `Number` type, meaning you can pass `int`,
`float64`, or any other standard numeric type interchangeably without manual conversion:

```go
p8.Spr(1, 10, 20)         // int coordinates
p8.Spr(1, 10.5, 20.0)     // float64 coordinates
p8.Spr(1, 10, 20.0)       // mixed - also valid
```

## Advanced / Internal APIs

PIGO8 also exports a number of lower-level types used internally for performance optimization:
sprite and pixel caches (`SpriteImageCache`, `SpritePixelCache`, `LRUCache`), batching systems
(`PixelBatchSystem`, `SpriteBatchSystem`), resource tracking (`ResourceManager`,
`ResourceLimits`), and render tuning (`RenderConfig`, `SetRenderConfig`). These exist to support
the engine's own rendering pipeline and advanced tuning scenarios, not everyday game code. They
are intentionally not covered in this reference; if you need them, browse the source directly
with `go doc github.com/drpaneas/pigo8 <Name>`.

## Reference Pages

- [Screen & Drawing](01-screen-drawing.md)
- [Colors & Palette](02-colors-palette.md)
- [Sprites](03-sprites.md)
- [Maps](04-maps.md)
- [Input](05-input.md)
- [Audio](06-audio.md)
- [Camera](07-camera.md)
- [Collision](08-collision.md)
- [Math](09-math.md)
- [Settings & Lifecycle](10-settings-lifecycle.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/00-overview.md
git commit -m "docs: add API reference overview page"
```

### Task 2.2: Write the Screen & Drawing reference page

**Files:**
- Create: `docs/reference/api/01-screen-drawing.md`

- [ ] **Step 1: Create the file using the real signatures/descriptions below**

Use this exact source material (captured via `go doc . <Name>` against the current codebase) to
write each entry. Follow the template shown for `Cls` and `Pset` exactly for the remaining
functions (`ClsRGBA`, `Pget`, `Line`, `Rect`, `Rectfill`, `Circ`, `Circfill`, `Print`, `Cursor`,
`GetScreenWidth`, `GetScreenHeight`).

```markdown
# Screen & Drawing

## `Cls(colorIndex ...int)`

Clears the current drawing screen with a specified PICO-8 color index. If no `colorIndex` is
provided, it defaults to 0 (black).

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `colorIndex` | `int` (optional, variadic) | PICO-8 color index (0-15). Defaults to 0 (black). |

**Example:**
```go
p8.Cls()     // clear to black
p8.Cls(12)   // clear to blue
```

**See also:** [Screen](../../10-graphics/01-screen.md)

## `ClsRGBA(clr color.RGBA)`

Clears the current drawing screen with a specified RGBA color instead of a PICO-8 palette index,
allowing any RGBA color to be used.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `clr` | `color.RGBA` | The RGBA color to clear the screen with. |

**Example:**
```go
p8.ClsRGBA(color.RGBA{R: 100, G: 150, B: 200, A: 255}) // custom blue
p8.ClsRGBA(color.RGBA{})                               // transparent black
```

**See also:** [Screen](../../10-graphics/01-screen.md)

## `Pset(x, y int, colorIndex ...int)`

Draws a single pixel at coordinates `(x, y)` using the specified PICO-8 color index, or the
current cursor color if omitted. `Pset` uses raw **screen** coordinates - unlike sprites and
shapes, it is not affected by the camera offset. Out-of-bounds coordinates or invalid color
indices are silently ignored (with a logged warning for invalid colors).

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `int` | Screen pixel coordinates. |
| `colorIndex` | `int` (optional, variadic) | PICO-8 color index (0-15). Defaults to the current cursor color. |

**Example:**
```go
p8.Cursor(0, 0, 8)   // set current color to red
p8.Pset(10, 20)      // draws a red pixel at (10, 20)
p8.Pset(50, 50, 12)  // draws a blue pixel at (50, 50), overriding the cursor color
```

**See also:** [Pixels](../../10-graphics/03-pixels.md)

## `Pget(x, y int) int`

Returns the PICO-8 color index (0-15) of the pixel at `(x, y)`. Returns 0 (black) if the
coordinates are out of bounds or the pixel doesn't exactly match a palette color.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `int` | Screen pixel coordinates to sample. |

**Example:**
```go
p8.Pset(10, 20, 8)
idx := p8.Pget(10, 20) // idx == 8
```

**See also:** [Pixels](../../10-graphics/03-pixels.md)

## `Line[X1, Y1, X2, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{})`

Draws a line between two points. Mimics PICO-8's `line(x1, y1, x2, y2, color)`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x1`, `y1` | `Number` | Starting point coordinates. |
| `x2`, `y2` | `Number` | Ending point coordinates. |
| `color` (via `options`) | `int` (optional) | PICO-8 color index (0-15). Defaults to the current draw color (7, white). |

**Example:**
```go
p8.Line(0, 0, 127, 127, 8) // diagonal red line across the screen
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Rect[X1, Y1, X2, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{})`

Draws an outline rectangle using two opposing corner points. Mimics PICO-8's `rect(x1, y1, x2, y2, color)`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x1`, `y1`, `x2`, `y2` | `Number` | Coordinates of two opposing corners. |
| `color` (via `options`) | `int` (optional) | PICO-8 color index (0-15). Defaults to the current draw color. |

**Example:**
```go
p8.Rect(10, 10, 40, 30, 11) // green outlined rectangle
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Rectfill[X1, Y1, X2, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{})`

Draws a filled rectangle using two opposing corner points. Mimics PICO-8's `rectfill(x1, y1, x2, y2, color)`.

**Parameters:** Same as `Rect`.

**Example:**
```go
p8.Rectfill(10, 10, 40, 30, 11) // filled green rectangle
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Circ[X, Y, R Number](x X, y Y, radius R, options ...interface{})`

Draws an outline circle. Mimics PICO-8's `circ(x, y, radius, color)`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `Number` | Center point coordinates. |
| `radius` | `Number` | Circle radius. |
| `color` (via `options`) | `int` (optional) | PICO-8 color index (0-15). Defaults to the current draw color. |

**Example:**
```go
p8.Circ(64, 64, 20, 9) // orange outlined circle at the center
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Circfill[X, Y, R Number](x X, y Y, radius R, options ...interface{})`

Draws a filled circle. Mimics PICO-8's `circfill(x, y, radius, color)`.

**Parameters:** Same as `Circ`.

**Example:**
```go
p8.Circfill(64, 64, 20, 9) // filled orange circle
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Print(s any, args ...int) (int, int)`

Draws `s` (converted via `fmt.Sprintf("%v", s)`) onto the screen at the implicit print cursor,
mimicking PICO-8's `print(str, [x, y], color)`, including cursor tracking. Returns the `(x, y)`
coordinates immediately after the printed string.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `s` | `any` | Value to print (formatted with `%v`). |
| `args` | `int` (optional, variadic) | `[]`: use cursor position/color. `[color]`: cursor position, given color. `[x, y]`: given position, cursor color. `[x, y, color]`: given position and color. |

**Returns:** the X and Y screen coordinates immediately after the printed text.

**Example:**
```go
p8.Cursor(0, 0, 6)
p8.Print("1 HELLO")            // prints at (0,0) in light gray, cursor moves to (0,6)
p8.Print("2 AT", 20, 20)       // prints at (20,20) in light gray
endX, endY := p8.Print("3 DONE") // prints at (20,26), returns next cursor position
```

**See also:** [Text](../../10-graphics/05-text.md)

## `Cursor(args ...int)`

Sets the implicit print cursor position and optionally the default draw color, mimicking
PICO-8's `cursor(x, y, color)`. Calling with no arguments resets the position to `(0, 0)` without
changing the color.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `args` | `int` (optional, variadic) | `[]`: reset to (0,0). `[x, y]`: set position. `[x, y, color]`: set position and draw color. |

**Example:**
```go
p8.Cursor(10, 20)     // move cursor to (10, 20)
p8.Cursor(30, 40, 5)  // move cursor and set draw color to 5
p8.Cursor()           // reset to (0, 0)
```

**See also:** [Text](../../10-graphics/05-text.md)

## `GetScreenWidth() int`

Returns the current logical screen width in pixels (128 by default, or the value configured via
`Settings.ScreenWidth`).

**Example:**
```go
w := p8.GetScreenWidth()
```

## `GetScreenHeight() int`

Returns the current logical screen height in pixels (128 by default, or the value configured via
`Settings.ScreenHeight`).

**Example:**
```go
h := p8.GetScreenHeight()
```

**See also:** [Screen](../../10-graphics/01-screen.md), [Settings](../../04-settings.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/01-screen-drawing.md
git commit -m "docs: add screen & drawing API reference page"
```

### Task 2.3: Write the Colors & Palette reference page

**Files:**
- Create: `docs/reference/api/02-colors-palette.md`

- [ ] **Step 1: Create the file**

Use this real source material for `Pal`, `Palt`, `Color`, `GetPaletteColor`, `GetPaletteSize`,
`SetPalette`, `SetPaletteColor` (from `go doc`), following the same template as Task 2.2:

```markdown
# Colors & Palette

## `Pal(args ...interface{})`

Configures draw palette mappings, mimicking PICO-8's `pal(c0, c1, p)`. When color `c0` is
requested for drawing, `c1` is used instead.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `args` | variadic | `[]`: reset all mappings. `[c0, c1]`: map c0 -> c1 for the draw palette. `[c0, c1, p]`: `p=0` maps the draw palette; `p=1` (screen palette post-processing) is not implemented and logs a warning. |

**Example:**
```go
p8.Pal(8, 12) // draw color 8 (red) as color 12 (blue) instead
p8.Pal()      // reset all palette mappings
```

**See also:** [Colors and Palette](../../10-graphics/02-colors.md)

## `Palt(args ...interface{})`

Sets color transparency for drawing. With no arguments, resets to default (only black
transparent).

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `color` | `int` | PICO-8 palette color index (0-15). |
| `transparent` | `bool` | `true` to make the color transparent, `false` to make it opaque. |

**Example:**
```go
p8.Spr(1, 10, 10)  // default transparency (black transparent)
p8.Palt(8, true)   // make red transparent
p8.Spr(1, 20, 10)  // now red pixels in the sprite are transparent
p8.Palt()          // reset to default transparency
```

**See also:** [Colors and Palette](../../10-graphics/02-colors.md)

## `Color(colorIndex int)`

Sets the current draw color used by subsequent drawing operations.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `colorIndex` | `int` | PICO-8 color index (0-15). |

**Example:**
```go
p8.Color(8)   // set current draw color to red
p8.Sset(10, 20) // draws a red pixel on the spritesheet at (10, 20)
```

**See also:** [Colors and Palette](../../10-graphics/02-colors.md)

## `GetPaletteColor(colorIndex int) color.Color`

Returns the `color.Color` at the given palette index, or `nil` if out of range.

**Example:**
```go
c := p8.GetPaletteColor(3)
```

## `GetPaletteSize() int`

Returns the current number of colors in the active palette.

**Example:**
```go
size := p8.GetPaletteSize()
p8.Print(fmt.Sprintf("Palette has %d colors", size), 10, 10, 7)
```

## `SetPalette(newPalette []color.Color)`

Replaces the entire color palette. Resizes the transparency array to match, resetting only index
0 as transparent by default.

**Example:**
```go
grayscale := []color.Color{
	color.RGBA{0, 0, 0, 255},
	color.RGBA{85, 85, 85, 255},
	color.RGBA{170, 170, 170, 255},
	color.RGBA{255, 255, 255, 255},
}
p8.SetPalette(grayscale)
```

**See also:** [Palette Effects](../../10-graphics/06-palette-effects.md)

## `SetPaletteColor(colorIndex int, newColor color.Color)`

Replaces a single palette color at the given index. No-op if the index is out of range.

**Example:**
```go
p8.SetPaletteColor(7, color.RGBA{200, 220, 255, 255}) // change white to light blue
```

**See also:** [Palette Effects](../../10-graphics/06-palette-effects.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/02-colors-palette.md
git commit -m "docs: add colors & palette API reference page"
```

### Task 2.4: Write the Sprites reference page

**Files:**
- Create: `docs/reference/api/03-sprites.md`

- [ ] **Step 1: Create the file**

Use this real source material for `Spr`, `Sspr`, `Sget`, `Sset`, `Fget`, `Fset`:

```markdown
# Sprites

## `Spr[SN, X, Y Number](spriteNumber SN, x X, y Y, options ...any)`

Draws sprite `spriteNumber` (and optionally a fractional block of surrounding sprites) at
`(x, y)`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `spriteNumber` | `Number` | Index of the top-left sprite to draw. |
| `x`, `y` | `Number` | Screen coordinates for the top-left corner. |
| `w`, `h` (via `options`) | `float64` or `int` (optional) | Width/height multiplier in sprites (default 1.0 each). |
| `flipX`, `flipY` (via `options`) | `bool` (optional) | Flip horizontally/vertically (default false). |

**Example:**
```go
p8.Spr(1, 10, 20)                   // draw sprite 1 at (10, 20)
p8.Spr(1, 50, 20, 1.5, 1.0)         // draw sprite 1 plus half of sprite 2
p8.Spr(0, 90, 20, 1.5, 1.5, true)   // 1.5x1.5 sprite block, flipped horizontally
```

**See also:** [Drawing Sprites](../../20-sprites/01-drawing-sprites.md)

## `Sspr[SX, SY, SW, SH, DX, DY Number](sx SX, sy SY, sw SW, sh SH, dx DX, dy DY, options ...any)`

Draws an arbitrary rectangular region of the spritesheet, with optional stretching and flipping.
Mimics PICO-8's `sspr(sx, sy, sw, sh, dx, dy, [dw, dh], [flip_x], [flip_y])`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `sx`, `sy` | `Number` | Source position on the spritesheet, in pixels. |
| `sw`, `sh` | `Number` | Source width/height, in pixels. |
| `dx`, `dy` | `Number` | Destination position on screen, in pixels. |
| `dw`, `dh` (via `options`) | `Number` (optional) | Destination width/height (defaults to `sw`/`sh`). |
| `flipX`, `flipY` (via `options`) | `bool` (optional) | Flip horizontally/vertically (default false). |

**Example:**
```go
p8.Sspr(8, 8, 16, 16, 10, 20)             // copy a 16x16 region as-is
p8.Sspr(8, 8, 16, 16, 10, 20, 32, 32)     // stretched to 32x32 on screen
p8.Sspr(8, 8, 16, 16, 10, 20, 16, 16, true, false) // flipped horizontally
```

**See also:** [Spritesheet Regions](../../20-sprites/02-spritesheet-regions.md)

## `Sget[X, Y Number](x X, y Y) int`

Returns the PICO-8 color index (0-15) of the pixel at `(x, y)` on the spritesheet. Returns 0 if
out of bounds.

**Example:**
```go
pixelColor := p8.Sget(10, 20)
```

## `Sset[X, Y Number](x X, y Y, colorIndex ...int)`

Sets the color of a pixel at `(x, y)` on the spritesheet. Uses the current draw color if
`colorIndex` is omitted.

**Example:**
```go
p8.Sset(10, 0, 8)  // red pixel at (10, 0) on the spritesheet
p8.Color(12)
p8.Sset(16, 0)     // blue pixel, using the current draw color
```

## `Fget(spriteNum int, flag ...int) (bitfield int, isSet bool)`

Returns a sprite's flag status. With no `flag` argument, returns the full 8-bit `bitfield`
(0-255). With a `flag` argument (0-7), `isSet` reports whether that specific flag is set.

**Example:**
```go
_, isSet := p8.Fget(1, 0)   // is flag 0 set on sprite 1?
allFlags, _ := p8.Fget(2)   // full bitfield for sprite 2
```

**See also:** [Sprite Flags](../../20-sprites/03-sprite-flags.md)

## `Fset(spriteNum int, flagOrValue interface{}, value ...interface{})`

Sets a sprite's flag status. If `flagOrValue` is a flag number (0-7), `value` sets that flag on
or off. If no `value` is given, `flagOrValue` is treated as a boolean or bitfield applied to all
flags.

**Example:**
```go
p8.Fset(1, 0, true)  // set flag 0 on sprite 1
p8.Fset(2, false)    // clear all flags on sprite 2
p8.Fset(2, 170)      // set flags 1,3,5,7 via bitfield (170 = 2+8+32+128)
```

**See also:** [Sprite Flags](../../20-sprites/03-sprite-flags.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/03-sprites.md
git commit -m "docs: add sprites API reference page"
```

### Task 2.5: Write the Maps reference page

**Files:**
- Create: `docs/reference/api/04-maps.md`

- [ ] **Step 1: Create the file**

```markdown
# Maps

## `Map(args ...any)`

Draws a rectangular region of the map to the screen.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `mx`, `my` | `int` (optional) | Map tile coordinates to start reading from (default 0, 0). |
| `sx`, `sy` | `int` (optional) | Screen pixel coordinates to draw at (default 0, 0). |
| `w`, `h` | `int` (optional) | Dimensions in tiles (default 16x16). |
| `layers` | `int` (optional) | Bitfield to filter drawn sprites by their flags (0 = draw all). |

**Example:**
```go
p8.Map()                          // draw the default 16x16 tile region at (0,0)
p8.Map(0, 0, 0, 0, 32, 32)        // draw a 32x32 tile region
p8.Map(0, 0, 0, 0, 32, 32, 1)     // only draw tiles whose sprite has flag 0 set
```

**See also:** [Drawing Maps](../../30-maps/01-drawing-maps.md)

## `Mget[C, R Number](column C, row R) int`

Returns the sprite number placed at the given map column/row.

**Example:**
```go
sprite := p8.Mget(5, 3)
```

## `Mset[C, R, S Number](column C, row R, sprite S)`

Sets the sprite number at the given map column/row.

**Example:**
```go
p8.Mset(5, 3, 1) // place sprite 1 at column 5, row 3
```

**See also:** [Map Data](../../30-maps/02-map-data.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/04-maps.md
git commit -m "docs: add maps API reference page"
```

### Task 2.6: Write the Input reference page

**Files:**
- Create: `docs/reference/api/05-input.md`

- [ ] **Step 1: Create the file**

```markdown
# Input

## `Btn(buttonIndex int, playerIndex ...int) bool`

Reports whether a button is currently held down, via keyboard (player 0 only), gamepad, mouse,
or gamepad axes. Mimics PICO-8's `btn()`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `buttonIndex` | `int` | PICO-8 button index (0-15) - see the button constants (`LEFT`, `RIGHT`, `UP`, `DOWN`, `O`, `X`, `ButtonStart`, `ButtonSelect`, and mouse/gamepad-specific constants). |
| `playerIndex` | `int` (optional) | Player index (0-7). Defaults to 0. |

**Example:**
```go
if p8.Btn(p8.LEFT) { g.x-- }
if p8.Btn(p8.RIGHT) { g.x++ }
```

**See also:** [Keyboard](../../40-input/01-keyboard.md), [Gamepad](../../40-input/02-gamepad.md)

## `Btnp(buttonIndex int, playerIndex ...int) bool`

Reports whether a button was **just pressed** this frame (transitioned from up to down, no
auto-repeat). Mimics PICO-8's `btnp()`. Keyboard input only applies to player 0; mouse input
applies to all player indices.

**Example:**
```go
if p8.Btnp(p8.X) {
	// jump
}
if p8.Btnp(p8.ButtonStart, 1) {
	// pause for player 1
}
```

**See also:** [Keyboard](../../40-input/01-keyboard.md)

## `GetMouseXY() (int, int)`

Returns the current mouse X and Y coordinates. Mimics PICO-8's `mouse()`.

**Example:**
```go
mx, my := p8.GetMouseXY()
p8.Circ(mx, my, 4, 8)
```

**See also:** [Mouse](../../40-input/03-mouse.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/05-input.md
git commit -m "docs: add input API reference page"
```

### Task 2.7: Write the Audio reference page

**Files:**
- Create: `docs/reference/api/06-audio.md`

- [ ] **Step 1: Create the file**

```markdown
# Audio

## `Music(n int, exclusive ...bool)`

Plays the audio file with ID `n`. If `n` is -1, stops all currently playing audio. If
`exclusive` is `true`, stops all other audio first.

**Example:**
```go
p8.Music(0)        // play audio 0 (mixes with anything already playing)
p8.Music(0, true)  // play audio 0, stopping everything else first
```

## `MusicLoop(n int, exclusive ...bool)`

Plays audio `n` in a loop. Equivalent to `MusicWithOptions(n, MusicOptions{Loop: true})`.

## `StopMusic(id int)`

Stops the audio file with the given ID. If `id` is -1, stops all audio.

## `MusicWithOptions(n int, opts MusicOptions)` / `MusicOptions`

Plays audio `n` using explicit options instead of variadic booleans.

**`MusicOptions` fields:**
| Field | Type | Description |
|-------|------|--------------|
| `Exclusive` | `bool` | If true, stops all other audio before playing. |
| `Loop` | `bool` | If true, loops indefinitely. |

**Example:**
```go
p8.MusicWithOptions(0, p8.MusicOptions{Loop: true, Exclusive: true})
```

## 32-bit float variants: `MusicF32`, `MusicLoopF32`, `StopMusicF32`, `MusicF32WithOptions`

Identical behavior to their non-`F32` counterparts, but use 32-bit float audio, the recommended
approach for Ebitengine v2.8+ for better performance and easier audio processing.

**Example:**
```go
p8.MusicF32(0)
p8.MusicLoopF32(1)
p8.StopMusicF32(-1) // stop all
```

**See also:** [Music and Sound](../../50-audio/00-music.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/06-audio.md
git commit -m "docs: add audio API reference page"
```

### Task 2.8: Write the Camera reference page

**Files:**
- Create: `docs/reference/api/07-camera.md`

- [ ] **Step 1: Create the file**

```markdown
# Camera

## `Camera(args ...any)`

Sets the camera position, offsetting all subsequent drawing operations (shapes, `Print`,
sprites, and maps). With no arguments, resets the camera to `(0, 0)`. Mimics PICO-8's
`camera(x, y)`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `any` (optional) | Horizontal/vertical camera offset. |

**Example:**
```go
p8.Camera(64, 32) // set camera to (64, 32)
p8.Camera()       // reset to (0, 0)

// Lock UI in place while the world scrolls:
p8.Camera()
p8.Print("SCORE: 1000", 2, 2)
p8.Camera(playerX-64, playerY-64)
p8.Map()
```

**See also:** [Camera](../../60-camera/00-camera.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/07-camera.md
git commit -m "docs: add camera API reference page"
```

### Task 2.9: Write the Collision reference page

**Files:**
- Create: `docs/reference/api/08-collision.md`

- [ ] **Step 1: Create the file**

```markdown
# Collision

## `ColorCollision[X, Y Number](x X, y Y, color int) bool`

Reports whether the pixel at `(x, y)` matches `color`. Returns `false` if the coordinates are
out of screen bounds (0-127) or the color index is invalid.

**Example:**
```go
if p8.ColorCollision(player.x, player.y, 3) {
	// touching a wall (color 3)
}
```

**See also:** [Color Collision](../../70-collision/01-color-collision.md)

## `MapCollision[X, Y Number](x X, y Y, flag int, size ...int) bool`

Reports whether a rectangular area starting at `(x, y)` overlaps any map tile whose sprite has
`flag` set. Internally resolves overlapping tiles via `Mget`/`Fget`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `Number` | Top-left pixel coordinates of the area to check. |
| `flag` | `int` | Sprite flag number (0-7) to check for on underlying tiles. |
| `size` | `int` (optional, variadic) | `[]`: 8x8 area. `[s]`: `s`x`s` area. `[w, h]`: `w`x`h` area. |

**Example:**
```go
if p8.MapCollision(player.x, player.y, p8.Flag0) {
	// collision with default 8x8 area
}
playerWidth, playerHeight := 14, 15
if p8.MapCollision(player.x, player.y, p8.Flag0, playerWidth, playerHeight) {
	// collision with a custom-sized area
}
```

**See also:** [Map Collision](../../70-collision/02-map-collision.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/08-collision.md
git commit -m "docs: add collision API reference page"
```

### Task 2.10: Write the Math reference page

**Files:**
- Create: `docs/reference/api/09-math.md`

- [ ] **Step 1: Create the file**

```markdown
# Math

## `Flr[T Number](a T) int`

Rounds down to the nearest whole integer. Mimics PICO-8's `flr()`.

**Example:**
```go
p8.Flr(1.99)  // 1
p8.Flr(-5.3)  // -6
```

## `Rnd[T Number](a T) int`

Returns a random integer in `[0, floor(a))`. Returns 0 if `a` is zero or negative. Mimics
PICO-8's `flr(rnd(a))`. Uses Go's `math/rand`; unlike PICO-8, results are not deterministic
across runs unless you explicitly seed the global source.

**Example:**
```go
p8.Rnd(5)    // 0, 1, 2, 3, or 4
p8.Rnd(5.9)  // 0-4 (floor(5.9) = 5)
```

## `Sqrt[T Number](a T) float64`

Returns the square root. Returns 0 for negative input (unlike `math.Sqrt`, which returns `NaN`),
matching PICO-8 behavior.

**Example:**
```go
p8.Sqrt(16) // 4.0
p8.Sqrt(-4) // 0.0
```

## `Sign(v float64) float64`

Returns the sign of `v`, with 0 treated as `+1`.

**Example:**
```go
p8.Sign(-3.5) // -1.0
p8.Sign(0)    // 1.0
```

## `Time() float64`

Returns the number of seconds elapsed since the game started, based on the number of `Update`
calls. Multiple calls within the same frame return the same value.

**Example:**
```go
frame := int(p8.Time() * 10) % 4
p8.Spr(1+frame, x, y) // simple 4-frame animation
```

**See also:** [Math Functions](../../80-math/00-math.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/09-math.md
git commit -m "docs: add math API reference page"
```

### Task 2.11: Write the Settings & Lifecycle reference page

**Files:**
- Create: `docs/reference/api/10-settings-lifecycle.md`

- [ ] **Step 1: Create the file**

```markdown
# Settings & Lifecycle

## `Cartridge` interface

```go
type Cartridge interface {
	Init()   // Called once at the start.
	Update() // Called every frame for logic.
	Draw()   // Called every frame for drawing.
}
```

Your game struct must implement this interface. See [The Game Loop](../../03-game-loop.md).

## `InsertGame(cartridge Cartridge)`

Registers your game implementation with PIGO8. Must be called before `Play()`/`PlayGameWith()`.

## `Play()`

Runs the console with default settings, using the cartridge registered via `InsertGame`.
Equivalent to `PlayGameWith(NewSettings())`.

## `PlayGameWith(settings *Settings)`

Runs the console with custom settings.

## `NewSettings() *Settings`

Creates a `Settings` object populated with defaults.

## `Settings` struct

| Field | Type | Default | Description |
|-------|------|---------|--------------|
| `ScaleFactor` | `int` | 4 | Integer window scaling factor. |
| `WindowTitle` | `string` | "PIGO-8 Game" | Window title bar text. |
| `TargetFPS` | `int` | 30 | Target ticks per second. |
| `ScreenWidth` | `int` | 128 | Logical screen width. |
| `ScreenHeight` | `int` | 128 | Logical screen height. |
| `Multiplayer` | `bool` | false | Enable multiplayer networking. |
| `Fullscreen` | `bool` | false | Start in fullscreen. |
| `ColorSpace` | `ebiten.ColorSpace` | default | Rendering color space. |
| `DisableHiDPI` | `bool` | false | Disable HiDPI scaling. |
| `SpriteImageCacheSize` | `int` | 256 | Max cached transparent sprite images. |
| `SpritePixelCacheSize` | `int` | 256 | Max cached sprite pixel data entries. |
| `SsprCacheSize` | `int` | 128 | Max cached `Sspr` source regions. |
| `MapCacheEnabled` | `bool` | true | Enable map tile caching. |
| `EnableFrameStats` | `bool` | false | Enable frame-level performance statistics. |

**Example:**
```go
settings := p8.NewSettings()
settings.WindowTitle = "My Game"
settings.ScaleFactor = 4
settings.TargetFPS = 60
settings.ScreenWidth = 160
settings.ScreenHeight = 144

p8.InsertGame(&game{})
p8.PlayGameWith(settings)
```

**See also:** [Settings](../../04-settings.md), [The Game Loop](../../03-game-loop.md)
```

- [ ] **Step 2: Commit**

```bash
git add docs/reference/api/10-settings-lifecycle.md
git commit -m "docs: add settings & lifecycle API reference page"
```

### Task 2.12: Wire the API Reference into SUMMARY.md and cross-link the cheatsheet

**Files:**
- Modify: `docs/SUMMARY.md`
- Modify: `docs/reference/cheatsheet.md`

- [ ] **Step 1: Add the API Reference section to `docs/SUMMARY.md`**

Replace:

```markdown
# Reference

- [Cheatsheet](reference/cheatsheet.md)
- [PICO-8 Comparison](reference/pico8-comparison.md)
```

with:

```markdown
# Reference

- [Cheatsheet](reference/cheatsheet.md)
- [API Reference](reference/api/00-overview.md)
  - [Screen & Drawing](reference/api/01-screen-drawing.md)
  - [Colors & Palette](reference/api/02-colors-palette.md)
  - [Sprites](reference/api/03-sprites.md)
  - [Maps](reference/api/04-maps.md)
  - [Input](reference/api/05-input.md)
  - [Audio](reference/api/06-audio.md)
  - [Camera](reference/api/07-camera.md)
  - [Collision](reference/api/08-collision.md)
  - [Math](reference/api/09-math.md)
  - [Settings & Lifecycle](reference/api/10-settings-lifecycle.md)
- [PICO-8 Comparison](reference/pico8-comparison.md)
```

- [ ] **Step 2: Add a pointer from the cheatsheet to the full reference**

At the top of `docs/reference/cheatsheet.md`, right after the `# PIGO8 Cheatsheet` heading, add:

```markdown
For full explanations, parameter tables, and examples, see the [API Reference](api/00-overview.md).
```

- [ ] **Step 3: Verify the mdBook builds with the new structure**

Run: `mdbook build`
Expected: builds successfully with no warnings about missing files referenced in `SUMMARY.md`.

- [ ] **Step 4: Commit**

```bash
git add docs/SUMMARY.md docs/reference/cheatsheet.md
git commit -m "docs: wire API reference into summary and cheatsheet"
```

---

## Phase 3: New Pages (FAQ, Testing)

### Task 3.1: Write the FAQ & Troubleshooting page

**Files:**
- Create: `docs/reference/faq.md`

- [ ] **Step 1: Create the file**

```markdown
# FAQ & Troubleshooting

## Installation & Build

### "cgo: C compiler not found" or similar CGO errors

Ebitengine (which PIGO8 is built on) requires CGO and a C compiler. Install one for your
platform:

- **macOS**: `xcode-select --install`
- **Linux**: `sudo apt install gcc` (Debian/Ubuntu) or `sudo dnf install gcc` (Fedora)
- **Windows**: install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MinGW-w64 and ensure it's on your `PATH`.

See [Installation](../01-installation.md) for full platform notes.

### Linux build fails with missing audio/graphics libraries

Install the required development libraries before building:

```bash
# Debian/Ubuntu
sudo apt install libasound2-dev libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install alsa-lib-devel mesa-libGL-devel xorg-x11-server-devel
```

### `go build` for WebAssembly fails with "GOOS=js GOARCH=wasm" errors

Make sure you're not manually setting `GOOS`/`GOARCH` in your shell environment before running
`cmd/webexport` - it sets these itself. If building manually, use:

```bash
GOOS=js GOARCH=wasm go build -o game.wasm .
```

and copy `wasm_exec.js` from your Go installation's `$(go env GOROOT)/lib/wasm/wasm_exec.js`
(Go 1.24+) or `misc/wasm/wasm_exec.js` (earlier versions). See
[Web Export](../90-advanced/02-web-export.md) for the full workflow, which automates this.

## Runtime

### The window opens but shows a blank/black screen

- Make sure you call `p8.Cls(...)` at the start of your `Draw()` method - without clearing, the
  screen may show stale or undefined content on some platforms.
- Confirm `Draw()` is actually being called: add a temporary `p8.Print("draw", 0, 0, 7)` to
  verify.
- If you're loading sprites/maps from `spritesheet.json`/`map.json`, confirm those files are in
  the same directory as your executable when running `go run .` (or embedded - see
  [Resource Embedding](../90-advanced/01-resource-embedding.md)).

### My sprite doesn't show up / looks wrong

- Sprite indices are 0-based. Sprite 0 is the top-left sprite of your spritesheet.
- Check `Palt()` - if the sprite's background color happens to match your current transparent
  color, parts of it may render as "invisible" unexpectedly.
- If you recently edited `spritesheet.json` with the [PIGO8 Editor](../editor.md), confirm you
  saved (the editor autosaves on switching modes, but verify the file's modification time).

### Audio doesn't play on Linux

Confirm `libasound2-dev` (or your distro's ALSA development package) is installed - see the
Linux build section above. If audio still doesn't play, check that your audio files follow the
`music0.wav`, `music1.wav`, ... naming convention expected by `Music()`/`MusicLoop()`.

### Browser build: audio doesn't play until I click something

This is expected - browsers block autoplaying audio until the user interacts with the page.
The web-exported page's virtual controls handle this automatically; audio starts working after
the first button press. See [Web Export](../90-advanced/02-web-export.md#audio-notes).

### Movement feels too fast/slow after changing `TargetFPS`

Fixed per-frame movement values (like `g.x++`) move faster at higher FPS. Use `p8.Time()` to
compute frame-rate-independent movement instead of relying on a fixed FPS. See
[The Game Loop](../03-game-loop.md#time-function).

## Where to Get Help

If your issue isn't covered here, check the
[GitHub issues](https://github.com/drpaneas/pigo8/issues) or open a new one with a minimal
reproduction.
```

- [ ] **Step 2: Add the FAQ page to `SUMMARY.md`**

In `docs/SUMMARY.md`, under the `# Reference` section, add the FAQ entry after the PICO-8
Comparison line:

```markdown
- [PICO-8 Comparison](reference/pico8-comparison.md)
- [FAQ & Troubleshooting](reference/faq.md)
```

- [ ] **Step 3: Verify and commit**

Run: `mdbook build`
Expected: builds with no warnings.

```bash
git add docs/reference/faq.md docs/SUMMARY.md
git commit -m "docs: add FAQ and troubleshooting page"
```

### Task 3.2: Write the Testing Your Game page

**Files:**
- Create: `docs/90-advanced/03-testing.md`

- [ ] **Step 1: Create the file**

```markdown
# Testing Your Game

PIGO8 games are plain Go structs implementing the `Cartridge` interface, so standard Go testing
applies directly to your game logic - the trick is keeping logic that can be tested (movement,
collision, state transitions) separate from code that can only run inside the game loop
(`Draw()` calls, which require an active screen).

## Testing Logic Decoupled from `Update`/`Draw`

Structure gameplay logic as plain methods that don't call PIGO8 drawing functions, so they can
be tested without running the game loop:

```go
type Player struct {
	X, Y  float64
	Score int
}

// Move is pure logic - no p8 calls - so it's directly testable.
func (p *Player) Move(dx, dy float64) {
	p.X += dx
	p.Y += dy
	if p.X < 0 {
		p.X = 0
	}
}
```

```go
func TestPlayerMove(t *testing.T) {
	p := &Player{X: 5, Y: 5}
	p.Move(-10, 0)
	if p.X != 0 {
		t.Errorf("expected X to clamp at 0, got %v", p.X)
	}
}
```

Your `Cartridge.Update()` method then becomes a thin adapter that reads input and calls these
tested methods:

```go
func (g *game) Update() {
	var dx, dy float64
	if p8.Btn(p8.LEFT) {
		dx = -1
	}
	if p8.Btn(p8.RIGHT) {
		dx = 1
	}
	g.player.Move(dx, dy)
}
```

## Testing Collision Logic

Collision helpers like `p8.MapCollision` and `p8.ColorCollision` need an initialized screen/map,
so wrap your collision *decisions* in testable functions instead of testing the PIGO8 calls
themselves:

```go
// canMoveTo is pure logic given a collision check result - test this directly.
func canMoveTo(hitWall bool, isJumping bool) bool {
	return !hitWall || isJumping
}

func TestCanMoveTo(t *testing.T) {
	if canMoveTo(true, false) {
		t.Error("should not be able to move into a wall while not jumping")
	}
	if !canMoveTo(true, true) {
		t.Error("should be able to move through a wall while jumping")
	}
}
```

## Testing State Transitions

Model game states as an enum and test transitions independently of rendering:

```go
type State int

const (
	StateMenu State = iota
	StatePlaying
	StateGameOver
)

func nextState(current State, playerHealth int, startPressed bool) State {
	switch current {
	case StateMenu:
		if startPressed {
			return StatePlaying
		}
	case StatePlaying:
		if playerHealth <= 0 {
			return StateGameOver
		}
	}
	return current
}

func TestNextState_GameOverOnZeroHealth(t *testing.T) {
	got := nextState(StatePlaying, 0, false)
	if got != StateGameOver {
		t.Errorf("expected StateGameOver, got %v", got)
	}
}
```

## Running Tests

Standard Go tooling works as expected:

```bash
go test ./...
go test -run TestPlayerMove -v
go test -cover ./...
```

## What Not to Unit Test

Don't try to unit test `Draw()` itself, or functions that call drawing primitives (`Spr`,
`Rect`, `Cls`, etc.) directly - they require an active Ebitengine screen and are better verified
visually by running the game (`go run .`) or via the [Web Export](02-web-export.md) for quick
browser checks.
```

- [ ] **Step 2: Add the Testing page to `SUMMARY.md`**

In `docs/SUMMARY.md`, under `# Advanced Topics`, add it after Web Export:

```markdown
- [Resource Embedding](90-advanced/01-resource-embedding.md)
- [Web Export](90-advanced/02-web-export.md)
- [Testing Your Game](90-advanced/03-testing.md)
- [PIGO8 Editor](editor.md)
```

- [ ] **Step 3: Verify and commit**

Run: `mdbook build`
Expected: builds with no warnings.

```bash
git add docs/90-advanced/03-testing.md docs/SUMMARY.md
git commit -m "docs: add testing your game page"
```

---

## Phase 4: Visual Capture Tooling

### Task 4.1: Extract shared WASM build logic into `internal/webbuild`

`cmd/webexport` currently has unexported build helpers (`buildWASM`, `copyWASMExec`,
`generateHTML`) that `cmd/docshots` also needs. Extract them into an internal package so both
tools share one implementation (DRY), following the existing small-focused-file convention.

**Files:**
- Create: `internal/webbuild/webbuild.go`
- Create: `internal/webbuild/webbuild_test.go`
- Modify: `cmd/webexport/main.go`
- Modify: `cmd/webexport/template.go` (move `generateHTML` and its template into `internal/webbuild`)

- [ ] **Step 1: Write the failing test**

```go
// internal/webbuild/webbuild_test.go
package webbuild

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestBuildWASMEnv(t *testing.T) {
	env := buildWASMEnv([]string{"GOOS=darwin", "GOARCH=arm64", "FOO=bar"})
	want := map[string]bool{"GOOS=js": true, "GOARCH=wasm": true, "FOO=bar": true}
	got := map[string]bool{}
	for _, e := range env {
		got[e] = true
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing expected env entry %q in %v", k, env)
		}
	}
	for _, e := range env {
		if e == "GOOS=darwin" || e == "GOARCH=arm64" {
			t.Errorf("stale GOOS/GOARCH entry %q should have been filtered", e)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/webbuild/... -v`
Expected: FAIL - package `webbuild` doesn't exist yet.

- [ ] **Step 3: Create `internal/webbuild/webbuild.go` with the extracted logic**

```go
// Package webbuild provides shared logic for compiling PIGO8 games to
// WebAssembly and generating a browser-playable page, used by both
// cmd/webexport and cmd/docshots.
package webbuild

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildWASM compiles the Go program in gameDir to a WebAssembly binary at outputPath.
func BuildWASM(gameDir, outputPath string) error {
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", outputPath, ".")
	cmd.Dir = gameDir
	cmd.Env = buildWASMEnv(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// buildWASMEnv returns a copy of env with GOOS/GOARCH forced to js/wasm,
// removing any pre-existing GOOS/GOARCH entries to avoid duplicates.
func buildWASMEnv(env []string) []string {
	filtered := make([]string, 0, len(env)+2)
	for _, e := range env {
		if !strings.HasPrefix(e, "GOOS=") && !strings.HasPrefix(e, "GOARCH=") {
			filtered = append(filtered, e)
		}
	}
	return append(filtered, "GOOS=js", "GOARCH=wasm")
}

// CopyWASMExec copies wasm_exec.js from the active Go installation into outputDir.
func CopyWASMExec(outputDir string) error {
	goroot, err := getGoRoot()
	if err != nil {
		return fmt.Errorf("failed to determine GOROOT: %w", err)
	}

	possiblePaths := []string{
		filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"),
		filepath.Join(goroot, "misc", "wasm", "wasm_exec.js"),
	}

	var wasmExecSrc string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			wasmExecSrc = path
			break
		}
	}
	if wasmExecSrc == "" {
		return fmt.Errorf("wasm_exec.js not found in GOROOT (%s). Tried: %v", goroot, possiblePaths)
	}

	return copyFile(wasmExecSrc, filepath.Join(outputDir, "wasm_exec.js"))
}

func getGoRoot() (string, error) {
	cmd := exec.Command("go", "env", "GOROOT")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func copyFile(src, dst string) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// EnsureExampleModule makes sure gameDir has its own go.mod pointing at the
// local pigo8 checkout via a replace directive, matching how CI builds each
// example (examples' go.mod/go.sum are gitignored, not committed).
func EnsureExampleModule(gameDir, repoRoot, modulePath string) error {
	if _, err := os.Stat(filepath.Join(gameDir, "go.mod")); err == nil {
		return nil // already has one
	}

	relPath, err := filepath.Rel(gameDir, repoRoot)
	if err != nil {
		return fmt.Errorf("computing relative path to repo root: %w", err)
	}

	initCmd := exec.Command("go", "mod", "init", modulePath)
	initCmd.Dir = gameDir
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}

	goModPath := filepath.Join(gameDir, "go.mod")
	f, err := os.OpenFile(goModPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, "\nreplace github.com/drpaneas/pigo8 => %s\n", filepath.ToSlash(relPath)); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = gameDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	return tidyCmd.Run()
}
```

- [ ] **Step 4: Move `generateHTML` and the HTML template into `internal/webbuild`**

Move the contents of `cmd/webexport/template.go` into a new `internal/webbuild/template.go`
(same package `webbuild`), renaming the exported entry point to `GenerateHTML(htmlPath, title
string) error` (capitalize to export it). Keep the template content byte-for-byte identical -
only the package name and function visibility change.

- [ ] **Step 5: Update `cmd/webexport/main.go` to use the shared package**

Replace the local `buildWASM`, `buildWASMEnv`, `copyWASMExec`, `getGoRoot`, `copyFile` functions
and their call sites in `cmd/webexport/main.go` with calls to `internal/webbuild`:

```go
import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/drpaneas/pigo8/internal/webbuild"
)
```

Replace:
```go
if err := buildWASM(gameDir, wasmPath); err != nil {
```
with:
```go
if err := webbuild.BuildWASM(gameDir, wasmPath); err != nil {
```

Replace:
```go
if err := copyWASMExec(outputDir); err != nil {
```
with:
```go
if err := webbuild.CopyWASMExec(outputDir); err != nil {
```

Replace:
```go
if err := generateHTML(htmlPath, title); err != nil {
```
with:
```go
if err := webbuild.GenerateHTML(htmlPath, title); err != nil {
```

Delete the now-unused local `buildWASM`, `buildWASMEnv`, `copyWASMExec`, `getGoRoot`, `copyFile`
functions from `cmd/webexport/main.go`, and delete `cmd/webexport/template.go` entirely (its
content now lives in `internal/webbuild/template.go`). Keep `serveHTTP` in `cmd/webexport/main.go`
- it's specific to the CLI, not shared build logic.

- [ ] **Step 6: Run tests to verify everything still passes**

Run: `go test ./internal/webbuild/... ./cmd/webexport/... -v`
Expected: PASS.

Run: `go build ./...`
Expected: builds cleanly with no errors.

- [ ] **Step 7: Manually verify `cmd/webexport` still works end-to-end**

Run:
```bash
go run ./cmd/webexport -game ./examples/hello_world -o /tmp/pigo8-webexport-check
ls /tmp/pigo8-webexport-check
```
Expected: `game.wasm`, `index.html`, and `wasm_exec.js` are created, matching pre-refactor
behavior.

- [ ] **Step 8: Commit**

```bash
git add internal/webbuild cmd/webexport
git commit -m "refactor: extract shared WASM build logic into internal/webbuild"
```

### Task 4.2: Add the `chromedp` dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Add the dependency**

Run:
```bash
go get github.com/chromedp/chromedp@latest
```

- [ ] **Step 2: Verify it resolves and the module still builds**

Run: `go build ./...`
Expected: builds cleanly (no code uses it yet, but the dependency should resolve without conflicts).

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add chromedp dependency for docs screenshot tooling"
```

### Task 4.3: Build the `cmd/docshots` manifest and GIF/PNG assembly logic

**Files:**
- Create: `cmd/docshots/manifest.go`
- Create: `cmd/docshots/frames.go`
- Create: `cmd/docshots/frames_test.go`

- [ ] **Step 1: Write the failing test for frame assembly**

```go
// cmd/docshots/frames_test.go
package main

import (
	"image"
	"image/color"
	"testing"
)

func TestFramesToGIF(t *testing.T) {
	// Two 4x4 frames: solid red, then solid blue.
	red := image.NewRGBA(image.Rect(0, 0, 4, 4))
	blue := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			red.Set(x, y, color.RGBA{R: 255, A: 255})
			blue.Set(x, y, color.RGBA{B: 255, A: 255})
		}
	}

	g, err := framesToGIF([]image.Image{red, blue}, 100)
	if err != nil {
		t.Fatalf("framesToGIF: %v", err)
	}
	if len(g.Image) != 2 {
		t.Fatalf("expected 2 encoded frames, got %d", len(g.Image))
	}
	if len(g.Delay) != 2 || g.Delay[0] != 10 {
		// GIF delay is in 1/100ths of a second; 100ms == 10.
		t.Errorf("expected delay [10, 10], got %v", g.Delay)
	}
}

func TestBestFrame(t *testing.T) {
	frames := []image.Image{
		image.NewRGBA(image.Rect(0, 0, 2, 2)),
		image.NewRGBA(image.Rect(0, 0, 2, 2)),
		image.NewRGBA(image.Rect(0, 0, 2, 2)),
	}
	got := bestFrame(frames)
	if got == nil {
		t.Fatal("expected a non-nil frame")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/docshots/... -v`
Expected: FAIL - `framesToGIF`/`bestFrame` undefined.

- [ ] **Step 3: Implement `frames.go`**

```go
// cmd/docshots/frames.go
package main

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
)

// framesToGIF converts a sequence of RGBA frames into an animated GIF,
// quantizing each frame to a fixed 256-color palette and showing each frame
// for delayMs milliseconds before advancing (looping indefinitely).
func framesToGIF(frames []image.Image, delayMs int) (*gif.GIF, error) {
	out := &gif.GIF{LoopCount: 0}
	delayHundredths := delayMs / 10
	if delayHundredths < 1 {
		delayHundredths = 1
	}

	for _, frame := range frames {
		bounds := frame.Bounds()
		paletted := image.NewPaletted(bounds, palette.Plan9)
		draw.Draw(paletted, bounds, frame, bounds.Min, draw.Src)

		out.Image = append(out.Image, paletted)
		out.Delay = append(out.Delay, delayHundredths)
	}

	return out, nil
}

// bestFrame picks a representative single frame for a static screenshot -
// currently the last captured frame, since by then any startup/loading
// artifacts have settled.
func bestFrame(frames []image.Image) image.Image {
	if len(frames) == 0 {
		return nil
	}
	return frames[len(frames)-1]
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/docshots/... -v`
Expected: PASS.

- [ ] **Step 5: Write the manifest**

```go
// cmd/docshots/manifest.go
package main

// InputStep represents holding a set of keys for a duration before releasing them.
type InputStep struct {
	Keys   []string // key names, e.g. "ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"
	HoldMs int
}

// CaptureJob describes one example to capture and where its output goes.
type CaptureJob struct {
	Name        string // used for logging and the output filename (without extension)
	ExampleDir  string // relative to the repo root, e.g. "examples/hello_world"
	Kind        string // "static" or "gif"
	Inputs      []InputStep
	CaptureMs   int // total capture window in milliseconds
	SampleMs    int // interval between frame samples in milliseconds
}

// manifest lists every example to capture for the documentation site.
// Each entry maps to a concrete existing example under examples/.
var manifest = []CaptureJob{
	{
		Name:       "hello-world",
		ExampleDir: "examples/hello_world",
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "animation",
		ExampleDir: "examples/animation",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   100,
	},
	{
		Name:       "spritesheet",
		ExampleDir: "examples/spritesheet",
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "map",
		ExampleDir: "examples/map",
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "map-layers",
		ExampleDir: "examples/map_layers",
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "camera",
		ExampleDir: "examples/camera/camera_example2",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowRight"}, HoldMs: 800},
			{Keys: []string{"ArrowDown"}, HoldMs: 800},
		},
	},
	{
		Name:       "palette",
		ExampleDir: "examples/palette",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
	},
	{
		Name:       "color-collision",
		ExampleDir: "examples/colorCollision",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowRight"}, HoldMs: 1000},
			{Keys: []string{"ArrowUp"}, HoldMs: 1000},
		},
	},
	{
		Name:       "map-collision",
		ExampleDir: "examples/gameboy",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowRight"}, HoldMs: 1500},
		},
	},
	{
		Name:       "pong",
		ExampleDir: "examples/pong",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowUp"}, HoldMs: 600},
			{Keys: []string{"ArrowDown"}, HoldMs: 1200},
			{Keys: []string{"ArrowUp"}, HoldMs: 600},
		},
	},
}
```

- [ ] **Step 6: Commit**

```bash
git add cmd/docshots/manifest.go cmd/docshots/frames.go cmd/docshots/frames_test.go
git commit -m "feat: add docshots manifest and frame/GIF assembly logic"
```

### Task 4.4: Build the headless capture logic and CLI entry point

**Files:**
- Create: `cmd/docshots/capture.go`
- Create: `cmd/docshots/main.go`

- [ ] **Step 1: Implement the headless capture logic**

```go
// cmd/docshots/capture.go
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"strings"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// arrowKeyCodes maps the key names used in the manifest to the JS keyCode /
// CDP key values needed to synthesize a real keydown/keyup pair for Ebiten's
// keyboard listener (Ebiten reads raw keydown/keyup events, not typed text).
var arrowKeyCodes = map[string]struct {
	Code            string
	Key             string
	WindowsVirtual  int64
	NativeVirtual   int64
}{
	"ArrowLeft":  {"ArrowLeft", "ArrowLeft", 37, 37},
	"ArrowUp":    {"ArrowUp", "ArrowUp", 38, 38},
	"ArrowRight": {"ArrowRight", "ArrowRight", 39, 39},
	"ArrowDown":  {"ArrowDown", "ArrowDown", 40, 40},
}

// holdKeys dispatches keydown events for the given keys, waits for holdMs,
// then dispatches the matching keyup events. Unknown key names are skipped.
func holdKeys(keys []string, holdMs int) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var downs []input.DispatchKeyEventParams
		for _, k := range keys {
			code, ok := arrowKeyCodes[k]
			if !ok {
				return fmt.Errorf("unsupported key %q in manifest input step", k)
			}
			downs = append(downs, *input.DispatchKeyEvent(input.KeyDown).
				WithCode(code.Code).
				WithKey(code.Key).
				WithWindowsVirtualKeyCode(code.WindowsVirtual).
				WithNativeVirtualKeyCode(code.NativeVirtual))
		}
		for _, d := range downs {
			if err := d.Do(ctx); err != nil {
				return err
			}
		}

		time.Sleep(time.Duration(holdMs) * time.Millisecond)

		for _, k := range keys {
			code := arrowKeyCodes[k]
			up := input.DispatchKeyEvent(input.KeyUp).
				WithCode(code.Code).
				WithKey(code.Key).
				WithWindowsVirtualKeyCode(code.WindowsVirtual).
				WithNativeVirtualKeyCode(code.NativeVirtual)
			if err := up.Do(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// captureCanvasPNG evaluates JS in the page to grab the current <canvas>
// contents as a base64-encoded PNG data URL, then decodes it into an image.Image.
func captureCanvasPNG(ctx context.Context) (image.Image, error) {
	var dataURL string
	script := `(() => {
		const c = document.querySelector('canvas');
		if (!c) return '';
		return c.toDataURL('image/png');
	})()`
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &dataURL)); err != nil {
		return nil, fmt.Errorf("evaluating canvas capture script: %w", err)
	}
	if dataURL == "" {
		return nil, fmt.Errorf("no canvas element found on page")
	}

	const prefix = "data:image/png;base64,"
	encoded := strings.TrimPrefix(dataURL, prefix)
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decoding canvas base64 data: %w", err)
	}

	img, _, err := image.Decode(strings.NewReader(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("decoding canvas PNG data: %w", err)
	}
	return img, nil
}

// captureJob runs a full capture session for one manifest entry against a
// page already served at pageURL, returning the captured frames in order.
func captureJob(ctx context.Context, job CaptureJob, pageURL string) ([]image.Image, error) {
	tabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, 30*time.Second)
	defer timeoutCancel()

	if err := chromedp.Run(tabCtx,
		chromedp.Navigate(pageURL),
		chromedp.WaitVisible("canvas", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // let the WASM game finish Init()
	); err != nil {
		return nil, fmt.Errorf("loading page %s: %w", pageURL, err)
	}

	var frames []image.Image
	elapsed := 0
	inputIdx := 0

	for elapsed < job.CaptureMs {
		if inputIdx < len(job.Inputs) {
			step := job.Inputs[inputIdx]
			if err := chromedp.Run(tabCtx, holdKeys(step.Keys, step.HoldMs)); err != nil {
				return nil, fmt.Errorf("dispatching input step %d: %w", inputIdx, err)
			}
			elapsed += step.HoldMs
			inputIdx++
			continue
		}

		frame, err := captureFrame(tabCtx)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)

		time.Sleep(time.Duration(job.SampleMs) * time.Millisecond)
		elapsed += job.SampleMs
	}

	// Always end with a final captured frame so "static" jobs have output
	// even if CaptureMs was fully consumed by input steps.
	frame, err := captureFrame(tabCtx)
	if err != nil {
		return nil, err
	}
	frames = append(frames, frame)

	return frames, nil
}

func captureFrame(ctx context.Context) (image.Image, error) {
	var img image.Image
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var innerErr error
		img, innerErr = captureCanvasPNG(ctx)
		return innerErr
	}))
	return img, err
}
```

- [ ] **Step 2: Implement the CLI entry point**

```go
// cmd/docshots/main.go

// Package main implements a tool that builds PIGO8 example games to
// WebAssembly, drives them headlessly via Chrome DevTools Protocol, and
// captures screenshots/GIFs used in the documentation site.
//
// Usage:
//
//	go run ./cmd/docshots                  # capture every job in the manifest
//	go run ./cmd/docshots -only hello-world # capture a single named job
package main

import (
	"context"
	"flag"
	"fmt"
	"image/gif"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/drpaneas/pigo8/internal/webbuild"
)

func main() {
	only := flag.String("only", "", "Name of a single manifest entry to capture (default: all)")
	outDir := flag.String("out", "docs/img/generated", "Output directory for captured images")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getting working directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "creating output dir: %v\n", err)
		os.Exit(1)
	}

	jobs := manifest
	if *only != "" {
		jobs = filterJobs(manifest, *only)
		if len(jobs) == 0 {
			fmt.Fprintf(os.Stderr, "no manifest entry named %q\n", *only)
			os.Exit(1)
		}
	}

	for _, job := range jobs {
		if err := runJob(repoRoot, job, *outDir); err != nil {
			fmt.Fprintf(os.Stderr, "job %q failed: %v\n", job.Name, err)
			os.Exit(1)
		}
		fmt.Printf("captured %s\n", job.Name)
	}
}

func filterJobs(all []CaptureJob, name string) []CaptureJob {
	var out []CaptureJob
	for _, j := range all {
		if j.Name == name {
			out = append(out, j)
		}
	}
	return out
}

func runJob(repoRoot string, job CaptureJob, outDir string) error {
	gameDir := filepath.Join(repoRoot, job.ExampleDir)
	buildDir, err := os.MkdirTemp("", "docshots-build-*")
	if err != nil {
		return fmt.Errorf("creating temp build dir: %w", err)
	}
	defer os.RemoveAll(buildDir)

	modulePath := "github.com/drpaneas/pigo8/examples/" + filepath.Base(job.ExampleDir)
	if err := webbuild.EnsureExampleModule(gameDir, repoRoot, modulePath); err != nil {
		return fmt.Errorf("ensuring example module: %w", err)
	}

	wasmPath := filepath.Join(buildDir, "game.wasm")
	if err := webbuild.BuildWASM(gameDir, wasmPath); err != nil {
		return fmt.Errorf("building wasm: %w", err)
	}
	if err := webbuild.CopyWASMExec(buildDir); err != nil {
		return fmt.Errorf("copying wasm_exec.js: %w", err)
	}
	if err := webbuild.GenerateHTML(filepath.Join(buildDir, "index.html"), job.Name); err != nil {
		return fmt.Errorf("generating html: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting listener: %w", err)
	}
	server := &http.Server{Handler: http.FileServer(http.Dir(buildDir))}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()

	pageURL := fmt.Sprintf("http://%s/index.html", listener.Addr().String())

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), chromedp.DefaultExecAllocatorOptions[:]...)
	defer allocCancel()
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	frames, err := captureJob(browserCtx, job, pageURL)
	if err != nil {
		return fmt.Errorf("capturing: %w", err)
	}
	if len(frames) == 0 {
		return fmt.Errorf("no frames captured")
	}

	switch job.Kind {
	case "static":
		return savePNG(bestFrame(frames), filepath.Join(outDir, job.Name+".png"))
	case "gif":
		g, err := framesToGIF(frames, job.SampleMs)
		if err != nil {
			return fmt.Errorf("assembling gif: %w", err)
		}
		return saveGIF(g, filepath.Join(outDir, job.Name+".gif"))
	default:
		return fmt.Errorf("unknown job kind %q", job.Kind)
	}
}

func savePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func saveGIF(g *gif.GIF, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, g)
}
```

Use these imports at the top of `cmd/docshots/main.go`:

```go
import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
	"github.com/drpaneas/pigo8/internal/webbuild"
)
```

- [ ] **Step 3: Build the tool**

Run: `go build ./cmd/docshots/...`
Expected: builds cleanly with no errors.

- [ ] **Step 4: Run unit tests**

Run: `go test ./cmd/docshots/... -v`
Expected: PASS (the `TestFramesToGIF` and `TestBestFrame` tests from Task 4.3 still pass; no
new unit tests are added for `capture.go`/`main.go` since they require a real browser and are
validated via the manual end-to-end check in Task 4.5 instead).

- [ ] **Step 5: Commit**

```bash
git add cmd/docshots
git commit -m "feat: add docshots headless capture CLI"
```

### Task 4.5: Validate `cmd/docshots` end-to-end on one example before running the full batch

This is the highest-risk piece of the whole plan (real browser automation) - validate it in
isolation before trusting it for all 10 manifest entries.

**Files:** none (validation only)

- [ ] **Step 1: Run a single capture job**

Run:
```bash
go run ./cmd/docshots -only hello-world -out /tmp/docshots-test
```
Expected: prints `captured hello-world` with no errors, and creates
`/tmp/docshots-test/hello-world.png`.

- [ ] **Step 2: Visually inspect the output**

Open `/tmp/docshots-test/hello-world.png` and confirm it shows the actual PIGO8 "hello, world!"
screen (dark blue background, white text), not a blank/black/error image.

- [ ] **Step 3: Run a GIF job and inspect it**

Run:
```bash
go run ./cmd/docshots -only animation -out /tmp/docshots-test
```
Expected: creates `/tmp/docshots-test/animation.gif` showing the animated sprites moving across
frames when viewed.

- [ ] **Step 4: If either check fails, debug before proceeding**

Common failure points to check, in order:
- Chromium not found: confirm `which chromium` (or `google-chrome`) resolves; if using a
  non-standard install, pass `chromedp.ExecPath("/path/to/chromium")` as an extra option in
  `chromedp.NewExecAllocator` in `cmd/docshots/main.go`.
- Blank/black captured image: increase the initial `chromedp.Sleep` in `captureJob` (Task 4.4)
  past 500ms - the WASM binary may need longer to load and call `Init()` on a slower machine.
  Also confirm the canvas selector `document.querySelector('canvas')` actually matches (Ebiten
  inserts the canvas after WASM init, not immediately on page load).
- `EnsureExampleModule` failures: run the failing example's `go mod init`/`go mod tidy` manually
  in its directory to see the underlying error (usually a version conflict resolved by deleting
  a stale `go.sum` in that example directory before retrying).

Do not proceed to Task 5 until both checks in Steps 1-3 pass.

---

## Phase 5: Generate and Embed Visuals

### Task 5.1: Generate all manifest visuals

**Files:** none (generates files under `docs/img/generated/`)

- [ ] **Step 1: Run the full batch**

Run: `go run ./cmd/docshots`
Expected: prints `captured <name>` for all 10 manifest entries with no errors, producing:
```
docs/img/generated/hello-world.png
docs/img/generated/animation.gif
docs/img/generated/spritesheet.png
docs/img/generated/map.png
docs/img/generated/map-layers.png
docs/img/generated/camera.gif
docs/img/generated/palette.gif
docs/img/generated/color-collision.gif
docs/img/generated/map-collision.gif
docs/img/generated/pong.gif
```

- [ ] **Step 2: Spot-check each file**

Open each generated file and confirm it shows real, recognizable gameplay/output from the
correct example (not a blank canvas or loading screen). Regenerate any individual job with
`go run ./cmd/docshots -only <name>` if it needs adjustment (e.g. a longer capture window for a
GIF that looks too static - adjust that job's `CaptureMs`/`Inputs` in `cmd/docshots/manifest.go`
and re-run).

- [ ] **Step 3: Commit the generated assets**

```bash
git add docs/img/generated
git commit -m "docs: generate example screenshots and gifs"
```

### Task 5.2: Embed the visuals into their target pages

**Files:**
- Modify: `docs/02-hello-world.md`
- Modify: `docs/20-sprites/01-drawing-sprites.md`
- Modify: `docs/20-sprites/02-spritesheet-regions.md`
- Modify: `docs/30-maps/01-drawing-maps.md`
- Modify: `docs/30-maps/02-map-data.md`
- Modify: `docs/60-camera/00-camera.md`
- Modify: `docs/10-graphics/06-palette-effects.md`
- Modify: `docs/70-collision/01-color-collision.md`
- Modify: `docs/70-collision/02-map-collision.md`
- Modify: `docs/tutorials/00-pong.md`

- [ ] **Step 1: Read each target page first to find a natural insertion point**

For each file, insert the image directly after the first `##` heading following the
introduction (i.e., after the reader has enough context to understand what they're looking at,
not before the first paragraph).

- [ ] **Step 2: Insert the correct relative image reference into each page**

Use this exact mapping (paths are relative to each markdown file's own directory, matching how
`docs/editor.md` already references top-level images):

| File | Markdown to insert |
|------|---------------------|
| `docs/02-hello-world.md` | `![Hello World running in PIGO8](img/generated/hello-world.png)` |
| `docs/20-sprites/01-drawing-sprites.md` | `![Sprite animation running in PIGO8](../img/generated/animation.gif)` |
| `docs/20-sprites/02-spritesheet-regions.md` | `![Spritesheet region example](../img/generated/spritesheet.png)` |
| `docs/30-maps/01-drawing-maps.md` | `![Map rendering example](../img/generated/map.png)` |
| `docs/30-maps/02-map-data.md` | `![Map with multiple layers](../img/generated/map-layers.png)` |
| `docs/60-camera/00-camera.md` | `![Camera panning example](../img/generated/camera.gif)` |
| `docs/10-graphics/06-palette-effects.md` | `![Palette effects example](../img/generated/palette.gif)` |
| `docs/70-collision/01-color-collision.md` | `![Color collision example](../img/generated/color-collision.gif)` |
| `docs/70-collision/02-map-collision.md` | `![Map collision example](../img/generated/map-collision.gif)` |
| `docs/tutorials/00-pong.md` | `![Pong gameplay](../img/generated/pong.gif)` |

- [ ] **Step 3: Verify the build renders images correctly**

Run: `mdbook build && mdbook serve` and open the 10 pages above in a browser to confirm every
image loads (no broken image icons) and is reasonably sized (mdBook's default theme scales
images to container width, so no manual sizing should be needed).

Stop the server with Ctrl+C when done.

- [ ] **Step 4: Commit**

```bash
git add docs/02-hello-world.md docs/20-sprites docs/30-maps docs/60-camera docs/10-graphics/06-palette-effects.md docs/70-collision docs/tutorials
git commit -m "docs: embed generated screenshots and gifs into relevant pages"
```

---

## Phase 6: Navigation, Consistency & Light Branding

### Task 6.1: Add "See Also" footers to existing conceptual pages

**Files:**
- Modify: `docs/03-game-loop.md`
- Modify: `docs/10-graphics/01-screen.md`
- Modify: `docs/10-graphics/02-colors.md`
- Modify: `docs/10-graphics/03-pixels.md`
- Modify: `docs/10-graphics/04-shapes.md`
- Modify: `docs/10-graphics/05-text.md`
- Modify: `docs/10-graphics/06-palette-effects.md`
- Modify: `docs/20-sprites/01-drawing-sprites.md`
- Modify: `docs/20-sprites/02-spritesheet-regions.md`
- Modify: `docs/20-sprites/03-sprite-flags.md`
- Modify: `docs/30-maps/01-drawing-maps.md`
- Modify: `docs/30-maps/02-map-data.md`
- Modify: `docs/40-input/01-keyboard.md`
- Modify: `docs/40-input/03-mouse.md`
- Modify: `docs/50-audio/00-music.md`
- Modify: `docs/60-camera/00-camera.md`
- Modify: `docs/70-collision/01-color-collision.md`
- Modify: `docs/70-collision/02-map-collision.md`
- Modify: `docs/80-math/00-math.md`

- [ ] **Step 1: Append a `## See Also` section to each file**

Add this exact footer to the end of each file, substituting the correct reference link per the
table below:

```markdown

## See Also

- [<Reference Page Title>](<relative path to matching docs/reference/api/*.md>)
```

| File | Reference link to add |
|------|------------------------|
| `docs/03-game-loop.md` | `[Settings & Lifecycle](reference/api/10-settings-lifecycle.md)` |
| `docs/10-graphics/01-screen.md` | `[Screen & Drawing](../reference/api/01-screen-drawing.md)` |
| `docs/10-graphics/02-colors.md` | `[Colors & Palette](../reference/api/02-colors-palette.md)` |
| `docs/10-graphics/03-pixels.md` | `[Screen & Drawing](../reference/api/01-screen-drawing.md)` |
| `docs/10-graphics/04-shapes.md` | `[Screen & Drawing](../reference/api/01-screen-drawing.md)` |
| `docs/10-graphics/05-text.md` | `[Screen & Drawing](../reference/api/01-screen-drawing.md)` |
| `docs/10-graphics/06-palette-effects.md` | `[Colors & Palette](../reference/api/02-colors-palette.md)` |
| `docs/20-sprites/01-drawing-sprites.md` | `[Sprites](../reference/api/03-sprites.md)` |
| `docs/20-sprites/02-spritesheet-regions.md` | `[Sprites](../reference/api/03-sprites.md)` |
| `docs/20-sprites/03-sprite-flags.md` | `[Sprites](../reference/api/03-sprites.md)` |
| `docs/30-maps/01-drawing-maps.md` | `[Maps](../reference/api/04-maps.md)` |
| `docs/30-maps/02-map-data.md` | `[Maps](../reference/api/04-maps.md)` |
| `docs/40-input/01-keyboard.md` | `[Input](../reference/api/05-input.md)` |
| `docs/40-input/03-mouse.md` | `[Input](../reference/api/05-input.md)` |
| `docs/50-audio/00-music.md` | `[Audio](../reference/api/06-audio.md)` |
| `docs/60-camera/00-camera.md` | `[Camera](../reference/api/07-camera.md)` |
| `docs/70-collision/01-color-collision.md` | `[Collision](../reference/api/08-collision.md)` |
| `docs/70-collision/02-map-collision.md` | `[Collision](../reference/api/08-collision.md)` |
| `docs/80-math/00-math.md` | `[Math](../reference/api/09-math.md)` |

Note: `docs/10-graphics/06-palette-effects.md`, `docs/20-sprites/01-drawing-sprites.md`,
`docs/20-sprites/02-spritesheet-regions.md`, `docs/60-camera/00-camera.md`,
`docs/70-collision/01-color-collision.md`, and `docs/70-collision/02-map-collision.md` already
had an image embedded in Task 5.2 - add the `See Also` section after that image, at the very end
of the file.

- [ ] **Step 2: Run the link checker to make sure every new link resolves**

Run: `go run ./cmd/linkcheck -dir docs`
Expected: `No broken links found.`

- [ ] **Step 3: Commit**

```bash
git add docs
git commit -m "docs: add See Also cross-links to conceptual pages"
```

### Task 6.2: Light branding via `book.toml`

**Files:**
- Modify: `book.toml`

- [ ] **Step 1: Add an `[output.html]` section wiring up the existing logo as favicon**

```toml
[book]
authors = ["drpaneas"]
language = "en"
multilingual = false
src = "docs" # <- this is where the markdown files should be
title = "PIGO8 Documentation"
description = "Technical reference for the PIGO8 library to help you make PICO-8 games using Go."

[build]
build-dir = "book" # <- default and expected by GH Actions. No need to create thid folder, Github will do.

[output.html]
favicon = "logo.png"
git-repository-url = "https://github.com/drpaneas/pigo8"
edit-url-template = "https://github.com/drpaneas/pigo8/edit/main/docs/{path}"
```

- [ ] **Step 2: Verify the build still works**

Run: `mdbook build`
Expected: builds successfully; open `book/index.html` in a browser and confirm the favicon tab
icon shows the PIGO8 logo, and a GitHub link icon appears in the page header.

- [ ] **Step 3: Commit**

```bash
git add book.toml
git commit -m "docs: wire up favicon and GitHub edit links in mdBook config"
```

---

## Phase 7: Final Validation

### Task 7.1: Full link check, build, and code quality pass

**Files:** none (validation only)

- [ ] **Step 1: Run the link checker against the full docs tree**

Run: `go run ./cmd/linkcheck -dir docs`
Expected: `No broken links found.`

- [ ] **Step 2: Run a full mdBook build**

Run: `mdbook build`
Expected: exits 0 with no warnings printed about missing SUMMARY.md targets or broken
references.

- [ ] **Step 3: Run the full Go test suite and linter**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: all pass. If `.golangci.yml` is configured for CI, also run: `golangci-lint run
./cmd/linkcheck/... ./cmd/docshots/... ./internal/webbuild/... ./cmd/webexport/...` and fix any
findings in the new/modified code.

- [ ] **Step 4: Verify every code sample that's a complete program compiles**

For each `go doc`-derived example in the new API Reference pages that shows a full `package
main` (there are none as standalone full programs in the reference pages themselves - they're
snippets - but double check the FAQ and Testing pages' full examples):

Run:
```bash
cd /tmp && mkdir pigo8-doc-verify && cd pigo8-doc-verify
go mod init pigo8-doc-verify
go get github.com/drpaneas/pigo8@latest
```

Then manually paste each full `package main` example from `docs/90-advanced/03-testing.md` and
confirm `go vet ./...` accepts the syntax (the testing examples are plain Go with no PIGO8 calls
in the tested logic, so this mainly confirms no typos).

- [ ] **Step 5: Manual spot-check via `mdbook serve`**

Run: `mdbook serve` and browse to:
- The new API Reference section (all 11 pages) - confirm formatting/tables render correctly.
- The FAQ page.
- The Testing Your Game page.
- At least 3 of the 10 pages with embedded images from Phase 5.

Stop the server with Ctrl+C when done.

- [ ] **Step 6: Final commit**

If Steps 3-4 required any fixes:

```bash
git add -A
git commit -m "fix: address final validation findings"
```

If no fixes were needed, no commit is necessary for this task.
