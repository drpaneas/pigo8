package pigo8

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

// Validate validates the sprite data for consistency and correctness
func (s *spriteData) Validate() error {
	// Validate dimensions
	if s.Width <= 0 || s.Height <= 0 {
		return fmt.Errorf("sprite %d has invalid dimensions: %dx%d (must be positive)", s.ID, s.Width, s.Height)
	}

	// Validate ID is non-negative
	if s.ID < 0 {
		return fmt.Errorf("sprite has invalid negative ID: %d", s.ID)
	}

	// Validate position is non-negative
	if s.X < 0 || s.Y < 0 {
		return fmt.Errorf("sprite %d has invalid negative position: (%d, %d)", s.ID, s.X, s.Y)
	}

	// Validate pixel data dimensions match declared dimensions
	if len(s.Pixels) != s.Height {
		return fmt.Errorf("sprite %d pixel data height mismatch: expected %d rows, got %d", s.ID, s.Height, len(s.Pixels))
	}

	for i, row := range s.Pixels {
		if len(row) != s.Width {
			return fmt.Errorf("sprite %d pixel data width mismatch at row %d: expected %d pixels, got %d", s.ID, i, s.Width, len(row))
		}

		// Validate color indices are within reasonable range (0-255 for extended palettes)
		for j, colorIndex := range row {
			if colorIndex < 0 || colorIndex > 255 {
				return fmt.Errorf("sprite %d has invalid color index %d at position (%d, %d): must be 0-255", s.ID, colorIndex, j, i)
			}
		}
	}

	// Validate flags bitfield is within reasonable range (8 bits = 0-255)
	if s.Flags.Bitfield < 0 || s.Flags.Bitfield > 255 {
		return fmt.Errorf("sprite %d has invalid flags bitfield %d: must be 0-255", s.ID, s.Flags.Bitfield)
	}

	// Validate individual flags array length
	if len(s.Flags.Individual) != 8 {
		return fmt.Errorf("sprite %d has invalid flags individual array length %d: must be 8", s.ID, len(s.Flags.Individual))
	}

	return nil
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
	// spritesheetColumns is the number of sprite columns in the sprite sheet
	// Default is 16 for standard PICO-8, 32 for custom palette
	spritesheetColumns = 16

	// spritesheetRows is the number of sprite rows in the sprite sheet
	// Default is 16 for standard PICO-8, 24 for custom palette
	spritesheetRows = 16

	// spritesheetWidth is the pixel width of the sprite sheet (columns * 8)
	spritesheetWidth = 128

	// spritesheetHeight is the pixel height of the sprite sheet (rows * 8)
	spritesheetHeight = 128

	// SpriteIDMapping maps original sprite IDs to loaded sprite indices for optimization
	// This allows empty and duplicate sprites to be mapped to existing sprites
	// Exported so it can be used by sprite lookup functions
	SpriteIDMapping map[int]int
)

// --- Target struct to hold processed sprite info ---

// spriteInfo holds the processed, ready-to-use sprite data.
// Exported for use in main.go.
type spriteInfo struct { // Exported
	ID    int
	Image *ebiten.Image
	Flags FlagsData
}

// --- Functions to load and process the spritesheet ---

// loadSpritesheetFromData processes sprite data provided as a byte slice.
// This allows users to load the spritesheet.json using go:embed or other methods
// in their own code (enabling build-time checks) and pass the data directly.
func loadSpritesheetFromData(data []byte) ([]spriteInfo, error) {
	return loadSpritesheetFromDataInternal(data, true)
}

// loadSpritesheetFromDataForTest is a test-specific version that skips pixel cache updates
func loadSpritesheetFromDataForTest(data []byte) ([]spriteInfo, error) {
	return loadSpritesheetFromDataInternal(data, false)
}

// validateAndUnmarshalSpritesheet validates and unmarshals the spritesheet data
func validateAndUnmarshalSpritesheet(data []byte) (*spriteSheet, error) {
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
	}

	return &sheet, nil
}

