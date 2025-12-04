# Transparency Effects Demo

A demonstration of advanced transparency and visual effects built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Effects Demo in your browser](https://drpaneas.github.io/pigo8/effects/)**

## 📖 Description

This example demonstrates advanced visual effects in PIGO8:
- Semi-transparent shadows under characters
- Ghost character with animated transparency
- Water overlay with wave patterns
- Fading title text effect
- Direct access to Ebiten for advanced rendering

### Controls

- **Arrow Keys**: Move the player
- **X Key**: Toggle help overlay

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/effects
go run .
```

## 📝 Source Code

```go
package main

import (
	"image/color"
	"math"
	"github.com/drpaneas/pigo8"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	tick      int
	fadeValue uint8
	ghostX    float64
	playerX   int
	playerY   int
}

func (g *Game) Init() {
	g.ghostX = 20
	g.playerX = 64
	g.playerY = 80
}

func (g *Game) Update() {
	g.tick++
	
	// Player movement
	if pigo8.Btn(pigo8.LEFT) && g.playerX > 10 {
		g.playerX--
	}
	// ... other directions
	
	// Animate ghost position
	g.ghostX += math.Sin(float64(g.tick)/20) * 0.5
	
	// Animate fade effect
	g.fadeValue = uint8(128 + 127*math.Sin(float64(g.tick)/30))
}

func (g *Game) Draw() {
	pigo8.Cls(1)
	
	// Get screen for direct Ebiten drawing
	screen := pigo8.CurrentScreen()
	
	// Draw player
	pigo8.Circfill(g.playerX, g.playerY, 5, 8)
	
	// Draw semi-transparent shadow
	shadowImg := ebiten.NewImage(12, 6)
	shadowImg.Fill(color.RGBA{0, 0, 0, 128})
	shadowOp := &ebiten.DrawImageOptions{}
	shadowOp.GeoM.Translate(float64(g.playerX-6), float64(g.playerY+6))
	shadowOp.Blend = ebiten.BlendSourceOver
	screen.DrawImage(shadowImg, shadowOp)
	
	// Draw ghost with animated transparency
	alpha := uint8(180 + 40*math.Sin(float64(g.tick)/15))
	// ... ghost drawing with transparency
	
	pigo8.Print("Transparency Effects Demo", 14, 5, 7)
}

func main() {
	pigo8.InsertGame(&Game{})
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO8 Transparency Effects Demo"
	pigo8.PlayGameWith(settings)
}
```

