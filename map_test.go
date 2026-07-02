package pigo8

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

func TestMgetAndMset(t *testing.T) {
	// Initialize the streaming system without loading spritesheet (which requires game context)
	streamingInitMutex.Lock()
	defer streamingInitMutex.Unlock()

	// Skip spritesheet loading for tests to avoid ReadPixels issue
	if currentSprites == nil {
		currentSprites = []spriteInfo{} // Empty sprites for testing
	}

	// Initialize the map system directly
	if err := initializeStreamingMapSystem(); err != nil {
		t.Fatalf("Failed to initialize streaming map system: %v", err)
	}

	streamingSystemInitialized = true

	// Create a byte slice for the entire default map dimensions
	testMapData := make([]byte, defaultPico8MapWidth*defaultPico8MapHeight)

	// Populate specific sprite data for testing
	// Simulating the old testMap structure within the dense map
	testMapData[0*defaultPico8MapWidth+0] = 1  // {X: 0, Y: 0, Sprite: 1}
	testMapData[1*defaultPico8MapWidth+1] = 2  // {X: 1, Y: 1, Sprite: 2}
	testMapData[2*defaultPico8MapWidth+2] = 3  // {X: 2, Y: 2, Sprite: 3}
	testMapData[7*defaultPico8MapWidth+5] = 42 // {X: 5, Y: 7, Sprite: 42} (Y is row, X is col)

	// Set this data as the current map
	SetMap(testMapData)

	// Test Mget with different coordinate types
	t.Run("Mget with different coordinate types", func(t *testing.T) {
		// Integer coordinates
		assert.Equal(t, 1, Mget(0, 0), "Mget(0, 0) should return sprite 1")
		assert.Equal(t, 2, Mget(1, 1), "Mget(1, 1) should return sprite 2")
		assert.Equal(t, 3, Mget(2, 2), "Mget(2, 2) should return sprite 3")
		assert.Equal(t, 42, Mget(5, 7), "Mget(5, 7) should return sprite 42")

		// Float coordinates (should be converted to int)
		assert.Equal(t, 1, Mget(0.9, 0.9), "Mget(0.9, 0.9) should return sprite 1")
		assert.Equal(t, 2, Mget(1.5, 1.5), "Mget(1.5, 1.5) should return sprite 2")

		// Empty cell (any cell not explicitly set should be 0)
		assert.Equal(t, 0, Mget(10, 10), "Mget for empty cell should return 0")
	})

	// Test Mget (some redundant, but good to keep for clarity from original test)
	t.Run("Mget", func(t *testing.T) {
		assert.Equal(t, 1, Mget(0, 0), "Mget(0, 0) should return sprite 1")
		assert.Equal(t, 42, Mget(5, 7), "Mget(5, 7) should return sprite 42")
	})

	// Test Mset
	t.Run("Mset basic functionality", func(t *testing.T) {
		// Set a new sprite at an empty location
		Mset(10, 10, 99)
		assert.Equal(t, 99, Mget(10, 10), "After Mset(10, 10, 99), Mget should return 99")

		// Update an existing cell
		Mset(0, 0, 55)
		assert.Equal(t, 55, Mget(0, 0), "After Mset(0, 0, 55), Mget should return 55")

		// Test with float coordinates and sprite number
		Mset(11.7, 11.7, 77.9)
		assert.Equal(t, 77, Mget(11, 11), "After Mset(11.7, 11.7, 77.9), Mget(11, 11) should return 77")
	})

	// Test Mset alias (Mset is not an alias anymore, but the test logic is valid for Mset)
	t.Run("Mset direct test", func(t *testing.T) { // Renamed from Mset alias
		Mset(12, 12, 88)
		assert.Equal(t, 88, Mget(12, 12), "After Mset(12, 12, 88), Mget should return 88")
	})

	// The "Map cell creation and update" sub-test is removed as it relied on the sparse
	// nature of the old map (currentMap.Cells). With a dense map, Mset always updates
	// an existing cell's value, and the concept of "adding" a cell is not applicable.
}

