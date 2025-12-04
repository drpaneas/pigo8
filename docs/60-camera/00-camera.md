# Camera

The camera offsets all drawing operations, creating the illusion of a scrolling world.

## Camera Function

```go
p8.Camera(x, y)  // Set camera offset
p8.Camera()      // Reset to (0, 0)
```

When the camera is at (x, y), everything draws shifted by (-x, -y).

## Following a Player

```go
func (g *game) Draw() {
    p8.Cls(0)
    
    // Center camera on player
    camX := g.playerX - 64  // Center of 128-pixel screen
    camY := g.playerY - 64
    
    p8.Camera(camX, camY)
    
    // Draw world
    p8.Map()
    p8.Spr(1, g.playerX, g.playerY)  // Use world coordinates
    
    // Draw UI (reset camera first)
    p8.Camera()
    p8.Print("SCORE: " + fmt.Sprint(g.score), 2, 2, 7)
}
```

## Camera Bounds

Keep the camera within map boundaries:

```go
func (g *game) updateCamera() {
    // Target camera position
    camX := g.playerX - 64
    camY := g.playerY - 64
    
    // Clamp to map boundaries
    mapWidth := 128 * 8   // 128 tiles × 8 pixels
    mapHeight := 128 * 8
    
    if camX < 0 { camX = 0 }
    if camY < 0 { camY = 0 }
    if camX > mapWidth - 128 { camX = mapWidth - 128 }
    if camY > mapHeight - 128 { camY = mapHeight - 128 }
    
    g.camX = camX
    g.camY = camY
}
```

## Smooth Camera

Add smoothing for a less jarring follow:

```go
func (g *game) updateCamera() {
    targetX := g.playerX - 64
    targetY := g.playerY - 64
    
    // Smooth interpolation (10% per frame)
    g.camX += (targetX - g.camX) * 0.1
    g.camY += (targetY - g.camY) * 0.1
}
```

## Screen Shake

Add impact effects:

```go
func (g *game) Draw() {
    p8.Cls(0)
    
    // Apply shake offset
    shakeX := 0.0
    shakeY := 0.0
    if g.shakeFrames > 0 {
        shakeX = float64(p8.Rnd(5)) - 2
        shakeY = float64(p8.Rnd(5)) - 2
        g.shakeFrames--
    }
    
    p8.Camera(g.camX + shakeX, g.camY + shakeY)
    // ... draw world
}
```

