package pigo8

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPixelBufferOperations(t *testing.T) {
	// Test pixel buffer initialization
	initPixelBuffer(128, 128)

	if pixelBufferWidth != 128 || pixelBufferHeight != 128 {
		t.Errorf("Pixel buffer dimensions incorrect: got %dx%d, want 128x128", pixelBufferWidth, pixelBufferHeight)
	}

	if len(pixelBuffer) != 128*128*4 {
		t.Errorf("Pixel buffer size incorrect: got %d, want %d", len(pixelBuffer), 128*128*4)
	}

	// Test setting pixels in buffer
	testColor := color.RGBA{255, 128, 64, 255}
	setPixelInBuffer(10, 20, testColor)

	if !bufferDirty {
		t.Error("Buffer should be marked as dirty after setting pixel")
	}

	// Test pixel values in buffer
	offset := (20*128 + 10) * 4
	if pixelBuffer[offset] != 255 || pixelBuffer[offset+1] != 128 || pixelBuffer[offset+2] != 64 || pixelBuffer[offset+3] != 255 {
		t.Errorf("Pixel values incorrect at offset %d: got [%d,%d,%d,%d], want [255,128,64,255]",
			offset, pixelBuffer[offset], pixelBuffer[offset+1], pixelBuffer[offset+2], pixelBuffer[offset+3])
	}
}

func TestSpriteModificationBatching(t *testing.T) {
	// Test sprite modification queuing
	testSprite := ebiten.NewImage(8, 8)
	testColor := color.RGBA{255, 0, 0, 255}

	queueSpriteModification(testSprite, 4, 4, testColor)

	spriteModMutex.Lock()
	mods := spriteModifications[testSprite]
	spriteModMutex.Unlock()

	if len(mods) != 1 {
		t.Errorf("Expected 1 modification, got %d", len(mods))
	}

	if mods[0].x != 4 || mods[0].y != 4 {
		t.Errorf("Modification coordinates incorrect: got (%d,%d), want (4,4)", mods[0].x, mods[0].y)
	}
}

func TestScreenPixelCacheOperations(t *testing.T) {
	// Test screen pixel cache initialization
	initScreenPixelCache(64, 64)

	if screenPixelCacheWidth != 64 || screenPixelCacheHeight != 64 {
		t.Errorf("Screen pixel cache dimensions incorrect: got %dx%d, want 64x64", screenPixelCacheWidth, screenPixelCacheHeight)
	}

	if len(screenPixelCache) != 64*64*4 {
		t.Errorf("Screen pixel cache size incorrect: got %d, want %d", len(screenPixelCache), 64*64*4)
	}

	// Test cache invalidation
	invalidateScreenPixelCache()
	if screenCacheValid {
		t.Error("Cache should be invalid after invalidation")
	}

	// Test cache statistics
	width, height, valid, size := GetScreenPixelCacheStats()
	if width != 64 || height != 64 {
		t.Errorf("Cache stats dimensions incorrect: got %dx%d, want 64x64", width, height)
	}
	if valid {
		t.Error("Cache should be invalid")
	}
	if size != 64*64*4 {
		t.Errorf("Cache stats size incorrect: got %d, want %d", size, 64*64*4)
	}
}

func TestSpritePixelCacheOperations(t *testing.T) {
	// Test sprite pixel cache operations
	testSprite := ebiten.NewImage(8, 8)
	spriteID := 42

	// Initialize caches first
	initializeCaches()

	// Initialize cache
	initSpritePixelCache(spriteID, testSprite)

	_, cacheSize, found := spritePixelCacheManager.Get(spriteID)

	if !found {
		t.Error("Sprite pixel cache should be initialized")
	}

	if cacheSize != 8*8*4 {
		t.Errorf("Sprite pixel cache size incorrect: got %d, want %d", cacheSize, 8*8*4)
	}

	// Test cache invalidation
	invalidateSpritePixelCache(spriteID)

	spritePixelCacheMutex.RLock()
	valid := spriteCacheValid[spriteID]
	spritePixelCacheMutex.RUnlock()

	if valid {
		t.Error("Sprite cache should be invalid after invalidation")
	}

	// Test cache clearing
	clearSpritePixelCache()

	initializeCaches()
	stats := spritePixelCacheManager.Stats()

	if stats.Size != 0 {
		t.Errorf("Sprite pixel cache should be empty after clearing, got %d entries", stats.Size)
	}

	// Test cache statistics
	totalSprites, validSprites, totalSize := GetSpritePixelCacheStats()
	if totalSprites != 0 || validSprites != 0 || totalSize != 0 {
		t.Errorf("Cache stats should be zero after clearing: got %d sprites, %d valid, %d size", totalSprites, validSprites, totalSize)
	}
}

func TestBatchReadingOptimizations(t *testing.T) {
	// Test that batch reading optimizations are properly integrated

	// Test screen pixel cache integration
	initPixelBuffer(32, 32)
	initScreenPixelCache(32, 32)

	// Test sprite pixel cache integration
	testSprite := ebiten.NewImage(8, 8)
	initSpritePixelCache(1, testSprite)

	// Verify cache systems are initialized
	if pixelBuffer == nil {
		t.Error("Pixel buffer should be initialized")
	}
	if screenPixelCache == nil {
		t.Error("Screen pixel cache should be initialized")
	}

	// Initialize caches first
	initializeCaches()

	// Initialize a specific sprite cache entry
	testSprite2 := ebiten.NewImage(8, 8)
	initSpritePixelCache(1, testSprite2)

	_, _, spriteCacheExists := spritePixelCacheManager.Get(1)

	if !spriteCacheExists {
		t.Error("Sprite pixel cache should be initialized")
	}
}

func TestSetScreenSize(t *testing.T) {
	// Test initial state
	initPixelBuffer(128, 128)

	// Test changing screen size
	setScreenSize(256, 192)

	if pixelBufferWidth != 256 || pixelBufferHeight != 192 {
		t.Errorf("Screen size change failed: got %dx%d, want 256x192", pixelBufferWidth, pixelBufferHeight)
	}

	if len(pixelBuffer) != 256*192*4 {
		t.Errorf("Pixel buffer size incorrect after resize: got %d, want %d", len(pixelBuffer), 256*192*4)
	}

	// Test that screenWidth and screenHeight were updated
	if screenWidth != 256 || screenHeight != 192 {
		t.Errorf("Global screen dimensions not updated: got %dx%d, want 256x192", screenWidth, screenHeight)
	}

	// Test setting a pixel in the new buffer
	testColor := color.RGBA{255, 255, 255, 255}
	setPixelInBuffer(100, 100, testColor)

	if !bufferDirty {
		t.Error("Buffer should be marked as dirty after setting pixel in new buffer")
	}
}
