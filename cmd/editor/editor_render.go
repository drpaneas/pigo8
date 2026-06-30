package main

import (
	"fmt"
	"image/color"
	"strconv"

	p8 "github.com/drpaneas/pigo8"
)

func (g *myGame) Draw() {
	p8.ClsRGBA(color.RGBA{25, 25, 25, 255})

	if g.mapMode {
		g.drawMapMode()
		return
	}

	g.updateDrawingCanvas()
	g.drawEditorCanvas()
	g.drawSpritesheetPanel()
	g.drawSelectionAndPalette()
}

// drawEditorCanvas draws the non‐map “editor” canvas
func (g *myGame) drawEditorCanvas() {
	const startX, startY = 10, 10
	endX := startX + 8*12 - 2
	endY := startY + 8*12 - 2

	gridPx := 8 * g.gridSize
	cell := max(1, 96/gridPx)
	// draw each cell
	for row := 0; row < gridPx; row++ {
		for col := 0; col < gridPx; col++ {
			color := getSquareColor(row, col)
			x := startX + col*cell
			y := startY + row*cell
			if cell > 1 {
				p8.Rectfill(x, y, x+cell-1, y+cell-1, color)
			} else {
				p8.Pset(x, y, color)
			}
		}
	}
	// hover text
	if g.hoverX >= 0 && g.hoverY >= 0 {
		p8.Print(
			fmt.Sprintf("pixel: (%d,%d)", g.hoverX, g.hoverY),
			startX, startY-10, g.getUIElementColor(),
		)
	}
	p8.Rect(startX-1, startY-1, endX+1, endY+1, g.getUIElementColor())
}

// drawSpritesheetPanel draws the spritesheet area and label
func (g *myGame) drawSpritesheetPanel() {
	sx, sy := spritesheetStartX, 10
	ex := sx + spriteSheetCols*spriteCellSize
	ey := sy + spriteSheetRows*spriteCellSize

	// draw each sprite tile
	for r := 0; r < spriteSheetRows; r++ {
		for c := 0; c < spriteSheetCols; c++ {
			baseX := sx + c*spriteCellSize
			baseY := sy + r*spriteCellSize
			for py := 0; py < 8; py++ {
				for px := 0; px < 8; px++ {
					col := spritesheet[r][c][py][px]
					if col == 0 {
						p8.Pset(baseX+px, baseY+py, 0)
					} else {
						p8.Pset(baseX+px, baseY+py, col)
					}
				}
			}
		}
	}

	// draw selection border
	if g.gridSize >= 1 {
		g.drawSelectionBorder(sx, sy)
	}
	p8.Rect(sx-1, sy-1, ex+1, ey+1, g.getUIElementColor())

	sizeText := map[int]string{1: "8x8", 2: "16x16", 4: "32x32"}[g.gridSize]
	p8.Print(
		fmt.Sprintf("spritesheet - sprite: %d - grid: %s",
			g.currentSprite, sizeText),
		sx, ey+4, g.getUIElementColor(),
	)
}

// drawSelectionBorder highlights the multi‐cell selection in the spritesheet
func (g *myGame) drawSelectionBorder(sx, sy int) {
	cols, rows := spriteSheetCols, spriteSheetRows
	base := g.currentSprite
	br := base / cols
	bc := base % cols

	x1 := sx + bc*spriteCellSize - 1
	y1 := sy + br*spriteCellSize - 1
	x2 := x1 + g.gridSize*spriteCellSize + 1
	y2 := y1 + g.gridSize*spriteCellSize + 1

	maxX := sx + cols*spriteCellSize
	maxY := sy + rows*spriteCellSize
	if x2 > maxX {
		x2 = maxX
	}
	if y2 > maxY {
		y2 = maxY
	}

	p8.Rect(x1, y1, x2, y2, g.getUIElementColor())
}

// drawSelectionAndPalette draws the selection and palette
func (g *myGame) drawSelectionAndPalette() {
	// spacing constants
	const offset = 15
	const paletteOffset = 40
	g.drawCheckboxes(10, 10+8*12-2+offset)
	g.drawPalette(10, 10+8*12-2+paletteOffset)
}

