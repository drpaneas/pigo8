package pigo8

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// PlayStation-Quality Performance Profiler
// ============================================================================
// Comprehensive performance metrics collection and analysis for game profiling.
// Designed for minimal overhead when disabled, maximum insight when enabled.
//
// Features:
// - Per-frame timing breakdown (Update, Draw, GPU sync)
// - Memory allocation tracking
// - Cache efficiency metrics
// - Draw call counting
// - Configurable sample windows for averaging
// ============================================================================

// ProfilerSection identifies which part of the frame is being profiled
type ProfilerSection int

const (
	ProfilerSectionUpdate ProfilerSection = iota
	ProfilerSectionDraw
	ProfilerSectionGPUSync
	ProfilerSectionPixelBatch
	ProfilerSectionSpriteRender
	ProfilerSectionMapRender
	ProfilerSectionShapeRender
	ProfilerSectionTextRender
	ProfilerSectionAudio
	ProfilerSectionTotal
)

func (ps ProfilerSection) String() string {
	names := []string{
		"Update", "Draw", "GPUSync", "PixelBatch", "SpriteRender",
		"MapRender", "ShapeRender", "TextRender", "Audio", "Total",
	}
	if int(ps) < len(names) {
		return names[ps]
	}
	return "Unknown"
}

// ProfilerSample represents a single timing sample
type ProfilerSample struct {
	Section  ProfilerSection
	Duration time.Duration
	Frame    int64
}

// ProfilerTimings holds timing data for all sections
type ProfilerTimings struct {
	Sections [ProfilerSectionTotal + 1]time.Duration
	Frame    int64
}

// PerformanceProfiler provides detailed performance analysis
type PerformanceProfiler struct {
	mu sync.RWMutex

	// Enabled state
	enabled atomic.Bool

	// Current frame tracking
	currentFrame     int64
	frameStartTime   time.Time
	sectionStartTime time.Time
	currentSection   ProfilerSection

	// Accumulated timings for current frame
	currentTimings ProfilerTimings

	// Historical data (ring buffer for rolling average)
	historySize   int
	history       []ProfilerTimings
	historyIndex  int
	historyFilled bool

	// Aggregate statistics
	totalFrames     int64
	totalTime       time.Duration
	longestFrame    time.Duration
	longestFrameNum int64

	// Draw call tracking
	drawCallsThisFrame int64
	totalDrawCalls     int64

	// Memory tracking
	lastMemStats     runtime.MemStats
	allocsThisFrame  uint64
	bytesThisFrame   uint64
	totalAllocs      uint64
	totalAllocBytes  uint64

	// GPU sync tracking
	gpuSyncCount     int64
	gpuSyncTimeTotal time.Duration

	// Cache metrics (per frame)
	cacheHitsThisFrame   int64
	cacheMissesThisFrame int64
}

// Global profiler
var (
	perfProfiler     *PerformanceProfiler
	perfProfilerOnce sync.Once
)

// GetPerformanceProfiler returns the singleton performance profiler
func GetPerformanceProfiler() *PerformanceProfiler {
	perfProfilerOnce.Do(func() {
		perfProfiler = &PerformanceProfiler{
			historySize: 120, // 4 seconds at 30fps, 2 seconds at 60fps
			history:     make([]ProfilerTimings, 120),
		}
	})
	return perfProfiler
}

// Enable enables or disables the profiler
func (pp *PerformanceProfiler) Enable(enabled bool) {
	pp.enabled.Store(enabled)
}

// IsEnabled returns whether the profiler is enabled
func (pp *PerformanceProfiler) IsEnabled() bool {
	return pp.enabled.Load()
}

// BeginFrame starts timing a new frame
func (pp *PerformanceProfiler) BeginFrame(frameNum int64) {
	if !pp.enabled.Load() {
		return
	}

	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.currentFrame = frameNum
	pp.frameStartTime = time.Now()
	pp.currentTimings = ProfilerTimings{Frame: frameNum}
	pp.drawCallsThisFrame = 0
	pp.cacheHitsThisFrame = 0
	pp.cacheMissesThisFrame = 0

	// Sample memory at frame start
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	pp.allocsThisFrame = m.Mallocs - pp.lastMemStats.Mallocs
	pp.bytesThisFrame = m.TotalAlloc - pp.lastMemStats.TotalAlloc
	pp.lastMemStats = m
}

// EndFrame finalizes timing for the current frame
func (pp *PerformanceProfiler) EndFrame() {
	if !pp.enabled.Load() {
		return
	}

	pp.mu.Lock()
	defer pp.mu.Unlock()

	// Calculate total frame time
	frameTime := time.Since(pp.frameStartTime)
	pp.currentTimings.Sections[ProfilerSectionTotal] = frameTime

	// Store in history
	pp.history[pp.historyIndex] = pp.currentTimings
	pp.historyIndex = (pp.historyIndex + 1) % pp.historySize
	if pp.historyIndex == 0 {
		pp.historyFilled = true
	}

	// Update aggregate stats
	pp.totalFrames++
	pp.totalTime += frameTime
	pp.totalDrawCalls += pp.drawCallsThisFrame
	pp.totalAllocs += pp.allocsThisFrame
	pp.totalAllocBytes += pp.bytesThisFrame

	if frameTime > pp.longestFrame {
		pp.longestFrame = frameTime
		pp.longestFrameNum = pp.currentFrame
	}
}

