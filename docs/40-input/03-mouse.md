# Mouse Input

## Mouse Position

```go
x, y := p8.GetMouseXY()
```

Returns the current mouse position in game coordinates.

### Example: Cursor

```go
func (g *game) Draw() {
    p8.Cls(0)
    
    mx, my := p8.GetMouseXY()
    p8.Circ(mx, my, 2, 7)  // Draw cursor
}
```

## Mouse Buttons

| Constant | Button |
|----------|--------|
| `p8.ButtonMouseLeft` | Left click |
| `p8.ButtonMouseRight` | Right click |
| `p8.ButtonMouseMiddle` | Middle click |
| `p8.ButtonMouseWheelUp` | Scroll up |
| `p8.ButtonMouseWheelDown` | Scroll down |

### Click Detection

```go
func (g *game) Update() {
    mx, my := p8.GetMouseXY()
    
    if p8.Btnp(p8.ButtonMouseLeft) {
        // Left mouse button just clicked
        g.handleClick(mx, my)
    }
    
    if p8.Btn(p8.ButtonMouseLeft) {
        // Left mouse button held (dragging)
        g.handleDrag(mx, my)
    }
}
```

### Scroll Wheel

```go
func (g *game) Update() {
    if p8.Btn(p8.ButtonMouseWheelUp) {
        g.zoom++
    }
    if p8.Btn(p8.ButtonMouseWheelDown) {
        g.zoom--
    }
}
```

## Example: Paint Program

```go
type game struct {
    lastX, lastY int
}

func (g *game) Update() {
    mx, my := p8.GetMouseXY()
    
    if p8.Btn(p8.ButtonMouseLeft) {
        // Draw line from last position to current
        p8.Line(g.lastX, g.lastY, mx, my, 7)
    }
    
    g.lastX = mx
    g.lastY = my
}
```

