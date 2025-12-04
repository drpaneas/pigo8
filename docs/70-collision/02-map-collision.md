# Map Collision

Check if a rectangular area overlaps with flagged map tiles.

## MapCollision Function

```go
p8.MapCollision(x, y, flag, width, height) bool
```

Returns `true` if any tile under the area has the specified flag set.

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| x | Number | required | Left edge (pixels) |
| y | Number | required | Top edge (pixels) |
| flag | int | required | Flag to check (0-7) |
| width | int | 8 | Area width (pixels) |
| height | int | 8 | Area height (pixels) |

## Basic Usage

```go
// Check 8×8 area at player position
if p8.MapCollision(g.playerX, g.playerY, 0) {
    // Hit a tile with flag 0 (solid)
}
```

## Custom Hitbox

```go
// Player is 6×12 pixels
playerWidth := 6
playerHeight := 12
if p8.MapCollision(g.x, g.y, 0, playerWidth, playerHeight) {
    // Collision
}
```

## Platformer Movement

```go
func (g *game) Update() {
    // Horizontal movement
    newX := g.x + g.velocityX
    if !p8.MapCollision(newX, g.y, 0, 8, 8) {
        g.x = newX
    } else {
        g.velocityX = 0
    }
    
    // Vertical movement
    newY := g.y + g.velocityY
    if !p8.MapCollision(g.x, newY, 0, 8, 8) {
        g.y = newY
    } else {
        if g.velocityY > 0 {
            g.onGround = true
        }
        g.velocityY = 0
    }
}
```

## Different Collision Types

Use different flags for different behaviors:

```go
const (
    FlagSolid    = 0  // Can't pass through
    FlagPlatform = 1  // Pass through from below
    FlagHazard   = 2  // Damages player
    FlagWater    = 3  // Slows movement
)

func (g *game) Update() {
    // Check solid collision
    if p8.MapCollision(g.x, g.y+1, FlagSolid) {
        g.onGround = true
    }
    
    // Check hazards
    if p8.MapCollision(g.x, g.y, FlagHazard) {
        g.takeDamage()
    }
    
    // Check water (for movement speed)
    g.inWater = p8.MapCollision(g.x, g.y, FlagWater)
}
```

