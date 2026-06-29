// Package main basic sprite editor
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	p8 "github.com/drpaneas/pigo8"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/spf13/afero"
)

const (
	// Sprite dimensions
	spriteSize = 8 // Size of each sprite in pixels

	// Grid sizes
	defaultGridSize = 1 // 8x8 grid (1 sprite)
	mediumGridSize  = 2 // 16x16 grid (4 sprites)
	largeGridSize   = 4 // 32x32 grid (16 sprites)

	// Editor map dimensions
	editorMapWidth  = 320
	editorMapHeight = 320

	// Default colors
	defaultColor     = 1 // Default color
	transparentColor = 0 // Transparent color

	// UI constants
	paletteColumns = 8 // Number of columns in the palette display
	numFlags       = 8 // Number of sprite flags

)

type myGame struct {
	currentColor  int
	currentSprite int
	hoverX        int       // X coordinate of the pixel being hovered over (-1 if none)
	hoverY        int       // Y coordinate of the pixel being hovered over (-1 if none)
	gridSize      int       // Size of the working grid (1=8x8, 2=16x16, 4=32x32, 8=64x64)
	lastWheelTime int64     // Last time the mouse wheel was scrolled or keyboard was used (for debouncing)
	mapMode       bool      // Whether we are in map mode
	copiedSprite  [8][8]int // Buffer for copied sprite data

	// Undo/Redo state
	undoStack      []string      // Stack of saved state filenames for undo
	redoStack      []string      // Stack of saved state filenames for redo
	fs             afero.Fs      // Virtual filesystem for state snapshots
	stateMutex     sync.Mutex    // Mutex for thread-safe access to state
	lastSaveTime   time.Time     // Last time a state was saved
	saveCooldown   time.Duration // Minimum time between saves
	undoInProgress bool          // Flag to prevent re-entrant undo/redo operations

	// Key state tracking
	lastUndoTime int64 // Last time undo was triggered
	lastRedoTime int64 // Last time redo was triggered
	keyCooldown  int64 // Minimum time between undo/redo actions in milliseconds

	// Map editor state
	mapViewport   mapViewport
	mapLastMouseX int
	mapLastMouseY int
	mapData       [editorMapHeight][editorMapWidth]int
}

type mapData struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Name        string    `json:"name"`
	Cells       []mapCell `json:"cells"`
}

type mapCell struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Sprite int `json:"sprite"`
}

// forEachSpritePixel iterates over every pixel in every sprite in the spritesheet
// and calls the provided function with the current coordinates
func forEachSpritePixel(fn func(row, col, r, c int)) {
	for row := 0; row < spriteSheetRows; row++ {
		for col := 0; col < spriteSheetCols; col++ {
			for r := 0; r < 8; r++ {
				for c := 0; c < 8; c++ {
					fn(row, col, r, c)
				}
			}
		}
	}
}

// forEachSelectedSprite iterates over each sprite in the current selection grid
// and calls the provided function with the sprite's row and column coordinates
func (g *myGame) forEachSelectedSprite(fn func(row, col int)) {
	baseRow := g.currentSprite / spriteSheetCols
	baseCol := g.currentSprite % spriteSheetCols
	size := g.gridSize
	if size < 1 {
		size = 1
	}
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			sprRow := baseRow + r
			sprCol := baseCol + c
			// Make sure we don't go out of bounds
			if sprRow >= 0 && sprRow < spriteSheetRows && sprCol >= 0 && sprCol < spriteSheetCols {
				fn(sprRow, sprCol)
			}
		}
	}
}

func (g *myGame) Init() {
	initSquareColors()

	// Initialize virtual filesystem and undo/redo stacks
	g.fs = afero.NewMemMapFs()
	g.undoStack = make([]string, 0)
	g.redoStack = make([]string, 0)
	g.saveCooldown = 100 * time.Millisecond

	// Initialize sprite flags to false
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			for flag := range numFlags {
				spriteFlags[row][col][flag] = false
			}
		}
	}

	// Initialize spritesheet (will also load from file if available)
	initSpritesheet()

	// Try to load map data from map.json
	hasFirstTryWorked := false
	if err := g.loadMapData(); err != nil {
		fmt.Println("No map.json found, starting with empty map")
		// Everytime you get out of map mode, save the map
		err := g.saveMapData()
		if err != nil {
			fmt.Println("Error saving map:", err)
			os.Exit(1)
		}
		fmt.Println("Map saved to map.json")
	} else {
		hasFirstTryWorked = true
	}

	if !hasFirstTryWorked {
		if err := g.loadMapData(); err != nil {
			fmt.Println("Could not create map.json")
			os.Exit(1)
		}
	}

	g.currentColor = defaultColor // Default color (usually red in PICO-8 palette)
	g.currentSprite = 1           // Default to first non-transparent sprite (sprite 0 is reserved)
	g.hoverX = -1                 // No hover initially
	g.hoverY = -1                 // No hover initially
	g.gridSize = defaultGridSize  // Start with 8x8 grid (1 sprite)
	g.mapViewport = defaultMapViewport()
	g.mapViewport.clamp(len(g.mapData[0]), len(g.mapData), g.mapScreenLayout().Canvas)

	// Ensure grid size is never less than 1
	if g.gridSize < defaultGridSize {
		g.gridSize = defaultGridSize
	}
	g.lastWheelTime = 0 // Initialize wheel time

	// Initialize the drawing canvas with the default sprite (1)
	g.updateDrawingCanvas()

	// Save the initial state to the undo stack
	if err := g.saveState(); err != nil {
		log.Printf("Failed to save initial state: %v", err)
	}
}