// BeginSection starts timing a specific section
func (pp *PerformanceProfiler) BeginSection(section ProfilerSection) {
	if !pp.enabled.Load() {
		return
	}

	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.currentSection = section
	pp.sectionStartTime = time.Now()
}

// EndSection finishes timing a specific section
func (pp *PerformanceProfiler) EndSection(section ProfilerSection) {
	if !pp.enabled.Load() {
		return
	}

	elapsed := time.Since(pp.sectionStartTime)

	pp.mu.Lock()
	defer pp.mu.Unlock()

	if section == pp.currentSection {
		pp.currentTimings.Sections[section] += elapsed
	}
}

// RecordDrawCall records a draw call
func (pp *PerformanceProfiler) RecordDrawCall() {
	if !pp.enabled.Load() {
		return
	}
	atomic.AddInt64(&pp.drawCallsThisFrame, 1)
}

// RecordGPUSync records a GPU synchronization event
func (pp *PerformanceProfiler) RecordGPUSync(duration time.Duration) {
	if !pp.enabled.Load() {
		return
	}
	atomic.AddInt64(&pp.gpuSyncCount, 1)
	pp.mu.Lock()
	pp.gpuSyncTimeTotal += duration
	pp.mu.Unlock()
}

// RecordCacheHit records a cache hit
func (pp *PerformanceProfiler) RecordCacheHit() {
	if !pp.enabled.Load() {
		return
	}
	atomic.AddInt64(&pp.cacheHitsThisFrame, 1)
}

// RecordCacheMiss records a cache miss
func (pp *PerformanceProfiler) RecordCacheMiss() {
	if !pp.enabled.Load() {
		return
	}
	atomic.AddInt64(&pp.cacheMissesThisFrame, 1)
}

// PerformanceReport holds a snapshot of profiler statistics
type PerformanceReport struct {
	// Frame timing
	AvgFrameTime    time.Duration `json:"avg_frame_time"`
	MaxFrameTime    time.Duration `json:"max_frame_time"`
	MinFrameTime    time.Duration `json:"min_frame_time"`
	CurrentFPS      float64       `json:"current_fps"`
	AvgFPS          float64       `json:"avg_fps"`
	TargetFPS       int           `json:"target_fps"`
	FrameBudget     time.Duration `json:"frame_budget"`
	BudgetUsedPct   float64       `json:"budget_used_pct"`

	// Section breakdown (percentage of frame)
	SectionBreakdown map[string]float64 `json:"section_breakdown"`

	// Draw calls
	DrawCallsPerFrame float64 `json:"draw_calls_per_frame"`
	TotalDrawCalls    int64   `json:"total_draw_calls"`

	// Memory
	AllocsPerFrame  float64 `json:"allocs_per_frame"`
	BytesPerFrame   float64 `json:"bytes_per_frame"`
	HeapAllocMB     float64 `json:"heap_alloc_mb"`
	HeapInUseMB     float64 `json:"heap_in_use_mb"`
	NumGC           uint32  `json:"num_gc"`

	// Cache efficiency
	CacheHitRate float64 `json:"cache_hit_rate"`

	// GPU sync
	GPUSyncCount int64         `json:"gpu_sync_count"`
	GPUSyncTime  time.Duration `json:"gpu_sync_time_total"`

	// Frame count
	TotalFrames int64 `json:"total_frames"`
}

// GetReport generates a performance report
func (pp *PerformanceProfiler) GetReport() PerformanceReport {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	report := PerformanceReport{
		TotalFrames:      pp.totalFrames,
		TotalDrawCalls:   pp.totalDrawCalls,
		GPUSyncCount:     pp.gpuSyncCount,
		GPUSyncTime:      pp.gpuSyncTimeTotal,
		SectionBreakdown: make(map[string]float64),
	}

	if pp.totalFrames == 0 {
		return report
	}

	// Calculate averages from history
	historyLen := pp.historySize
	if !pp.historyFilled {
		historyLen = pp.historyIndex
	}
	if historyLen == 0 {
		historyLen = 1
	}

	var totalFrameTime time.Duration
	var minFrame, maxFrame time.Duration
	var sectionTotals [ProfilerSectionTotal + 1]time.Duration

	first := true
	for i := 0; i < historyLen; i++ {
		h := pp.history[i]
		ft := h.Sections[ProfilerSectionTotal]
		totalFrameTime += ft

		if first || ft < minFrame {
			minFrame = ft
		}
		if first || ft > maxFrame {
			maxFrame = ft
		}
		first = false

		for s := ProfilerSection(0); s <= ProfilerSectionTotal; s++ {
			sectionTotals[s] += h.Sections[s]
		}
	}

	report.AvgFrameTime = totalFrameTime / time.Duration(historyLen)
	report.MinFrameTime = minFrame
	report.MaxFrameTime = maxFrame

	if report.AvgFrameTime > 0 {
		report.CurrentFPS = float64(time.Second) / float64(report.AvgFrameTime)
	}

	if pp.totalTime > 0 {
		report.AvgFPS = float64(pp.totalFrames) * float64(time.Second) / float64(pp.totalTime)
	}

	// Section breakdown as percentage
	totalSectionTime := sectionTotals[ProfilerSectionTotal]
	if totalSectionTime > 0 {
		for s := ProfilerSection(0); s < ProfilerSectionTotal; s++ {
			pct := float64(sectionTotals[s]) / float64(totalSectionTime) * 100
			report.SectionBreakdown[s.String()] = pct
		}
	}

	// Draw calls per frame
	report.DrawCallsPerFrame = float64(pp.totalDrawCalls) / float64(pp.totalFrames)

	// Memory stats
	report.AllocsPerFrame = float64(pp.totalAllocs) / float64(pp.totalFrames)
	report.BytesPerFrame = float64(pp.totalAllocBytes) / float64(pp.totalFrames)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	report.HeapAllocMB = float64(m.HeapAlloc) / 1024 / 1024
	report.HeapInUseMB = float64(m.HeapInuse) / 1024 / 1024
	report.NumGC = m.NumGC

	// Cache hit rate
	totalCacheOps := pp.cacheHitsThisFrame + pp.cacheMissesThisFrame
	if totalCacheOps > 0 {
		report.CacheHitRate = float64(pp.cacheHitsThisFrame) / float64(totalCacheOps) * 100
	}

	return report
}