// updateSpritesheetDimensions updates global spritesheet dimensions from the sheet data
func updateSpritesheetDimensions(sheet *spriteSheet) {
	// Check for custom spritesheet dimensions in the JSON file
	if sheet.SpriteSheetColumns > 0 && sheet.SpriteSheetRows > 0 {
		// Update the global sprite sheet dimensions
		spritesheetColumns = sheet.SpriteSheetColumns
		spritesheetRows = sheet.SpriteSheetRows

		// If width and height are specified, use them directly
		if sheet.SpriteSheetWidth > 0 && sheet.SpriteSheetHeight > 0 {
			spritesheetWidth = sheet.SpriteSheetWidth
			spritesheetHeight = sheet.SpriteSheetHeight
		} else {
			// Otherwise calculate them from columns and rows (assuming 8x8 sprites)
			spritesheetWidth = spritesheetColumns * 8
			spritesheetHeight = spritesheetRows * 8
		}

		log.Printf("Custom spritesheet dimensions detected: %dx%d sprites (%dx%d pixels)",
			spritesheetColumns, spritesheetRows, spritesheetWidth, spritesheetHeight)
	}
}

// validateSpritePixelData validates that the first sprite has pixel data
func validateSpritePixelData(sheet *spriteSheet) {
	// Check if pixel data is present for the first sprite (if any)
	if len(sheet.Sprites) > 0 && len(sheet.Sprites[0].Pixels) == 0 {
		log.Printf(
			"Warning: First sprite has empty pixel data after unmarshalling. Check JSON tags, especially for 'pixels'.",
		)
	}
}

// isEmptySprite checks if a sprite is empty (no pixel data or all pixels are 0)
func isEmptySprite(spriteData spriteData) bool {
	// Check if pixel data is empty for this specific sprite
	if len(spriteData.Pixels) == 0 ||
		(len(spriteData.Pixels) > 0 && len(spriteData.Pixels[0]) == 0) {
		return true
	}

	// Check if sprite is empty (all pixels are 0)
	return isSpriteEmpty(spriteData)
}

// createSpriteInfo creates a spriteInfo from spriteData
func createSpriteInfo(spriteData spriteData, updatePixelCache bool) (spriteInfo, error) {
	// Create a new Ebiten image for the sprite
	img := ebiten.NewImage(spriteData.Width, spriteData.Height)

	// Create pixel buffer for batch operations
	pixels := make([]byte, spriteData.Width*spriteData.Height*4)

	// Iterate over pixels and set colors based on the palette
	for y, row := range spriteData.Pixels {
		for x, colorIndex := range row {
			// Use Pico8Palette (defined in screen.go, same package)
			if colorIndex >= 0 && colorIndex < len(pico8Palette) {
				// PICO-8 color 0 is often transparent
				if colorIndex != 0 {
					offset := (y*spriteData.Width + x) * 4
					r, g, b, a := pico8Palette[colorIndex].RGBA()
					pixels[offset] = uint8(r >> 8)   // Red
					pixels[offset+1] = uint8(g >> 8) // Green
					pixels[offset+2] = uint8(b >> 8) // Blue
					pixels[offset+3] = uint8(a >> 8) // Alpha
				}
			} else {
				log.Printf("Warning: Sprite %d has out-of-range color index %d at (%d, %d) - using transparent pixel", spriteData.ID, colorIndex, x, y)
			}
		}
	}

	// Upload all pixels to GPU in one operation
	img.WritePixels(pixels)

	// Create the SpriteInfo struct
	info := spriteInfo{
		ID:    spriteData.ID,
		Image: img,
		Flags: spriteData.Flags,
	}

	// Initialize sprite pixel cache for batch reading operations
	initSpritePixelCache(spriteData.ID, img)
	if updatePixelCache {
		updateSpritePixelCache(spriteData.ID, img)
	}

	return info, nil
}

// loadSpritesheetFromDataInternal is the internal implementation
func loadSpritesheetFromDataInternal(data []byte, updatePixelCache bool) ([]spriteInfo, error) {
	sheet, err := validateAndUnmarshalSpritesheet(data)
	if err != nil {
		return nil, err
	}

	// Return empty slice if no sprites found
	if len(sheet.Sprites) == 0 {
		return []spriteInfo{}, nil
	}

	updateSpritesheetDimensions(sheet)
	validateSpritePixelData(sheet)

	return processSprites(sheet, updatePixelCache)
}