// TestMgetWithMapFile tests Mget with a real map file
func TestMgetWithMapFile(t *testing.T) {
	// Skip this test if we're not in an environment with a map.json file
	// This check is important because the test manipulates map.json
	if _, err := os.Stat("map.json"); os.IsNotExist(err) {
		t.Skip("Skipping test: map.json not found. This test requires a map.json to exist (it will be temporarily overwritten).")
	}

	// Define local structs that match the map.json format for marshalling
	type tempJSONMapCell struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Sprite int `json:"sprite"`
	}
	type tempJSONMapData struct {
		Version     string            `json:"version"`
		Description string            `json:"description"`
		Width       int               `json:"width"`
		Height      int               `json:"height"`
		Name        string            `json:"name"`
		Cells       []tempJSONMapCell `json:"cells"`
	}

	// Create data for a temporary map.json
	tempMapFileContent := tempJSONMapData{
		Version:     "1.0",
		Description: "Temp Test Map for TestMgetWithMapFile",
		Width:       16, // Using a small, specific size for this test map file
		Height:      16,
		Name:        "temp_map_from_file_test",
		Cells: []tempJSONMapCell{
			{X: 3, Y: 4, Sprite: 123},
			{X: 5, Y: 6, Sprite: 456},
		},
	}

	// Marshal the temporary map data to JSON
	tempMapJSON, err := json.MarshalIndent(tempMapFileContent, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal temp map data: %v", err)
	}

	// Save the original map.json if it exists
	var originalMapJSON []byte
	var originalMapExists bool
	if _, err := os.Stat("map.json"); err == nil {
		originalMapExists = true
		originalMapJSON, err = os.ReadFile("map.json")
		if err != nil {
			t.Fatalf("Failed to read original map.json: %v", err)
		}
	}

	// Write the temporary map.json
	err = os.WriteFile("map.json", tempMapJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write temp map.json: %v", err)
	}

	// Restore the original map.json (or remove temp) after the test
	t.Cleanup(func() {
		if originalMapExists {
			err := os.WriteFile("map.json", originalMapJSON, 0644)
			if err != nil {
				t.Logf("Warning: Failed to restore original map.json: %v", err)
			}
		} else {
			if err := os.Remove("map.json"); err != nil {
				t.Logf("Warning: Failed to remove temporary map.json created by test: %v", err)
			}
		}
		// Also, reset the streaming system state after this test so other tests are not affected
		streamingInitMutex.Lock()
		streamingSystemInitialized = false
		worldMapStream = nil
		activeTileBufferInstance = nil
		mapCacheIsValid = false
		streamingInitMutex.Unlock()
	})

	// CRITICAL: Force the streaming system to re-initialize to load the new map.json
	streamingInitMutex.Lock()
	streamingSystemInitialized = false
	worldMapStream = nil           // Clear any map data from previous tests/SetMap
	activeTileBufferInstance = nil // Clear active buffer
	mapCacheIsValid = false        // Invalidate draw cache
	streamingInitMutex.Unlock()
	// EnsureStreamingSystemInitialized() will be called by Mget/Mset if not already initialized

	// Test Mget with the file-based map
	// Note: The map loaded from map.json will have dimensions 16x16 as per tempMapFileContent.
	// Mget/Mset will operate within these world dimensions.
	assert.Equal(t, 123, Mget(3, 4), "Mget(3, 4) should return sprite 123 from file")
	assert.Equal(t, 456, Mget(5, 6), "Mget(5, 6) should return sprite 456 from file")
	assert.Equal(t, 0, Mget(9, 9), "Mget(9, 9) should return 0 for empty cell in file-loaded map")

	// Test Mset with the file-based map
	Mset(9, 9, 789)
	assert.Equal(t, 789, Mget(9, 9), "After Mset(9, 9, 789), Mget should return 789 in file-loaded map")

	// Test Mget for an out-of-bounds coordinate according to the 16x16 map loaded from the file.
	// The streaming system's worldMapStream might be larger (DefaultPico8MapWidth x Height) if initializeStreamingMapSystem
	// decided to create a default one first, but the *loaded data* from map.json is 16x16.
	// Mget should respect the dimensions of the loaded map data if map.json dictated smaller dimensions.
	// However, our current initializeStreamingMapSystem prioritizes map.json dimensions for worldMapStream.
	assert.Equal(t, 0, Mget(20, 20), "Mget(20, 20) should be 0 as it's outside the 16x16 map loaded from file")
}

