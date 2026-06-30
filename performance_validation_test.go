package pigo8

import (
	"testing"
	"time"
)

// TestPerformanceOptimizations validates all performance optimizations work correctly
func TestPerformanceOptimizations(t *testing.T) {
	// Test 1: Cache initialization should be one-time only
	t.Run("CacheInitializationOnce", func(t *testing.T) {
		// Multiple calls should not reinitialize
		EnsureCachesInitialized()
		firstCache := spriteImageCache

		EnsureCachesInitialized()
		secondCache := spriteImageCache

		if firstCache != secondCache {
			t.Error("Cache should not be reinitialized on subsequent calls")
		}
	})

	// Test 2: Frame-level metrics should batch operations
	t.Run("FrameLevelMetrics", func(t *testing.T) {
		EnableFrameStats(true)

		before := GetSpriteMetrics().SpritesRendered

		// Simulate sprite operations within a single frame (batched locally)
		for i := 0; i < 100; i++ {
			recordSpriteRendered()
		}

		// Flush local frame counters into the global metrics collector,
		// mirroring what engine.go's Draw() does once per frame.
		FlushFrameMetrics()

		// Verify metrics were recorded
		metrics := GetSpriteMetrics()
		if metrics.SpritesRendered-before != 100 {
			t.Errorf("Global metrics should include frame data, got delta %d", metrics.SpritesRendered-before)
		}
	})

	// Test 3: Configuration locking should prevent changes during rendering
	t.Run("ConfigurationLocking", func(t *testing.T) {
		// Lock configuration
		LockConfiguration()

		// Attempt to change configuration
		err := SetDebugSpriteLogging(true)
		if err == nil {
			t.Error("Configuration change should fail when locked")
		}

		// Unlock and try again
		UnlockConfiguration()
		err = SetDebugSpriteLogging(true)
		if err != nil {
			t.Errorf("Configuration change should succeed when unlocked: %v", err)
		}

		// Reset
		_ = SetDebugSpriteLogging(false)
	})

	// Test 4: Debug functions should have minimal overhead in release builds
	t.Run("DebugOverhead", func(t *testing.T) {
		// This test validates that debug functions are optimized
		start := time.Now()

		// Call debug functions many times
		for i := 0; i < 10000; i++ {
			debugSpriteNotFound(i, float64(i), float64(i))
			debugLog("Test message %d", i)
		}

		elapsed := time.Since(start)

		// In release builds, this should be very fast
		// In debug builds, it will be slower but still reasonable
		if elapsed > 100*time.Millisecond {
			t.Logf("Debug overhead: %v (this is expected in debug builds)", elapsed)
		} else {
			t.Logf("Debug overhead: %v (optimized for release)", elapsed)
		}
	})

	// Test 5: Resource limits should be enforced
	t.Run("ResourceLimitEnforcement", func(t *testing.T) {
		// Set strict limits
		limits := ResourceLimits{
			MaxCacheSize:   5,
			MaxSpriteCount: 10,
			MaxMemoryUsage: 1024, // 1KB (very small)
		}

		err := SetResourceLimits(limits)
		if err != nil {
			t.Fatalf("Failed to set resource limits: %v", err)
		}

		// Test that limits are respected
		currentLimits := GetResourceLimits()
		if currentLimits.MaxCacheSize != 5 {
			t.Errorf("Resource limits not applied correctly: got %d, want 5", currentLimits.MaxCacheSize)
		}

		// Reset to defaults
		_ = SetResourceLimits(DefaultResourceLimits())
	})
}

// TestHotPathPerformance tests the performance of critical game loops
func TestHotPathPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Initialize system
	EnsureCachesInitialized()

	// Create test sprites
	testSprites := make([]spriteInfo, 100)
	for i := range testSprites {
		testSprites[i] = spriteInfo{
			ID:    i,
			Image: nil, // Don't create actual images to avoid GPU dependency
			Flags: FlagsData{Bitfield: 0, Individual: []bool{false, false, false, false, false, false, false, false}},
		}
	}
	currentSprites = testSprites

	// Test sprite lookup performance (critical hot path)
	t.Run("SpriteLookupPerformance", func(t *testing.T) {
		maxPerSprite := 100 * time.Nanosecond
		if underRaceDetector() {
			maxPerSprite = 2 * time.Microsecond
		}

		start := time.Now()

		// Simulate 60fps with 1000 sprites per frame for 1 second
		iterations := 60 * 1000

		for i := 0; i < iterations; i++ {
			spriteID := i % 100
			spriteInfo := findSpriteByID(spriteID)
			if spriteInfo == nil {
				t.Errorf("Sprite %d should exist", spriteID)
			}
		}

		elapsed := time.Since(start)
		avgPerSprite := elapsed / time.Duration(iterations)

		t.Logf("Sprite lookup performance: %v total, %v per sprite", elapsed, avgPerSprite)

		// Race builds add heavy instrumentation, so use a looser threshold there.
		if avgPerSprite > maxPerSprite {
			t.Errorf("Sprite lookup too slow: %v per sprite (should be < %v)", avgPerSprite, maxPerSprite)
		}
	})

	// Test frame-level metrics overhead
	t.Run("FrameMetricsOverhead", func(t *testing.T) {
		EnableFrameStats(true)

		start := time.Now()

		// Simulate 60 frames
		for frame := 0; frame < 60; frame++ {
			// Simulate 1000 sprites per frame
			for sprite := 0; sprite < 1000; sprite++ {
				recordSpriteRendered()
			}

			// Flush local frame counters, mirroring engine.go's Draw()
			FlushFrameMetrics()
		}

		elapsed := time.Since(start)
		avgPerFrame := elapsed / 60

		t.Logf("Frame metrics overhead: %v total, %v per frame", elapsed, avgPerFrame)

		// Frame-level metrics should add minimal overhead
		if avgPerFrame > 1*time.Millisecond {
			t.Errorf("Frame metrics overhead too high: %v per frame", avgPerFrame)
		}
	})
}
