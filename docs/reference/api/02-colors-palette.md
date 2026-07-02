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
| `color` | `int` | PICO-8 color index (0-15). |
| `transparent` | `bool` | `true` to make the color transparent, `false` to make it opaque. |

**Note:** the real signature is variadic (`args ...interface{}`); `color` and `transparent` are extracted positionally from `args` rather than being named parameters (the same pattern used by `Cursor`/`Print`).

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

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `colorIndex` | `int` | PICO-8 color index (0-15) to look up. |

**Example:**
```go
c := p8.GetPaletteColor(3)
```

**See also:** [Colors and Palette](../../10-graphics/02-colors.md)

## `GetPaletteSize() int`

Returns the current number of colors in the active palette.

**Example:**
```go
size := p8.GetPaletteSize()
p8.Print(fmt.Sprintf("Palette has %d colors", size), 10, 10, 7)
```

**See also:** [Colors and Palette](../../10-graphics/02-colors.md)

## `SetPalette(newPalette []color.Color)`

Replaces the entire color palette. Resizes the transparency array to match, resetting only index
0 as transparent by default.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `newPalette` | `[]color.Color` | The full replacement palette. |

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

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `colorIndex` | `int` | PICO-8 color index (0-15) to replace. |
| `newColor` | `color.Color` | The new color to place at that index. |

**Example:**
```go
p8.SetPaletteColor(7, color.RGBA{200, 220, 255, 255}) // change white to light blue
```

**See also:** [Palette Effects](../../10-graphics/06-palette-effects.md)
