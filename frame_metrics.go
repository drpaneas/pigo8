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

// Global frame tracking
var (
	currentFrame     *FrameStats
	frameCounter     int64
	frameStatsActive bool = false
)

// BeginSpriteFrame starts frame-level metrics collection
// This should be called once per frame, not per sprite
func BeginSpriteFrame() *FrameStats {
	if !frameStatsActive {
		return nil
	}

	frameNum := atomic.AddInt64(&frameCounter, 1)
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

	// Record frame-level metrics to global collector
	metricsCollector.RecordFrameMetrics(frame, frameTime)

	currentFrame = nil
}

// Frame-level sprite counting (no timing overhead)
func recordSpriteRendered() {
	if currentFrame != nil {
		atomic.AddInt64(&currentFrame.SpritesRendered, 1)
	}
}

func recordCacheHit() {
	if currentFrame != nil {
		atomic.AddInt64(&currentFrame.CacheHits, 1)
	}
}

func recordCacheMiss() {
	if currentFrame != nil {
		atomic.AddInt64(&currentFrame.CacheMisses, 1)
	}
}

func recordValidationError() {
	if currentFrame != nil {
		atomic.AddInt64(&currentFrame.ValidationErrors, 1)
	}
}

func recordSpriteSkipped() {
	if currentFrame != nil {
		atomic.AddInt64(&currentFrame.SpritesSkipped, 1)
	}
}

func recordSpriteLoaded() {
	if currentFrame != nil {
		atomic.AddInt64(&currentFrame.SpritesLoaded, 1)
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
