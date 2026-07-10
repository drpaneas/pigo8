package main

import (
	"encoding/json"
	"fmt"
	"os"

	p8 "github.com/drpaneas/pigo8"
)

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

// spritesheetFileReadable reports whether spritesheet.json already exists
// and can be read. Reading it (rather than os.Stat) is deliberate: on
// GOOS=js (WASM, e.g. running in a browser), os.Stat returns a generic
// "not implemented" error rather than a proper ErrNotExist, so
// os.IsNotExist can't distinguish "missing file" from "real error" there -
// the editor would otherwise treat every startup in a browser as a fatal
// error before the game loop even starts. Any read failure (missing,
// unreadable, or platform quirk) is treated the same way: fall back to
// creating a fresh default spritesheet, matching the existing resilience
// pattern already used by loadSpritesheet/initSpritesheet elsewhere in
// this file.
func spritesheetFileReadable() bool {
	_, err := os.ReadFile("spritesheet.json")
	return err == nil
}

func initPico8Spritesheet() error {
	if spritesheetFileReadable() {
		return nil
	}

	// Create a temporary spritesheet.json file that PIGO8 can load. This
	// already pushes the full sprite grid into PIGO8's engine directly
	// (see createTempSpritesheet), so no further per-pixel Sset() sync is
	// needed here.
	createTempSpritesheet()

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

	// Push the full grid into PIGO8's engine directly from the in-memory
	// JSON, rather than relying on Sset()'s lazy-load to read it back from
	// disk later. That round trip silently fails in environments with no
	// real filesystem (e.g. WASM in a browser): the write below appears to
	// fail non-fatally, but the file then doesn't exist to read back
	// either, so the engine would fall back to a minimal built-in default
	// with no mapping for any sprite ID beyond it - every subsequent
	// Sset() call from drawing (and Spr() call from placing sprites on the
	// map) would silently no-op with a "non-existent sprite ID" warning.
	if err := p8.LoadSpritesheetFromBytes(jsonData); err != nil {
		fmt.Println("Error loading temporary spritesheet into PIGO8:", err)
	}

	// Best-effort: also persist to disk so native runs get a real
	// spritesheet.json file. Failure here (e.g. no writable filesystem) is
	// non-fatal - the engine already has the data via LoadSpritesheetFromBytes
	// above regardless of whether this succeeds.
	if err := os.WriteFile("spritesheet.json", jsonData, 0644); err != nil {
		fmt.Println("Error writing temporary spritesheet file:", err)
	}
}

func updateMapSprites(spriteIndex int) {
	// Scan through the loaded world map (from p8.MapSize(), not the fixed
	// editorMapWidth x editorMapHeight editor buffer) and update any
	// instances of this sprite. The editor buffer is intentionally larger
	// than any real project's map so the editor can grow into it; scanning
	// past the actual world bounds makes every Mset() below get rejected
	// (and logged) as out-of-bounds for no benefit, since those cells will
	// never be part of the loaded map.
	worldWidth, worldHeight := p8.MapSize()
	if worldWidth <= 0 || worldHeight <= 0 {
		worldWidth, worldHeight = editorMapWidth, editorMapHeight
	}
	worldWidth = min(worldWidth, editorMapWidth)
	worldHeight = min(worldHeight, editorMapHeight)

	for y := 0; y < worldHeight; y++ {
		for x := 0; x < worldWidth; x++ {
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
