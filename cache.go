package pigo8

import (
	"container/list"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// LRUCache implements a thread-safe LRU cache with size limits
type LRUCache[K comparable, V any] struct {
	maxSize int
	items   map[K]*list.Element
	order   *list.List
	mutex   sync.RWMutex

	// Metrics
	hits      int64
	misses    int64
	evictions int64
}

// CacheEntry represents an entry in the LRU cache
type CacheEntry[K comparable, V any] struct {
	key        K
	value      V
	accessTime time.Time
}

// NewLRUCache creates a new LRU cache with the specified maximum size
func NewLRUCache[K comparable, V any](maxSize int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		maxSize: maxSize,
		items:   make(map[K]*list.Element),
		order:   list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var zero V

	if elem, exists := c.items[key]; exists {
		// Move to front (most recently used)
		c.order.MoveToFront(elem)
		entry := elem.Value.(*CacheEntry[K, V])
		entry.accessTime = time.Now()
		c.hits++
		return entry.value, true
	}

	c.misses++
	return zero, false
}

// Put adds or updates a value in the cache
func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, exists := c.items[key]; exists {
		// Update existing entry
		c.order.MoveToFront(elem)
		entry := elem.Value.(*CacheEntry[K, V])
		entry.value = value
		entry.accessTime = time.Now()
		return
	}

	// Add new entry
	entry := &CacheEntry[K, V]{
		key:        key,
		value:      value,
		accessTime: time.Now(),
	}

	elem := c.order.PushFront(entry)
	c.items[key] = elem

	// Evict if necessary
	if c.order.Len() > c.maxSize {
		c.evictLRU()
	}
}

// evictLRU removes the least recently used item
func (c *LRUCache[K, V]) evictLRU() {
	elem := c.order.Back()
	if elem != nil {
		c.order.Remove(elem)
		entry := elem.Value.(*CacheEntry[K, V])
		delete(c.items, entry.key)
		c.evictions++
	}
}

// Clear removes all items from the cache
func (c *LRUCache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[K]*list.Element)
	c.order = list.New()
}

// Size returns the current number of items in the cache
func (c *LRUCache[K, V]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
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
		Size:      len(c.items),
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
