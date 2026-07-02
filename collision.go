package pigo8

import (
	"log"
	"sync"
)

// ColorCollision checks if the pixel at coordinates (x, y) matches the specified color.
// Returns true if the pixel color matches the given color, false otherwise.
//
// Parameters:
//   - x: The x-coordinate to check (0-127), can be any numeric type
//   - y: The y-coordinate to check (0-127), can be any numeric type
//   - color: The PICO-8 color index to check against (0-15)
//
// Returns:
//   - bool: true if the pixel at (x, y) matches the specified color, false otherwise
//
// If the coordinates are outside the screen bounds (0-127), the function returns false.
// If the color index is invalid (not 0-15), the function returns false.
//
// Example:
//
//	// Check if the pixel at (10, 20) is red (color 8)
//	if ColorCollision(10, 20, 8) {
//	    // Pixel is red, handle collision
//	}
//
//	// Check if the player is touching a wall (color 3) using float coordinates
//	if ColorCollision(player.x, player.y, 3) {
//	    // Player is touching a wall, prevent movement
//	}
func ColorCollision[X Number, Y Number](x X, y Y, color int) bool {
	// Convert coordinates to int
	xInt := int(float64(x))
	yInt := int(float64(y))

	// Validate coordinates are within screen bounds
	if xInt < 0 || xInt >= GetScreenWidth() || yInt < 0 || yInt >= GetScreenHeight() {
		log.Println("ColorCollision: Invalid coordinates:", xInt, yInt)
		return false
	}

	// Validate color index
	if color < 0 || color >= len(pico8Palette) {
		log.Printf("ColorCollision: Invalid color index: %d. Palette has %d colors.", color, len(pico8Palette))
		return false
	}

	// Get the color at the specified coordinates
	pixelColor := Pget(xInt, yInt)

	// Return true if the pixel color matches the specified color
	return pixelColor == color
}

// Add flag caching for collision detection
var (
	flagCache      = make(map[int]map[int]bool) // spriteID -> flag -> isSet
	flagCacheMutex sync.RWMutex
)

// getCachedFlag returns the cached flag value for a sprite, or computes and caches it
func getCachedFlag(spriteID, flag int) bool {
	flagCacheMutex.RLock()
	if spriteFlags, exists := flagCache[spriteID]; exists {
		if isSet, flagExists := spriteFlags[flag]; flagExists {
			flagCacheMutex.RUnlock()
			return isSet
		}
	}
	flagCacheMutex.RUnlock()

	// Compute flag value
	_, isSet := Fget(spriteID, flag)

	// Cache the result
	flagCacheMutex.Lock()
	if flagCache[spriteID] == nil {
		flagCache[spriteID] = make(map[int]bool)
	}
	flagCache[spriteID][flag] = isSet
	flagCacheMutex.Unlock()

	return isSet
}

// ClearFlagCache clears the flag cache (call when sprites change)
func ClearFlagCache() {
	flagCacheMutex.Lock()
	flagCache = make(map[int]map[int]bool)
	flagCacheMutex.Unlock()
}

// MapCollision checks if a rectangular area, starting at pixel coordinates (x, y) and with a given width and height,
// overlaps with any map tiles that have the specified flag set.
//
// Parameters:
//   - x: The x-coordinate of the top-left corner of the area to check (pixel units).
//   - y: The y-coordinate of the top-left corner of the area to check (pixel units).
//   - flag: The sprite flag number (0-7) to check for on underlying map tiles.
//   - size: (optional) Variadic integers defining the collision area's dimensions in pixels:
//   - No argument: defaults to an 8x8 pixel area.
//   - One argument `s`: defines an `s`x`s` pixel square area.
//   - Two arguments `w, h`: defines a `w`x`h` pixel rectangular area.
//     (Additional arguments are ignored).
//
// Returns:
//   - bool: true if any map tile overlapping the specified area has the given flag set, false otherwise.
//
// Behavior:
// The function first determines the width and height of the collision area based on the `size` parameters.
// It then calculates the range of map tiles (which are 8x8 pixels) that this rectangular area overlaps.
// For each map tile within this range, it retrieves the sprite ID using Mget() and then checks
// if the specified `flag` is set on that sprite using Fget(). If a tile with the
// target flag is found within the area, the function immediately returns true.
// If all overlapping tiles are checked and none have the flag set, it returns false.
//
// Example:
//
//	// Check if the 8x8 area at (player.x, player.y) collides with a tile having flag 0 set
//	if MapCollision(player.x, player.y, 0) {
//	    // Collision detected
//	}
//
//	// Check if a 14x15 pixel player area collides with a tile having flag 0 set
//	playerWidth := 14
//	playerHeight := 15
//	if MapCollision(player.x, player.y, 0, playerWidth, playerHeight) {
//	    // Collision with the rectangular player area detected
//	}
//
//	// Check if a 16x16 pixel area collides with a tile having flag 1 set
//	if MapCollision(enemy.x, enemy.y, 1, 16) { // Assumes square enemy
//	    // Collision detected
//	}
func MapCollision[X Number, Y Number](x X, y Y, flag int, size ...int) bool {
	objectWidth := 8  // Default width in pixels
	objectHeight := 8 // Default height in pixels

	if len(size) > 0 {
		if size[0] > 0 {
			objectWidth = size[0]
		}
		if len(size) > 1 && size[1] > 0 {
			objectHeight = size[1]
		} else if size[0] > 0 { // If only one size is given, assume square
			objectHeight = size[0]
		}
	}

	fx := float64(x)
	fy := float64(y)

	// Determine the range of map tiles the object overlaps
	tileXStart := Flr(fx / 8.0)
	tileYStart := Flr(fy / 8.0)
	tileXEnd := Flr((fx + float64(objectWidth) - 1) / 8.0)
	tileYEnd := Flr((fy + float64(objectHeight) - 1) / 8.0)

	// Check each tile in the overlapping range
	for ty := tileYStart; ty <= tileYEnd; ty++ {
		for tx := tileXStart; tx <= tileXEnd; tx++ {
			// Get sprite number from map at (tx, ty)
			spriteID := Mget(tx, ty)

			// If spriteID is 0 (empty tile) or less (out of bounds/error), skip
			if spriteID <= 0 {
				continue
			}

			// Use cached flag lookup for better performance
			if getCachedFlag(spriteID, flag) {
				return true // Collision detected
			}
		}
	}

	return false // No collision found
}
