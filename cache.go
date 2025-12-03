package pigo8

import (
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

// currentFrameTick holds the current frame tick for LRU tracking.
// Updated once per frame via SetFrameTick() to avoid time.Now() overhead.
// Uses atomic operations to prevent data races between frame updates and cache reads.
var currentFrameTick atomic.Int64

// SetFrameTick updates the current frame tick. Call this once per frame
// (typically at the start of Update or Draw) with ebiten.Tick().
// Thread-safe via atomic store.
func SetFrameTick(tick int64) {
	currentFrameTick.Store(tick)
}

// getFrameTick returns the current frame tick atomically.
func getFrameTick() int64 {
	return currentFrameTick.Load()
}

// lruEntry represents an entry in the slice-based LRU cache
type lruEntry[K comparable, V any] struct {
	key        K
	value      V
	lastAccess int64 // Frame tick, not time.Time (zero-cost update)
	valid      bool  // Whether this slot is in use
}

// LRUCache implements a thread-safe LRU cache using slices instead of container/list.
// This eliminates allocations from linked list nodes and time.Now() calls.
type LRUCache[K comparable, V any] struct {
	maxSize int
	items   map[K]int        // key -> index in entries slice
	entries []lruEntry[K, V] // pre-allocated entry storage
	order   []int            // LRU order (most recent at end)
	size    int              // current number of valid entries
	mutex   sync.RWMutex

	// Metrics
	hits      int64
	misses    int64
	evictions int64
}

// NewLRUCache creates a new LRU cache with the specified maximum size
func NewLRUCache[K comparable, V any](maxSize int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		maxSize: maxSize,
		items:   make(map[K]int, maxSize),
		entries: make([]lruEntry[K, V], maxSize),
		order:   make([]int, 0, maxSize),
		size:    0,
	}
}

// Get retrieves a value from the cache
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var zero V

	if idx, exists := c.items[key]; exists {
		// Update access time with frame tick (atomic read for thread safety)
		c.entries[idx].lastAccess = getFrameTick()
		// Move to end of order (most recently used)
		c.moveToEnd(idx)
		c.hits++
		return c.entries[idx].value, true
	}

	c.misses++
	return zero, false
}

// moveToEnd moves the given index to the end of the order slice (most recently used)
func (c *LRUCache[K, V]) moveToEnd(idx int) {
	// Find position in order slice
	for i, orderIdx := range c.order {
		if orderIdx == idx {
			// Remove from current position and append to end
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, idx)
			return
		}
	}
}

// Put adds or updates a value in the cache
func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if key already exists
	if idx, exists := c.items[key]; exists {
		// Update existing entry (no allocation)
		c.entries[idx].value = value
		c.entries[idx].lastAccess = getFrameTick()
		c.moveToEnd(idx)
		return
	}

	// Need to add new entry
	var idx int

	if c.size < c.maxSize {
		// Use next available slot (no allocation, entries pre-allocated)
		idx = c.size
		c.size++
	} else {
		// Evict least recently used (first in order slice)
		idx = c.order[0]
		// Remove old key from items map
		delete(c.items, c.entries[idx].key)
		// Remove from front of order
		c.order = c.order[1:]
		c.evictions++
	}

	// Set entry data (reusing slot, no allocation)
	c.entries[idx] = lruEntry[K, V]{
		key:        key,
		value:      value,
		lastAccess: getFrameTick(),
		valid:      true,
	}
	c.items[key] = idx
	c.order = append(c.order, idx)
}

// Clear removes all items from the cache
func (c *LRUCache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Reset without reallocating
	c.items = make(map[K]int, c.maxSize)
	c.order = c.order[:0] // Reuse slice capacity
	c.size = 0
	// Mark all entries as invalid
	for i := range c.entries {
		c.entries[i].valid = false
	}
}

// Size returns the current number of items in the cache
func (c *LRUCache[K, V]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.size
}

// Stats returns cache statistics
func (c *LRUCache[K, V]) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Size:      c.size,
		MaxSize:   c.maxSize,
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		HitRate:   hitRate,
	}
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Size      int     `json:"size"`
	MaxSize   int     `json:"max_size"`
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	HitRate   float64 `json:"hit_rate"`
}

// SpriteImageCache is a specialized cache for sprite images
type SpriteImageCache struct {
	cache *LRUCache[*ebiten.Image, *ebiten.Image]
}

// NewSpriteImageCache creates a new sprite image cache
func NewSpriteImageCache(maxSize int) *SpriteImageCache {
	return &SpriteImageCache{
		cache: NewLRUCache[*ebiten.Image, *ebiten.Image](maxSize),
	}
}

// Get retrieves a cached sprite image
func (s *SpriteImageCache) Get(original *ebiten.Image) (*ebiten.Image, bool) {
	return s.cache.Get(original)
}

// Put caches a sprite image
func (s *SpriteImageCache) Put(original, cached *ebiten.Image) {
	s.cache.Put(original, cached)
}

// Clear clears the cache
func (s *SpriteImageCache) Clear() {
	s.cache.Clear()
}

// Stats returns cache statistics
func (s *SpriteImageCache) Stats() CacheStats {
	return s.cache.Stats()
}

// SpritePixelCache is a specialized cache for sprite pixel data
type SpritePixelCache struct {
	cache *LRUCache[int, []byte]
	sizes *LRUCache[int, int]
}

// NewSpritePixelCache creates a new sprite pixel cache
func NewSpritePixelCache(maxSize int) *SpritePixelCache {
	return &SpritePixelCache{
		cache: NewLRUCache[int, []byte](maxSize),
		sizes: NewLRUCache[int, int](maxSize),
	}
}

// Get retrieves cached pixel data
func (s *SpritePixelCache) Get(spriteID int) ([]byte, int, bool) {
	pixels, hasPixels := s.cache.Get(spriteID)
	size, hasSize := s.sizes.Get(spriteID)

	if hasPixels && hasSize {
		return pixels, size, true
	}
	return nil, 0, false
}

// Put caches pixel data
func (s *SpritePixelCache) Put(spriteID int, pixels []byte, size int) {
	s.cache.Put(spriteID, pixels)
	s.sizes.Put(spriteID, size)
}

// Clear clears the cache
func (s *SpritePixelCache) Clear() {
	s.cache.Clear()
	s.sizes.Clear()
}

// Stats returns cache statistics
func (s *SpritePixelCache) Stats() CacheStats {
	return s.cache.Stats()
}