// Define the sprite structure to match PIGO8's format
type spriteData struct {
	ID     int          `json:"id"`
	X      int          `json:"x"`
	Y      int          `json:"y"`
	Width  int          `json:"width"`
	Height int          `json:"height"`
	Used   bool         `json:"used"`
	Flags  p8.FlagsData `json:"flags"`
	Pixels [][]int      `json:"pixels"`
}

// Define the spritesheet structure
type spriteSheetData struct {
	// Custom spritesheet dimensions
	SpriteSheetColumns int          `json:"SpriteSheetColumns"`
	SpriteSheetRows    int          `json:"SpriteSheetRows"`
	SpriteSheetWidth   int          `json:"SpriteSheetWidth"`
	SpriteSheetHeight  int          `json:"SpriteSheetHeight"`
	Sprites            []spriteData `json:"sprites"`
}

// isSpriteEmpty checks if a sprite at the given row and column is completely empty (all pixels are 0)
func isSpriteEmpty(row, col int) bool {
	for r := 0; r < spriteSize; r++ {
		for c := 0; c < spriteSize; c++ {
			if spritesheet[row][col][r][c] != 0 {
				return false
			}
		}
	}
	return true
}

// convertSpriteToData converts a sprite at the given row and column to PIGO8's spriteData format
func convertSpriteToData(row, col int) spriteData {
	// Calculate sprite index
	spriteIndex := row*spriteSheetCols + col

	// Check if sprite is empty
	isEmpty := isSpriteEmpty(row, col)

	// Create a new sprite
	sprite := spriteData{
		ID:     spriteIndex,
		X:      col * spriteSize,
		Y:      row * spriteSize,
		Width:  spriteSize,
		Height: spriteSize,
		Used:   !isEmpty, // Only mark non-empty sprites as used
		Flags: p8.FlagsData{
			Bitfield:   0, // Will be calculated below
			Individual: make([]bool, 8),
		},
		Pixels: make([][]int, 8),
	}

	// Fill in the flags
	bitfield := 0
	for i := 0; i < numFlags; i++ {
		flagValue := spriteFlags[row][col][i]
		sprite.Flags.Individual[i] = flagValue
		if flagValue {
			bitfield |= 1 << i
		}
	}
	sprite.Flags.Bitfield = bitfield

	// Initialize pixel data
	for r := 0; r < spriteSize; r++ {
		sprite.Pixels[r] = make([]int, spriteSize)
		for c := 0; c < spriteSize; c++ {
			sprite.Pixels[r][c] = spritesheet[row][col][r][c]
		}
	}

	return sprite
}

// applySpriteData applies PIGO8's spriteData format to the spritesheet at the given position
func applySpriteData(sprite spriteData) {
	// Calculate row and column from sprite ID
	row := sprite.ID / spriteSheetCols
	col := sprite.ID % spriteSheetCols

	// Make sure the sprite is within bounds
	if row >= 0 && row < spriteSheetRows && col >= 0 && col < spriteSheetCols {
		// Load pixel data
		for r := range 8 {
			for c := range 8 {
				// Make sure we have pixel data for this position
				if r < len(sprite.Pixels) && c < len(sprite.Pixels[r]) {
					spritesheet[row][col][r][c] = sprite.Pixels[r][c]
				}
			}
		}

		// Load flag data
		for i := range 8 {
			if i < len(sprite.Flags.Individual) {
				// Get the flag value from the individual array
				flagValue := sprite.Flags.Individual[i]
				// Store it in the spriteFlags array
				spriteFlags[row][col][i] = flagValue
			}
		}
	}
}

