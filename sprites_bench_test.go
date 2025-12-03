package pigo8

import (
	"runtime"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkSpriteHashGeneration benchmarks sprite hash generation performance
func BenchmarkSpriteHashGeneration(b *testing.B) {
	pixels := make([][]int, 8)
	for i := range pixels {
		pixels[i] = make([]int, 8)
		for j := range pixels[i] {
			pixels[i][j] = (i + j) % 16
		}
	}

	flags := FlagsData{
		Bitfield:   0,
		Individual: []bool{false, false, false, false, false, false, false, false},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateOptimizedSpriteHashInternal(pixels, flags)
	}
}

// BenchmarkLRUCacheOperations benchmarks LRU cache performance
func BenchmarkLRUCacheOperations(b *testing.B) {
	cache := NewLRUCache[int, string](1000)

	b.Run("Put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Put(i%1000, "test_value")
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			cache.Put(i, "test_value")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(i % 1000)
		}
	})

	b.Run("Mixed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				cache.Put(i%1000, "test_value")
			} else {
				cache.Get(i % 1000)
			}
		}
	})
}

// BenchmarkSpriteValidation benchmarks sprite validation performance
func BenchmarkSpriteValidation(b *testing.B) {
	sprite := spriteData{
		ID:     1,
		X:      0,
		Y:      0,
		Width:  8,
		Height: 8,
		Used:   true,
		Flags: FlagsData{
			Bitfield:   0,
			Individual: []bool{false, false, false, false, false, false, false, false},
		},
		Pixels: make([][]int, 8),
	}

	for i := range sprite.Pixels {
		sprite.Pixels[i] = make([]int, 8)
		for j := range sprite.Pixels[i] {
			sprite.Pixels[i][j] = (i + j) % 16
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sprite.Validate()
	}
}

// BenchmarkMetricsCollection benchmarks metrics collection performance
func BenchmarkMetricsCollection(b *testing.B) {
	b.Run("RecordSpriteRendered", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metricsCollector.RecordSpriteRendered()
		}
	})

	b.Run("RecordRenderTime", func(b *testing.B) {
		duration := 1 * time.Millisecond
		for i := 0; i < b.N; i++ {
			metricsCollector.RecordRenderTime(duration)
		}
	})

	b.Run("GetSpriteMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GetSpriteMetrics()
		}
	})
}

// BenchmarkConfigurationOperations benchmarks configuration changes
func BenchmarkConfigurationOperations(b *testing.B) {
	b.Run("SetDebugSpriteLogging", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SetDebugSpriteLogging(i%2 == 0)
		}
	})

	b.Run("GetRenderConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GetRenderConfig()
		}
	})
}

// StressTestConcurrentAccess tests concurrent access to sprite system
func TestStressTestConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Initialize system
	_ = SetDebugSpriteLogging(false) // Reduce log noise

	const numGoroutines = 10
	const operationsPerGoroutine = 1000

	done := make(chan bool, numGoroutines)

	// Start multiple goroutines performing different operations
	for i := 0; i < numGoroutines; i++ {
		go func(_ int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				switch j % 4 {
				case 0:
					// Configuration changes
					_ = SetDebugSpriteLogging(j%2 == 0)
				case 1:
					// Metrics collection
					metricsCollector.RecordSpriteRendered()
				case 2:
					// Cache operations
					initializeCaches()
					if spriteImageCache != nil {
						img := ebiten.NewImage(8, 8)
						spriteImageCache.Put(img, img)
					}
				case 3:
					// Hash operations
					pixels := [][]int{{1, 2}, {3, 4}}
					flags := FlagsData{Bitfield: 0, Individual: []bool{false, false, false, false, false, false, false, false}}
					generateOptimizedSpriteHashInternal(pixels, flags)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(30 * time.Second):
			t.Fatal("Stress test timed out")
		}
	}

	t.Logf("Stress test completed: %d goroutines × %d operations", numGoroutines, operationsPerGoroutine)
}

// TestMemoryLeakDetection tests for memory leaks in cache operations
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Initialize caches
	initializeCaches()

	// Record initial memory
	var initialMem, finalMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMem)

	// Perform many cache operations
	for i := 0; i < 10000; i++ {
		img := ebiten.NewImage(8, 8)
		spriteImageCache.Put(img, img)

		pixels := make([]byte, 64*4)
		spritePixelCacheManager.Put(i, pixels, 64)

		// Occasionally clear caches to test cleanup
		if i%1000 == 0 {
			ClearSpriteCache()
		}
	}

	// Force cleanup and measure final memory
	ClearSpriteCache()
	runtime.GC()
	runtime.ReadMemStats(&finalMem)

	memoryGrowth := int64(finalMem.Alloc) - int64(initialMem.Alloc)
	t.Logf("Memory growth: %d bytes", memoryGrowth)

	// Allow some memory growth but flag excessive growth
	if memoryGrowth > 50*1024*1024 { // 50MB threshold
		t.Errorf("Excessive memory growth detected: %d bytes", memoryGrowth)
	}
}

