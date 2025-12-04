# Drawing Sprites

## Spr - Draw a Sprite

The most common way to draw sprites:

```go
p8.Spr(spriteNumber, x, y)
p8.Spr(spriteNumber, x, y, width, height)
p8.Spr(spriteNumber, x, y, width, height, flipX)
p8.Spr(spriteNumber, x, y, width, height, flipX, flipY)
```

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| spriteNumber | int | required | Sprite ID from spritesheet |
| x | Number | required | Screen X position |
| y | Number | required | Screen Y position |
| width | float64 | 1.0 | Width in sprite units (1 = 8 pixels) |
| height | float64 | 1.0 | Height in sprite units |
| flipX | bool | false | Flip horizontally |
| flipY | bool | false | Flip vertically |

### Basic Usage

```go
// Draw sprite 1 at position (60, 60)
p8.Spr(1, 60, 60)

// Draw with float coordinates (for smooth movement)
p8.Spr(1, g.playerX, g.playerY)
```

### Multi-Tile Sprites

Draw sprites that span multiple tiles:

```go
// Draw a 2×2 sprite (16×16 pixels)
// This draws sprites 1, 2, 17, 18 as a block
p8.Spr(1, 50, 50, 2, 2)

// Draw a 1.5 width (12 pixels wide)
p8.Spr(1, 50, 50, 1.5, 1)
```

### Flipping

Mirror sprites for left/right or up/down facing:

```go
// Player facing right
p8.Spr(1, g.x, g.y)

// Player facing left (flip horizontally)
p8.Spr(1, g.x, g.y, 1, 1, true)

// Upside down
p8.Spr(1, g.x, g.y, 1, 1, false, true)
```

### Animation Example

```go
type game struct {
    frame     int
    animFrame int
}

func (g *game) Update() {
    g.frame++
    // Change animation frame every 8 game frames
    if g.frame % 8 == 0 {
        g.animFrame = (g.animFrame + 1) % 4
    }
}

func (g *game) Draw() {
    p8.Cls(0)
    // Sprites 1-4 are animation frames
    p8.Spr(1 + g.animFrame, 60, 60)
}
```

## Transparency

By default, black pixels (color 0) are transparent. To change this:

```go
// Make color 0 opaque (black is drawn)
p8.Palt(0, false)

// Make color 8 (red) transparent
p8.Palt(8, true)

// Reset to default
p8.Palt()
```