// processSprites handles the main sprite processing logic
func processSprites(sheet *spriteSheet, updatePixelCache bool) ([]spriteInfo, error) {
	var loadedSprites []spriteInfo
	localSpriteIDMapping := make(map[int]int)
	spriteHashes := make(map[uint64]int) // hash -> first sprite index with this content (uint64 for performance)

	// Clear the global sprite hash table to ensure it's in sync with the local spriteHashes map.
	// This prevents bugs where duplicates are detected against stale entries from previous batches,
	// which would cause incorrect isDuplicate=false returns when the hash isn't in spriteHashes.
	spriteHashTable.Clear()
	uniqueLoaded := 0
	duplicatesMapped := 0
	emptiesMappedTo0 := 0

	// First pass: load unique sprites and build hash map
	for _, spriteData := range sheet.Sprites {
		if !spriteData.Used {
			continue // Skip unused sprites
		}

		// Validate sprite data with timing
		validationStart := time.Now()
		if err := spriteData.Validate(); err != nil {
			metricsCollector.RecordValidationTime(time.Since(validationStart))
			recordValidationError()

			if isStrictValidationEnabled() {
				metricsCollector.RecordStrictValidation()
				return nil, fmt.Errorf("strict validation failed for sprite %d: %w", spriteData.ID, err)
			}
			log.Printf("Warning: Skipping invalid sprite: %v", err)
			continue
		}
		metricsCollector.RecordValidationTime(time.Since(validationStart))

		// Handle empty sprites
		if isEmptySprite(spriteData) {
			localSpriteIDMapping[spriteData.ID] = 0
			emptiesMappedTo0++
			continue
		}

		// Handle duplicate detection for sprites without flags (only if optimization is enabled)
		if isOptimizationEnabled() && spriteData.Flags.Bitfield == 0 {
			existingIndex, isDuplicate, hash := checkForDuplicateWithCollisionDetection(spriteData, spriteHashes)
			if isDuplicate {
				localSpriteIDMapping[spriteData.ID] = existingIndex
				duplicatesMapped++
				recordSpriteSkipped()
				continue
			}
			// Store hash for future duplicate detection
			// IMPORTANT: Use the hash returned by checkForDuplicateWithCollisionDetection
			// (which may be collision-adjusted) to avoid overwriting entries on hash collisions
			spriteHashes[hash] = len(loadedSprites)
		}

		// Load unique sprite
		info, err := createSpriteInfo(spriteData, updatePixelCache)
		if err != nil {
			return nil, fmt.Errorf("failed to create sprite %d: %w", spriteData.ID, err)
		}

		loadedSprites = append(loadedSprites, info)
		localSpriteIDMapping[spriteData.ID] = len(loadedSprites) - 1
		uniqueLoaded++
		recordSpriteLoaded()
	}

	// Post-process sprites and finalize
	return finalizeSprites(loadedSprites, localSpriteIDMapping, uniqueLoaded, duplicatesMapped, emptiesMappedTo0, updatePixelCache, sheet)
}

// fillMissingIDs fills in missing sprite IDs up to sheet capacity
func fillMissingIDs(localSpriteIDMapping map[int]int) {
	if spritesheetRows > 0 && spritesheetColumns > 0 {
		maxSpriteID := spritesheetRows * spritesheetColumns
		for id := 0; id < maxSpriteID; id++ {
			if _, exists := localSpriteIDMapping[id]; !exists {
				// Map missing IDs to sprite 0 (transparent)
				localSpriteIDMapping[id] = 0
			}
		}
	}
}

// ensureTransparentSprite ensures sprite 0 exists as a transparent sprite
func ensureTransparentSprite(loadedSprites []spriteInfo, localSpriteIDMapping map[int]int, updatePixelCache bool) ([]spriteInfo, map[int]int, int) {
	uniqueLoaded := 0

	// Ensure sprite 0 exists as a transparent sprite, but only if we have sprites to work with
	// or if we're in production mode (updatePixelCache = true)
	if updatePixelCache && (len(loadedSprites) == 0 || (len(loadedSprites) > 0 && loadedSprites[0].ID != 0)) {
		// Create a transparent sprite 0
		transparentSprite := createTransparentSprite()
		// Insert at the beginning
		loadedSprites = append([]spriteInfo{transparentSprite}, loadedSprites...)
		// Update all mappings to account for the shift
		for id, index := range localSpriteIDMapping {
			localSpriteIDMapping[id] = index + 1
		}
		// Map sprite 0 to the new transparent sprite
		localSpriteIDMapping[0] = 0
		uniqueLoaded++
	} else if len(loadedSprites) > 0 && loadedSprites[0].ID != 0 {
		// In test mode, only create transparent sprite if we have other sprites but no sprite 0
		transparentSprite := createTransparentSprite()
		// Insert at the beginning
		loadedSprites = append([]spriteInfo{transparentSprite}, loadedSprites...)
		// Update all mappings to account for the shift
		for id, index := range localSpriteIDMapping {
			localSpriteIDMapping[id] = index + 1
		}
		// Map sprite 0 to the new transparent sprite
		localSpriteIDMapping[0] = 0
		uniqueLoaded++
	}

	return loadedSprites, localSpriteIDMapping, uniqueLoaded
}

