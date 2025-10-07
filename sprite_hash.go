package pigo8

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"sync"
	"time"
)

// SpriteHashEntry represents a sprite hash table entry with collision detection
type SpriteHashEntry struct {
	Hash      string    `json:"hash"`
	Pixels    [][]int   `json:"pixels"`
	Flags     FlagsData `json:"flags"`
	SpriteID  int       `json:"sprite_id"`
	CreatedAt time.Time `json:"created_at"`
}

// SpriteHashTable manages sprite hashes with collision detection
type SpriteHashTable struct {
	entries map[string]*SpriteHashEntry
	mutex   sync.RWMutex
}

// NewSpriteHashTable creates a new sprite hash table
func NewSpriteHashTable() *SpriteHashTable {
	return &SpriteHashTable{
		entries: make(map[string]*SpriteHashEntry),
	}
}

// generateOptimizedSpriteHashWithTiming creates a hash with timing metrics
func generateOptimizedSpriteHashWithTiming(pixels [][]int, flags FlagsData) (string, time.Duration) {
	start := time.Now()
	hash := generateOptimizedSpriteHashInternal(pixels, flags)
	duration := time.Since(start)

	// Record metrics
	metricsCollector.RecordHashTime(duration)

	return hash, duration
}

// generateOptimizedSpriteHashInternal is the internal hash generation function
func generateOptimizedSpriteHashInternal(pixels [][]int, flags FlagsData) string {
	// Use FNV-1a hash for fast sprite deduplication
	hasher := fnv.New64a()

	// Write pixel data directly as bytes to avoid string allocation
	for _, row := range pixels {
		for _, pixel := range row {
			// Convert int to 4 bytes (little endian)
			pixelBytes := [4]byte{
				byte(pixel),
				byte(pixel >> 8),
				byte(pixel >> 16),
				byte(pixel >> 24),
			}
			hasher.Write(pixelBytes[:])
		}
	}

	// Write flags data
	// Write bitfield as 4 bytes
	flagBytes := [4]byte{
		byte(flags.Bitfield),
		byte(flags.Bitfield >> 8),
		byte(flags.Bitfield >> 16),
		byte(flags.Bitfield >> 24),
	}
	hasher.Write(flagBytes[:])

	// Write individual flags as bytes
	for _, flag := range flags.Individual {
		if flag {
			hasher.Write([]byte{1})
		} else {
			hasher.Write([]byte{0})
		}
	}

	return fmt.Sprintf("%x", hasher.Sum64())
}

// AddEntry adds a sprite hash entry with collision detection
func (h *SpriteHashTable) AddEntry(pixels [][]int, flags FlagsData, spriteID int) (string, bool) {
	hash, _ := generateOptimizedSpriteHashWithTiming(pixels, flags)

	h.mutex.Lock()
	defer h.mutex.Unlock()

	if existing, exists := h.entries[hash]; exists {
		// Check for actual collision vs duplicate
		if !h.isActualDuplicate(existing, pixels, flags) {
			// This is a hash collision!
			metricsCollector.RecordHashCollision()

			// Generate a collision-resistant hash by including sprite ID
			collisionHash := fmt.Sprintf("%s_%d", hash, spriteID)
			h.entries[collisionHash] = &SpriteHashEntry{
				Hash:      collisionHash,
				Pixels:    pixels,
				Flags:     flags,
				SpriteID:  spriteID,
				CreatedAt: time.Now(),
			}
			return collisionHash, false
		}
		// This is a true duplicate
		return hash, true
	}

	// New unique sprite
	h.entries[hash] = &SpriteHashEntry{
		Hash:      hash,
		Pixels:    pixels,
		Flags:     flags,
		SpriteID:  spriteID,
		CreatedAt: time.Now(),
	}

	return hash, false
}

// isActualDuplicate performs deep comparison to detect true duplicates vs hash collisions
func (h *SpriteHashTable) isActualDuplicate(existing *SpriteHashEntry, pixels [][]int, flags FlagsData) bool {
	// Compare flags first (faster)
	if existing.Flags.Bitfield != flags.Bitfield {
		return false
	}

	if !reflect.DeepEqual(existing.Flags.Individual, flags.Individual) {
		return false
	}

	// Compare pixel dimensions
	if len(existing.Pixels) != len(pixels) {
		return false
	}

	// Compare pixel data
	for i, row := range existing.Pixels {
		if len(row) != len(pixels[i]) {
			return false
		}
		for j, pixel := range row {
			if pixel != pixels[i][j] {
				return false
			}
		}
	}

	return true
}

// GetEntry retrieves a hash entry
func (h *SpriteHashTable) GetEntry(hash string) (*SpriteHashEntry, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	entry, exists := h.entries[hash]
	return entry, exists
}

// Clear clears all hash entries
func (h *SpriteHashTable) Clear() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.entries = make(map[string]*SpriteHashEntry)
}

// Stats returns hash table statistics
func (h *SpriteHashTable) Stats() HashTableStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return HashTableStats{
		EntryCount:     len(h.entries),
		CollisionCount: int(metricsCollector.hashCollisions),
	}
}

// HashTableStats holds hash table statistics
type HashTableStats struct {
	EntryCount     int `json:"entry_count"`
	CollisionCount int `json:"collision_count"`
}

// Global sprite hash table
var spriteHashTable = NewSpriteHashTable()

// checkForDuplicateWithCollisionDetection checks for sprite duplicates with collision handling
func checkForDuplicateWithCollisionDetection(spriteData spriteData, spriteHashes map[string]int) (int, bool) {
	hash, isDuplicate := spriteHashTable.AddEntry(spriteData.Pixels, spriteData.Flags, spriteData.ID)

	if isDuplicate {
		if existingIndex, found := spriteHashes[hash]; found {
			return existingIndex, true
		}
	}

	return -1, false
}
