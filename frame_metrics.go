package pigo8

import (
	"sync/atomic"
	"time"
)

// FrameStats holds per-frame performance statistics
type FrameStats struct {
	SpritesRendered  int64
	CacheMisses      int64
	CacheHits        int64
	ValidationErrors int64
	SpritesSkipped   int64
	SpritesLoaded    int64
	StartTime        time.Time
	EndTime          time.Time
	FrameNumber      int64
}

// ===== Optimization 8: Local counters for per-frame accumulation =====
// These avoid atomic ops in the hot path (single-threaded Draw)

// localFrameCounters holds per-frame counters (not atomic - single threaded in Draw)
type localFrameCounters struct {
	spritesRendered  int64
	cacheHits        int64
	cacheMisses      int64
	validationErrors int64
	spritesSkipped   int64
	spritesLoaded    int64
}

// localCounters is the thread-local metrics for the current frame
var localCounters localFrameCounters

// Global frame tracking
var (
	currentFrame     *FrameStats
	frameCounter     int64
	frameStatsActive = false
)

// BeginSpriteFrame starts frame-level metrics collection
// This should be called once per frame, not per sprite
func BeginSpriteFrame() *FrameStats {
	if !frameStatsActive {
		return nil
	}

	// Use current frame number (already incremented by IncrementFrameCounter in Update())
	frameNum := atomic.LoadInt64(&frameCounter)
	currentFrame = &FrameStats{
		StartTime:   time.Now(),
		FrameNumber: frameNum,
	}
	return currentFrame
}

// EndSpriteFrame completes frame-level metrics and records to global collector
func EndSpriteFrame(frame *FrameStats) {
	if frame == nil || !frameStatsActive {
		return
	}

	frame.EndTime = time.Now()
	frameTime := frame.EndTime.Sub(frame.StartTime)

	// Note: frame stats are already populated by record* functions when currentFrame is active
	// Reset local counters
	localCounters = localFrameCounters{}

	// Record frame-level metrics to global collector
	metricsCollector.RecordFrameMetrics(frame, frameTime)

	currentFrame = nil
}

// Frame-level sprite counting - Optimization 8: No atomic ops in hot path
// These just increment local counters; atomics happen once per frame in EndSpriteFrame/FlushFrameMetrics
// When currentFrame is active, we also update it for backward compatibility with tests

func recordSpriteRendered() {
	localCounters.spritesRendered++
	// Also update currentFrame for backward compatibility
	if currentFrame != nil {
		currentFrame.SpritesRendered++
	}
}

func recordCacheHit() {
	localCounters.cacheHits++
	if currentFrame != nil {
		currentFrame.CacheHits++
	}
}

func recordCacheMiss() {
	localCounters.cacheMisses++
	if currentFrame != nil {
		currentFrame.CacheMisses++
	}
}

func recordValidationError() {
	localCounters.validationErrors++
	if currentFrame != nil {
		currentFrame.ValidationErrors++
	}
}

func recordSpriteSkipped() {
	localCounters.spritesSkipped++
	if currentFrame != nil {
		currentFrame.SpritesSkipped++
	}
}

func recordSpriteLoaded() {
	localCounters.spritesLoaded++
	if currentFrame != nil {
		currentFrame.SpritesLoaded++
	}
}

// EnableFrameStats enables frame-level statistics collection
func EnableFrameStats(enabled bool) {
	frameStatsActive = enabled
}

// IsFrameStatsEnabled returns whether frame stats are enabled
func IsFrameStatsEnabled() bool {
	return frameStatsActive
}

// GetCurrentFrameNumber returns the current frame number
func GetCurrentFrameNumber() int64 {
	return atomic.LoadInt64(&frameCounter)
}

// IncrementFrameCounter increments the frame counter and returns the new value.
// This should be called once per frame in Update(), regardless of whether
// frame stats are enabled, to ensure the LRU cache has valid frame ticks.
func IncrementFrameCounter() int64 {
	return atomic.AddInt64(&frameCounter, 1)
}

// FlushFrameMetrics flushes local frame metrics to global counters.
// Call this at the end of Draw() if not using BeginSpriteFrame/EndSpriteFrame.
func FlushFrameMetrics() {
	if localCounters.spritesRendered > 0 {
		atomic.AddInt64(&metricsCollector.spritesRendered, localCounters.spritesRendered)
	}
	if localCounters.cacheHits > 0 {
		atomic.AddInt64(&metricsCollector.cacheHits, localCounters.cacheHits)
	}
	if localCounters.cacheMisses > 0 {
		atomic.AddInt64(&metricsCollector.cacheMisses, localCounters.cacheMisses)
	}
	if localCounters.validationErrors > 0 {
		atomic.AddInt64(&metricsCollector.validationErrors, localCounters.validationErrors)
	}
	if localCounters.spritesSkipped > 0 {
		atomic.AddInt64(&metricsCollector.spritesSkipped, localCounters.spritesSkipped)
	}
	if localCounters.spritesLoaded > 0 {
		atomic.AddInt64(&metricsCollector.spritesLoaded, localCounters.spritesLoaded)
	}

	// Reset local counters
	localCounters = localFrameCounters{}
}
