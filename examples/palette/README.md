# Dynamic Palette Demo

A demonstration of custom color palettes built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Palette Demo in your browser](https://drpaneas.github.io/pigo8/palette/)**

## 📖 Description

This example demonstrates how to use custom palettes in PIGO8:
- Creating custom color palettes programmatically with `p8.SetPalette()`
- Using a limited 4-color green palette for retro aesthetics
- Drawing with custom palette colors

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/palette
go run .
```

## 📝 Source Code

```go
package main

import (
	"image/color"
	"github.com/drpaneas/pigo8"
)

type Game struct {
	greenPalette []color.Color
}

func NewGame() *Game {
	return &Game{
		greenPalette: []color.Color{
			color.RGBA{38, 52, 35, 255},    // Dark green
			color.RGBA{83, 112, 47, 255},   // Medium green
			color.RGBA{166, 182, 61, 255},  // Light green
			color.RGBA{241, 243, 192, 255}, // Pale yellow
		},
	}
}

func (g *Game) Init() {
	pigo8.SetPalette(g.greenPalette)
}

func (g *Game) Update() {}

func (g *Game) Draw() {
	pigo8.Cls(3) // Pale yellow background
	pigo8.Print("custom palette demo", 25, 10, 0)
	
	centerX := pigo8.GetScreenWidth() / 2
	centerY := pigo8.GetScreenHeight() / 2
	
	// Draw concentric circles with custom colors
	pigo8.Circfill(centerX, centerY, 40, 0) // Dark green
	pigo8.Circfill(centerX, centerY, 30, 1) // Medium green
	pigo8.Circfill(centerX, centerY, 20, 2) // Light green
}

func main() {
	pigo8.InsertGame(NewGame())
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO8 Dynamic Palette Demo"
	pigo8.PlayGameWith(settings)
}
```

