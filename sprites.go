package pigo8

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/exp/constraints"
)

// SpriteInfo definition is now expected in spritesheet.go within this package

// Number covers ints, floats
type Number interface {
	constraints.Integer | constraints.Float
}

// Note: currentScreen, currentSprites, and currentDrawColor are defined in engine.go

// colorRGBATransparent is a transparent RGBA color used for clearing images
var colorRGBATransparent = color.RGBA{0, 0, 0, 0}

// Add sprite caching for transparent versions
var (
	// Global sprite cache for transparent versions using LRU
	spriteImageCache *SpriteImageCache

	// Sprite pixel cache for batch reading operations using LRU
	spritePixelCacheManager *SpritePixelCache
	spriteCacheValid        = make(map[int]bool) // spriteID -> cache validity
	spritePixelCacheMutex   sync.RWMutex

	// Debug configuration
	debugSpriteLogging = false
	debugMutex         sync.RWMutex

	// Sprite optimization configuration
	optimizeSprites = true
	optimizeMutex   sync.RWMutex

	// Sprite validation configuration
	strictValidation = false
	validationMutex  sync.RWMutex

	// Configuration locking for rendering safety
	configurationLocked = false
	configLockMutex     sync.RWMutex

	// Resource limits
	resourceLimits    = DefaultResourceLimits()
	limitsInitialized = false

	// Cache initialization
	cacheInitOnce sync.Once
)

// ClearSpriteCache clears the sprite cache (useful for memory management)
func ClearSpriteCache() {
	if spriteImageCache != nil {
		spriteImageCache.Clear()
	}
	if spritePixelCacheManager != nil {
		spritePixelCacheManager.Clear()
	}
}

// queueSpriteModification queues a pixel modification for batch processing
func queueSpriteModification(sprite *ebiten.Image, x, y int, clr color.Color) {
	spriteModMutex.Lock()
	defer spriteModMutex.Unlock()

	spriteModifications[sprite] = append(spriteModifications[sprite], pixelMod{x, y, clr})

	// Invalidate sprite pixel cache since we're modifying the sprite
	// Find sprite ID by searching through currentSprites
	for _, spriteInfo := range currentSprites {
		if spriteInfo.Image == sprite {
			invalidateSpritePixelCache(spriteInfo.ID)
			break
		}
	}
}

// flushSpriteModifications applies all pending sprite modifications in batch
func flushSpriteModifications() {
	spriteModMutex.Lock()
	defer spriteModMutex.Unlock()

	for sprite, mods := range spriteModifications {
		if len(mods) > 0 {
			// Get current pixels from GPU
			width := sprite.Bounds().Dx()
			height := sprite.Bounds().Dy()
			pixels := make([]byte, width*height*4)
			sprite.ReadPixels(pixels)

			// Apply all modifications to the pixel buffer
			for _, mod := range mods {
				if mod.x >= 0 && mod.x < width && mod.y >= 0 && mod.y < height {
					offset := (mod.y*width + mod.x) * 4
					r, g, b, a := mod.color.RGBA()
					pixels[offset] = uint8(r >> 8)   // Red
					pixels[offset+1] = uint8(g >> 8) // Green
					pixels[offset+2] = uint8(b >> 8) // Blue
					pixels[offset+3] = uint8(a >> 8) // Alpha
				}
			}

			// Upload all changes back to GPU in one operation
			sprite.WritePixels(pixels)

			// Update sprite pixel cache after modifications
			// Find sprite ID by searching through currentSprites
			for _, spriteInfo := range currentSprites {
				if spriteInfo.Image == sprite {
					updateSpritePixelCache(spriteInfo.ID, sprite)
					break
				}
			}
		}
	}

	// Clear the modifications map
	spriteModifications = make(map[*ebiten.Image][]pixelMod)
}

// initSpritePixelCache initializes the sprite pixel cache for batch reading operations
func initSpritePixelCache(spriteID int, sprite *ebiten.Image) {
	spritePixelCacheMutex.Lock()
	defer spritePixelCacheMutex.Unlock()

	width := sprite.Bounds().Dx()
	height := sprite.Bounds().Dy()
	cacheSize := width * height * 4

	// Check if we need to update the cache (cache should be initialized by now)
	if spritePixelCacheManager != nil {
		if _, currentSize, found := spritePixelCacheManager.Get(spriteID); !found || currentSize != cacheSize {
			pixels := make([]byte, cacheSize)
			spritePixelCacheManager.Put(spriteID, pixels, cacheSize)
			spriteCacheValid[spriteID] = false
		}
	}
}

// updateSpritePixelCache reads all pixels from a sprite into the cache
func updateSpritePixelCache(spriteID int, sprite *ebiten.Image) {
	spritePixelCacheMutex.Lock()
	defer spritePixelCacheMutex.Unlock()

	// Get cached pixels (cache should be initialized by now)
	if spritePixelCacheManager == nil {
		spriteCacheValid[spriteID] = false
		return
	}

	pixels, cacheSize, found := spritePixelCacheManager.Get(spriteID)
	if sprite == nil || !found || cacheSize == 0 {
		spriteCacheValid[spriteID] = false
		return
	}

	// Read all pixels from GPU in one batch operation
	sprite.ReadPixels(pixels)
	spriteCacheValid[spriteID] = true
}

// invalidateSpritePixelCache marks a sprite's pixel cache as invalid
func invalidateSpritePixelCache(spriteID int) {
	spritePixelCacheMutex.Lock()
	defer spritePixelCacheMutex.Unlock()
	spriteCacheValid[spriteID] = false
}

// clearSpritePixelCache clears all sprite pixel caches
func clearSpritePixelCache() {
	spritePixelCacheMutex.Lock()
	defer spritePixelCacheMutex.Unlock()

	if spritePixelCacheManager != nil {
		spritePixelCacheManager.Clear()
	}
	spriteCacheValid = make(map[int]bool)
}

// GetSpritePixelCacheStats returns statistics about sprite pixel caches
func GetSpritePixelCacheStats() (totalSprites int, validSprites int, totalSize int) {
	spritePixelCacheMutex.RLock()
	defer spritePixelCacheMutex.RUnlock()

	if spritePixelCacheManager != nil {
		stats := spritePixelCacheManager.Stats()
		totalSprites = stats.Size
	} else {
		totalSprites = 0
	}
	validSprites = 0
	totalSize = 0

	// Count valid sprites
	for spriteID, valid := range spriteCacheValid {
		if valid {
			validSprites++
			if _, size, found := spritePixelCacheManager.Get(spriteID); found {
				totalSize += size
			}
		}
	}

	return totalSprites, validSprites, totalSize
}

// ForceUpdateSpritePixelCache forces an update of all sprite pixel caches
func ForceUpdateSpritePixelCache() {
	if currentSprites == nil {
		return
	}

	for _, sprite := range currentSprites {
		updateSpritePixelCache(sprite.ID, sprite.Image)
	}
}

// ClearAllCaches clears all pixel caches (screen and sprites)
func ClearAllCaches() {
	clearSpritePixelCache()
	invalidateScreenPixelCache()
	ClearSpriteCache()
}
