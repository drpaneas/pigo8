package pigo8

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Resource lifecycle management: tracking, disposal, and memory pressure handling.

// ResourceType identifies the type of tracked resource
type ResourceType int

// Resource types for lifecycle tracking.
const (
	ResourceTypeImage ResourceType = iota
	ResourceTypeShader
	ResourceTypeAudio
	ResourceTypeBuffer
)

func (rt ResourceType) String() string {
	switch rt {
	case ResourceTypeImage:
		return "Image"
	case ResourceTypeShader:
		return "Shader"
	case ResourceTypeAudio:
		return "Audio"
	case ResourceTypeBuffer:
		return "Buffer"
	default:
		return "Unknown"
	}
}

// TrackedResource represents a resource being tracked for lifecycle management
type TrackedResource struct {
	ID        int64
	Type      ResourceType
	Size      int64 // Estimated size in bytes
	CreatedAt time.Time
	Name      string // Optional debug name
	disposed  bool
}

// ResourceManager handles lifecycle of GPU and memory resources
type ResourceManager struct {
	mu sync.RWMutex

	// Resource tracking
	resources map[int64]*TrackedResource
	nextID    int64

	// Memory tracking
	totalAllocated int64
	peakAllocated  int64

	// Thresholds
	memoryWarningThreshold int64 // Bytes - warn when exceeded
	memoryLimitThreshold   int64 // Bytes - force cleanup when exceeded

	// Stats
	allocations   int64
	deallocations int64
	cleanupRuns   int64

	// Shutdown state
	shuttingDown bool
}

// Global resource manager
var (
	resourceManager     *ResourceManager
	resourceManagerOnce sync.Once
)

// GetResourceManager returns the singleton resource manager
func GetResourceManager() *ResourceManager {
	resourceManagerOnce.Do(func() {
		resourceManager = &ResourceManager{
			resources:              make(map[int64]*TrackedResource, 256),
			memoryWarningThreshold: 64 * 1024 * 1024,  // 64 MB warning
			memoryLimitThreshold:   128 * 1024 * 1024, // 128 MB force cleanup
		}
	})
	return resourceManager
}

// Track registers a resource for lifecycle tracking
func (rm *ResourceManager) Track(rtype ResourceType, size int64, name string) int64 {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.shuttingDown {
		return -1
	}

	rm.nextID++
	id := rm.nextID

	rm.resources[id] = &TrackedResource{
		ID:        id,
		Type:      rtype,
		Size:      size,
		CreatedAt: time.Now(),
		Name:      name,
		disposed:  false,
	}

	rm.totalAllocated += size
	if rm.totalAllocated > rm.peakAllocated {
		rm.peakAllocated = rm.totalAllocated
	}

	atomic.AddInt64(&rm.allocations, 1)

	// Check memory pressure
	if rm.totalAllocated > rm.memoryWarningThreshold {
		log.Printf("ResourceManager: Memory warning - %d MB allocated (peak: %d MB)",
			rm.totalAllocated/1024/1024, rm.peakAllocated/1024/1024)
	}

	return id
}

// Release marks a resource as disposed and updates tracking
func (rm *ResourceManager) Release(id int64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	res, exists := rm.resources[id]
	if !exists || res.disposed {
		return
	}

	res.disposed = true
	rm.totalAllocated -= res.Size
	delete(rm.resources, id)

	atomic.AddInt64(&rm.deallocations, 1)
}

// ForceCleanup triggers aggressive resource cleanup
func (rm *ResourceManager) ForceCleanup() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	atomic.AddInt64(&rm.cleanupRuns, 1)

	// Clear all caches
	ClearAllCaches()
	ClearSsprCache()
	ClearFlagCache()

	// Suggest GC run
	runtime.GC()

	log.Printf("ResourceManager: Force cleanup completed. Allocated: %d MB",
		rm.totalAllocated/1024/1024)
}

// CheckMemoryPressure returns true if memory usage is above warning threshold
func (rm *ResourceManager) CheckMemoryPressure() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.totalAllocated > rm.memoryWarningThreshold
}

// Stats returns resource manager statistics
func (rm *ResourceManager) Stats() ResourceStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return ResourceStats{
		TotalAllocated:  rm.totalAllocated,
		PeakAllocated:   rm.peakAllocated,
		ActiveResources: len(rm.resources),
		Allocations:     atomic.LoadInt64(&rm.allocations),
		Deallocations:   atomic.LoadInt64(&rm.deallocations),
		CleanupRuns:     atomic.LoadInt64(&rm.cleanupRuns),
	}
}

// ResourceStats holds resource manager statistics
type ResourceStats struct {
	TotalAllocated  int64 `json:"total_allocated"`
	PeakAllocated   int64 `json:"peak_allocated"`
	ActiveResources int   `json:"active_resources"`
	Allocations     int64 `json:"allocations"`
	Deallocations   int64 `json:"deallocations"`
	CleanupRuns     int64 `json:"cleanup_runs"`
}

