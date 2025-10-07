//go:build !debug

package pigo8

// Release build optimizations - minimal overhead

// debugSpriteNotFound is a no-op in release builds
func debugSpriteNotFound(spriteID int, x, y float64) {
	// No-op for performance
}

// debugEnabled always returns false in release builds
func debugEnabled() bool {
	return false
}

// recordDebugMetrics is a no-op in release builds
func recordDebugMetrics() {
	// No-op for performance
}

// debugLog is a no-op in release builds
func debugLog(format string, args ...interface{}) {
	// No-op for performance
}

// validateSpriteInDebug is a no-op in release builds
func validateSpriteInDebug(spriteID int) bool {
	return true // Always valid in release
}
