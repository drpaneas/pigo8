# Space Invaders

A classic Space Invaders game built with [PIGO8](https://github.com/drpaneas/pigo8), inspired by the 1978 arcade hit.

## 🎮 Play Online

**[▶️ Play Space Invaders in your browser](https://drpaneas.github.io/pigo8/space_invaders/)**

## 📖 Description

This example demonstrates a complete game implementation in PIGO8. It shows:
- Player movement and shooting mechanics
- Enemy AI with random shooting patterns
- Collision detection between bullets and aliens
- Score tracking and lives system
- Game over and restart functionality
- Sound effects with `p8.Music()`

### Controls

- **Arrow Keys**: Move spaceship left/right
- **Z (O button)**: Shoot / Restart game

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/space_invaders
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

import (
	"fmt"
	"github.com/drpaneas/pigo8"
)

type Game struct {
	playerX, playerY int
	lives            int
	bullets          []bullet
	aliens           []alien
	score            int
	gameOver         bool
}

func (g *Game) Init() {
	g.playerX = 64
	g.playerY = 120
	g.lives = 7
	g.initAliens()
}

func (g *Game) Update() {
	// Player movement
	if pigo8.Btn(pigo8.LEFT) && g.playerX > 8 {
		g.playerX -= 2
	}
	if pigo8.Btn(pigo8.RIGHT) && g.playerX < 120 {
		g.playerX += 2
	}
	
	// Shooting
	if pigo8.Btnp(pigo8.O) {
		g.bullets = append(g.bullets, bullet{x: g.playerX, y: g.playerY - 8})
	}
	
	// Update bullets and check collisions...
}

func (g *Game) Draw() {
	pigo8.Cls(0)
	// Draw player triangle
	pigo8.Line(g.playerX+4, g.playerY-8, g.playerX, g.playerY, 7)
	pigo8.Line(g.playerX+4, g.playerY-8, g.playerX+8, g.playerY, 7)
	// Draw aliens, bullets, UI...
	pigo8.Print(fmt.Sprintf("score: %d", g.score), 4, 4, 7)
}

func main() {
	pigo8.InsertGame(NewGame())
	pigo8.Play()
}
```
