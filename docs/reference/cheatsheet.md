# PIGO8 Cheatsheet

For full explanations, parameter tables, and examples, see the [API Reference](api/00-overview.md).

## Quick Reference

### Initialization

```go
p8.InsertGame(&myGame{})  // Register game
p8.Play()                 // Start with defaults
p8.PlayGameWith(settings) // Start with custom settings
```

### Screen

```go
p8.Cls(color)            // Clear screen
p8.GetScreenWidth()      // Get width (128)
p8.GetScreenHeight()     // Get height (128)
```

### Drawing

```go
p8.Pset(x, y, color)     // Set pixel
p8.Pget(x, y)            // Get pixel color
p8.Line(x1, y1, x2, y2, color)
p8.Rect(x1, y1, x2, y2, color)
p8.Rectfill(x1, y1, x2, y2, color)
p8.Circ(x, y, r, color)
p8.Circfill(x, y, r, color)
p8.Print(text, x, y, color)
```

### Sprites

```go
p8.Spr(n, x, y)                    // Draw sprite
p8.Spr(n, x, y, w, h, flipX, flipY)
p8.Sspr(sx, sy, sw, sh, dx, dy)    // Draw region
p8.Sget(x, y)                      // Get spritesheet pixel
p8.Sset(x, y, color)               // Set spritesheet pixel
bitfield, isSet := p8.Fget(sprite, flag)  // Get flag (returns 2 values)
p8.Fset(sprite, flag, value)              // Set sprite flag
```

### Maps

```go
p8.Map(mx, my, sx, sy, w, h, layers)
p8.Mget(x, y)            // Get tile sprite
p8.Mset(x, y, sprite)    // Set tile sprite
```

### Input

```go
p8.Btn(button)           // Is button held?
p8.Btnp(button)          // Was button just pressed?
p8.GetMouseXY()          // Get mouse position
// Buttons: LEFT, RIGHT, UP, DOWN, O (Z), X (X)
// ButtonStart (Enter), ButtonMouseLeft, etc.
```

### Audio

```go
p8.Music(n)              // Play audio once
p8.Music(n, true)        // Play exclusively
p8.MusicLoop(n)          // Play in a loop
p8.StopMusic(n)          // Stop specific
p8.StopMusic(-1)         // Stop all
// F32 variants also exist: MusicF32, MusicLoopF32, StopMusicF32
```

### Camera

```go
p8.Camera(x, y)          // Set offset
p8.Camera()              // Reset to (0,0)
```

### Palette

```go
p8.Pal(c0, c1)           // Swap colors
p8.Pal()                 // Reset palette swap
p8.Palt(color, transparent)  // Set transparency
p8.Palt()                // Reset transparency
p8.Color(c)              // Set draw color
```

### Collision

```go
p8.ColorCollision(x, y, color)
p8.MapCollision(x, y, flag, w, h)
```

### Math

```go
p8.Flr(n)                // Floor (returns int)
p8.Rnd(n)                // Random 0 to n-1 (returns int)
p8.Sqrt(n)               // Square root (returns float64)
p8.Sign(float64(n))      // -1.0 or 1.0 (takes float64)
p8.Time()                // Seconds elapsed (returns float64)
```

### Settings

```go
settings := p8.NewSettings()
settings.ScreenWidth = 160
settings.ScreenHeight = 144
settings.ScaleFactor = 4
settings.TargetFPS = 60
settings.WindowTitle = "My Game"
settings.Fullscreen = true
settings.DisableHiDPI = true  // Default: true
```

## Color Palette

| 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
|---|---|---|---|---|---|---|---|
| Black | DarkBlue | DarkPurple | DarkGreen | Brown | DarkGray | LightGray | White |

| 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 |
|---|---|---|---|---|---|---|---|
| Red | Orange | Yellow | Green | Blue | Indigo | Pink | Peach |

## Button Constants

| Direction | Face | Menu | Mouse |
|-----------|------|------|-------|
| `LEFT` (0) | `O` (4) | `ButtonStart` (6) | `ButtonMouseLeft` |
| `RIGHT` (1) | `X` (5) | `ButtonSelect` (7) | `ButtonMouseRight` |
| `UP` (2) | | | `ButtonMouseMiddle` |
| `DOWN` (3) | | | |

## Game Structure

```go
type game struct {
    // Your state
}

func (g *game) Init()   { /* once at start */ }
func (g *game) Update() { /* every frame */ }
func (g *game) Draw()   { /* every frame */ }

func main() {
    p8.InsertGame(&game{})
    p8.Play()
}
```

## Common Patterns

### Movement

```go
if p8.Btn(p8.LEFT)  { g.x-- }
if p8.Btn(p8.RIGHT) { g.x++ }
if p8.Btn(p8.UP)    { g.y-- }
if p8.Btn(p8.DOWN)  { g.y++ }
```

### Animation

```go
frame := int(p8.Time() * 10) % 4
p8.Spr(1 + frame, x, y)
```

### Tile Collision

```go
tileX := p8.Flr(g.x / 8)
tileY := p8.Flr(g.y / 8)
if p8.MapCollision(g.x, g.y, 0) {
    // Hit solid tile
}
```

