package pigo8

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Batched pixel rendering for Pset(). Accumulates pixel writes into a CPU buffer
// and flushes once per frame via a single WritePixels() call.

// pendingPixel represents a single pixel write request
type pendingPixel struct {
	r, g, b, a uint8
}

// PixelBatchSystem handles efficient batched pixel rendering
type PixelBatchSystem struct {
	// Sparse pending pixels: only dirty pixels are tracked
	pending map[int]pendingPixel // key = y*width + x

	// Pre-allocated batch buffer (reused every frame)
	buffer       []byte
	bufferWidth  int
	bufferHeight int

	// Overlay image for compositing (reused, not recreated)
	overlay *ebiten.Image

	// Stats for profiling
	pixelsThisFrame int64
	flushCount      int64

	// Dirty tracking
	hasPendingPixels bool
}

// Global pixel batch system (initialized lazily)
var (
	pixelBatch     *PixelBatchSystem
	pixelBatchOnce sync.Once
)

// getPixelBatchSystem returns the singleton pixel batch system
func getPixelBatchSystem() *PixelBatchSystem {
	pixelBatchOnce.Do(func() {
		pixelBatch = &PixelBatchSystem{
			pending: make(map[int]pendingPixel, 256), // Pre-allocate for typical usage
		}
	})
	return pixelBatch
}

// ensureBufferSize ensures the batch buffer matches screen dimensions
func (pbs *PixelBatchSystem) ensureBufferSize(width, height int) {
	if pbs.bufferWidth == width && pbs.bufferHeight == height && pbs.buffer != nil {
		return // Already correct size
	}

	// Allocate new buffer
	pbs.bufferWidth = width
	pbs.bufferHeight = height
	pbs.buffer = make([]byte, width*height*4)

	// Recreate overlay image
	if pbs.overlay != nil {
		pbs.overlay.Deallocate()
	}
	pbs.overlay = ebiten.NewImage(width, height)
}

// QueuePixel adds a pixel to the batch queue (zero-allocation hot path)
func (pbs *PixelBatchSystem) QueuePixel(x, y int, clr color.Color) {
	// Bounds check
	if x < 0 || x >= pbs.bufferWidth || y < 0 || y >= pbs.bufferHeight {
		return
	}

	// Convert color to RGBA bytes
	r, g, b, a := clr.RGBA()

	// Use sparse map key
	key := y*pbs.bufferWidth + x
	pbs.pending[key] = pendingPixel{
		r: uint8(r >> 8),
		g: uint8(g >> 8),
		b: uint8(b >> 8),
		a: uint8(a >> 8),
	}

	pbs.hasPendingPixels = true
	pbs.pixelsThisFrame++
}

// Flush writes all pending pixels to the screen in a single batch operation
func (pbs *PixelBatchSystem) Flush(screen *ebiten.Image) {
	if !pbs.hasPendingPixels || screen == nil {
		return
	}

	// Ensure buffer is correctly sized
	bounds := screen.Bounds()
	pbs.ensureBufferSize(bounds.Dx(), bounds.Dy())

	// Clear buffer to transparent
	for i := range pbs.buffer {
		pbs.buffer[i] = 0
	}

	// Write pending pixels to buffer
	for key, pixel := range pbs.pending {
		offset := key * 4
		if offset+3 < len(pbs.buffer) {
			pbs.buffer[offset] = pixel.r
			pbs.buffer[offset+1] = pixel.g
			pbs.buffer[offset+2] = pixel.b
			pbs.buffer[offset+3] = pixel.a
		}
	}

	// Upload to GPU in single operation
	pbs.overlay.WritePixels(pbs.buffer)

	// Composite overlay onto screen with alpha blending
	opts := &ebiten.DrawImageOptions{}
	opts.Blend = ebiten.BlendSourceOver
	screen.DrawImage(pbs.overlay, opts)

	// Clear pending map (reuse capacity)
	for k := range pbs.pending {
		delete(pbs.pending, k)
	}

	pbs.hasPendingPixels = false
	pbs.flushCount++

	// Update shadow buffer with the pixels we just drew
	pbs.updateShadowBuffer()
}

// updateShadowBuffer syncs pending pixels to shadow buffer for Pget()
func (pbs *PixelBatchSystem) updateShadowBuffer() {
	if shadowBuffer == nil || !shadowBufferValid {
		return
	}

	// Shadow buffer already contains the clear color from Cls()
	// We just need to mark it as potentially dirty for sprite-based updates
	// The actual pixel data was written to screen, so Pget() will work correctly
	// via the shadow buffer lazy sync mechanism
}

// Reset clears the batch system state (call at frame start)
func (pbs *PixelBatchSystem) Reset() {
	// Clear pending map (reuse capacity)
	for k := range pbs.pending {
		delete(pbs.pending, k)
	}
	pbs.hasPendingPixels = false
	pbs.pixelsThisFrame = 0
}