// FormatReport returns a human-readable performance report string
func (pp *PerformanceProfiler) FormatReport() string {
	r := pp.GetReport()

	return fmt.Sprintf(`=== PIGO-8 Performance Report ===
Frames: %d | FPS: %.1f (avg: %.1f)
Frame Time: avg=%.2fms min=%.2fms max=%.2fms

Section Breakdown:
  Update:  %5.1f%%
  Draw:    %5.1f%%
  Sprites: %5.1f%%
  Map:     %5.1f%%
  Shapes:  %5.1f%%
  Pixels:  %5.1f%%
  GPU:     %5.1f%%

Draw Calls: %.1f/frame | Total: %d
Memory: %.2f MB heap | %.0f allocs/frame | %.0f bytes/frame
Cache Hit Rate: %.1f%%
GPU Syncs: %d (%.2fms total)
`,
		r.TotalFrames, r.CurrentFPS, r.AvgFPS,
		float64(r.AvgFrameTime)/float64(time.Millisecond),
		float64(r.MinFrameTime)/float64(time.Millisecond),
		float64(r.MaxFrameTime)/float64(time.Millisecond),
		r.SectionBreakdown["Update"],
		r.SectionBreakdown["Draw"],
		r.SectionBreakdown["SpriteRender"],
		r.SectionBreakdown["MapRender"],
		r.SectionBreakdown["ShapeRender"],
		r.SectionBreakdown["PixelBatch"],
		r.SectionBreakdown["GPUSync"],
		r.DrawCallsPerFrame, r.TotalDrawCalls,
		r.HeapAllocMB, r.AllocsPerFrame, r.BytesPerFrame,
		r.CacheHitRate,
		r.GPUSyncCount, float64(r.GPUSyncTime)/float64(time.Millisecond),
	)
}

// Reset clears all profiler statistics
func (pp *PerformanceProfiler) Reset() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.currentFrame = 0
	pp.historyIndex = 0
	pp.historyFilled = false
	pp.totalFrames = 0
	pp.totalTime = 0
	pp.longestFrame = 0
	pp.longestFrameNum = 0
	pp.drawCallsThisFrame = 0
	pp.totalDrawCalls = 0
	pp.allocsThisFrame = 0
	pp.bytesThisFrame = 0
	pp.totalAllocs = 0
	pp.totalAllocBytes = 0
	pp.gpuSyncCount = 0
	pp.gpuSyncTimeTotal = 0
	pp.cacheHitsThisFrame = 0
	pp.cacheMissesThisFrame = 0

	// Clear history
	for i := range pp.history {
		pp.history[i] = ProfilerTimings{}
	}
}

// ============================================================================
// Convenience Profiling Functions
// ============================================================================

// ProfileSection is a helper for profiling a code section with defer
func ProfileSection(section ProfilerSection) func() {
	pp := GetPerformanceProfiler()
	pp.BeginSection(section)
	return func() {
		pp.EndSection(section)
	}
}

// ProfileFrame is a helper for profiling a complete frame
func ProfileFrame(frameNum int64) func() {
	pp := GetPerformanceProfiler()
	pp.BeginFrame(frameNum)
	return func() {
		pp.EndFrame()
	}
}

// EnableProfiling enables performance profiling
func EnableProfiling(enabled bool) {
	GetPerformanceProfiler().Enable(enabled)
}

// GetProfilingReport returns the current performance report
func GetProfilingReport() PerformanceReport {
	return GetPerformanceProfiler().GetReport()
}

// PrintProfilingReport prints the performance report to log
func PrintProfilingReport() {
	fmt.Println(GetPerformanceProfiler().FormatReport())
}

