//go:build debug

package pigo8

import (
	"log"
	"runtime"
)

// Debug build - full debugging capabilities

// debugSpriteNotFound logs detailed sprite miss information
func debugSpriteNotFound(spriteID int, x, y float64) {
	if isDebugEnabled() {
		log.Printf("Debug: Sprite %d not found in Spr() [Frame:%d, Caller:%s, Pos:(%.2f,%.2f)]",
			spriteID, GetCurrentFrameNumber(), getCallerName(), x, y)
	}
}

// debugEnabled returns the actual debug state
func debugEnabled() bool {
	return isDebugEnabled()
}

// recordDebugMetrics records debug-specific metrics
func recordDebugMetrics() {
	// Could add debug-specific metric collection here
}

// debugLog performs conditional debug logging
func debugLog(format string, args ...interface{}) {
	if isDebugEnabled() {
		log.Printf("Debug: "+format, args...)
	}
}

// validateSpriteInDebug performs additional validation in debug builds
func validateSpriteInDebug(spriteID int) bool {
	// Could add additional validation logic here
	return true
}

// getCallerName returns the actual calling function name
func getCallerName() string {
	if pc, _, _, ok := runtime.Caller(3); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			return fn.Name()
		}
	}
	return "unknown"
}