// loadSpritesheet loads the spritesheet from spritesheet.json if it exists
func loadSpritesheet() error {
	// Check if spritesheet.json exists
	data, err := os.ReadFile("spritesheet.json")
	if err != nil {
		// File doesn't exist or can't be read, just use the default empty spritesheet
		fmt.Println("No spritesheet.json found, using empty spritesheet")
		return err
	}

	// Parse the JSON data
	var sheet spriteSheetData
	err = json.Unmarshal(data, &sheet)
	if err != nil {
		return fmt.Errorf("error parsing spritesheet.json: %w", err)
	}

	// Load the sprites into the spritesheet
	for _, sprite := range sheet.Sprites {
		applySpriteData(sprite)
	}

	fmt.Println("Loaded spritesheet from spritesheet.json")
	return nil
}

// saveSpritesheet saves the current spritesheet to a JSON file
func saveSpritesheet() error {
	// Create the spritesheet structure following the PIGO8 format
	sheet := spriteSheetData{
		SpriteSheetColumns: spriteSheetCols,
		SpriteSheetRows:    spriteSheetRows,
		SpriteSheetWidth:   spriteSheetCols * spriteSize, // Each sprite is 8x8 pixels
		SpriteSheetHeight:  spriteSheetRows * spriteSize, // Each sprite is 8x8 pixels
		Sprites:            make([]spriteData, 0, spriteSheetRows*spriteSheetCols),
	}

	// Convert only non-empty sprites
	savedCount := 0
	skippedCount := 0
	for row := 0; row < spriteSheetRows; row++ {
		for col := 0; col < spriteSheetCols; col++ {
			if !isSpriteEmpty(row, col) {
				sheet.Sprites = append(sheet.Sprites, convertSpriteToData(row, col))
				savedCount++
			} else {
				skippedCount++
			}
		}
	}

	fmt.Printf("Spritesheet saved: %d sprites saved, %d empty sprites skipped\n", savedCount, skippedCount)

	return saveJSONToFile("spritesheet.json", sheet)
}

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

var (
	width           = 48  // Increased to accommodate the larger spritesheet and more space
	height          = 27  // Increased to accommodate the taller spritesheet
	mapViewWidth    = 128 // Default map viewport width in pixels (16 sprites)
	mapViewHeight   = 128 // Default map viewport height in pixels (16 sprites)
	unit            = 8
	spriteCellSize  = 8  // Size of each sprite cell in the spritesheet
	spriteSheetCols = 32 // Number of columns in the spritesheet
	spriteSheetRows = 24 // Number of rows in the spritesheet

	// Position of the spritesheet grid
	spritesheetStartX = 120 // Position spritesheet (adjusted 10px to the left)
)

var squareColors [64][64]int      // Up to 64x64 grid to store square colors
var spritesheet [24][32][8][8]int // 24x32 grid of 8x8 sprites
var spriteFlags [24][32][8]bool   // Flags for each sprite [row][col][flag0-7]

func initSquareColors() {
	for row := range 64 {
		for col := range 64 {
			squareColors[row][col] = 0 // Default color
		}
	}
}

func initSpritesheet() {
	// Try to load the spritesheet from a file
	if err := loadSpritesheet(); err != nil {
		// Only initialize with transparent colors if loading failed
		forEachSpritePixel(func(row, col, r, c int) {
			// Initialize with transparent color (0)
			spritesheet[row][col][r][c] = 0
		})
	}

	// Initialize the spritesheet in PIGO8 with our current data
	forEachSpritePixel(func(row, col, r, c int) {
		// Calculate the absolute pixel position
		px := col*8 + c
		py := row*8 + r
		// Set the pixel color in PIGO8
		p8.Sset(px, py, spritesheet[row][col][r][c])
	})
}

func initPico8Spritesheet() error {
	// Check if spritesheet.json already exists
	if _, err := os.Stat("spritesheet.json"); err == nil {
		// File exists, no need to create it
		return nil
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("error checking spritesheet.json: %w", err)
	}

	// Create a temporary spritesheet.json file that PIGO8 can load
	createTempSpritesheet()

	// Now initialize all sprites with our data
	forEachSpritePixel(func(row, col, r, c int) {
		// Calculate the absolute pixel position
		px := col*8 + c
		py := row*8 + r
		// Set the pixel color in PIGO8
		p8.Sset(px, py, spritesheet[row][col][r][c])
	})

	return nil
}