// Shutdown prepares the resource manager for application exit
func (rm *ResourceManager) Shutdown() {
	rm.mu.Lock()
	rm.shuttingDown = true
	rm.mu.Unlock()

	// Log any leaked resources in debug builds
	rm.mu.RLock()
	if len(rm.resources) > 0 {
		log.Printf("ResourceManager: Shutdown with %d resources still tracked:", len(rm.resources))
		for _, res := range rm.resources {
			if !res.disposed {
				log.Printf("  - %s[%d] %s: %d bytes (created: %v ago)",
					res.Type, res.ID, res.Name, res.Size, time.Since(res.CreatedAt))
			}
		}
	}
	rm.mu.RUnlock()
}

// ============================================================================
// Thread-Safe Palette System
// ============================================================================
// Copy-on-write pattern for palette access to prevent data races.
// ============================================================================

// PaletteSnapshot is an immutable copy of palette state
type PaletteSnapshot struct {
	Colors       []ColorRGBA
	Transparency []bool
	DrawMap      []int
	Version      int64
}

// ColorRGBA is a simple RGBA color struct for copy-on-write
type ColorRGBA struct {
	R, G, B, A uint8
}

// Global palette version for change detection
var (
	paletteVersion atomic.Int64
)

// GetPaletteSnapshot returns a thread-safe snapshot of the current palette
// This is safe to use from any goroutine (e.g., for collision detection in Update())
func GetPaletteSnapshot() *PaletteSnapshot {
	// Read under lock
	colorToIndexMapMutex.RLock()
	defer colorToIndexMapMutex.RUnlock()

	snapshot := &PaletteSnapshot{
		Colors:       make([]ColorRGBA, len(pico8Palette)),
		Transparency: make([]bool, len(paletteTransparency)),
		DrawMap:      make([]int, len(drawPaletteMap)),
		Version:      paletteVersion.Load(),
	}

	// Copy colors
	for i, c := range pico8Palette {
		r, g, b, a := c.RGBA()
		snapshot.Colors[i] = ColorRGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}
	}

	// Copy transparency
	copy(snapshot.Transparency, paletteTransparency)

	// Copy draw map
	copy(snapshot.DrawMap, drawPaletteMap)

	return snapshot
}

// IsPaletteChanged checks if palette has changed since a given version
func IsPaletteChanged(version int64) bool {
	return paletteVersion.Load() != version
}

// NotifyPaletteChanged should be called when palette is modified
// This is called internally by SetPalette, SetPaletteColor, Pal, Palt
func NotifyPaletteChanged() {
	paletteVersion.Add(1)
}

// ============================================================================
// Disposable Image Wrapper
// ============================================================================
// Wraps ebiten.Image with lifecycle tracking for proper resource management.
// ============================================================================

// DisposableImage wraps an ebiten.Image with lifecycle tracking
type DisposableImage struct {
	*ebiten.Image
	resourceID int64
	disposed   bool
	mu         sync.RWMutex
}

// NewDisposableImage creates a tracked image
func NewDisposableImage(width, height int, name string) *DisposableImage {
	img := ebiten.NewImage(width, height)
	size := int64(width * height * 4) // RGBA bytes

	return &DisposableImage{
		Image:      img,
		resourceID: GetResourceManager().Track(ResourceTypeImage, size, name),
		disposed:   false,
	}
}

// Dispose releases the image resources
func (di *DisposableImage) Dispose() {
	di.mu.Lock()
	defer di.mu.Unlock()

	if di.disposed {
		return
	}

	di.disposed = true
	di.Deallocate()
	GetResourceManager().Release(di.resourceID)
}

// IsDisposed returns true if the image has been disposed
func (di *DisposableImage) IsDisposed() bool {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.disposed
}

// ============================================================================
// Automatic Memory Pressure Handler
// ============================================================================
// Periodically checks memory and triggers cleanup when needed.
// ============================================================================

var (
	memoryWatcherRunning atomic.Bool
	memoryWatcherStop    chan struct{}
)

// StartMemoryWatcher starts background memory pressure monitoring
func StartMemoryWatcher(checkInterval time.Duration) {
	if !memoryWatcherRunning.CompareAndSwap(false, true) {
		return // Already running
	}

	memoryWatcherStop = make(chan struct{})

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rm := GetResourceManager()
				if rm.CheckMemoryPressure() {
					log.Println("MemoryWatcher: High memory pressure detected, running cleanup...")
					rm.ForceCleanup()
				}
			case <-memoryWatcherStop:
				memoryWatcherRunning.Store(false)
				return
			}
		}
	}()
}

// StopMemoryWatcher stops the background memory watcher
func StopMemoryWatcher() {
	if memoryWatcherRunning.Load() {
		close(memoryWatcherStop)
	}
}
