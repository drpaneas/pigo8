# The Game Loop

Every game runs a continuous loop: read input, update state, render graphics. PIGO8 handles the loop for you through three methods.

## Init, Update, Draw

```go
type Cartridge interface {
    Init()   // Called once when the game starts
    Update() // Called every frame (default: 30 times/second)
    Draw()   // Called every frame after Update
}
```

### Init

Runs once before the first frame. Use it to:

- Set initial positions
- Load level data
- Initialize scores

```go
func (g *game) Init() {
    g.playerX = 64
    g.playerY = 64
    g.score = 0
}
```

### Update

Runs every frame before `Draw`. Use it to:

- Read input
- Move game objects
- Check collisions
- Update game state

```go
func (g *game) Update() {
    if p8.Btn(p8.LEFT) {
        g.playerX--
    }
    if p8.Btn(p8.RIGHT) {
        g.playerX++
    }
}
```

**Important**: Never draw in `Update`. Drawing operations only work in `Draw`.

### Draw

Runs every frame after `Update`. Use it to:

- Clear the screen
- Draw backgrounds, sprites, and UI
- Render everything visible

```go
func (g *game) Draw() {
    p8.Cls(0)                          // Always clear first
    p8.Spr(1, g.playerX, g.playerY)    // Draw player sprite
    p8.Print(g.score, 2, 2, 7)         // Draw score
}
```

## Frame Timing

By default, PIGO8 runs at **30 FPS** (frames per second). Each frame:

1. `Update()` is called
2. `Draw()` is called
3. The screen is displayed
4. Wait until the next frame is due

At 30 FPS, each frame is ~33 milliseconds. At 60 FPS, each frame is ~16 milliseconds.

## Changing FPS

Use settings to change the frame rate:

```go
func main() {
    settings := p8.NewSettings()
    settings.TargetFPS = 60  // Smoother animation
    
    p8.InsertGame(&game{})
    p8.PlayGameWith(settings)
}
```

Higher FPS means smoother animation but faster game speed if you're using fixed movement values. Consider using delta time for frame-rate independent movement.

## Time Function

Get the elapsed time since the game started:

```go
func (g *game) Update() {
    elapsed := p8.Time()  // Returns seconds as float64
    
    // Do something every 2 seconds
    if int(elapsed) % 2 == 0 {
        // ...
    }
}
```

## Complete Example: Moving Square

```go
package main

import p8 "github.com/drpaneas/pigo8"

type game struct {
    x, y float64
}

func (g *game) Init() {
    g.x = 60
    g.y = 60
}

func (g *game) Update() {
    if p8.Btn(p8.LEFT)  { g.x-- }
    if p8.Btn(p8.RIGHT) { g.x++ }
    if p8.Btn(p8.UP)    { g.y-- }
    if p8.Btn(p8.DOWN)  { g.y++ }
}

func (g *game) Draw() {
    p8.Cls(0)
    p8.Rectfill(g.x, g.y, g.x+8, g.y+8, 8)  // Red square
}

func main() {
    p8.InsertGame(&game{})
    p8.Play()
}
```

Arrow keys move the red square. The game loop makes it responsive.