// TestResourceLimitEnforcement tests resource limit enforcement
func TestResourceLimitEnforcement(t *testing.T) {
	// Set strict limits
	limits := ResourceLimits{
		MaxCacheSize:      10,
		MaxSpriteCount:    100,
		ValidationTimeout: 1 * time.Second,
		MaxMemoryUsage:    1024 * 1024, // 1MB
	}

	err := SetResourceLimits(limits)
	if err != nil {
		t.Fatalf("Failed to set resource limits: %v", err)
	}

	// Test cache size enforcement
	initializeCaches()

	// Add more items than the limit
	for i := 0; i < 20; i++ {
		img := ebiten.NewImage(8, 8)
		spriteImageCache.Put(img, img)
	}

	stats := spriteImageCache.Stats()
	if stats.Size > limits.MaxCacheSize {
		t.Errorf("Cache size (%d) exceeds limit (%d)", stats.Size, limits.MaxCacheSize)
	}

	// Test resource usage checking
	violations := CheckResourceUsage()
	t.Logf("Resource violations: %v", violations)
}

// BenchmarkHashCollisionHandling benchmarks hash collision detection
func BenchmarkHashCollisionHandling(b *testing.B) {
	hashTable := NewSpriteHashTable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pixels := [][]int{{i % 16, (i + 1) % 16}, {(i + 2) % 16, (i + 3) % 16}}
		flags := FlagsData{
			Bitfield:   i % 256,
			Individual: []bool{i%2 == 0, false, false, false, false, false, false, false},
		}

		hashTable.AddEntry(pixels, flags, i)
	}
}

// BenchmarkSpriteRenderingHotPath benchmarks the critical sprite rendering path
func BenchmarkSpriteRenderingHotPath(b *testing.B) {
	// Initialize system
	initializeCaches()

	// Create test sprites
	testSprites := make([]spriteInfo, 10)
	for i := range testSprites {
		img := ebiten.NewImage(8, 8)
		testSprites[i] = spriteInfo{
			ID:    i,
			Image: img,
			Flags: FlagsData{Bitfield: 0, Individual: []bool{false, false, false, false, false, false, false, false}},
		}
	}
	currentSprites = testSprites

	b.Run("SpriteRenderLoop_NoMetrics", func(b *testing.B) {
		// Disable frame stats for pure performance test
		EnableFrameStats(false)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate typical sprite rendering loop
			spriteID := i % 10

			// This tests the core sprite lookup logic (hot path)
			spriteInfo := findSpriteByID(spriteID)
			if spriteInfo != nil {
				// Just test the lookup, not the GPU operations
				_ = spriteInfo.ID
			}
		}
	})

	b.Run("SpriteRenderLoop_WithMetrics", func(b *testing.B) {
		// Enable frame stats
		EnableFrameStats(true)
		frame := BeginSpriteFrame()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			spriteID := i % 10

			spriteInfo := findSpriteByID(spriteID)
			if spriteInfo != nil {
				recordSpriteRendered()
				// Just test the lookup with metrics, not GPU operations
				_ = spriteInfo.ID
			}
		}

		EndSpriteFrame(frame)
	})
}

// BenchmarkMemoryAllocationProfile profiles memory allocations in hot paths
func BenchmarkMemoryAllocationProfile(b *testing.B) {
	initializeCaches()

	b.Run("SpriteCreation_Allocations", func(b *testing.B) {
		b.ReportAllocs()

		// Test allocation patterns without GPU operations
		for i := 0; i < b.N; i++ {
			// Simulate sprite info creation
			info := spriteInfo{
				ID:    i,
				Image: nil, // Don't create actual images in benchmark
				Flags: FlagsData{Bitfield: 0, Individual: []bool{false, false, false, false, false, false, false, false}},
			}
			_ = info
		}
	})

	b.Run("CacheOperations_Allocations", func(b *testing.B) {
		b.ReportAllocs()

		// Test cache operations without GPU
		for i := 0; i < b.N; i++ {
			spriteID := i % 100
			pixels := make([]byte, 64*4) // 8x8 sprite

			if _, _, exists := spritePixelCacheManager.Get(spriteID); !exists {
				spritePixelCacheManager.Put(spriteID, pixels, 64)
			}
		}
	})
}
