package pigo8

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// defaultPico8MapWidth defines the default width of the PIGO-8 map in tiles if not specified in map.json.
	defaultPico8MapWidth = 128
	// defaultPico8MapHeight defines the default height of the PIGO-8 map in tiles if not specified in map.json.
	defaultPico8MapHeight = 128

	// activeTileBufferWidthInTiles defines the width of the streaming buffer in tiles.
	// Should be larger than screen width in tiles. E.g., Screen=32tiles -> Buffer=64tiles.
	activeTileBufferWidthInTiles = 64
	// activeTileBufferHeightInTiles defines the height of the streaming buffer in tiles.
	// Should be larger than screen height in tiles. E.g., Screen=30tiles -> Buffer=60tiles.
	activeTileBufferHeightInTiles = 60
)

// --- Structs for map.json parsing (sparse format) ---
type mapCellJSON struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Sprite int `json:"sprite"`
}

type mapDataJSON struct {
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Width       int           `json:"width"`  // World width in tiles
	Height      int           `json:"height"` // World height in tiles
	Name        string        `json:"name"`
	Cells       []mapCellJSON `json:"cells"` // Sparse list of non-zero tiles
}

// --- Structs for Streaming Map System ---

// tilemapStream holds the entire world's map data.
// Data is stored as a 1D slice, indexed by [y * WorldWidthInTiles + x].
type tilemapStream struct {
	Data               []int // Dense representation of the entire world map
	WorldWidthInTiles  int
	WorldHeightInTiles int
}

// activeTileBuffer holds the currently active (buffered) portion of the map.
// Data is stored as a 1D slice, indexed by [y * WidthInTiles + x] (local buffer coordinates).
type activeTileBuffer struct {
	Data           []int // Dense representation of the buffered region
	BufferWorldX   int   // Top-left tile X-coordinate of this buffer in the world
	BufferWorldY   int   // Top-left tile Y-coordinate of this buffer in the world
	WidthInTiles   int   // Width of this buffer in tiles (e.g., ActiveTileBufferWidthInTiles)
	HeightInTiles  int   // Height of this buffer in tiles (e.g., ActiveTileBufferHeightInTiles)
	IsRegionLoaded bool  // True if this buffer currently holds valid map data
}

// newActiveTileBuffer creates a new active tile buffer using the memory pool
func newActiveTileBuffer(width, height int) *activeTileBuffer {
	bufferSize := width * height
	return &activeTileBuffer{
		Data:          getMapBuffer(bufferSize),
		WidthInTiles:  width,
		HeightInTiles: height,
	}
}

var (
	// Streaming Map System
	worldMapStream             *tilemapStream
	activeTileBufferInstance   *activeTileBuffer
	streamingSystemInitialized bool
	streamingInitErr           error
	streamingInitializer       = initializeStreamingMapSystem
	streamingInitMutex         sync.Mutex
	worldMapMutex              sync.RWMutex // Protects worldMapStream
	activeBufferMutex          sync.RWMutex // Protects activeTileBufferInstance

	spriteInfoMap map[int]*spriteInfo // Preserved

	// Map Caching (Preserved)
	mapCacheImage                *ebiten.Image
	mapCacheIsValid              bool
	mapCacheDrawnForWorldTileX   int
	mapCacheDrawnForWorldTileY   int
	mapCacheWidthInTiles         int
	mapCacheHeightInTiles        int
	mapCacheRenderedLayers       int
	mapCacheRenderedScreenWidth  int
	mapCacheRenderedScreenHeight int

	// mapCacheEnabled mirrors Settings.MapCacheEnabled (set via
	// SetMapCacheEnabled during PlayGameWith). When false, drawMapRegion
	// rebuilds its composited tile image from scratch every call instead of
	// reusing it across frames - useful for maps that change every frame,
	// where a stale cache would otherwise show outdated tiles.
	mapCacheEnabled = true

	// Memory monitoring (Preserved)
	lastMemoryUsage uint64
	memoryMutex     sync.Mutex

	// Add memory pool for map buffers
	mapBufferPool = sync.Pool{
		New: func() interface{} {
			// Create a new buffer with default size
			buf := make([]int, activeTileBufferWidthInTiles*activeTileBufferHeightInTiles)
			return &buf
		},
	}
)

// getMapBuffer gets a buffer from the pool or creates a new one
func getMapBuffer(size int) []int {
	bufferPtr := mapBufferPool.Get().(*[]int)
	buffer := *bufferPtr
	if cap(buffer) < size {
		// If the pooled buffer is too small, create a new one
		return make([]int, size)
	}
	// Resize the buffer to the requested size
	return buffer[:size]
}

