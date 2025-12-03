package pigo8

import (
	"hash/fnv"
	"reflect"
	"sync"
	"time"
)

// SpriteHashEntry represents a sprite hash table entry with collision detection
type SpriteHashEntry struct {
	Hash      uint64    `json:"hash"`
	Pixels    [][]int   `json:"pixels"`
	Flags     FlagsData `json:"flags"`
	SpriteID  int       `json:"sprite_id"`
	CreatedAt time.Time `json:"created_at"`
}

// SpriteHashTable manages sprite hashes with collision detection
type SpriteHashTable struct {
	entries          map[uint64]*SpriteHashEntry
	mutex            sync.RWMutex
	collisionCounter uint64 // Monotonically increasing counter for unique collision hashes
}

// NewSpriteHashTable creates a new sprite hash table
func NewSpriteHashTable() *SpriteHashTable {
	return &SpriteHashTable{
		entries:          make(map[uint64]*SpriteHashEntry),
		collisionCounter: 0,
	}
}

// generateOptimizedSpriteHashWithTiming creates a hash with timing metrics
func generateOptimizedSpriteHashWithTiming(pixels [][]int, flags FlagsData) (uint64, time.Duration) {
	start := time.Now()
	hash := generateOptimizedSpriteHashInternal(pixels, flags)
	duration := time.Since(start)

	// Record metrics
	metricsCollector.RecordHashTime(duration)

	return hash, duration
}

// generateOptimizedSpriteHashInternal is the internal hash generation function.
// Returns uint64 directly to avoid string allocation overhead.
func generateOptimizedSpriteHashInternal(pixels [][]int, flags FlagsData) uint64 {
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

	return hasher.Sum64()
}

// AddEntry adds a sprite hash entry with collision detection
func (h *SpriteHashTable) AddEntry(pixels [][]int, flags FlagsData, spriteID int) (uint64, bool) {
	hash, _ := generateOptimizedSpriteHashWithTiming(pixels, flags)

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// ALWAYS search all entries for duplicates first, regardless of whether
	// the original hash slot is free. This ensures correctness even if:
	// - A duplicate exists at a collision-adjusted hash
	// - The hash function has edge cases
	// - Delete operations are added in the future
	for existingHash, entry := range h.entries {
		if h.isActualDuplicate(entry, pixels, flags) {
			return existingHash, true
		}
	}

	// No duplicate found - add the new entry
	_, exists := h.entries[hash]

	// Fast path: original hash slot is free
	if !exists {
		h.entries[hash] = &SpriteHashEntry{
			Hash:      hash,
			Pixels:    pixels,
			Flags:     flags,
			SpriteID:  spriteID,
			CreatedAt: time.Now(),
		}
		return hash, false
	}

	// No duplicate found anywhere - create collision-adjusted entry
	metricsCollector.RecordHashCollision()

	// Generate a unique collision hash using a counter to avoid overwrites.
	// Keep trying until we find a free slot (handles edge cases where collision
	// hashes themselves collide with existing entries).
	var collisionHash uint64
	for {
		h.collisionCounter++
		collisionHash = hash ^ (h.collisionCounter << 32)

		// Check if this slot is free
		if _, taken := h.entries[collisionHash]; !taken {
			break
		}
		// Slot is taken, try next counter value
	}

	h.entries[collisionHash] = &SpriteHashEntry{
		Hash:      collisionHash,
		Pixels:    pixels,
		Flags:     flags,
		SpriteID:  spriteID,
		CreatedAt: time.Now(),
	}
	return collisionHash, false
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
func (h *SpriteHashTable) GetEntry(hash uint64) (*SpriteHashEntry, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	entry, exists := h.entries[hash]
	return entry, exists
}

// Clear clears all hash entries and resets the collision counter
func (h *SpriteHashTable) Clear() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.entries = make(map[uint64]*SpriteHashEntry)
	h.collisionCounter = 0
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

// checkForDuplicateWithCollisionDetection checks for sprite duplicates with collision handling.
// Returns:
//   - existingIndex: the index of the duplicate sprite if found, -1 otherwise
//   - isDuplicate: true if a duplicate was found
//   - hash: the hash used for this sprite (may be collision-adjusted)
//
// The returned hash should be used when storing in spriteHashes to ensure
// collision-adjusted hashes are stored correctly.
//
// Note: The global spriteHashTable should be cleared at the start of each processSprites()
// call to ensure it stays in sync with the local spriteHashes map. If AddEntry reports
// a duplicate but the hash isn't in spriteHashes, it indicates an inconsistency bug.
func checkForDuplicateWithCollisionDetection(spriteData spriteData, spriteHashes map[uint64]int) (existingIndex int, isDuplicate bool, hash uint64) {
	hash, isDup := spriteHashTable.AddEntry(spriteData.Pixels, spriteData.Flags, spriteData.ID)

	if isDup {
		if idx, found := spriteHashes[hash]; found {
			return idx, true, hash
		}
		// Bug detection: AddEntry found a duplicate in the global table, but the hash
		// isn't in the local spriteHashes map. This indicates the global table has
		// stale entries from a previous batch (spriteHashTable.Clear() was not called).
		// Treat as new sprite to avoid invalid index, but log for debugging.
		debugLog("Warning: sprite %d detected as duplicate (hash %x) but not found in local map - treating as new sprite", spriteData.ID, hash)
	}

	return -1, false, hash
}
