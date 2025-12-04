# Screen Functions

## Cls - Clear Screen

`Cls(colorIndex)` fills the entire screen with a single color.

```go
func (g *game) Draw() {
    p8.Cls(0)   // Clear to black
    // ... draw everything else
}
```

**Always call Cls at the start of Draw.** Otherwise, you'll see ghosting from the previous frame.

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| colorIndex | int | 0 | Color index (0-15) |

### Examples

```go
p8.Cls()     // Clear to black (color 0)
p8.Cls(1)    // Clear to dark blue
p8.Cls(12)   // Clear to blue
```

## ClsRGBA - Clear with Custom Color

For colors outside the palette, use `ClsRGBA`:

```go
import "image/color"

func (g *game) Draw() {
    p8.ClsRGBA(color.RGBA{R: 50, G: 50, B: 80, A: 255})
    // ...
}
```

## Screen Dimensions

Get the current screen size at runtime:

```go
width := p8.GetScreenWidth()   // Default: 128
height := p8.GetScreenHeight() // Default: 128
```

Useful for centering objects or handling custom resolutions:

```go
func (g *game) Draw() {
    p8.Cls(0)
    
    // Center a message
    msg := "game over"
    x := (p8.GetScreenWidth() - len(msg)*4) / 2
    y := p8.GetScreenHeight() / 2
    p8.Print(msg, x, y, 8)
}
```

