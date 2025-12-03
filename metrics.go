package pigo8

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ===== Optimization 8: Batched Per-Frame Metrics =====
// See frame_metrics.go for the per-frame local counters.
// The record* functions there use atomic ops only when frameStatsActive is true,
// avoiding overhead when frame stats are disabled.

// SpriteMetrics holds comprehensive sprite system performance metrics
type SpriteMetrics struct {
	// Cache performance
	ImageCacheStats CacheStats `json:"image_cache"`
	PixelCacheStats CacheStats `json:"pixel_cache"`

	// Sprite operations
	SpritesRendered   int64   `json:"sprites_rendered"`
	SpritesLoaded     int64   `json:"sprites_loaded"`
	SpritesSkipped    int64   `json:"sprites_skipped"`
	DeduplicationRate float64 `json:"deduplication_rate"`

	// Validation metrics
	ValidationTime    time.Duration `json:"validation_time"`
	ValidationErrors  int64         `json:"validation_errors"`
	StrictValidations int64         `json:"strict_validations"`

	// Hash performance
	HashCollisions   int64         `json:"hash_collisions"`
	HashComputeTime  time.Duration `json:"hash_compute_time"`
	HashComputeCount int64         `json:"hash_compute_count"`

	// Memory usage
	MemoryUsage int64 `json:"memory_usage_bytes"`
	SpriteCount int   `json:"sprite_count"`

	// Performance timing
	AverageRenderTime time.Duration `json:"average_render_time"`
	AverageLoadTime   time.Duration `json:"average_load_time"`

	// System info
	LastUpdated time.Time `json:"last_updated"`
}

// MetricsCollector manages sprite system metrics
type MetricsCollector struct {
	// Atomic counters for thread-safe updates
	spritesRendered   int64
	spritesLoaded     int64
	spritesSkipped    int64
	validationErrors  int64
	strictValidations int64
	hashCollisions    int64
	hashComputeCount  int64

	// Cache performance counters (accumulated from per-frame metrics)
	cacheHits   int64
	cacheMisses int64

	// Timing accumulators
	totalValidationTime time.Duration
	totalHashTime       time.Duration
	totalRenderTime     time.Duration
	totalLoadTime       time.Duration

	// Counters for averages
	renderCount int64
	loadCount   int64

	mutex sync.RWMutex
}

// Global metrics collector
var metricsCollector = &MetricsCollector{}

// RecordSpriteRendered increments the sprites rendered counter
// Note: For hot path, use recordSpriteRendered() (local counter) + FlushMetrics()
func (m *MetricsCollector) RecordSpriteRendered() {
	atomic.AddInt64(&m.spritesRendered, 1)
}

// RecordSpriteLoaded increments the sprites loaded counter
func (m *MetricsCollector) RecordSpriteLoaded() {
	atomic.AddInt64(&m.spritesLoaded, 1)
}

// RecordSpriteSkipped increments the sprites skipped counter
func (m *MetricsCollector) RecordSpriteSkipped() {
	atomic.AddInt64(&m.spritesSkipped, 1)
}

// RecordValidationError increments the validation error counter
func (m *MetricsCollector) RecordValidationError() {
	atomic.AddInt64(&m.validationErrors, 1)
}

// RecordStrictValidation increments the strict validation counter
func (m *MetricsCollector) RecordStrictValidation() {
	atomic.AddInt64(&m.strictValidations, 1)
}

// RecordHashCollision increments the hash collision counter
func (m *MetricsCollector) RecordHashCollision() {
	atomic.AddInt64(&m.hashCollisions, 1)
}

// RecordValidationTime adds to the total validation time
func (m *MetricsCollector) RecordValidationTime(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalValidationTime += duration
}

// RecordHashTime adds to the total hash computation time
// Note: This is now a no-op since we removed time.Now() from hash generation
func (m *MetricsCollector) RecordHashTime(_ time.Duration) {
	// Intentionally empty - hash timing removed for performance
	// Keep method for API compatibility
}

// RecordRenderTime adds to the total render time
func (m *MetricsCollector) RecordRenderTime(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalRenderTime += duration
	atomic.AddInt64(&m.renderCount, 1)
}

// RecordFrameMetrics records frame-level metrics (batch operation)
func (m *MetricsCollector) RecordFrameMetrics(frame *FrameStats, frameTime time.Duration) {
	if frame == nil {
		return
	}

	// Batch update all counters
	atomic.AddInt64(&m.spritesRendered, frame.SpritesRendered)
	atomic.AddInt64(&m.spritesLoaded, frame.SpritesLoaded)
	atomic.AddInt64(&m.spritesSkipped, frame.SpritesSkipped)
	atomic.AddInt64(&m.validationErrors, frame.ValidationErrors)
	atomic.AddInt64(&m.cacheHits, frame.CacheHits)
	atomic.AddInt64(&m.cacheMisses, frame.CacheMisses)

	// Record frame timing
	m.mutex.Lock()
	m.totalRenderTime += frameTime
	atomic.AddInt64(&m.renderCount, 1)
	m.mutex.Unlock()
}