// finalizeSprites completes sprite processing with post-processing steps
func finalizeSprites(loadedSprites []spriteInfo, localSpriteIDMapping map[int]int, uniqueLoaded, duplicatesMapped, emptiesMappedTo0 int, updatePixelCache bool, sheet *spriteSheet) ([]spriteInfo, error) {
	fillMissingIDs(localSpriteIDMapping)

	var additionalUnique int
	loadedSprites, localSpriteIDMapping, additionalUnique = ensureTransparentSprite(loadedSprites, localSpriteIDMapping, updatePixelCache)
	uniqueLoaded += additionalUnique

	// Update global sprite ID mapping
	SpriteIDMapping = localSpriteIDMapping

	// Log optimization statistics
	logOptimizationStats(uniqueLoaded, duplicatesMapped, emptiesMappedTo0)

	if len(loadedSprites) == 0 && len(sheet.Sprites) > 0 {
		log.Printf(
			"Warning: No 'used' sprites were processed. Check the 'used' field in your spritesheet data.",
		)
	}

	return loadedSprites, nil
}

// logOptimizationStats logs sprite optimization statistics
func logOptimizationStats(uniqueLoaded, duplicatesMapped, emptiesMappedTo0 int) {
	theoreticalTotal := spritesheetRows * spritesheetColumns
	if theoreticalTotal > 0 {
		reductionPercent := float64(theoreticalTotal-uniqueLoaded) / float64(theoreticalTotal) * 100
		log.Printf("Sprite optimization: %d unique loaded, %d duplicates mapped, %d empties mapped to 0 (%.1f%% reduction from theoretical %d)",
			uniqueLoaded, duplicatesMapped, emptiesMappedTo0, reductionPercent, theoreticalTotal)
	} else {
		log.Printf("Sprite optimization: %d unique loaded, %d duplicates mapped, %d empties mapped to 0",
			uniqueLoaded, duplicatesMapped, emptiesMappedTo0)
	}
}

// loadSpritesheet tries to load spritesheet.json from the current directory, then from common locations,
// then from custom embedded resources, and finally falls back to default embedded resources.
func loadSpritesheet() ([]spriteInfo, error) {
	return loadSpritesheetInternal(true)
}

// loadSpritesheetForTest is a test-specific version that skips pixel cache updates
func loadSpritesheetForTest() ([]spriteInfo, error) {
	return loadSpritesheetInternal(false)
}

// loadSpritesheetInternal is the internal implementation
func loadSpritesheetInternal(updatePixelCache bool) ([]spriteInfo, error) {
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
	sprites, err := loadSpritesheetFromDataInternal(data, updatePixelCache)
	if err != nil {
		return nil, fmt.Errorf("error processing spritesheet data: %w", err)
	}

	// Log when spritesheet is loaded
	fileSize := float64(len(data)) / 1024
	log.Printf("Spritesheet: %d sprites (%.1f KB)", len(sprites), fileSize)

	return sprites, nil
}

// isSpriteEmpty checks if a sprite has all pixels set to 0 (transparent)
func isSpriteEmpty(sprite spriteData) bool {
	for _, row := range sprite.Pixels {
		for _, pixel := range row {
			if pixel != 0 {
				return false
			}
		}
	}
	return true
}

// createTransparentSprite creates a transparent sprite with ID 0
func createTransparentSprite() spriteInfo {
	// Create an 8x8 transparent image
	img := ebiten.NewImage(8, 8)
	// No need to set pixels, they default to transparent

	return spriteInfo{
		ID:    0,
		Image: img,
		Flags: FlagsData{
			Bitfield:   0,
			Individual: make([]bool, 8),
		},
	}
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