// returnMapBuffer returns a buffer to the pool
func returnMapBuffer(buffer []int) {
	// Clear the buffer before returning to pool
	for i := range buffer {
		buffer[i] = 0
	}
	mapBufferPool.Put(&buffer)
}

// logMemory logs the current memory usage with a label
func logMemory(label string, forceLog bool) {
	memoryMutex.Lock()
	defer memoryMutex.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert to MB for comparison and display
	currentUsageMB := float64(m.Alloc) / 1024 / 1024

	// If this is the first check, just store the value
	if lastMemoryUsage == 0 {
		lastMemoryUsage = m.Alloc
		log.Printf("PICO-8 Memory: %.2f MB", currentUsageMB)
		return
	}

	// Calculate memory difference
	diffBytes := int64(m.Alloc) - int64(lastMemoryUsage)
	diffMB := float64(diffBytes) / 1024 / 1024

	// Log if forced or if memory changed significantly
	if forceLog || diffMB > 1.0 || diffMB < -1.0 {
		log.Printf("PICO-8 Memory (%s): %.2f MB (change: %.2f MB)", label, currentUsageMB, diffMB)
		lastMemoryUsage = m.Alloc
	}
}

// loadAndParseMapJSON reads and parses a map JSON file.
func loadAndParseMapJSON(filename string) (*mapDataJSON, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		commonLocations := []string{
			filepath.Join("assets", filename),
			filepath.Join("resources", filename),
			filepath.Join("data", filename),
			filepath.Join("static", filename),
		}
		found := false
		for _, location := range commonLocations {
			data, err = os.ReadFile(location)
			if err == nil {
				log.Printf("Loaded map JSON from %s", location)
				found = true
				break
			}
		}
		if !found {
			log.Printf("Map JSON file '%s' not found in common locations, trying embedded resources", filename)
			// Assuming tryLoadEmbeddedMap() returns []byte, error and is defined elsewhere
			embeddedData, embErr := tryLoadEmbeddedMap()
			if embErr != nil {
				return nil, fmt.Errorf("map file '%s' not found locally and failed to load embedded map: %w", filename, embErr)
			}
			data = embeddedData
			log.Printf("Loaded map JSON from embedded resources.")
		}
	} else {
		log.Printf("Loaded map JSON from %s", filename)
	}

	var jsonData mapDataJSON
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse map JSON from %s: %w", filename, err)
	}

	// Validate and default map dimensions if necessary
	if jsonData.Width <= 0 {
		log.Printf("Warning: map JSON from %s has invalid width %d. Using default %d.", filename, jsonData.Width, defaultPico8MapWidth)
		jsonData.Width = defaultPico8MapWidth
	}
	if jsonData.Height <= 0 {
		log.Printf("Warning: map JSON from %s has invalid height %d. Using default %d.", filename, jsonData.Height, defaultPico8MapHeight)
		jsonData.Height = defaultPico8MapHeight
	}
	log.Printf("Parsed map JSON from %s: Version=%s, Desc=%s, Name=%s, Size=%dx%d, Cells=%d",
		filename, jsonData.Version, jsonData.Description, jsonData.Name, jsonData.Width, jsonData.Height, len(jsonData.Cells))

	return &jsonData, nil
}

// Map draws a rectangular region of the PICO-8 map to the screen.
// Optional args: [mx, my, sx, sy, w, h, layers]
//   - mx, my: map tile coordinates in tiles (defaults 0,0)
//   - sx, sy: screen pixel coordinates to draw at (defaults 0,0)
//   - w, h: dimensions in tiles (defaults 16x16)
//   - layers: bitfield to filter sprites by their flags (0 = draw all)
func Map(args ...any) {
	// Default map coordinates
	mx, my := 0, 0

	// If arguments are provided, extract map coordinates
	if len(args) >= 1 {
		if mxVal, ok := args[0].(int); ok {
			mx = mxVal
		} else if mxVal, ok := args[0].(float64); ok {
			mx = int(mxVal)
		}
	}

	if len(args) >= 2 {
		if myVal, ok := args[1].(int); ok {
			my = myVal
		} else if myVal, ok := args[1].(float64); ok {
			my = int(myVal)
		}
	}

	// Pass remaining arguments to the generic implementation
	var remainingArgs []any
	if len(args) > 2 {
		remainingArgs = args[2:]
	}

	// Call the generic implementation
	mapG(mx, my, remainingArgs...)
}

