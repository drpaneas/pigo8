package pigo8

import "fmt"

// Engine integration functions for optimal performance

// InitializeSpriteSystem initializes the sprite system for optimal performance
// Call this once at application startup, before any sprite operations
func InitializeSpriteSystem() {
	// Initialize caches once at startup
	EnsureCachesInitialized()

	// Enable frame stats by default in debug builds
	EnableFrameStats(debugEnabled())
}

// OptimizeSpritesForGame pre-warms caches and optimizes sprite data for game performance
func OptimizeSpritesForGame() error {
	// Ensure sprites are loaded
	if currentSprites == nil {
		sprites, err := loadSpritesheet()
		if err != nil {
			return err
		}
		currentSprites = sprites
	}

	// Pre-warm sprite cache to avoid cache misses during gameplay
	PrewarmSpriteCache(currentSprites)

	// Pre-initialize pixel caches for commonly used sprites
	for _, sprite := range currentSprites {
		if sprite.Image != nil {
			initSpritePixelCache(sprite.ID, sprite.Image)
			updateSpritePixelCache(sprite.ID, sprite.Image)
		}
	}

	return nil
}

// GetPerformanceReport returns a formatted performance report for debugging
func GetPerformanceReport() string {
	metrics := GetSpriteMetrics()

	return fmt.Sprintf(`
PIGO8 Sprite System Performance Report
=====================================
Sprites Rendered: %d
Cache Hit Rate: %.2f%%
Memory Usage: %.2f MB
Deduplication Rate: %.2f%%
Average Render Time: %v
Hash Collisions: %d
Validation Errors: %d

Cache Statistics:
- Image Cache: %d/%d entries (%.1f%% full)
- Pixel Cache: %d/%d entries (%.1f%% full)
`,
		metrics.SpritesRendered,
		metrics.ImageCacheStats.HitRate*100,
		float64(metrics.MemoryUsage)/(1024*1024),
		metrics.DeduplicationRate*100,
		metrics.AverageRenderTime,
		metrics.HashCollisions,
		metrics.ValidationErrors,
		metrics.ImageCacheStats.Size,
		metrics.ImageCacheStats.MaxSize,
		float64(metrics.ImageCacheStats.Size)/float64(metrics.ImageCacheStats.MaxSize)*100,
		metrics.PixelCacheStats.Size,
		metrics.PixelCacheStats.MaxSize,
		float64(metrics.PixelCacheStats.Size)/float64(metrics.PixelCacheStats.MaxSize)*100,
	)
}
