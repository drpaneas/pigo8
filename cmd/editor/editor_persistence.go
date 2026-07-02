package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	p8 "github.com/drpaneas/pigo8"
	"github.com/spf13/afero"
)

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

// buildSpritesheetJSON constructs the in-memory JSON representation of the
// current spritesheet (spritesheet[]/spriteFlags[], matching PIGO8's file
// format). Used both to persist the spritesheet to disk and to push it
// directly into PIGO8's engine via p8.LoadSpritesheetFromBytes.
func buildSpritesheetJSON() ([]byte, error) {
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

	return json.MarshalIndent(sheet, "", "  ")
}

// saveSpritesheet pushes the current spritesheet into PIGO8's engine
// directly from in-memory JSON, then best-effort persists the same bytes to
// spritesheet.json on disk.
//
// The engine sync happens directly from bytes (via p8.LoadSpritesheetFromBytes)
// rather than by writing to disk and letting PIGO8 read it back later -
// that round trip silently fails wherever there's no real filesystem (e.g.
// WASM in a browser): the write below can fail non-fatally, and a later
// read then finds nothing, leaving the engine's copy stale. The disk write
// is still attempted, for native runs where callers expect a real
// spritesheet.json file to exist afterward.
func saveSpritesheet() error {
	jsonData, err := buildSpritesheetJSON()
	if err != nil {
		return fmt.Errorf("error marshaling spritesheet: %w", err)
	}

	if err := p8.LoadSpritesheetFromBytes(jsonData); err != nil {
		log.Printf("Error loading spritesheet into PIGO8: %v", err)
	}

	if err := os.WriteFile("spritesheet.json", jsonData, 0644); err != nil {
		return fmt.Errorf("error writing spritesheet.json: %w", err)
	}
	return nil
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