// RecordLoadTime adds to the total load time
func (m *MetricsCollector) RecordLoadTime(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalLoadTime += duration
	atomic.AddInt64(&m.loadCount, 1)
}

// GetSpriteMetrics returns comprehensive sprite system metrics
func GetSpriteMetrics() SpriteMetrics {
	metricsCollector.mutex.RLock()
	defer metricsCollector.mutex.RUnlock()

	// Get cache stats
	var imageCacheStats, pixelCacheStats CacheStats
	if spriteImageCache != nil {
		imageCacheStats = spriteImageCache.Stats()
	}
	if spritePixelCacheManager != nil {
		pixelCacheStats = spritePixelCacheManager.Stats()
	}

	// Calculate averages
	var avgRenderTime, avgLoadTime time.Duration

	renderCount := atomic.LoadInt64(&metricsCollector.renderCount)
	if renderCount > 0 {
		avgRenderTime = metricsCollector.totalRenderTime / time.Duration(renderCount)
	}

	loadCount := atomic.LoadInt64(&metricsCollector.loadCount)
	if loadCount > 0 {
		avgLoadTime = metricsCollector.totalLoadTime / time.Duration(loadCount)
	}

	hashCount := atomic.LoadInt64(&metricsCollector.hashComputeCount)

	// Calculate deduplication rate
	loaded := atomic.LoadInt64(&metricsCollector.spritesLoaded)
	skipped := atomic.LoadInt64(&metricsCollector.spritesSkipped)
	deduplicationRate := float64(0)
	if loaded+skipped > 0 {
		deduplicationRate = float64(skipped) / float64(loaded+skipped)
	}

	// Get memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SpriteMetrics{
		ImageCacheStats:   imageCacheStats,
		PixelCacheStats:   pixelCacheStats,
		SpritesRendered:   atomic.LoadInt64(&metricsCollector.spritesRendered),
		SpritesLoaded:     loaded,
		SpritesSkipped:    skipped,
		DeduplicationRate: deduplicationRate,
		ValidationTime:    metricsCollector.totalValidationTime,
		ValidationErrors:  atomic.LoadInt64(&metricsCollector.validationErrors),
		StrictValidations: atomic.LoadInt64(&metricsCollector.strictValidations),
		HashCollisions:    atomic.LoadInt64(&metricsCollector.hashCollisions),
		HashComputeTime:   0, // No longer tracked for performance
		HashComputeCount:  hashCount,
		MemoryUsage:       int64(memStats.Alloc),
		SpriteCount:       len(currentSprites),
		AverageRenderTime: avgRenderTime,
		AverageLoadTime:   avgLoadTime,
		LastUpdated:       time.Now(),
	}
}

// ResetSpriteMetrics resets all sprite metrics to zero
func ResetSpriteMetrics() {
	metricsCollector.mutex.Lock()
	defer metricsCollector.mutex.Unlock()

	atomic.StoreInt64(&metricsCollector.spritesRendered, 0)
	atomic.StoreInt64(&metricsCollector.spritesLoaded, 0)
	atomic.StoreInt64(&metricsCollector.spritesSkipped, 0)
	atomic.StoreInt64(&metricsCollector.validationErrors, 0)
	atomic.StoreInt64(&metricsCollector.strictValidations, 0)
	atomic.StoreInt64(&metricsCollector.hashCollisions, 0)
	atomic.StoreInt64(&metricsCollector.hashComputeCount, 0)
	atomic.StoreInt64(&metricsCollector.cacheHits, 0)
	atomic.StoreInt64(&metricsCollector.cacheMisses, 0)
	atomic.StoreInt64(&metricsCollector.renderCount, 0)
	atomic.StoreInt64(&metricsCollector.loadCount, 0)

	metricsCollector.totalValidationTime = 0
	metricsCollector.totalHashTime = 0
	metricsCollector.totalRenderTime = 0
	metricsCollector.totalLoadTime = 0

	// Also reset local frame counters
	localCounters = localFrameCounters{}
}

// ResourceLimits defines resource usage limits for the sprite system
type ResourceLimits struct {
	MaxCacheSize      int           `json:"max_cache_size"`
	MaxSpriteCount    int           `json:"max_sprite_count"`
	ValidationTimeout time.Duration `json:"validation_timeout"`
	MaxMemoryUsage    int64         `json:"max_memory_usage_bytes"`
}

// DefaultResourceLimits returns sensible default resource limits
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		MaxCacheSize:      1000,  // 1000 cached sprites
		MaxSpriteCount:    10000, // 10000 total sprites
		ValidationTimeout: 5 * time.Second,
		MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
	}
}

// CheckResourceLimits checks if current usage exceeds limits
func CheckResourceLimits(limits ResourceLimits) []string {
	var violations []string
	metrics := GetSpriteMetrics()

	if metrics.ImageCacheStats.Size > limits.MaxCacheSize {
		violations = append(violations, "Image cache size exceeds limit")
	}

	if metrics.SpriteCount > limits.MaxSpriteCount {
		violations = append(violations, "Sprite count exceeds limit")
	}

	if metrics.MemoryUsage > limits.MaxMemoryUsage {
		violations = append(violations, "Memory usage exceeds limit")
	}

	return violations
}
