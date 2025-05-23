package pigo8

import (
	"encoding/json" // Keep color import
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Structs to match spritesheet.json ---

// FlagsData holds sprite flag information.
// Exported because it's part of the exported SpriteInfo struct.
type FlagsData struct { // Exported
	Bitfield   int    `json:"bitfield"`
	Individual []bool `json:"individual"`
}

// spriteData holds the raw data for a single sprite from JSON.
// Kept internal as it's only used during loading.
type spriteData struct { // Internal
	ID     int       `json:"id"`
	X      int       `json:"x"`
	Y      int       `json:"y"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Pixels [][]int   `json:"pixels"`
	Flags  FlagsData `json:"flags"` // Uses exported FlagsData
	Used   bool      `json:"used"`
}

// spriteSheet holds the overall structure of the JSON file.
// Kept internal.
type spriteSheet struct { // Internal
	// Custom spritesheet dimensions (optional)
	SpriteSheetColumns int          `json:"SpriteSheetColumns,omitempty"`
	SpriteSheetRows    int          `json:"SpriteSheetRows,omitempty"`
	SpriteSheetWidth   int          `json:"SpriteSheetWidth,omitempty"`
	SpriteSheetHeight  int          `json:"SpriteSheetHeight,omitempty"`
	Sprites            []spriteData `json:"sprites"`
}

// --- Sprite sheet dimensions ---

// Default sprite sheet dimensions (16x16 sprites)
var (
	// SpritesheetColumns is the number of sprite columns in the sprite sheet
	// Default is 16 for standard PICO-8, 32 for custom palette
	SpritesheetColumns = 16

	// SpritesheetRows is the number of sprite rows in the sprite sheet
	// Default is 16 for standard PICO-8, 24 for custom palette
	SpritesheetRows = 16

	// SpritesheetWidth is the pixel width of the sprite sheet (columns * 8)
	SpritesheetWidth = 128

	// SpritesheetHeight is the pixel height of the sprite sheet (rows * 8)
	SpritesheetHeight = 128
)

// --- Target struct to hold processed sprite info ---

// SpriteInfo holds the processed, ready-to-use sprite data.
// Exported for use in main.go.
type SpriteInfo struct { // Exported
	ID    int
	Image *ebiten.Image
	Flags FlagsData
}

// --- Functions to load and process the spritesheet ---

// loadSpritesheetFromData processes sprite data provided as a byte slice.
// This allows users to load the spritesheet.json using go:embed or other methods
// in their own code (enabling build-time checks) and pass the data directly.
func loadSpritesheetFromData(data []byte) ([]SpriteInfo, error) {
	// Basic check if data is empty
	if len(data) == 0 {
		return nil, fmt.Errorf("provided spritesheet data is empty")
	}

	// Unmarshal the JSON data
	var sheet spriteSheet
	err := json.Unmarshal(data, &sheet)
	if err != nil {
		// Return a clear error about unmarshalling
		return nil, fmt.Errorf("error unmarshalling provided spritesheet data: %w", err)
	}

	// Add a check to see if sprites were loaded
	if len(sheet.Sprites) == 0 {
		// Log warning here as it's about content, not file loading
		log.Printf(
			"Warning: No sprites found after unmarshalling spritesheet data. Check JSON format and tags.",
		)
		// Return empty slice, not necessarily an error
		return []SpriteInfo{}, nil
	}

	// Check for custom spritesheet dimensions in the JSON file
	if sheet.SpriteSheetColumns > 0 && sheet.SpriteSheetRows > 0 {
		// Update the global sprite sheet dimensions
		SpritesheetColumns = sheet.SpriteSheetColumns
		SpritesheetRows = sheet.SpriteSheetRows

		// If width and height are specified, use them directly
		if sheet.SpriteSheetWidth > 0 && sheet.SpriteSheetHeight > 0 {
			SpritesheetWidth = sheet.SpriteSheetWidth
			SpritesheetHeight = sheet.SpriteSheetHeight
		} else {
			// Otherwise calculate them from columns and rows (assuming 8x8 sprites)
			SpritesheetWidth = SpritesheetColumns * 8
			SpritesheetHeight = SpritesheetRows * 8
		}

		log.Printf("Custom spritesheet dimensions detected: %dx%d sprites (%dx%d pixels)",
			SpritesheetColumns, SpritesheetRows, SpritesheetWidth, SpritesheetHeight)
	}
	// Check if pixel data is present for the first sprite (if any)
	if len(sheet.Sprites) > 0 && len(sheet.Sprites[0].Pixels) == 0 {
		log.Printf(
			"Warning: First sprite has empty pixel data after unmarshalling. Check JSON tags, especially for 'pixels'.",
		)
	}

	// Process used sprites
	var loadedSprites []SpriteInfo
	for _, spriteData := range sheet.Sprites {
		if !spriteData.Used {
			continue // Skip unused sprites
		}

		// Check if pixel data is empty for this specific sprite
		if len(spriteData.Pixels) == 0 ||
			(len(spriteData.Pixels) > 0 && len(spriteData.Pixels[0]) == 0) {
			log.Printf(
				"Warning: Skipping sprite %d due to missing or empty pixel data.",
				spriteData.ID,
			)
			continue
		}

		// Create a new Ebiten image for the sprite
		img := ebiten.NewImage(spriteData.Width, spriteData.Height)

		// Iterate over pixels and set colors based on the palette
		for y, row := range spriteData.Pixels {
			for x, colorIndex := range row {
				// Use Pico8Palette (defined in screen.go, same package)
				if colorIndex >= 0 && colorIndex < len(Pico8Palette) {
					// PICO-8 color 0 is often transparent
					if colorIndex != 0 {
						img.Set(x, y, Pico8Palette[colorIndex])
					}
				} else {
					log.Printf("Warning: Sprite %d has out-of-range color index %d at (%d, %d)", spriteData.ID, colorIndex, x, y)
				}
			}
		}

		// Create the SpriteInfo struct
		info := SpriteInfo{
			ID:    spriteData.ID,
			Image: img,
			Flags: spriteData.Flags,
		}
		loadedSprites = append(loadedSprites, info)
	}

	if len(loadedSprites) == 0 &&
		len(sheet.Sprites) > 0 { // Only warn if sprites existed but none were 'used'
		log.Printf(
			"Warning: No 'used' sprites were processed. Check the 'used' field in your spritesheet data.",
		)
	}

	return loadedSprites, nil
}

// loadSpritesheet tries to load spritesheet.json from the current directory, then from common locations,
// then from custom embedded resources, and finally falls back to default embedded resources.
func loadSpritesheet() ([]SpriteInfo, error) {
	const spritesheetFilename = "spritesheet.json"

	// First try to load from the file system
	data, err := os.ReadFile(spritesheetFilename)
	if err != nil {
		// Check common alternative locations
		commonLocations := []string{
			filepath.Join("assets", spritesheetFilename),
			filepath.Join("resources", spritesheetFilename),
			filepath.Join("data", spritesheetFilename),
			filepath.Join("static", spritesheetFilename),
		}

		for _, location := range commonLocations {
			data, err = os.ReadFile(location)
			if err == nil {
				log.Printf("Loaded spritesheet from %s", location)
				break
			}
		}

		// If still not found, try embedded resources
		if err != nil {
			log.Printf("Spritesheet file not found in common locations, trying embedded resources")
			embeddedData, embErr := tryLoadEmbeddedSpritesheet()
			if embErr != nil {
				return nil, fmt.Errorf("failed to load embedded spritesheet: %w", embErr)
			}
			data = embeddedData
		}
	} else {
		log.Printf("Using spritesheet file from current directory: %s", spritesheetFilename)
	}

	// Log memory after reading file
	logMemory("after reading spritesheet file", false)

	// Process the spritesheet data
	sprites, err := loadSpritesheetFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error processing spritesheet data: %w", err)
	}

	// Log when spritesheet is loaded
	fileSize := float64(len(data)) / 1024
	log.Printf("Spritesheet: %d sprites (%.1f KB)", len(sprites), fileSize)

	return sprites, nil
}

// LoadSpritesheet loads sprite data from a specific JSON file and updates the
// engine's active spritesheet (currentSprites).
// This function is intended to be called by user code (e.g., an editor) to reload
// the spritesheet at runtime.
func LoadSpritesheet(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading spritesheet file %s: %w", filename, err)
	}

	newSprites, err := loadSpritesheetFromData(data)
	if err != nil {
		return fmt.Errorf("error processing spritesheet data from %s: %w", filename, err)
	}

	// Update the package-level currentSprites variable (defined in engine.go)
	currentSprites = newSprites
	log.Printf("Successfully loaded and updated spritesheet from %s. %d sprites processed.", filename, len(currentSprites))
	return nil
}
