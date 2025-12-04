# Color Collision Demo

A demonstration of pixel-perfect color collision detection built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Color Collision Demo in your browser](https://drpaneas.github.io/pigo8/colorCollision/)**

## 📖 Description

This example demonstrates how to implement pixel-perfect collision detection in PIGO8:
- Using `p8.ColorCollision()` to detect collisions with specific colors
- Moving a player through a maze without passing through walls
- Checking collision before applying movement

### Controls

- **Arrow Keys**: Move the player pixel

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/colorCollision
go run .
```

## 📝 Source Code

```go
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type Player struct {
	x, y, speed    float64
	collisionColor int
}

type Game struct {
	player Player
}

func (g *Game) Init() {
	g.player = Player{
		x:              10,
		y:              10,
		speed:          1,
		collisionColor: 10, // Yellow walls
	}
}

func (g *Game) Update() {
	beforeX := g.player.x
	beforeY := g.player.y

	// Check collision AFTER moving, revert if collision detected
	if p8.Btn(p8.LEFT) {
		g.player.x -= g.player.speed
		if p8.ColorCollision(g.player.x, g.player.y, g.player.collisionColor) {
			g.player.x = beforeX
		}
	}
	// ... similar for other directions
}

func (g *Game) Draw() {
	p8.Cls()
	p8.Rectfill(g.player.x, g.player.y, g.player.x, g.player.y, 12)
	
	// Draw maze walls with collision color
	p8.Line(30, 30, 30, 100, g.player.collisionColor)
	// ... more walls
}

func main() {
	p8.InsertGame(&Game{})
	p8.Play()
}
```