// resetMapCacheState snapshots and restores the map-cache-related package
// state so tests can freely mutate it without leaking into other tests.
func resetMapCacheState(t *testing.T) {
	t.Helper()
	origEnabled := mapCacheEnabled
	origValid := mapCacheIsValid
	origImage := mapCacheImage
	origX := mapCacheDrawnForWorldTileX
	origY := mapCacheDrawnForWorldTileY
	origW := mapCacheWidthInTiles
	origH := mapCacheHeightInTiles
	origLayers := mapCacheRenderedLayers
	origScreenW := mapCacheRenderedScreenWidth
	origScreenH := mapCacheRenderedScreenHeight
	t.Cleanup(func() {
		mapCacheEnabled = origEnabled
		mapCacheIsValid = origValid
		mapCacheImage = origImage
		mapCacheDrawnForWorldTileX = origX
		mapCacheDrawnForWorldTileY = origY
		mapCacheWidthInTiles = origW
		mapCacheHeightInTiles = origH
		mapCacheRenderedLayers = origLayers
		mapCacheRenderedScreenWidth = origScreenW
		mapCacheRenderedScreenHeight = origScreenH
	})
}

func TestSetMapCacheEnabled(t *testing.T) {
	resetMapCacheState(t)

	SetMapCacheEnabled(false)
	assert.False(t, mapCacheEnabled)

	SetMapCacheEnabled(true)
	assert.True(t, mapCacheEnabled)
}

func TestIsMapCacheValid(t *testing.T) {
	resetMapCacheState(t)

	// A cache that exactly matches every parameter is only considered valid
	// when caching is enabled and every dimension/position/layer matches.
	setValidCache := func() {
		mapCacheEnabled = true
		mapCacheIsValid = true
		mapCacheImage = ebiten.NewImage(8, 8)
		mapCacheDrawnForWorldTileX = 1
		mapCacheDrawnForWorldTileY = 2
		mapCacheWidthInTiles = 3
		mapCacheHeightInTiles = 4
		mapCacheRenderedLayers = 5
		mapCacheRenderedScreenWidth = 128
		mapCacheRenderedScreenHeight = 128
	}

	t.Run("matching parameters with caching enabled is valid", func(t *testing.T) {
		setValidCache()
		assert.True(t, isMapCacheValid(1, 2, 3, 4, 5, 128, 128))
	})

	t.Run("disabling caching forces invalid regardless of matching parameters", func(t *testing.T) {
		setValidCache()
		mapCacheEnabled = false
		assert.False(t, isMapCacheValid(1, 2, 3, 4, 5, 128, 128),
			"MapCacheEnabled=false must force a rebuild every call, not reuse a stale cache")
	})

	t.Run("mismatched map position invalidates the cache", func(t *testing.T) {
		setValidCache()
		assert.False(t, isMapCacheValid(99, 2, 3, 4, 5, 128, 128))
	})

	t.Run("mismatched screen size invalidates the cache", func(t *testing.T) {
		setValidCache()
		assert.False(t, isMapCacheValid(1, 2, 3, 4, 5, 256, 256))
	})

	t.Run("no image yet is invalid", func(t *testing.T) {
		setValidCache()
		mapCacheImage = nil
		assert.False(t, isMapCacheValid(1, 2, 3, 4, 5, 128, 128))
	})

	t.Run("mapCacheIsValid false is invalid even if everything else matches", func(t *testing.T) {
		setValidCache()
		mapCacheIsValid = false
		assert.False(t, isMapCacheValid(1, 2, 3, 4, 5, 128, 128))
	})
}

func TestMgetReturnsZeroWhenStreamingInitializerFails(t *testing.T) {
	streamingInitMutex.Lock()
	streamingSystemInitialized = false
	worldMapStream = nil
	activeTileBufferInstance = nil
	mapCacheIsValid = false
	streamingInitErr = nil
	streamingInitMutex.Unlock()

	originalStreamingInitializer := streamingInitializer
	streamingInitializer = func() error {
		return errors.New("boom")
	}
	defer func() {
		streamingInitializer = originalStreamingInitializer
	}()

	if got := Mget(0, 0); got != 0 {
		t.Fatalf("expected zero tile when initialization fails, got %d", got)
	}
}
