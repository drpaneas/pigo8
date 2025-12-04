# Pause Menu Demo

A demonstration of the built-in pause menu functionality built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Pause Menu Demo in your browser](https://drpaneas.github.io/pigo8/pause_menu/)**

## 📖 Description

This example demonstrates the pause system in PIGO8:
- Using the built-in pause menu with the Start button
- Starfield background effect with blinking stars
- Basic player movement
- The pause state is handled internally by the engine

### Controls

- **Arrow Keys**: Move the spaceship
- **Start (Enter)**: Pause/unpause the game

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/pause_menu
go run .
```

## 📝 Source Code

```go
package main

import (
	"math/rand"
	p8 "github.com/drpaneas/pigo8"
)

type Game struct {
	playerX, playerY int
	stars            []Star
}

type Star struct {
	x, y, color, speed, blink int
	active                    bool
}

func (g *Game) Init() {
	g.playerX = p8.GetScreenWidth() / 2
	g.playerY = p8.GetScreenHeight() / 2
	
	g.stars = make([]Star, 50)
	for i := range g.stars {
		g.stars[i] = Star{
			x:      rand.Intn(p8.GetScreenWidth()),
			y:      rand.Intn(p8.GetScreenHeight()),
			color:  5 + rand.Intn(3),
			speed:  1 + rand.Intn(2),
			blink:  rand.Intn(30),
			active: true,
		}
	}
}

func (g *Game) Update() {
	// Update stars with blinking effect
	for i := range g.stars {
		g.stars[i].y += g.stars[i].speed
		if g.stars[i].y > p8.GetScreenHeight() {
			g.stars[i].y = 0
			g.stars[i].x = rand.Intn(p8.GetScreenWidth())
		}
		g.stars[i].blink--
		if g.stars[i].blink <= 0 {
			g.stars[i].active = !g.stars[i].active
			g.stars[i].blink = rand.Intn(30)
		}
	}
	
	// Player movement
	if p8.Btn(p8.LEFT) && g.playerX > 0 {
		g.playerX--
	}
	if p8.Btn(p8.RIGHT) && g.playerX < p8.GetScreenWidth()-8 {
		g.playerX++
	}
	// ... other directions
}

func (g *Game) Draw() {
	p8.Cls(0)
	
	// Draw stars
	for _, star := range g.stars {
		if star.active {
			p8.Pset(star.x, star.y, star.color)
		}
	}
	
	p8.Spr(1, g.playerX, g.playerY)
	p8.Print("PRESS START TO PAUSE", 10, 2, 7)
}

func main() {
	p8.InsertGame(&Game{})
	settings := p8.NewSettings()
	settings.WindowTitle = "PIGO8 Pause Menu Demo"
	p8.PlayGameWith(settings)
}
```

