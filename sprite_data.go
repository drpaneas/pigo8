package pigo8

import (
	"image/color"
	"log"
)

// Sget returns the PICO-8 color index (0-15) of the pixel at the specified coordinates on the spritesheet.
// This mimics PICO-8's sget(x, y) function.
//
// x: the distance from the left side of the spritesheet (in pixels).
// y: the distance from the top side of the spritesheet (in pixels).
//
// Returns:
//   - int: The color index (0-15) of the pixel at the specified coordinates.
//   - If the coordinates are out of bounds or no sprite is found, returns 0.
//
// Example:
//
//	pixel_color := Sget(10, 20) // Returns color index (0-15) if pixel exists
func Sget[X Number, Y Number](x X, y Y) int {
	// Convert generic x, y to required types
	px := int(x)
	py := int(y)

	// Ensure spritesheet is loaded
	if currentSprites == nil {
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Failed to load spritesheet for Sget(): %v", err)
			return 0 // Return 0 if spritesheet couldn't be loaded
		}
		currentSprites = loaded
	}

	// In PICO-8, sprites are arranged in a grid on the spritesheet
	// Each sprite is 8x8 pixels, and the spritesheet is 128x128 pixels (16x16 sprites)
	// Find which sprite contains the specified pixel coordinates
	spriteX := px / 8                                   // Determine which sprite column contains the pixel
	spriteY := py / 8                                   // Determine which sprite row contains the pixel
	spriteCellID := calculateSpriteID(spriteX, spriteY) // Calculate sprite ID based on dynamic dimensions

	// Calculate the pixel position within the sprite
	localX := px % 8 // X position within the sprite (0-7)
	localY := py % 8 // Y position within the sprite (0-7)

	// Find the sprite with the matching ID
	for _, sprite := range currentSprites {
		if sprite.ID == spriteCellID {
			// Try to get pixel from cache first (batch reading optimization)
			spritePixelCacheMutex.RLock()
			if spriteCacheValid[spriteCellID] {
				if pixels, cacheSize, found := spritePixelCacheManager.Get(spriteCellID); found && cacheSize > 0 {
					offset := (localY*8 + localX) * 4
					if offset+3 < len(pixels) {
						r := pixels[offset]
						g := pixels[offset+1]
						b := pixels[offset+2]
						a := pixels[offset+3]
						spritePixelCacheMutex.RUnlock()

						// Create color from RGBA values
						pixelColor := color.RGBA{r, g, b, a}

						// Find the matching color in the PICO-8 palette
						for i, c := range pico8Palette {
							if colorEquals(pixelColor, c) {
								recordCacheHit()
								return i // Return the color index (0-15)
							}
						}
						return 0
					}
				}
			}
			spritePixelCacheMutex.RUnlock()

			recordCacheMiss()

			// Fallback to individual pixel read if cache is not available
			pixelColor := sprite.Image.At(localX, localY)

			// Find the matching color in the PICO-8 palette
			for i, c := range pico8Palette {
				if colorEquals(pixelColor, c) {
					return i // Return the color index (0-15)
				}
			}
			// If no matching color found, return 0 (transparent/black)
			return 0
		}
	}

	// If no matching pixel was found, return 0
	return 0
}