// mapG is the generic version of Map that accepts any number type for coordinates.
// The mx and my coordinates can be any integer or float type (e.g., int, float64)
// due to the use of generics [MX Number, MY Number]. They are converted internally
// to integers for map calculations.
//
// Optional args: [sx, sy, w, h, layers]
//   - sx, sy: screen pixel coordinates to draw at (defaults 0,0)
//   - w, h: dimensions in tiles (defaults 16x16)
//   - layers: bitfield to filter sprites by their flags (0 = draw all)
//
// Usage:
//
//	mapG(0, 0) // Draw map at (0,0) with default size
//	mapG(mx, my) // Draw map at specified coordinates
//	mapG(mx, my, sx, sy) // Draw map at specified coordinates with screen offset
//	mapG(mx, my, sx, sy, w, h) // Draw map with custom dimensions
//	mapG(mx, my, sx, sy, w, h, layers) // Draw map with layer filtering
func mapG[MX Number, MY Number](mx MX, my MY, args ...any) {
	if err := ensureStreamingSystemInitialized(); err != nil {
		log.Printf("Map(): streaming system unavailable: %v", err)
		return
	}

	// Convert generic mx, my to required types
	mapX := int(mx)
	mapY := int(my)

	// Parse arguments
	sx, sy, wTiles, hTiles, layers := parseMapArgs(args)
	if wTiles <= 0 || hTiles <= 0 {
		return
	}

	// Draw the map region
	drawMapRegion(mapX, mapY, sx, sy, wTiles, hTiles, layers)
}

// initializeStreamingMapSystem sets up the TilemapStream and ActiveTileBuffer.
// It should be called only once, typically by EnsureStreamingSystemInitialized.
func initializeStreamingMapSystem() error {
	logMemory("before streaming map init", true)

	const mapFilename = "map.json"
	jsonData, err := loadAndParseMapJSON(mapFilename)

	worldWidth := defaultPico8MapWidth
	worldHeight := defaultPico8MapHeight

	if err == nil && jsonData != nil {
		if jsonData.Width > 0 {
			worldWidth = jsonData.Width
		}
		if jsonData.Height > 0 {
			worldHeight = jsonData.Height
		}
		log.Printf("Initializing streaming map system with world size: %dx%d from %s", worldWidth, worldHeight, mapFilename)
	} else {
		log.Printf("Failed to load map '%s' or map data is invalid: %v. Initializing with default world size: %dx%d", mapFilename, err, worldWidth, worldHeight)
	}

	worldMapMutex.Lock()
	worldMapStream = &tilemapStream{
		Data:               make([]int, worldWidth*worldHeight),
		WorldWidthInTiles:  worldWidth,
		WorldHeightInTiles: worldHeight,
	}

	if jsonData != nil && len(jsonData.Cells) > 0 {
		log.Printf("Populating TilemapStream with %d cells from %s", len(jsonData.Cells), mapFilename)
		populatedTiles := 0
		for _, cell := range jsonData.Cells {
			if cell.X >= 0 && cell.X < worldMapStream.WorldWidthInTiles &&
				cell.Y >= 0 && cell.Y < worldMapStream.WorldHeightInTiles {
				worldMapStream.Data[cell.Y*worldMapStream.WorldWidthInTiles+cell.X] = cell.Sprite
				populatedTiles++
			} else {
				log.Printf("Warning: cell data out of bounds in %s: (%d, %d) for sprite %d. World size: %dx%d. Skipping cell.",
					mapFilename, cell.X, cell.Y, cell.Sprite, worldMapStream.WorldWidthInTiles, worldMapStream.WorldHeightInTiles)
			}
		}
		log.Printf("Finished populating TilemapStream. %d cells processed, %d tiles set.", len(jsonData.Cells), populatedTiles)
	} else if err == nil {
		log.Printf("Map JSON '%s' loaded but contains no cell data (or jsonData is nil after parse attempt). World map will be default (empty).", mapFilename)
	}
	worldMapMutex.Unlock()

	activeBufferMutex.Lock()
	activeTileBufferInstance = newActiveTileBuffer(activeTileBufferWidthInTiles, activeTileBufferHeightInTiles)
	activeTileBufferInstance.BufferWorldX = -1
	activeTileBufferInstance.BufferWorldY = -1
	activeTileBufferInstance.IsRegionLoaded = false
	activeBufferMutex.Unlock()

	log.Println("Streaming map system initialized.")
	logMemory("after streaming map init", true)
	return nil
}