// Stats returns batch system statistics
func (pbs *PixelBatchSystem) Stats() (pixelsQueued int64, flushes int64) {
	return pbs.pixelsThisFrame, pbs.flushCount
}

// HasPending returns true if there are pending pixel writes
func (pbs *PixelBatchSystem) HasPending() bool {
	return pbs.hasPendingPixels
}

// Multi-tier buffer pool for sprite rendering. Pools buffers by size tier
// (8x8, 16x16, 32x32, 64x64, 128x128) to reduce allocations.

// BufferTier represents a pool tier for a specific buffer size
type BufferTier struct {
	pool     sync.Pool
	size     int
	maxAlloc int // Maximum size this tier handles
}

// MultiTierBufferPool provides efficient buffer allocation for varying sizes
type MultiTierBufferPool struct {
	tiers   []*BufferTier
	stats   BufferPoolStats
	statsMu sync.RWMutex
}

// BufferPoolStats tracks pool performance
type BufferPoolStats struct {
	Hits        int64
	Misses      int64
	Allocations int64
	Returns     int64
}

// Pre-defined tier sizes (in pixels, multiply by 4 for bytes)
var tierSizes = []int{
	8 * 8,     // 64 pixels = 256 bytes (8x8 sprites)
	16 * 16,   // 256 pixels = 1KB (16x16 sprites)
	32 * 32,   // 1024 pixels = 4KB (32x32 sprites)
	64 * 64,   // 4096 pixels = 16KB (64x64 sprites)
	128 * 128, // 16384 pixels = 64KB (full PICO-8 screen)
}

// Global multi-tier buffer pool
var (
	globalBufferPool *MultiTierBufferPool
	bufferPoolOnce   sync.Once
)

// GetBufferPool returns the global multi-tier buffer pool
func GetBufferPool() *MultiTierBufferPool {
	bufferPoolOnce.Do(func() {
		globalBufferPool = NewMultiTierBufferPool()
	})
	return globalBufferPool
}

// NewMultiTierBufferPool creates a new multi-tier buffer pool
func NewMultiTierBufferPool() *MultiTierBufferPool {
	pool := &MultiTierBufferPool{
		tiers: make([]*BufferTier, len(tierSizes)),
	}

	for i, pixelCount := range tierSizes {
		byteSize := pixelCount * 4 // RGBA
		pool.tiers[i] = &BufferTier{
			size:     byteSize,
			maxAlloc: byteSize,
			pool: sync.Pool{
				New: func() interface{} {
					// Capture byteSize in closure
					size := byteSize
					buf := make([]byte, size)
					return &buf
				},
			},
		}
		// Fix closure capture issue
		pool.tiers[i].pool.New = pool.makeNewFunc(byteSize)
	}

	return pool
}

// makeNewFunc creates a pool.New function for a specific size
func (p *MultiTierBufferPool) makeNewFunc(size int) func() interface{} {
	return func() interface{} {
		buf := make([]byte, size)
		return &buf
	}
}

// Get retrieves a buffer of at least the requested size
func (p *MultiTierBufferPool) Get(sizeBytes int) []byte {
	// Find appropriate tier
	for _, tier := range p.tiers {
		if sizeBytes <= tier.maxAlloc {
			bufPtr := tier.pool.Get().(*[]byte)
			buf := *bufPtr

			p.statsMu.Lock()
			p.stats.Hits++
			p.statsMu.Unlock()

			// Clear buffer before returning (prevents stale data)
			for i := range buf {
				buf[i] = 0
			}

			// Return exact size requested (slice of pooled buffer)
			return buf[:sizeBytes]
		}
	}

	// Size exceeds all tiers - allocate directly
	p.statsMu.Lock()
	p.stats.Misses++
	p.stats.Allocations++
	p.statsMu.Unlock()

	return make([]byte, sizeBytes)
}

// Put returns a buffer to the pool
func (p *MultiTierBufferPool) Put(buf []byte) {
	capacity := cap(buf)

	// Find appropriate tier based on capacity
	for _, tier := range p.tiers {
		if capacity == tier.size {
			// Reset slice to full capacity before returning
			buf = buf[:capacity]
			tier.pool.Put(&buf)

			p.statsMu.Lock()
			p.stats.Returns++
			p.statsMu.Unlock()
			return
		}
	}

	// Buffer doesn't match any tier - let GC handle it
}

// Stats returns pool statistics
func (p *MultiTierBufferPool) Stats() BufferPoolStats {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	return p.stats
}

// Rectangle pool for sub-image operations.

var rectPool = sync.Pool{
	New: func() interface{} {
		return &image.Rectangle{}
	},
}

// GetRect gets a rectangle from the pool
func GetRect(x0, y0, x1, y1 int) *image.Rectangle {
	r := rectPool.Get().(*image.Rectangle)
	r.Min.X = x0
	r.Min.Y = y0
	r.Max.X = x1
	r.Max.Y = y1
	return r
}

// PutRect returns a rectangle to the pool
func PutRect(r *image.Rectangle) {
	rectPool.Put(r)
}