// colorEquals compares two colors for equality
func colorEquals(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// Color sets the current draw color to be used by subsequent drawing operations.
// The color parameter should be a number from 0 to 15 corresponding to the PICO-8 palette.
//
// Example:
//
//	Color(8) // Set current draw color to red (color 8)
//	Sset(10, 20) // Draw a red pixel at (10, 20) on the spritesheet
func Color(colorIndex int) {
	// Clamp color index to valid range (0-15)
	if colorIndex < 0 {
		colorIndex = 0
	} else if colorIndex >= len(pico8Palette) {
		colorIndex = len(pico8Palette) - 1
	}

	// Update both color variables to keep them in sync
	currentDrawColor = colorIndex
	cursorColor = colorIndex
}

// Sset sets the color of a pixel at the specified coordinates on the spritesheet.
// If the optional color parameter is not provided, it uses the current draw color.
//
// x: the distance from the left side of the spritesheet (in pixels).
// y: the distance from the top side of the spritesheet (in pixels).
// colorIndex: (optional) a color number from 0 to 15.
//
// Example:
//
//	Sset(10, 0, 8) // Draw a red pixel at (10,0) on the spritesheet
//	Color(12)
//	Sset(16, 0) // Draw a blue pixel at (16,0) using the current draw color
func Sset[X Number, Y Number](x X, y Y, colorIndex ...int) {
	// Convert generic x, y to required types
	px := int(x)
	py := int(y)

	// Determine which color to use
	colorToUse := currentDrawColor
	if len(colorIndex) > 0 {
		colorToUse = colorIndex[0]
		// Clamp color index to valid range (0-15)
		if colorToUse < 0 {
			colorToUse = 0
		} else if colorToUse >= len(pico8Palette) {
			colorToUse = len(pico8Palette) - 1
		}
	}

	// Ensure spritesheet is loaded
	if currentSprites == nil {
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Failed to load spritesheet for Sset(): %v", err)
			return // Can't set pixel if spritesheet couldn't be loaded
		}
		currentSprites = loaded
	}

	// In PICO-8, sprites are arranged in a grid on the spritesheet
	// Each sprite is 8x8 pixels, and the spritesheet is 128x128 pixels (16x16 sprites)
	// Find which sprite contains the specified pixel coordinates
	spriteX := px / 8                                   // Determine which sprite column contains the pixel
	spriteY := py / 8                                   // Determine which sprite row contains the pixel
	spriteCellID := calculateSpriteID(spriteX, spriteY) // Calculate sprite ID based on dynamic dimensions

	// Calculate the pixel position within the sprite
	localX := px % 8 // X position within the sprite (0-7)
	localY := py % 8 // Y position within the sprite (0-7)

	// Find the sprite with the matching ID
	for i := range currentSprites {
		sprite := &currentSprites[i]
		if sprite.ID == spriteCellID {
			// Queue sprite modification instead of immediate GPU upload
			queueSpriteModification(sprite.Image, localX, localY, pico8Palette[colorToUse])
			return
		}
	}

	// If no sprite with the matching ID was found, log a warning
	log.Printf("Warning: Sset() called for non-existent sprite ID %d at position (%d, %d)", spriteCellID, px, py)
}