// ensureStreamingSystemInitialized guarantees that the streaming map system is set up.
// This function is responsible for calling initializeStreamingMapSystem once,
// loading spritesheets, and setting up map cache parameters.
// It should be called by map-accessing functions like Mget, Mset, Map.
func ensureStreamingSystemInitialized() error {
	if streamingSystemInitialized {
		return nil
	}

	streamingInitMutex.Lock()
	defer streamingInitMutex.Unlock()

	if streamingSystemInitialized {
		return nil
	}
	if streamingInitErr != nil {
		return streamingInitErr
	}

	log.Println("EnsureStreamingSystemInitialized: Initializing...")

	if currentSprites == nil {
		var err error
		currentSprites, err = loadSpritesheet()
		if err != nil {
			log.Printf("EnsureStreamingSystemInitialized: Failed to load spritesheet: %v. Map operations might be affected.", err)
		} else {
			log.Println("EnsureStreamingSystemInitialized: Spritesheet loaded.")

			if len(spriteInfoMap) == 0 {
				spriteInfoMap = make(map[int]*spriteInfo, len(currentSprites))
				for i := range currentSprites {
					info := &currentSprites[i]
					spriteInfoMap[info.ID] = info
				}
				log.Println("EnsureStreamingSystemInitialized: Sprite info map populated.")
			}
		}
	}

	if err := streamingInitializer(); err != nil {
		streamingInitErr = fmt.Errorf("failed to initialize streaming map system: %w", err)
		log.Printf("EnsureStreamingSystemInitialized: %v", streamingInitErr)
		return streamingInitErr
	}

	mapCacheIsValid = false
	if GetScreenWidth() > 0 && GetScreenHeight() > 0 {
		mapCacheWidthInTiles = GetScreenWidth() / 8
		mapCacheHeightInTiles = GetScreenHeight() / 8
	} else {
		log.Printf("EnsureStreamingSystemInitialized: ScreenWidth/ScreenHeight not available or zero. Using default map cache dimensions.")
		mapCacheWidthInTiles = defaultPico8MapWidth / 2
		mapCacheHeightInTiles = defaultPico8MapHeight / 2
	}
	if mapCacheWidthInTiles <= 0 {
		mapCacheWidthInTiles = 16
	}
	if mapCacheHeightInTiles <= 0 {
		mapCacheHeightInTiles = 16
	}

	log.Printf("EnsureStreamingSystemInitialized: Map cache parameters set (Width: %d tiles, Height: %d tiles).", mapCacheWidthInTiles, mapCacheHeightInTiles)

	log.Println("EnsureStreamingSystemInitialized: System ready.")
	streamingInitErr = nil
	streamingSystemInitialized = true
	return nil
}

// parseMapArgs parses the optional arguments for the Map functions
// Returns screen x, screen y, width in tiles, height in tiles, and layers bitfield
func parseMapArgs(args []any) (sx, sy, wTiles, hTiles, layers int) {
	// Default parameters
	sx, sy = 0, 0
	wTiles = defaultPico8MapWidth
	hTiles = defaultPico8MapHeight
	layers = 0

	// Process optional arguments
	if len(args) >= 1 {
		if sxVal, ok := args[0].(int); ok {
			sx = sxVal
		} else if sxVal, ok := args[0].(float64); ok {
			sx = int(sxVal)
		}
	}
	if len(args) >= 2 {
		if syVal, ok := args[1].(int); ok {
			sy = syVal
		} else if syVal, ok := args[1].(float64); ok {
			sy = int(syVal)
		}
	}
	if len(args) >= 3 {
		if wVal, ok := args[2].(int); ok {
			wTiles = wVal
		} else if wVal, ok := args[2].(float64); ok {
			wTiles = int(wVal)
		}
	}
	if len(args) >= 4 {
		if hVal, ok := args[3].(int); ok {
			hTiles = hVal
		} else if hVal, ok := args[3].(float64); ok {
			hTiles = int(hVal)
		}
	}
	if len(args) >= 5 {
		if layerVal, ok := args[4].(int); ok {
			layers = layerVal
		} else if layerVal, ok := args[4].(float64); ok {
			layers = int(layerVal)
		}
	}
	if len(args) > 5 {
		log.Printf("Warning: Map() called with too many arguments (%d), expected max 5 ([sx,sy,w,h,layers]).", len(args))
	}

	return sx, sy, wTiles, hTiles, layers
}

