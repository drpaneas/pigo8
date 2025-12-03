package pigo8

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// RenderConfig holds rendering configuration options
type RenderConfig struct {
	DebugSprites     bool // Enable debug logging for sprite operations (default: false)
	OptimizeSprites  bool // Enable sprite optimization during loading (default: true)
	StrictValidation bool // Enable strict sprite validation - fail on invalid sprites (default: false)
}

// SetDebugSpriteLogging controls whether debug information is logged for sprite operations.
// This is useful for debugging sprite lookup issues and performance analysis.
//
// This function is thread-safe.
//
// Parameters:
//   - enabled: true to enable debug logging, false to disable (default)
//
// Example:
//
//	// Enable debug logging
//	p8.SetDebugSpriteLogging(true)
func SetDebugSpriteLogging(enabled bool) error {
	return safeConfigurationChange(func() {
		debugMutex.Lock()
		defer debugMutex.Unlock()
		debugSpriteLogging = enabled
	})
}

// SetOptimizeSprites controls whether sprite optimization (deduplication) is enabled during loading.
// When enabled, duplicate sprites are detected and mapped to reduce memory usage.
//
// This function is thread-safe.
//
// Parameters:
//   - enabled: true to enable sprite optimization (default), false to disable
//
// Example:
//
//	// Disable sprite optimization for debugging
//	p8.SetOptimizeSprites(false)
func SetOptimizeSprites(enabled bool) error {
	return safeConfigurationChange(func() {
		optimizeMutex.Lock()
		defer optimizeMutex.Unlock()
		optimizeSprites = enabled
	})
}

// SetStrictValidation controls whether strict sprite validation is enabled.
// When enabled, invalid sprites cause loading to fail with an error.
// When disabled, invalid sprites are skipped with a warning (default behavior).
//
// This function is thread-safe.
//
// Parameters:
//   - enabled: true to enable strict validation, false for lenient validation (default)
//
// Example:
//
//	// Enable strict validation for development
//	p8.SetStrictValidation(true)
func SetStrictValidation(enabled bool) error {
	return safeConfigurationChange(func() {
		validationMutex.Lock()
		defer validationMutex.Unlock()
		strictValidation = enabled
	})
}

// SetRenderConfig applies a complete render configuration in a thread-safe manner.
//
// Parameters:
//   - config: RenderConfig struct with desired settings
//
// Example:
//
//	config := p8.RenderConfig{
//		DebugSprites:     true,
//		OptimizeSprites:  true,
//		StrictValidation: false,
//	}
//	p8.SetRenderConfig(config)
func SetRenderConfig(config RenderConfig) error {
	if err := SetDebugSpriteLogging(config.DebugSprites); err != nil {
		return fmt.Errorf("failed to set debug sprites: %w", err)
	}
	if err := SetOptimizeSprites(config.OptimizeSprites); err != nil {
		return fmt.Errorf("failed to set optimize sprites: %w", err)
	}
	if err := SetStrictValidation(config.StrictValidation); err != nil {
		return fmt.Errorf("failed to set strict validation: %w", err)
	}
	return nil
}

// GetRenderConfig returns the current render configuration in a thread-safe manner.
func GetRenderConfig() RenderConfig {
	debugMutex.RLock()
	optimizeMutex.RLock()
	validationMutex.RLock()
	defer func() {
		validationMutex.RUnlock()
		optimizeMutex.RUnlock()
		debugMutex.RUnlock()
	}()

	return RenderConfig{
		DebugSprites:     debugSpriteLogging,
		OptimizeSprites:  optimizeSprites,
		StrictValidation: strictValidation,
	}
}

// isDebugEnabled returns whether debug logging is enabled (thread-safe)
func isDebugEnabled() bool {
	debugMutex.RLock()
	defer debugMutex.RUnlock()
	return debugSpriteLogging
}

// isOptimizationEnabled returns whether sprite optimization is enabled (thread-safe)
func isOptimizationEnabled() bool {
	optimizeMutex.RLock()
	defer optimizeMutex.RUnlock()
	return optimizeSprites
}

// isStrictValidationEnabled returns whether strict validation is enabled (thread-safe)
func isStrictValidationEnabled() bool {
	validationMutex.RLock()
	defer validationMutex.RUnlock()
	return strictValidation
}

// initializeCaches initializes the sprite cache systems (thread-safe, one-time only)
func initializeCaches() {
	cacheInitOnce.Do(func() {
		if !limitsInitialized {
			resourceLimits = DefaultResourceLimits()
			limitsInitialized = true
		}

		spriteImageCache = NewSpriteImageCache(resourceLimits.MaxCacheSize)
		spritePixelCacheManager = NewSpritePixelCache(resourceLimits.MaxCacheSize)
	})
}

// EnsureCachesInitialized ensures caches are initialized (call this at startup)
// This prevents initialization overhead in hot paths
func EnsureCachesInitialized() {
	initializeCaches()
}

