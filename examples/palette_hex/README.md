# Palette from File Demo

A demonstration of loading custom palettes from hex files built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Palette Hex Demo in your browser](https://drpaneas.github.io/pigo8/palette_hex/)**

## 📖 Description

This example demonstrates loading custom palettes from files in PIGO8:
- Loading palettes from `palette.hex` files
- Using `p8.GetPaletteSize()` to work with variable palette sizes
- Using `p8.GetPaletteColor()` to inspect color values
- Displaying a color grid with palette information

### Controls

- **Z (O button)**: Cycle through colors

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/palette_hex
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .

import (
	"fmt"
	"github.com/drpaneas/pigo8"
)

type Game struct {
	currentColor int
}

func (g *Game) Init() {}

func (g *Game) Update() {
	if pigo8.Btn(pigo8.O) {
		g.currentColor = (g.currentColor + 1) % pigo8.GetPaletteSize()
	}
}

func (g *Game) Draw() {
	pigo8.Cls(0)
	
	paletteSize := pigo8.GetPaletteSize()
	pigo8.Print("load palette from file", 20, 4, 7)
	pigo8.Print(fmt.Sprintf("total colors: %d", paletteSize), 20, 12, 7)
	
	// Draw color grid
	cellSize := 8
	cols := 8
	for i := range paletteSize {
		x := 16 + (i%cols)*cellSize
		y := 32 + (i/cols)*cellSize
		pigo8.Rectfill(x, y, x+cellSize-1, y+cellSize-1, i)
		
		if i == g.currentColor {
			pigo8.Rect(x-1, y-1, x+cellSize, y+cellSize, 7)
		}
	}
	
	// Display current color info
	clr := pigo8.GetPaletteColor(g.currentColor)
	r, gb, b, a := clr.RGBA()
	pigo8.Print(fmt.Sprintf("color %d: rgba(%d,%d,%d,%d)", 
		g.currentColor, r>>8, gb>>8, b>>8, a>>8), 20, 100, 7)
}

func main() {
	pigo8.InsertGame(&Game{})
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO-8 Custom Palette Example"
	pigo8.PlayGameWith(settings)
}
```

