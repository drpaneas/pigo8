//go:build !debug

package pigo8

// Release build optimizations - minimal overhead

// debugSpriteNotFound is a no-op in release builds
func debugSpriteNotFound(_ int, _, _ float64) {
	// No-op for performance
}

// debugLog is a no-op in release builds
func debugLog(_ string, _ ...interface{}) {
	// No-op for performance
}