// PrewarmSpriteCache pre-creates transparent versions of sprites to avoid cache misses
func PrewarmSpriteCache(sprites []spriteInfo) {
	if spriteImageCache == nil {
		return
	}

	for _, sprite := range sprites {
		if sprite.Image != nil {
			// Pre-create transparent version
			_ = createTransparentSpriteImage(sprite.Image)
		}
	}
}

// LockConfiguration prevents configuration changes during rendering
// This should be called at the start of each frame to prevent visual glitches
func LockConfiguration() {
	configLockMutex.Lock()
	defer configLockMutex.Unlock()
	configurationLocked = true
}

// UnlockConfiguration allows configuration changes after rendering
// This should be called at the end of each frame
func UnlockConfiguration() {
	configLockMutex.Lock()
	defer configLockMutex.Unlock()
	configurationLocked = false
}

// isConfigurationLocked checks if configuration is currently locked
func isConfigurationLocked() bool {
	configLockMutex.RLock()
	defer configLockMutex.RUnlock()
	return configurationLocked
}

// safeConfigurationChange executes a configuration change only if not locked
func safeConfigurationChange(changeFn func()) error {
	if isConfigurationLocked() {
		return fmt.Errorf("configuration is locked during rendering - changes not allowed")
	}
	changeFn()
	return nil
}

// SetResourceLimits configures resource usage limits for the sprite system
func SetResourceLimits(limits ResourceLimits) error {
	return safeConfigurationChange(func() {
		resourceLimits = limits

		// Reinitialize caches with new limits if they exist
		if spriteImageCache != nil {
			// Note: This is a simplified approach - in production you might want to preserve existing cache entries
			spriteImageCache = NewSpriteImageCache(limits.MaxCacheSize)
		}
		if spritePixelCacheManager != nil {
			spritePixelCacheManager = NewSpritePixelCache(limits.MaxCacheSize)
		}
	})
}

// GetResourceLimits returns the current resource limits
func GetResourceLimits() ResourceLimits {
	return resourceLimits
}

// CheckResourceUsage checks current resource usage against limits and returns warnings
func CheckResourceUsage() []string {
	return CheckResourceLimits(resourceLimits)
}

// ForceGarbageCollection forces garbage collection and cache cleanup
func ForceGarbageCollection() {
	// Clear caches to free memory
	ClearSpriteCache()

	// Clear hash table
	spriteHashTable.Clear()

	// Force Go garbage collection
	runtime.GC()

	if isDebugEnabled() {
		metrics := GetSpriteMetrics()
		log.Printf("Debug: Forced garbage collection - Memory usage: %d bytes", metrics.MemoryUsage)
	}
}

// ReloadSprites reloads the sprite system from disk - critical for game development workflow
func ReloadSprites() error {
	start := time.Now()
	defer func() {
		metricsCollector.RecordLoadTime(time.Since(start))
	}()

	if isDebugEnabled() {
		log.Printf("Debug: Starting sprite hot-reload...")
	}

	// Clear all caches first
	ClearSpriteCache()
	spriteHashTable.Clear()

	// Reset sprite pixel cache validity
	spritePixelCacheMutex.Lock()
	spriteCacheValid = make(map[int]bool)
	spritePixelCacheMutex.Unlock()

	// Reload sprites from disk
	newSprites, err := loadSpritesheet()
	if err != nil {
		return fmt.Errorf("failed to reload sprites: %w", err)
	}

	// Update global sprite reference (thread-safe)
	currentSpritesMu.Lock()
	currentSprites = newSprites
	currentSpritesMu.Unlock()

	// Invalidate sprite ID index so it gets rebuilt with the new sprites
	InvalidateSpriteIDIndex()

	if isDebugEnabled() {
		log.Printf("Debug: Hot-reload completed - loaded %d sprites in %v",
			len(newSprites), time.Since(start))
	}

	return nil
}

// WatchSpritesheetFile watches the spritesheet file for changes and auto-reloads
// This is useful during development but should be disabled in production
func WatchSpritesheetFile(filePath string, callback func(error)) {
	if !isDebugEnabled() {
		log.Printf("Warning: Sprite file watching requires debug mode to be enabled")
		return
	}

	go func() {
		var lastModTime time.Time

		for {
			if stat, err := os.Stat(filePath); err == nil {
				if !lastModTime.IsZero() && stat.ModTime().After(lastModTime) {
					if isDebugEnabled() {
						log.Printf("Debug: Detected spritesheet file change, reloading...")
					}

					if reloadErr := ReloadSprites(); reloadErr != nil {
						if callback != nil {
							callback(reloadErr)
						}
					} else {
						if callback != nil {
							callback(nil)
						}
					}
				}
				lastModTime = stat.ModTime()
			}

			time.Sleep(1 * time.Second) // Check every second
		}
	}()
}