// SetMapCacheEnabled enables or disables the composited-image cache used by
// Map()/drawMapRegion. Mirrors Settings.MapCacheEnabled; also callable at
// runtime for games that need to toggle it (e.g. right before drawing a map
// that's regenerated every frame, where a stale cache would show outdated
// tiles).
func SetMapCacheEnabled(enabled bool) {
	mapCacheEnabled = enabled
}

// isMapCacheValid reports whether the existing mapCacheImage can be reused
// as-is for a draw call with the given parameters, rather than being
// rebuilt. Always false when mapCacheEnabled is false, so every call forces
// a fresh rebuild instead of reusing a (potentially stale) prior frame.
func isMapCacheValid(mapX, mapY, wTiles, hTiles, layers, screenWidth, screenHeight int) bool {
	return mapCacheEnabled &&
		mapCacheIsValid &&
		mapCacheImage != nil &&
		mapCacheDrawnForWorldTileX == mapX &&
		mapCacheDrawnForWorldTileY == mapY &&
		mapCacheWidthInTiles == wTiles &&
		mapCacheHeightInTiles == hTiles &&
		mapCacheRenderedLayers == layers &&
		mapCacheRenderedScreenWidth == screenWidth &&
		mapCacheRenderedScreenHeight == screenHeight
}

// drawMapRegion draws a region of the map to the screen using a cache
func drawMapRegion(mapX, mapY, sx, sy, wTiles, hTiles, layers int) {
	if wTiles <= 0 || hTiles <= 0 {
		return
	}

	// Check cache validity
	cacheIsCurrentlyValid := isMapCacheValid(mapX, mapY, wTiles, hTiles, layers, GetScreenWidth(), GetScreenHeight())

	if !cacheIsCurrentlyValid {
		// Invalidate and rebuild cache
		requiredCacheWidth := wTiles * 8
		requiredCacheHeight := hTiles * 8

		// Ensure mapCacheImage exists and is the correct size
		if mapCacheImage == nil || mapCacheImage.Bounds().Dx() != requiredCacheWidth || mapCacheImage.Bounds().Dy() != requiredCacheHeight {
			if mapCacheImage != nil {
				mapCacheImage.Deallocate() // Dispose old image before creating new
			}
			mapCacheImage = ebiten.NewImage(requiredCacheWidth, requiredCacheHeight)
		} else {
			mapCacheImage.Clear() // Clear existing image for redraw
		}

		if mapCacheImage == nil { // Still nil after attempt to create
			log.Println("Error: Failed to create or clear mapCacheImage")
			return
		}

		// Iterate through the tiles that should be visible in the cache
		for ty := 0; ty < hTiles; ty++ {
			for tx := 0; tx < wTiles; tx++ {
				worldTileX := mapX + tx
				worldTileY := mapY + ty

				spriteID := Mget(worldTileX, worldTileY) // Mget handles map boundaries
				if spriteID == 0 {                       // Empty tile or out of bounds according to Mget's logic
					continue
				}

				// Layer check
				if layers > 0 {
					flagBits, _ := Fget(spriteID)
					if flagBits&layers == 0 {
						continue
					}
				}

				tileImg := getSpriteImage(spriteID) // GetSpriteImage handles nil if sprite not found
				if tileImg != nil {
					// Use transparency shader for PICO-8 style color-key transparency
					// (black pixels rendered as transparent)
					shader := getTransparencyShader()
					if shader != nil {
						var shaderOpts ebiten.DrawRectShaderOptions
						shaderOpts.GeoM.Translate(float64(tx*8), float64(ty*8))
						shaderOpts.Images[0] = tileImg
						bounds := tileImg.Bounds()
						mapCacheImage.DrawRectShader(bounds.Dx(), bounds.Dy(), shader, &shaderOpts)
					} else {
						// Fallback to regular DrawImage if shader unavailable
						opts := &ebiten.DrawImageOptions{}
						opts.Filter = ebiten.FilterNearest
						opts.GeoM.Translate(float64(tx*8), float64(ty*8))
						mapCacheImage.DrawImage(tileImg, opts)
					}
				}
			}
		}

		// Update cache state variables
		mapCacheDrawnForWorldTileX = mapX
		mapCacheDrawnForWorldTileY = mapY
		mapCacheWidthInTiles = wTiles
		mapCacheHeightInTiles = hTiles
		mapCacheRenderedLayers = layers
		mapCacheRenderedScreenWidth = GetScreenWidth()
		mapCacheRenderedScreenHeight = GetScreenHeight()
		mapCacheIsValid = true
		// log.Printf("Map cache rebuilt for world (%d,%d) screen (%d,%d) tiles %dx%d layers %d screen_dims %dx%d", mapX, mapY, sx, sy, wTiles, hTiles, layers, ScreenWidth, ScreenHeight)
	}

	// Draw the (now valid) cache to the screen
	screenToDrawOn := CurrentScreen() // Get the main screen from engine
	if screenToDrawOn == nil || mapCacheImage == nil {
		// log.Println("Warning: Cannot draw map cache, screenToDrawOn or mapCacheImage is nil.")
		return
	}

	drawOpts := &ebiten.DrawImageOptions{}
	drawOpts.Filter = ebiten.FilterNearest
	// Apply the global PIGO-8 camera offset. sx and sy are the screen coordinates
	// passed to Map() (e.g., 0,0 if Map() is called with no arguments).
	// cameraX and cameraY are the global offsets from the pigo8.Camera() function.
	finalScreenX := float64(sx) - cameraX
	finalScreenY := float64(sy) - cameraY
	drawOpts.GeoM.Translate(finalScreenX, finalScreenY)
	screenToDrawOn.DrawImage(mapCacheImage, drawOpts)

	// Mark shadow buffer dirty since we drew to the screen
	MarkShadowBufferDirtyFromSprite()
}