// createTempSpritesheet creates a temporary spritesheet.json file
// that PIGO8 can load to initialize its sprite system
func createTempSpritesheet() {
	// Create a basic spritesheet structure
	var sheet spriteSheetData

	// Set dimensions
	sheet.SpriteSheetColumns = spriteSheetCols
	sheet.SpriteSheetRows = spriteSheetRows
	sheet.SpriteSheetWidth = spriteSheetCols * 8
	sheet.SpriteSheetHeight = spriteSheetRows * 8

	// Create sprites array
	sprites := make([]spriteData, spriteSheetRows*spriteSheetCols)

	// Create all sprites first
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			spriteIndex := row*spriteSheetCols + col

			// Create a sprite with basic data
			sprite := spriteData{
				ID:     spriteIndex,
				X:      col * 8,
				Y:      row * 8,
				Width:  8,
				Height: 8,
				Used:   true,
				Flags: p8.FlagsData{
					Bitfield:   0,
					Individual: make([]bool, 8),
				},
				Pixels: make([][]int, 8),
			}

			// Initialize pixel arrays
			for r := range 8 {
				sprite.Pixels[r] = make([]int, 8)
			}

			// Add to sprites array
			sprites[spriteIndex] = sprite
		}
	}

	// Fill in pixel data
	forEachSpritePixel(func(row, col, r, c int) {
		spriteIndex := row*spriteSheetCols + col
		sprites[spriteIndex].Pixels[r][c] = spritesheet[row][col][r][c]
	})

	// Set sprites in sheet
	sheet.Sprites = sprites

	// Convert to JSON
	jsonData, err := json.MarshalIndent(sheet, "", "  ")
	if err != nil {
		fmt.Println("Error creating temporary spritesheet JSON:", err)
		return
	}

	// Write to file
	err = os.WriteFile("spritesheet.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing temporary spritesheet file:", err)
		return
	}

	// The spritesheet.json file will be loaded automatically the next time
	// a sprite-related function like Spr() or Sset() is called
}

func updateMapSprites(spriteIndex int) {
	// Scan through the entire map and update any instances of this sprite
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			// Check if this map cell uses the modified sprite
			if p8.Mget(x, y) == spriteIndex {
				// Force a redraw of this sprite by setting it to itself
				p8.Mset(x, y, spriteIndex)
			}
		}
	}
}

func getSquareColor(row, col int) int {
	if row < 0 || row >= 64 || col < 0 || col >= 64 {
		return -1
	}
	return squareColors[row][col]
}

func setSquareColor(row, col, color int) {
	// Ensure the coordinates are within the grid
	if row < 0 || row >= 64 || col < 0 || col >= 64 {
		return
	}
	// Update the color in the squareColors array
	squareColors[row][col] = color
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

// saveJSONToFile saves any data structure to a JSON file with proper indentation
func saveJSONToFile(filename string, data any) error {
	// Convert to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing %s: %w", filename, err)
	}

	return nil
}

// loadJSONFromFile loads a JSON file into the provided data structure
func loadJSONFromFile(filename string, data any) error {
	// Read the file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	// Parse the JSON data
	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("error parsing %s: %w", filename, err)
	}

	return nil
}

// convertMapToData converts the game's map data to the PIGO8 MapData format
func (g *myGame) convertMapToData() mapData {
	mapData := mapData{
		Version:     "1.0",
		Description: "Map created with PIGO8 editor",
		Width:       editorMapWidth,
		Height:      editorMapHeight,
		Name:        "map",
		Cells:       []mapCell{},
	}

	// Convert our map data to PIGO8's format
	for y := range g.mapData {
		for x := range g.mapData[y] {
			sprite := g.mapData[y][x]
			// Only save non-zero sprites to keep the file size smaller
			if sprite != 0 {
				mapData.Cells = append(mapData.Cells, mapCell{
					X:      x,
					Y:      y,
					Sprite: sprite,
				})
			}
		}
	}

	return mapData
}

// applyMapData applies the PIGO8 MapData format to the game's map
func (g *myGame) applyMapData(mapData mapData) {
	// Initialize map with zeros
	for y := range g.mapData {
		for x := range g.mapData[y] {
			g.mapData[y][x] = 0
		}
	}

	// Load the cells into our map data
	for _, cell := range mapData.Cells {
		// Make sure coordinates are within bounds
		if cell.X >= 0 && cell.X < editorMapWidth && cell.Y >= 0 && cell.Y < editorMapHeight {
			g.mapData[cell.Y][cell.X] = cell.Sprite
			// Also update the PIGO8 map
			p8.Mset(cell.X, cell.Y, cell.Sprite)
		}
	}

	fmt.Printf("Loaded map: %dx%d tiles. View: %dx%d pixels (%dx%d tiles). %d cells\n", mapData.Width, mapData.Height, mapViewWidth, mapViewHeight, mapViewWidth/unit, mapViewHeight/unit, len(mapData.Cells))
}

// saveMapData saves the current map to map.json
func (g *myGame) saveMapData() error {
	mapData := g.convertMapToData()
	return saveJSONToFile("map.json", mapData)
}

// loadMapData loads the map from map.json if it exists
func (g *myGame) loadMapData() error {
	var mapData mapData
	if err := loadJSONFromFile("map.json", &mapData); err != nil {
		return err
	}

	g.applyMapData(mapData)
	return nil
}

