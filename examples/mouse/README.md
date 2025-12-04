# Mouse Input Demo

A pixel art drawing application demonstrating mouse input built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Mouse Demo in your browser](https://drpaneas.github.io/pigo8/mouse/)**

## 📖 Description

This example demonstrates mouse input handling in PIGO8:
- Getting mouse position with `p8.GetMouseXY()`
- Detecting mouse button clicks with `p8.Btn()` and `p8.Btnp()`
- Mouse wheel input for brush size control
- Drawing on a canvas with variable brush sizes
- Color palette selection

### Controls

- **Left Mouse Button**: Draw
- **Middle Mouse Button**: Erase
- **Right Mouse Button**: Select color from palette
- **Mouse Wheel**: Change brush size
- **X Key**: Clear canvas

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/mouse
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
	canvas    [][]int
	drawColor int
	brushSize int
}

func (g *Game) Init() {
	g.canvas = make([][]int, pigo8.GetScreenHeight())
	for i := range g.canvas {
		g.canvas[i] = make([]int, pigo8.GetScreenWidth())
	}
	g.drawColor = 7 // White
	g.brushSize = 1
}

func (g *Game) Update() {
	mouseX, mouseY := pigo8.GetMouseXY()
	
	// Brush size with mouse wheel
	if pigo8.Btn(pigo8.ButtonMouseWheelUp) {
		g.brushSize = min(g.brushSize+1, 10)
	}
	if pigo8.Btn(pigo8.ButtonMouseWheelDown) {
		g.brushSize = max(g.brushSize-1, 1)
	}
	
	// Drawing
	if pigo8.Btn(pigo8.ButtonMouseLeft) {
		g.drawCircle(mouseX, mouseY, g.brushSize, g.drawColor)
	}
	
	// Erasing
	if pigo8.Btn(pigo8.ButtonMouseMiddle) {
		g.drawCircle(mouseX, mouseY, g.brushSize, 0)
	}
	
	// Color selection
	if pigo8.Btnp(pigo8.ButtonMouseRight) {
		// Select color from palette...
	}
}

func (g *Game) Draw() {
	pigo8.Cls(0)
	// Draw canvas, palette, cursor...
	mouseX, mouseY := pigo8.GetMouseXY()
	pigo8.Circ(mouseX, mouseY, float64(g.brushSize), 7)
	pigo8.Print(fmt.Sprintf("mouse: %d,%d", mouseX, mouseY), 2, 118, 7)
}

func main() {
	pigo8.InsertGame(&Game{})
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO-8 Mouse Example"
	pigo8.PlayGameWith(settings)
}
```

