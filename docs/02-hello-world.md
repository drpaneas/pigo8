# Hello World

Let's build the simplest possible PIGO8 game: a screen that displays "Hello, World!"

## The Complete Game

Create `main.go`:

```go
package main

import p8 "github.com/drpaneas/pigo8"

type game struct{}

func (g *game) Init()   {}
func (g *game) Update() {}

func (g *game) Draw() {
    p8.Cls(1)                          // Clear screen with dark blue
    p8.Print("hello, world!", 40, 60)  // Draw text at (40, 60)
}

func main() {
    p8.InsertGame(&game{})
    p8.Play()
}
```

Run it:

```bash
go run .
```

## What's Happening?

### The Game Struct

```go
type game struct{}
```

Your game state lives in a struct. For now it's empty, but you'll add player position, score, and other state here.

### The Cartridge Interface

PIGO8 expects your game to implement three methods:

```go
func (g *game) Init()   {}  // Called once at startup
func (g *game) Update() {}  // Called every frame for logic
func (g *game) Draw()   {}  // Called every frame for rendering
```

This is the **game loop**—the heartbeat of every game.

### Inserting and Playing

```go
p8.InsertGame(&game{})  // Register your game
p8.Play()               // Start the engine
```

Think of it like inserting a cartridge into a console.

## Understanding the Draw Function

```go
func (g *game) Draw() {
    p8.Cls(1)                          // Clear to color 1 (dark blue)
    p8.Print("hello, world!", 40, 60)  // White text by default
}
```

- `Cls(colorIndex)` clears the screen. Color 1 is dark blue in the default palette.
- `Print(text, x, y)` draws text. Coordinates start at (0, 0) in the top-left corner.

## Next Steps

You now have a working PIGO8 game. In the next chapter, we'll explore the game loop in detail and add interactivity.