// Refactored Update method with reduced cyclomatic complexity and integrated save-on-toggle logic
func (g *myGame) Update() {
	g.toggleMapMode()
	g.handleUndoRedo() // Handle undo/redo in both modes
	if g.mapMode {
		g.handleMapMode()
	} else {
		g.handleEditorMode()
	}
}

// toggleMapMode flips mapMode and saves data on entry/exit
func (g *myGame) toggleMapMode() {
	if p8.Btnp(p8.X) {
		// Save current state before switching modes to ensure all changes are captured
		if err := g.saveState(); err != nil {
			log.Printf("Error saving state before mode switch: %v", err)
		}

		g.mapMode = !g.mapMode
		if g.mapMode {
			// When entering map mode, save the current spritesheet first
			if err := saveSpritesheet(); err != nil {
				log.Printf("Error saving spritesheet before entering map mode: %v", err)
				// Decide if this is a fatal error or if we can proceed
				// For now, we'll log and continue, but PIGO-8 might not have the latest sprites
			}
			// Then, instruct PIGO-8 to reload this spritesheet
			if err := p8.LoadSpritesheet("spritesheet.json"); err != nil {
				log.Printf("Error loading spritesheet into PIGO-8: %v", err)
				// Similar to above, log and continue for now
			}
		} else {
			if err := g.saveMapData(); err != nil {
				fmt.Println("Error saving map:", err)
				os.Exit(1)
			}
		}
	}
}

func (g *myGame) eraseAt(x, y int) {
	if g.inBounds(x, y) {
		// Only save if this cell is non-zero (actual change)
		if g.mapData[y][x] != 0 {
			log.Printf("Erasing at (%d,%d) - was %d", x, y, g.mapData[y][x])
			p8.Mset(x, y, 0)
			g.mapData[y][x] = 0
			g.saveCurrentStateIfNeeded()
			log.Printf("After erase - map[%d][%d] = %d, p8.Mget = %d",
				y, x, g.mapData[y][x], p8.Mget(x, y))
		}
	}
}

func (g *myGame) inBounds(x, y int) bool {
	mapWidth, mapHeight := g.mapDataBounds()
	return x >= 0 && x < mapWidth && y >= 0 && y < mapHeight
}

func (g *myGame) placeGridSprites(x, y int) {
	w, h := g.gridSize, g.gridSize
	if w < 1 {
		w, h = 1, 1
	}
	base := g.currentSprite
	changed := false

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			tx, ty := x+dx, y+dy
			if !g.inBounds(tx, ty) {
				log.Printf("  Out of bounds: (%d,%d)", tx, ty)
				continue
			}
			row := base/spriteSheetCols + dy
			col := base%spriteSheetCols + dx
			if row < spriteSheetRows && col < spriteSheetCols {
				idx := row*spriteSheetCols + col
				// Only mark as changed if we're actually changing the value
				if g.mapData[ty][tx] != idx {
					p8.Mset(tx, ty, idx)
					g.mapData[ty][tx] = idx
					changed = true
				}
			} else {
				log.Printf("  Invalid sprite position: row=%d, col=%d (max %d,%d)",
					row, col, spriteSheetRows-1, spriteSheetCols-1)
			}
		}
	}

	// Only save state if something was actually changed
	if changed {
		g.saveCurrentStateIfNeeded()
	}
}

// saveState saves the current state to a temporary file and updates the undo stack
func (g *myGame) saveState() error {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	// Don't save too frequently
	if time.Since(g.lastSaveTime) < g.saveCooldown {
		log.Println("Skipping saveState: too soon since last save")
		return nil
	}

	// If we have a redo stack, clear it when making new changes after undo
	if len(g.redoStack) > 0 {
		// Clean up old redo states
		for _, filename := range g.redoStack {
			if err := g.fs.Remove(filename); err != nil && !os.IsNotExist(err) {
				log.Printf("Error removing redo state file %s: %v", filename, err)
			}
		}
		g.redoStack = g.redoStack[:0]
	}

	// Create a unique filename
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("state_%d.json", timestamp)

	// Create state data
	state := struct {
		Spritesheet   [24][32][8][8]int
		SpriteFlags   [24][32][8]bool
		MapData       [editorMapHeight][editorMapWidth]int
		CurrentSprite int
		CurrentColor  int
	}{
		Spritesheet:   spritesheet,
		SpriteFlags:   spriteFlags,
		MapData:       g.mapData,
		CurrentSprite: g.currentSprite,
		CurrentColor:  g.currentColor,
	}

	// Marshal to JSON
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}

	// Save to virtual filesystem
	if err := afero.WriteFile(g.fs, filename, data, 0644); err != nil {
		return fmt.Errorf("error saving state: %w", err)
	}

	// Update undo stack and clear redo stack
	g.undoStack = append(g.undoStack, filename)
	g.redoStack = g.redoStack[:0] // Clear redo stack

	// Limit undo stack size (keep last 50 states)
	if len(g.undoStack) > 50 {
		// Remove oldest state file
		oldest := g.undoStack[0]
		if err := g.fs.Remove(oldest); err != nil && !os.IsNotExist(err) {
			log.Printf("Error removing old state file %s: %v", oldest, err)
		}
		g.undoStack = g.undoStack[1:]
	}

	g.lastSaveTime = time.Now()
	return nil
}

