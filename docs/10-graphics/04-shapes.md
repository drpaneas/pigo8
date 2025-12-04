# Drawing Shapes

## Rectangles

### Rect - Outline Rectangle

```go
p8.Rect(x1, y1, x2, y2, color)
```

Draws a 1-pixel outline rectangle from corner (x1, y1) to corner (x2, y2).

```go
p8.Rect(10, 10, 50, 30, 7)  // White outline
```

### Rectfill - Filled Rectangle

```go
p8.Rectfill(x1, y1, x2, y2, color)
```

Draws a filled rectangle.

```go
p8.Rectfill(10, 10, 50, 30, 8)  // Red filled rectangle
```

## Circles

### Circ - Outline Circle

```go
p8.Circ(x, y, radius, color)
```

Draws a 1-pixel outline circle centered at (x, y).

```go
p8.Circ(64, 64, 20, 12)  // Blue circle at center, radius 20
```

### Circfill - Filled Circle

```go
p8.Circfill(x, y, radius, color)
```

Draws a filled circle.

```go
p8.Circfill(64, 64, 20, 11)  // Green filled circle
```

## Lines

```go
p8.Line(x1, y1, x2, y2, color)
```

Draws a line from (x1, y1) to (x2, y2).

```go
p8.Line(0, 0, 127, 127, 7)    // Diagonal white line
p8.Line(64, 0, 64, 127, 5)    // Vertical gray line
```

> **Note:** Unlike `Rect` and `Circ`, the `Line` function does not apply camera offset. If you're using `Camera()` for scrolling, lines will remain fixed on screen.

## Color Parameter

All shape functions accept an optional color parameter:

- **With color**: Uses the specified color
- **Without color**: Uses the current draw color (set by `Color()` or `Cursor()`)

```go
p8.Color(8)         // Set current color to red
p8.Rect(10, 10, 50, 30)  // Draws in red
p8.Rect(60, 10, 100, 30, 12)  // Draws in blue (override)
```

## Example: Simple Scene

```go
func (g *game) Draw() {
    p8.Cls(12)  // Sky blue background
    
    // Sun
    p8.Circfill(100, 20, 12, 10)  // Yellow sun
    
    // Ground
    p8.Rectfill(0, 100, 127, 127, 3)  // Dark green grass
    
    // House
    p8.Rectfill(40, 70, 80, 100, 4)   // Brown walls
    p8.Rectfill(55, 85, 65, 100, 5)   // Gray door
    
    // Roof (triangle using lines)
    p8.Line(35, 70, 60, 50, 8)   // Left roof edge
    p8.Line(60, 50, 85, 70, 8)   // Right roof edge
    p8.Line(35, 70, 85, 70, 8)   // Bottom edge
}
```