// loadRegionIntoActiveBuffer loads the specified region of the world map into the active tile buffer.
// It attempts to center the buffer around targetWorldX, targetWorldY.
// This function acquires necessary locks.
func loadRegionIntoActiveBuffer(targetWorldX, targetWorldY int) error {
	activeBufferMutex.Lock()
	defer activeBufferMutex.Unlock()
	worldMapMutex.RLock() // Read-only access to worldMapStream needed
	defer worldMapMutex.RUnlock()

	if worldMapStream == nil {
		return fmt.Errorf("loadRegionIntoActiveBuffer: worldMapStream is nil, system not properly initialized")
	}
	if activeTileBufferInstance == nil {
		return fmt.Errorf("loadRegionIntoActiveBuffer: activeTileBufferInstance is nil, system not properly initialized")
	}
	if activeTileBufferInstance.WidthInTiles <= 0 || activeTileBufferInstance.HeightInTiles <= 0 {
		return fmt.Errorf("loadRegionIntoActiveBuffer: activeTileBufferInstance has invalid dimensions (%dx%d)", activeTileBufferInstance.WidthInTiles, activeTileBufferInstance.HeightInTiles)
	}

	newBufferWorldX := targetWorldX - activeTileBufferInstance.WidthInTiles/2
	newBufferWorldY := targetWorldY - activeTileBufferInstance.HeightInTiles/2

	if worldMapStream.WorldWidthInTiles > activeTileBufferInstance.WidthInTiles {
		newBufferWorldX = min(newBufferWorldX, worldMapStream.WorldWidthInTiles-activeTileBufferInstance.WidthInTiles)
	} else {
		newBufferWorldX = 0
	}
	newBufferWorldX = max(0, newBufferWorldX)

	if worldMapStream.WorldHeightInTiles > activeTileBufferInstance.HeightInTiles {
		newBufferWorldY = min(newBufferWorldY, worldMapStream.WorldHeightInTiles-activeTileBufferInstance.HeightInTiles)
	} else {
		newBufferWorldY = 0
	}
	newBufferWorldY = max(0, newBufferWorldY)

	requiredBufferSize := activeTileBufferInstance.WidthInTiles * activeTileBufferInstance.HeightInTiles
	if activeTileBufferInstance.Data == nil || len(activeTileBufferInstance.Data) != requiredBufferSize {
		// Return old buffer to pool if it exists
		if activeTileBufferInstance.Data != nil {
			returnMapBuffer(activeTileBufferInstance.Data)
		}
		// Get new buffer from pool
		activeTileBufferInstance.Data = getMapBuffer(requiredBufferSize)
	}

	for y := 0; y < activeTileBufferInstance.HeightInTiles; y++ {
		for x := 0; x < activeTileBufferInstance.WidthInTiles; x++ {
			worldX := newBufferWorldX + x
			worldY := newBufferWorldY + y
			bufferIndex := y*activeTileBufferInstance.WidthInTiles + x

			if worldX >= 0 && worldX < worldMapStream.WorldWidthInTiles &&
				worldY >= 0 && worldY < worldMapStream.WorldHeightInTiles {
				worldIndex := worldY*worldMapStream.WorldWidthInTiles + worldX
				activeTileBufferInstance.Data[bufferIndex] = worldMapStream.Data[worldIndex]
			} else {
				activeTileBufferInstance.Data[bufferIndex] = 0
			}
		}
	}

	activeTileBufferInstance.BufferWorldX = newBufferWorldX
	activeTileBufferInstance.BufferWorldY = newBufferWorldY
	activeTileBufferInstance.IsRegionLoaded = true

	// log.Printf("Loaded region into active buffer. Origin: (%d,%d), Target: (%d,%d)", newBufferWorldX, newBufferWorldY, targetWorldX, targetWorldY)
	// logMemory("after loadRegionIntoActiveBuffer", false)
	return nil
}