// syncMapDataToPigo8 updates PICO-8's internal map memory to match g.mapData
// using the editor's full map dimensions.
func (g *myGame) syncMapDataToPigo8() {
	nonZeroTiles := 0

	for y := 0; y < editorMapHeight; y++ {
		for x := 0; x < editorMapWidth; x++ {
			spriteID := g.mapData[y][x]
			if spriteID < 0 {
				spriteID = 0
			}
			p8.Mset(x, y, spriteID)
			if spriteID != 0 {
				nonZeroTiles++
			}
		}
	}

	log.Printf("Synced full map to PIGO8 using Mset. Non-zero tiles: %d", nonZeroTiles)
}

// loadState loads a state from the virtual filesystem
func (g *myGame) loadState(filename string) error {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	// Read from virtual filesystem
	data, err := afero.ReadFile(g.fs, filename)
	if err != nil {
		return fmt.Errorf("error reading state: %w", err)
	}

	// Unmarshal state
	var state struct {
		Spritesheet   [24][32][8][8]int
		SpriteFlags   [24][32][8]bool
		MapData       [editorMapHeight][editorMapWidth]int
		CurrentSprite int
		CurrentColor  int
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("error unmarshaling state: %w", err)
	}

	// Apply state
	spritesheet = state.Spritesheet
	spriteFlags = state.SpriteFlags
	g.mapData = state.MapData
	g.currentSprite = state.CurrentSprite
	g.currentColor = state.CurrentColor

	// Update the display
	g.updateDrawingCanvas()
	g.syncMapDataToPigo8() // Sync map data to PICO-8's internal map memory
	updateMapSprites(-1)   // Update all sprites

	return nil
}

// undo reverts to the previous state
func (g *myGame) undo() {

	if len(g.undoStack) < 2 {
		log.Println("Not enough states to undo")
		return // Need at least 2 states to undo (current + previous)
	}

	// Pop the current state and move to redo stack
	current := g.undoStack[len(g.undoStack)-1]
	g.redoStack = append(g.redoStack, current)
	g.undoStack = g.undoStack[:len(g.undoStack)-1]

	// Load previous state
	if len(g.undoStack) > 0 {
		prevState := g.undoStack[len(g.undoStack)-1]
		if err := g.loadState(prevState); err != nil {
			log.Printf("Error undoing: %v", err)
		}
	}
}

// redo re-applies the next state
func (g *myGame) redo() {
	log.Printf("Redo called. Stack sizes - undo: %d, redo: %d", len(g.undoStack), len(g.redoStack))

	if len(g.redoStack) == 0 {
		log.Println("Nothing to redo")
		return
	}

	// Pop from redo stack and push to undo stack
	next := g.redoStack[len(g.redoStack)-1]
	g.redoStack = g.redoStack[:len(g.redoStack)-1]
	g.undoStack = append(g.undoStack, next)

	log.Printf("Redoing state from %s", next)
	// Load the state
	if err := g.loadState(next); err != nil {
		log.Printf("Error redoing: %v", err)
	} else {
		log.Printf("Redo successful. New stack sizes - undo: %d, redo: %d",
			len(g.undoStack), len(g.redoStack))
	}
}

// saveCurrentStateIfNeeded saves the current state if enough time has passed
func (g *myGame) saveCurrentStateIfNeeded() {
	if time.Since(g.lastSaveTime) >= g.saveCooldown {
		if err := g.saveState(); err != nil {
			log.Printf("Error saving state: %v", err)
		}
	}
}

// canTriggerAction checks if enough time has passed since the last action
func (g *myGame) canTriggerAction(lastActionTime *int64) bool {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-*lastActionTime < 200 { // 200ms cooldown
		return false
	}
	*lastActionTime = now
	return true
}

