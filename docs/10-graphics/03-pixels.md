# Pixel Manipulation

## Pset - Set Pixel

Draw a single pixel at (x, y):

```go
p8.Pset(64, 64, 8)  // Red pixel at center
p8.Pset(10, 20)     // Use current draw color
```

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| x | int | X coordinate |
| y | int | Y coordinate |
| color | int (optional) | Color index (0-15) |

> **Note:** Unlike shape functions, `Pset` and `Pget` take `int` parameters, not generic Number types. Use `int()` to convert if needed.

## Pget - Get Pixel

Read the color index at (x, y):

```go
color := p8.Pget(64, 64)  // Returns 0-15
```

Returns 0 if coordinates are out of bounds.

### Use Cases

- **Collision detection**: Check if a pixel is a certain color
- **Color picking**: Find what color is at a position
- **Procedural generation**: Read and modify existing pixels

## Example: Starfield

```go
type game struct {
    stars [][2]int  // x, y pairs
}

func (g *game) Init() {
    // Create 50 random stars
    for i := 0; i < 50; i++ {
        x := p8.Rnd(128)
        y := p8.Rnd(128)
        g.stars = append(g.stars, [2]int{x, y})
    }
}

func (g *game) Draw() {
    p8.Cls(0)
    for _, star := range g.stars {
        // Random twinkle: white or light gray
        color := 7
        if p8.Rnd(10) < 3 {
            color = 6
        }
        p8.Pset(star[0], star[1], color)
    }
}
```

## Color Function

Set the current draw color used by functions when no color is specified:

```go
p8.Color(8)        // Set current color to red
p8.Pset(10, 10)    // Draws red (uses current color)
p8.Pset(20, 20, 12) // Draws blue (overrides current color)
```