// Mget returns the sprite number at the specified map coordinates.
// This mimics PICO-8's mget(column, row) function.
func Mget[C Number, R Number](column C, row R) int {
	if err := ensureStreamingSystemInitialized(); err != nil {
		log.Printf("Mget: streaming system unavailable: %v", err)
		return 0
	}

	col := int(column)
	r := int(row)

	worldMapMutex.RLock()
	if worldMapStream == nil {
		log.Printf("Mget: worldMapStream is nil. Streaming system not initialized.")
		worldMapMutex.RUnlock()
		return 0
	}
	worldWidth := worldMapStream.WorldWidthInTiles
	worldHeight := worldMapStream.WorldHeightInTiles
	worldMapMutex.RUnlock()

	if col < 0 || col >= worldWidth || r < 0 || r >= worldHeight {
		return 0
	}

	activeBufferMutex.RLock()
	tileInBuff := activeTileBufferInstance != nil &&
		activeTileBufferInstance.IsRegionLoaded &&
		col >= activeTileBufferInstance.BufferWorldX && col < activeTileBufferInstance.BufferWorldX+activeTileBufferInstance.WidthInTiles &&
		r >= activeTileBufferInstance.BufferWorldY && r < activeTileBufferInstance.BufferWorldY+activeTileBufferInstance.HeightInTiles

	if tileInBuff {
		bufferX := col - activeTileBufferInstance.BufferWorldX
		bufferY := r - activeTileBufferInstance.BufferWorldY
		val := activeTileBufferInstance.Data[bufferY*activeTileBufferInstance.WidthInTiles+bufferX]
		activeBufferMutex.RUnlock()
		return val
	}
	activeBufferMutex.RUnlock()

	if err := loadRegionIntoActiveBuffer(col, r); err != nil {
		log.Printf("Mget: Error loading region for (%d,%d): %v", col, r, err)
		return 0
	}

	activeBufferMutex.RLock()
	defer activeBufferMutex.RUnlock()

	if activeTileBufferInstance == nil || !activeTileBufferInstance.IsRegionLoaded {
		log.Printf("Mget: Active buffer still not loaded after loadRegionIntoActiveBuffer for (%d,%d)", col, r)
		return 0
	}

	if col < activeTileBufferInstance.BufferWorldX || col >= activeTileBufferInstance.BufferWorldX+activeTileBufferInstance.WidthInTiles ||
		r < activeTileBufferInstance.BufferWorldY || r >= activeTileBufferInstance.BufferWorldY+activeTileBufferInstance.HeightInTiles {
		log.Printf("Mget: Target tile (%d,%d) NOT in buffer after load. Buffer: (%d,%d) %dx%d. This is unexpected.",
			col, r, activeTileBufferInstance.BufferWorldX, activeTileBufferInstance.BufferWorldY, activeTileBufferInstance.WidthInTiles, activeTileBufferInstance.HeightInTiles)
		return 0
	}

	bufferX := col - activeTileBufferInstance.BufferWorldX
	bufferY := r - activeTileBufferInstance.BufferWorldY
	if bufferX < 0 || bufferX >= activeTileBufferInstance.WidthInTiles || bufferY < 0 || bufferY >= activeTileBufferInstance.HeightInTiles {
		log.Printf("Mget: Calculated local buffer coordinates (%d,%d) are out of bounds for buffer size %dx%d. World: (%d,%d). This is a critical error.",
			bufferX, bufferY, activeTileBufferInstance.WidthInTiles, activeTileBufferInstance.HeightInTiles, col, r)
		return 0
	}

	return activeTileBufferInstance.Data[bufferY*activeTileBufferInstance.WidthInTiles+bufferX]
}