// handleUndoRedo manages the undo/redo key presses with proper debouncing
func (g *myGame) handleUndoRedo() {
	if g.undoInProgress { // Prevent re-entrant calls
		return
	}

	// Initialize key cooldown if not set
	if g.keyCooldown == 0 {
		g.keyCooldown = 200 // 200ms cooldown by default
	}

	// Check for Cmd+Z (Undo) or Cmd+Shift+Z (Redo)
	if ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				// Redo with Cmd+Shift+Z
				if g.canTriggerAction(&g.lastRedoTime) {
					g.undoInProgress = true
					defer func() { g.undoInProgress = false }()
					g.redo()
				}
			} else {
				// Undo with Cmd+Z
				if g.canTriggerAction(&g.lastUndoTime) {
					g.undoInProgress = true
					defer func() { g.undoInProgress = false }()
					g.undo()
				}
			}
		}
	}
}

// -------------------- Editor Mode --------------------
func (g *myGame) handleEditorMode() {
	mx, my := p8.GetMouseXY()
	g.toggleSpriteFlags(mx, my)
	g.handleDrawingGrid(mx, my)
	g.handleSpriteSelection(mx, my)
	g.handlePaletteSelection(mx, my)
	g.handleWheel()
	g.handleKeyboardNavigation()
	g.handleCopyPaste()

	// Handle undo/redo with proper debouncing
	g.handleUndoRedo()
}

func (g *myGame) toggleSpriteFlags(mx, my int) {
	// Use the same coordinates and size as in drawCheckboxes
	const checkboxSize = 8
	const offset = 15 // Match the offset from drawSelectionAndPalette

	// Get the same base coordinates used in drawCheckboxes
	baseX := 10
	baseY := 10 + 8*12 - 2 + offset // Add the offset to match drawCheckboxes position

	for i := 0; i < 8; i++ {
		checkboxX := baseX + i*checkboxSize*3/2
		checkboxY := baseY
		if mx >= checkboxX && mx < checkboxX+checkboxSize && my >= checkboxY && my < checkboxY+checkboxSize && p8.Btnp(p8.ButtonMouseLeft) {
			g.toggleFlagAtIndex(i)
		}
	}
}

func (g *myGame) toggleFlagAtIndex(i int) {
	g.saveCurrentStateIfNeeded()

	base := g.currentSprite
	r, c := base/spriteSheetCols, base%spriteSheetCols
	cur := spriteFlags[r][c][i]
	for dr := range g.safeGridSize() {
		for dc := range g.safeGridSize() {
			rr, cc := r+dr, c+dc
			if rr < spriteSheetRows && cc < spriteSheetCols {
				spriteFlags[rr][cc][i] = !cur
			}
		}
	}
	if err := saveSpritesheet(); err != nil {
		log.Printf("Error saving spritesheet after toggling flag: %v", err)
	}
}

func (g *myGame) safeGridSize() int {
	if g.gridSize < 1 {
		return 1
	}
	return g.gridSize
}

func (g *myGame) handleDrawingGrid(mx, my int) {
	const gx, gy, size = 10, 10, 8
	gridPx := size * g.gridSize
	cell := max(1, 96/gridPx)
	row := (my - gy) / cell
	col := (mx - gx) / cell

	if row < 0 || row >= gridPx || col < 0 || col >= gridPx {
		g.hoverX, g.hoverY = -1, -1
		return
	}
	g.updateHover(row, col)
	if p8.Btn(p8.ButtonMouseLeft) {
		g.drawAt(row, col, g.currentColor)
	} else if p8.Btn(p8.ButtonMouseRight) {
		g.drawAt(row, col, 0)
	}
}

func (g *myGame) updateHover(row, col int) {
	base := g.currentSprite
	r := base/spriteSheetCols + row/8
	c := base%spriteSheetCols + col/8
	pr, pc := row%8, col%8
	g.hoverX, g.hoverY = c*8+pc, r*8+pr
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

func (g *myGame) handleSpriteSelection(mx, my int) {
	row := (my - 10) / spriteCellSize
	col := (mx - spritesheetStartX) / spriteCellSize
	if row >= 0 && row < spriteSheetRows && col >= 0 && col < spriteSheetCols && p8.Btnp(p8.ButtonMouseLeft) {
		idx := row*spriteSheetCols + col
		if idx > 0 {
			g.currentSprite = idx
		}
		g.updateDrawingCanvas()
	}
}

func (g *myGame) handlePaletteSelection(mx, my int) {
	const gx, gy = 10, 10 + 8*12 - 2 + 40
	row := (my - gy) / 12
	col := (mx - gx) / 12
	colors := p8.GetPaletteSize()
	if row >= 0 && col >= 0 && row*8+col < colors && p8.Btnp(p8.ButtonMouseLeft) {
		g.currentColor = row*8 + col
	}
}

func (g *myGame) handleWheel() {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-g.lastWheelTime <= 150 { // 150ms debounce for wheel and keyboard
		return
	}
	if p8.Btnp(p8.ButtonMouseWheelUp) && g.gridSize < 4 {
		g.gridSize = min(g.gridSize*2, 8)
		g.lastWheelTime = now
		g.updateDrawingCanvas()
	} else if p8.Btnp(p8.ButtonMouseWheelDown) && g.gridSize > 1 {
		g.gridSize = max(1, g.gridSize/2)
		g.lastWheelTime = now
		g.updateDrawingCanvas()
	}
}

// handleKeyboardNavigation handles keyboard arrow key navigation between sprites
// copySprite copies the current sprite data to the clipboard
func (g *myGame) copySprite() {
	r, c := g.currentSprite/spriteSheetCols, g.currentSprite%spriteSheetCols
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			g.copiedSprite[y][x] = spritesheet[r][c][y][x]
		}
	}
}