// Fset sets the flag status of a sprite.
// If flag is provided, sets that specific flag to the value.
// If flag is not provided, sets all flags according to the value (either a boolean or a bitfield).
//
// spriteNum: the sprite number to modify.
// flagOrValue: either the flag number (0-7) or a boolean/bitfield value.
// value: (optional) true/false to turn the flag on/off.
//
// Example:
//
//	// Set flag 0 to true on sprite 1
//	Fset(1, 0, true)
//
//	// Set all flags off on sprite 2
//	Fset(2, false)
//
//	// Set flags 1,3,5,7 on sprite 2 using a bitfield (170 = 2+8+32+128)
//	Fset(2, 170)
func Fset(spriteNum int, flagOrValue interface{}, value ...interface{}) {
	// Lazy-load sprites if needed
	if currentSprites == nil {
		sprites, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Fset() called but failed to load spritesheet: %v", err)
			return
		}
		currentSprites = sprites
	}

	// Find the sprite with the matching ID
	var spriteIndex = -1
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNum {
			spriteIndex = i
			break
		}
	}

	// If sprite not found, return
	if spriteIndex == -1 {
		log.Printf("Warning: Fset() called with invalid sprite number: %d", spriteNum)
		return
	}

	// Case 1: Setting a specific flag
	if flagNum, ok := flagOrValue.(int); ok && len(value) > 0 {
		// Validate flag index
		if flagNum < 0 || flagNum >= 8 {
			log.Printf("Warning: Fset() called with invalid flag number: %d", flagNum)
			return
		}

		// Get the boolean value
		var boolValue bool
		switch v := value[0].(type) {
		case bool:
			boolValue = v
		case int:
			boolValue = v != 0
		default:
			log.Printf("Warning: Fset() called with invalid value type: %T", value[0])
			return
		}

		// Set the flag
		currentSprites[spriteIndex].Flags.Individual[flagNum] = boolValue

		// Update the bitfield
		if boolValue {
			// Set the bit
			currentSprites[spriteIndex].Flags.Bitfield |= 1 << flagNum
		} else {
			// Clear the bit
			currentSprites[spriteIndex].Flags.Bitfield &= ^(1 << flagNum)
		}
		return
	}

	// Case 2: Setting all flags with a boolean
	if boolValue, ok := flagOrValue.(bool); ok {
		// Set all flags to the same value
		for i := 0; i < 8; i++ {
			currentSprites[spriteIndex].Flags.Individual[i] = boolValue
		}

		// Update the bitfield
		if boolValue {
			currentSprites[spriteIndex].Flags.Bitfield = 255 // All bits set
		} else {
			currentSprites[spriteIndex].Flags.Bitfield = 0 // All bits cleared
		}
		return
	}

	// Case 3: Setting flags with a bitfield
	if intValue, ok := flagOrValue.(int); ok {
		// Clamp the value to valid range (0-255)
		if intValue < 0 {
			intValue = 0
		} else if intValue > 255 {
			intValue = 255
		}

		// Set the bitfield
		currentSprites[spriteIndex].Flags.Bitfield = intValue

		// Update individual flags
		for i := 0; i < 8; i++ {
			currentSprites[spriteIndex].Flags.Individual[i] = (intValue & (1 << i)) != 0
		}
		return
	}

	log.Printf("Warning: Fset() called with invalid arguments: %v, %v", flagOrValue, value)
}

// Fget returns the flag status of a sprite.
// Returns:
// - bitfield: the entire bitfield of all flags (0-255)
// - isSet: true if the specific flag is set (only meaningful when a flag is provided)
//
// spriteNum: the sprite number to check.
// flag: (optional) the flag number (0-7) to check.
//
// When no flag is specified, only check the bitfield value and ignore isSet.
// When a flag is specified, check isSet for that specific flag's status.
//
// Example:
//
//	// Check if flag 0 is set on sprite 1
//	_, isSet := Fget(1, 0) // Returns true or false in isSet
//
//	// Get all flags for sprite 2 as a bitfield
//	allFlags, _ := Fget(2) // Returns an integer (0-255) in allFlags
func Fget(spriteNum int, flag ...int) (bitfield int, isSet bool) {
	// Lazy-load sprites if needed
	if currentSprites == nil {
		sprites, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Fget() called but failed to load spritesheet: %v", err)
			return 0, false
		}
		currentSprites = sprites
	}

	// Find the sprite with the matching ID
	var spriteInfo *spriteInfo
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNum {
			spriteInfo = &currentSprites[i]
			break
		}
	}

	// If sprite not found, return zero values (reduce log noise for frequent queries)
	if spriteInfo == nil {
		// Only log in debug builds to reduce noise
		debugLog("Fget() called for non-existent sprite ID %d", spriteNum)
		return 0, false
	}

	// Get the entire bitfield
	bitfield = spriteInfo.Flags.Bitfield

	// If a specific flag is requested, check that flag
	if len(flag) > 0 {
		flagNum := flag[0]

		// Validate flag number (0-7)
		if flagNum < 0 || flagNum > 7 {
			log.Printf("Warning: Fget() called with invalid flag number %d. Valid range is 0-7.", flagNum)
			return bitfield, false
		}

		// Check if the specific flag is set
		bitMask := 1 << flagNum
		isSet = (bitfield & bitMask) != 0
	}

	return bitfield, isSet
}