// Mset sets the sprite number at the specified map coordinates.
// This mimics PICO-8's mset(column, row, sprite) function.
func Mset[C Number, R Number, S Number](column C, row R, sprite S) {
	if err := ensureStreamingSystemInitialized(); err != nil {
		log.Printf("Mset: streaming system unavailable: %v", err)
		return
	}

	col := int(column)
	r := int(row)
	spriteNum := int(sprite)

	if spriteNum < 0 {
		log.Printf("Mset: Invalid sprite number %d for (%d,%d). Must be >= 0.", spriteNum, col, r)
		return
	}

	worldMapMutex.Lock()
	if worldMapStream == nil {
		log.Printf("Mset: worldMapStream is nil. Streaming system not initialized.")
		worldMapMutex.Unlock()
		return
	}

	if col < 0 || col >= worldMapStream.WorldWidthInTiles || r < 0 || r >= worldMapStream.WorldHeightInTiles {
		log.Printf("Mset: Coordinates (%d,%d) are out of world bounds (%dx%d).",
			col, r, worldMapStream.WorldWidthInTiles, worldMapStream.WorldHeightInTiles)
		worldMapMutex.Unlock()
		return
	}

	worldMapStream.Data[r*worldMapStream.WorldWidthInTiles+col] = spriteNum
	worldMapMutex.Unlock()

	activeBufferMutex.Lock()
	if activeTileBufferInstance != nil && activeTileBufferInstance.IsRegionLoaded &&
		col >= activeTileBufferInstance.BufferWorldX && col < activeTileBufferInstance.BufferWorldX+activeTileBufferInstance.WidthInTiles &&
		r >= activeTileBufferInstance.BufferWorldY && r < activeTileBufferInstance.BufferWorldY+activeTileBufferInstance.HeightInTiles {

		bufferX := col - activeTileBufferInstance.BufferWorldX
		bufferY := r - activeTileBufferInstance.BufferWorldY
		activeTileBufferInstance.Data[bufferY*activeTileBufferInstance.WidthInTiles+bufferX] = spriteNum
	}
	activeBufferMutex.Unlock()

	mapCacheIsValid = false
	// log.Printf("Mset: Set tile at (%d,%d) to sprite %d. Map cache invalidated.", col, r, spriteNum)
}

// SetMap directly sets the entire PICO-8 map data from a byte slice.
// The data slice should contain DefaultPico8MapHeight * DefaultPico8MapWidth bytes,
// representing sprite IDs in row-major order.
func SetMap(data []byte) {
	if err := ensureStreamingSystemInitialized(); err != nil {
		log.Printf("SetMap: streaming system unavailable: %v", err)
		return
	}

	expectedLen := defaultPico8MapWidth * defaultPico8MapHeight
	if len(data) != expectedLen {
		log.Printf("Warning: SetMap received data of incorrect length. Expected %d, got %d", expectedLen, len(data))
		return
	}

	worldMapMutex.Lock()
	// Ensure worldMapStream is initialized with default dimensions if it's nil or has different dimensions
	if worldMapStream == nil || worldMapStream.WorldWidthInTiles != defaultPico8MapWidth || worldMapStream.WorldHeightInTiles != defaultPico8MapHeight {
		log.Printf("SetMap: Initializing/resetting worldMapStream to default dimensions (%dx%d).", defaultPico8MapWidth, defaultPico8MapHeight)
		worldMapStream = &tilemapStream{
			Data:               make([]int, defaultPico8MapWidth*defaultPico8MapHeight),
			WorldWidthInTiles:  defaultPico8MapWidth,
			WorldHeightInTiles: defaultPico8MapHeight,
		}
	} else {
		// If it exists and has correct dimensions, clear existing data by re-making the slice
		worldMapStream.Data = make([]int, defaultPico8MapWidth*defaultPico8MapHeight)
	}

	for i := 0; i < expectedLen; i++ {
		worldMapStream.Data[i] = int(data[i])
	}
	worldMapMutex.Unlock()

	activeBufferMutex.Lock()
	if activeTileBufferInstance != nil {
		activeTileBufferInstance.IsRegionLoaded = false // Invalidate buffer as world map changed
		// log.Printf("SetMap: Active tile buffer invalidated.")
	}
	activeBufferMutex.Unlock()

	mapCacheIsValid = false
	log.Printf("SetMap: World map data updated from byte slice. Active buffer and map cache invalidated.")
}