// pasteSprite pastes the copied sprite data to the current sprite
func (g *myGame) pasteSprite() {
	r, c := g.currentSprite/spriteSheetCols, g.currentSprite%spriteSheetCols
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			spritesheet[r][c][y][x] = g.copiedSprite[y][x]
		}
	}
	updateMapSprites(g.currentSprite)
	g.updateDrawingCanvas()
}

// handleCopyPaste handles keyboard shortcuts for copy and paste
func (g *myGame) handleCopyPaste() {
	// Check for CMD+C (Copy)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) && (ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl)) {
		g.copySprite()
	}

	// Check for CMD+V (Paste)
	if inpututil.IsKeyJustPressed(ebiten.KeyV) && (ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl)) {
		g.saveCurrentStateIfNeeded()
		g.pasteSprite()
	}
}

func (g *myGame) handleKeyboardNavigation() {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-g.lastWheelTime <= 150 { // 150ms debounce for keyboard navigation
		return
	}

	currentRow := g.currentSprite / spriteSheetCols
	currentCol := g.currentSprite % spriteSheetCols
	moved := false

	switch {
	case p8.Btnp(p8.LEFT) && currentCol > 0:
		g.currentSprite--
		moved = true
	case p8.Btnp(p8.RIGHT) && currentCol < spriteSheetCols-1:
		g.currentSprite++
		moved = true
	case p8.Btnp(p8.UP) && currentRow > 0:
		g.currentSprite -= spriteSheetCols
		moved = true
	case p8.Btnp(p8.DOWN) && currentRow < spriteSheetRows-1:
		g.currentSprite += spriteSheetCols
		moved = true
	}

	if moved {
		g.lastWheelTime = now
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

func main() {
	// Store initial default map view dimensions (pixels) from global vars
	initialDefaultMapViewWidthPx := mapViewWidth
	initialDefaultMapViewHeightPx := mapViewHeight

	// Calculate UI overhead in tiles based on initial global defaults for editor and map view
	// global 'width' and 'height' are editor total tiles (e.g., 48, 27)
	// global 'unit' is pixels per tile (e.g., 8)
	uiWidthOverheadInTiles := width - (initialDefaultMapViewWidthPx / unit)
	uiHeightOverheadInTiles := height - (initialDefaultMapViewHeightPx / unit)

	// Parse command line flags
	// Default values for flags are the initialDefaultMapViewWidth/Height
	widthFlag := flag.Int("w", initialDefaultMapViewWidthPx, "map viewport width in pixels")
	heightFlag := flag.Int("h", initialDefaultMapViewHeightPx, "map viewport height in pixels")
	flag.Parse()

	// Update global map viewport dimensions (pixels) from flags
	mapViewWidth = *widthFlag
	mapViewHeight = *heightFlag

	// Ensure new map viewport dimensions are multiples of 'unit' (sprite size)
	mapViewWidth = (mapViewWidth / unit) * unit
	mapViewHeight = (mapViewHeight / unit) * unit

	// Recalculate global editor window dimensions (tiles: global 'width', 'height')
	// based on the new map view dimensions (now in global mapViewWidth/Height) and the UI overhead
	width = (mapViewWidth / unit) + uiWidthOverheadInTiles
	height = (mapViewHeight / unit) + uiHeightOverheadInTiles

	// Initialize spritesheet if it doesn't exist
	if err := initPico8Spritesheet(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing spritesheet: %v\n", err)
		os.Exit(1)
	}

	if width > 256 || height > 256 {
		log.Printf("Editor window size is too large: %dx%d tiles.\n", width, height)
		os.Exit(1)
	}

	settings := p8.NewSettings()
	// These use the recalculated global 'width' and 'height'
	settings.ScreenWidth = width * unit
	settings.ScreenHeight = height * unit
	settings.ScaleFactor = 3
	p8.InsertGame(&myGame{})
	p8.PlayGameWith(settings)
}