// drawCheckboxes draws the 8 checkboxes for sprite flags
func (g *myGame) drawCheckboxes(x, y int) {
	checkboxSize := 8 // Smaller checkboxes

	// Draw 8 checkboxes in a row
	for i := range 8 {
		checkboxX := x + i*checkboxSize*3/2 // Space them out a bit
		checkboxY := y

		// Draw checkbox outline
		p8.Rect(checkboxX, checkboxY, checkboxX+checkboxSize-1, checkboxY+checkboxSize-1, g.getUIElementColor())

		// Check the flag state across all selected sprites
		allTrue := true
		allFalse := true

		// Check flag state for all selected sprites
		g.forEachSelectedSprite(func(sprRow, sprCol int) {
			if spriteFlags[sprRow][sprCol][i] {
				allFalse = false // At least one is true
			} else {
				allTrue = false // At least one is false
			}
		})

		// Fill the checkbox based on state
		if allTrue {
			// All sprites have this flag set - fill with solid color
			p8.Rectfill(checkboxX+2, checkboxY+2, checkboxX+checkboxSize-3, checkboxY+checkboxSize-3, g.getUIElementColor())
		} else if !allFalse {
			// Mixed state - some sprites have this flag set, others don't - show a pattern
			p8.Rectfill(checkboxX+2, checkboxY+2, checkboxX+checkboxSize-3, checkboxY+checkboxSize-3, g.getUIElementColor()) // Use a different color
			// Draw a pattern to indicate mixed state
			p8.Line(checkboxX+2, checkboxY+2, checkboxX+checkboxSize-3, checkboxY+checkboxSize-3, g.getUIElementColor())
			p8.Line(checkboxX+checkboxSize-3, checkboxY+2, checkboxX+2, checkboxY+checkboxSize-3, g.getUIElementColor())
		}

		// Draw flag number
		p8.Print(strconv.Itoa(i), checkboxX+1, checkboxY+checkboxSize+2, g.getUIElementColor())
	}

	// Draw label
	p8.Print("flags", x, y-10, g.getUIElementColor())
}

// drawPalette draws the color palette below the grid
func (g *myGame) drawPalette(x, y int) {
	// Get the total palette size
	totalColors := p8.GetPaletteSize()

	// Always use 8 columns for the palette
	colorsPerRow := 8

	// Draw each color in the palette
	for i := range totalColors {
		// Calculate the row and column for this color
		row := i / colorsPerRow
		col := i % colorsPerRow

		// Calculate the position
		px := x + col*12
		py := y + row*12

		// Draw the color square
		p8.Rectfill(px, py, px+10, py+10, i)

		// Highlight the currently selected color with a white border
		if i == g.currentColor {
			p8.Rect(px-1, py-1, px+11, py+11, g.getUIElementColor()) // White highlight border
		}
	}
}

// updateDrawingCanvas updates the drawing canvas to show the selected sprites based on current grid size
func (g *myGame) updateDrawingCanvas() {
	// Clear the drawing canvas
	for row := range 64 {
		for col := range 64 {
			squareColors[row][col] = 0
		}
	}

	// Copy the sprites to the drawing canvas
	g.forEachSelectedSprite(func(srcRow, srcCol int) {
		// Calculate relative position from base sprite
		baseRow := g.currentSprite / spriteSheetCols
		baseCol := g.currentSprite % spriteSheetCols
		r := srcRow - baseRow
		c := srcCol - baseCol

		// Copy the sprite to the drawing canvas
		for pixelRow := 0; pixelRow < 8; pixelRow++ {
			for pixelCol := 0; pixelCol < 8; pixelCol++ {
				// Calculate destination position in drawing canvas
				dstRow := r*8 + pixelRow
				dstCol := c*8 + pixelCol

				// Copy the pixel
				squareColors[dstRow][dstCol] = spritesheet[srcRow][srcCol][pixelRow][pixelCol]
			}
		}
	})
}

func (g *myGame) drawAt(row, col, colorIndex int) {
	base := g.currentSprite
	r := base/spriteSheetCols + row/8
	c := base%spriteSheetCols + col/8
	pr, pc := row%8, col%8

	if spritesheet[r][c][pr][pc] != colorIndex {
		setSquareColor(row, col, colorIndex)
		spritesheet[r][c][pr][pc] = colorIndex
		// ... (mutate all state you want to track for undo)
		g.saveCurrentStateIfNeeded()
		p8.Sset(c*8+pc, r*8+pr, colorIndex)
		updateMapSprites(r*spriteSheetCols + c)
		g.updateDrawingCanvas()
	}
}

// getUIElementColor returns the appropriate color for UI elements based on the active palette.
// It returns 7 (white) if the default PICO-8 palette is active, otherwise defaultColor (1).
func (g *myGame) getUIElementColor() int {
	if p8.IsDefaultPico8PaletteActive() {
		return 7 // PICO-8 white
	}
	return defaultColor // Defined as 1 (dark-blue)
}
