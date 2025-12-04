# Custom Resolution Demo

A Game Boy-style demo showcasing custom screen resolution built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Custom Resolution Demo in your browser](https://drpaneas.github.io/pigo8/customResolution/)**

## 📖 Description

This example demonstrates how to use custom screen resolutions in PIGO8:
- Setting custom screen dimensions (160x144 for Game Boy style)
- Using `p8.GetScreenWidth()` and `p8.GetScreenHeight()` for responsive code
- Creating a starfield effect
- Drawing screen borders

### Controls

- **Arrow Keys**: Move the spaceship

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/customResolution
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
	stars            []star
}

type star struct {
	x, y, color, speed int
}

func (g *Game) Init() {
	g.playerX = 80
	g.playerY = 120
	
	// Create stars using screen dimensions
	g.stars = make([]star, 50)
	for i := range g.stars {
		g.stars[i] = star{
			x:     rand.Intn(p8.GetScreenWidth()),
			y:     rand.Intn(p8.GetScreenHeight()),
			color: 6 + rand.Intn(2),
			speed: 1 + rand.Intn(2),
		}
	}
}

func (g *Game) Update() {
	// Update stars with screen bounds
	for i := range g.stars {
		g.stars[i].y += g.stars[i].speed
		if g.stars[i].y > p8.GetScreenHeight() {
			g.stars[i].y = 0
		}
	}
	
	// Handle movement with custom bounds
	if p8.Btn(p8.RIGHT) && g.playerX < p8.GetScreenWidth()-4 {
		g.playerX++
	}
	// ... other directions
}

func (g *Game) Draw() {
	p8.Cls(1)
	// Draw stars and player...
	p8.Print("game boy style", 50, 5, 7)
	p8.Print("160 x 144", 60, 14, 7)
}

func main() {
	settings := p8.NewSettings()
	settings.ScreenWidth = 160
	settings.ScreenHeight = 144
	settings.WindowTitle = "Game Boy Style Demo"
	
	p8.InsertGame(&Game{})
	p8.PlayGameWith(settings)
}
```

