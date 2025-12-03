package pigo8

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Spr draws a potentially fractional rectangular region of sprites,
// using the internal `currentScreen` and `currentSprites` variables.
//
// The x and y coordinates can be any integer or float type (e.g., int, float64)
// due to the use of generics [X Number, Y Number]. They are converted internally
// to float64 for drawing calculations.
//
// screen:          REMOVED (uses internal currentScreen)
// sprites:         REMOVED (uses internal currentSprites)
// spriteNumber:    The index (int) for the top-left sprite of the block.
// x:               Screen X coordinate (any Number type) for the top-left corner.
// y:               Screen Y coordinate (any Number type) for the top-left corner.
// options...:      Optional parameters (w, h, flipX, flipY)
//   - w (float64 or int): Width multiplier (default 1.0). Handled via interface{}.
//   - h (float64 or int): Height multiplier (default 1.0). Handled via interface{}.
//   - flipX (bool):       Flip horizontally (default false). Handled via interface{}.
//   - flipY (bool):       Flip vertically (default false). Handled via interface{}.
//
// Usage:
//
//	Spr(spriteNumber, x, y)
//	Spr(spriteNumber, x, y, w, h)
//	Spr(spriteNumber, x, y, w, h, flipX)
//	Spr(spriteNumber, x, y, w, h, flipX, flipY)
//
// Example:
//
//	var ix, iy int = 10, 20
//	var fx, fy float64 = 30.5, 20.0
//
//	// Draw sprite 1 at (10, 20) using int coordinates
//	Spr(1, ix, iy)
//
//	// Draw sprite 1 at (30.5, 20.0) using float64 coordinates
//	Spr(1, fx, fy)
//
//	// Draw sprite 1 at (10, 20.0) using mixed int/float64 coordinates
//	Spr(1, ix, fy)
//
//	// Draw sprite 1 and the left half of sprite 2 (w=1.5)
//	Spr(1, 50, 20, 1.5, 1.0)
//
//	// Draw a 1.5w x 1.5h block starting at sprite 0
//	Spr(0, 70, 20, 1.5, 1.5)
//
//	// Draw the same 1.5 x 1.5 block, flipped horizontally
//	Spr(0, 90, 20, 1.5, 1.5, true)
//
//	// Draw sprite 0 using a float sprite number (truncated to 0)
//	Spr(0.7, 110, 20)
//
//	// Explicitly specify generic types if needed (rarely necessary)
//	Spr[int, float64](1, 10, 20.5)
//
//	// Explicitly specify all generic types
//	Spr[float64, int, float64](1.2, 10, 20.5) // spriteNumber becomes 1
func Spr[SN Number, X Number, Y Number](spriteNumber SN, x X, y Y, options ...any) {
	// Convert generic spriteNumber, x, y to required types
	spriteNumInt := int(spriteNumber) // Cast sprite number to int
	fx := float64(x)
	fy := float64(y)

	// Apply camera offset before using coordinates for drawing
	screenFx, screenFy := applyCameraOffset(fx, fy)

	// Use internal package variables set by engine.Draw
	if currentScreen == nil {
		log.Println("Warning: Spr() called before screen was ready.")
		return
	}

	// --- Lazy Loading Logic ---
	if currentSprites == nil {
		loaded, err := loadSpritesheet() // Call the loading function from spritesheet.go
		if err != nil {
			log.Fatalf("Fatal: Failed to load required spritesheet for Spr(): %v", err)
		}
		currentSprites = loaded // Store successfully loaded sprites
	}

	// Find the sprite by ID or index
	spriteInfo := findSpriteByID(spriteNumInt)
	if spriteInfo == nil {
		// No sprite found with this ID or at this index
		debugSpriteNotFound(spriteNumInt, fx, fy)
		return
	}

	// Record sprite rendered (frame-level, minimal overhead)
	recordSpriteRendered()

	// Parse optional arguments
	scaleW, scaleH, flipX, flipY := parseSprOptions(options)

	// Get sprite dimensions
	tileImage := spriteInfo.Image
	spriteWidth := float64(tileImage.Bounds().Dx())
	spriteHeight := float64(tileImage.Bounds().Dy())

	// Create a transparent version of the sprite
	tempImage := createTransparentSpriteImage(tileImage)

	// Calculate final dimensions
	destWidth := spriteWidth * scaleW
	destHeight := spriteHeight * scaleH

	// Setup drawing options
	opts := setupDrawOptions(screenFx, screenFy, destWidth, destHeight, scaleW, scaleH, flipX, flipY)

	// Draw the sprite
	currentScreen.DrawImage(tempImage, opts)
}

// findSpriteByID finds a sprite by its ID using sprite mapping optimization
func findSpriteByID(spriteNumInt int) *spriteInfo {
	// Handle sprite ID 0 as transparent sentinel - always allow it even if not present
	if spriteNumInt == 0 {
		// Try to find sprite 0 first
		for i := range currentSprites {
			if currentSprites[i].ID == 0 {
				return &currentSprites[i]
			}
		}
		// If sprite 0 doesn't exist, return nil (safe no-op)
		return nil
	}

	// Use sprite ID mapping if available
	if SpriteIDMapping != nil {
		if mappedIndex, exists := SpriteIDMapping[spriteNumInt]; exists {
			// Handle mapping to 0 (transparent)
			if mappedIndex == 0 {
				// Find sprite with ID 0 or return nil for safe no-op
				for i := range currentSprites {
					if currentSprites[i].ID == 0 {
						return &currentSprites[i]
					}
				}
				return nil // Safe no-op if sprite 0 not present
			}

			// Return the mapped sprite by index
			if mappedIndex >= 0 && mappedIndex < len(currentSprites) {
				return &currentSprites[mappedIndex]
			}
		}
		// If not found in mapping, treat as non-existent (return nil for safe no-op)
		return nil
	}

	// Fallback to original behavior if no mapping is available
	var spriteInfo *spriteInfo
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNumInt {
			spriteInfo = &currentSprites[i]
			break
		}
	}

	if spriteInfo == nil {
		// If we can't find a sprite with the exact ID, try to use the array index as a fallback
		if spriteNumInt >= 0 && spriteNumInt < len(currentSprites) {
			spriteInfo = &currentSprites[spriteNumInt]
		}
	}

	return spriteInfo
}

// parseSprOptions parses the optional arguments for the Spr function
func parseSprOptions(options []any) (scaleW float64, scaleH float64, flipX bool, flipY bool) {
	// Default values
	scaleW = 1.0
	scaleH = 1.0
	flipX = false
	flipY = false

	// Process optional width multiplier (arg 1)
	if len(options) > 0 && options[0] != nil {
		switch val := options[0].(type) {
		case int:
			scaleW = float64(val)
		case float64:
			scaleW = val
		default:
			log.Printf("Warning: Spr() optional arg 1: expected float64 or int (width multiplier), got %T (%v)", options[0], options[0])
		}
	}

	// Process optional height multiplier (arg 2)
	if len(options) > 1 && options[1] != nil {
		switch val := options[1].(type) {
		case int:
			scaleH = float64(val)
		case float64:
			scaleH = val
		default:
			log.Printf("Warning: Spr() optional arg 2: expected float64 or int (height multiplier), got %T (%v)", options[1], options[1])
		}
	}

	// Process optional flipX (arg 3)
	if len(options) > 2 && options[2] != nil {
		switch val := options[2].(type) {
		case bool:
			flipX = val
		default:
			log.Printf("Warning: Spr() optional arg 3: expected bool (flipX), got %T (%v)", options[2], options[2])
		}
	}

	// Process optional flipY (arg 4)
	if len(options) > 3 && options[3] != nil {
		switch val := options[3].(type) {
		case bool:
			flipY = val
		default:
			log.Printf("Warning: Spr() optional arg 4: expected bool (flipY), got %T (%v)", options[3], options[3])
		}
	}

	// Warn if too many arguments
	if len(options) > 4 {
		log.Printf("Warning: Spr() called with too many arguments (%d), expected max 6 (num, x, y, w, h, fx, fy).", len(options)+3)
	}

	return scaleW, scaleH, flipX, flipY
}

// createTransparentSpriteImage creates a transparent version of a sprite, with caching
func createTransparentSpriteImage(tileImage *ebiten.Image) *ebiten.Image {
	// Check cache first (caches are initialized once via sync.Once)
	if cached, exists := spriteImageCache.Get(tileImage); exists {
		recordCacheHit()
		return cached
	}

	recordCacheMiss()

	// Create new transparent image
	width := tileImage.Bounds().Dx()
	height := tileImage.Bounds().Dy()
	tempImage := ebiten.NewImage(width, height)

	// Get all pixels from source image in batch
	sourcePixels := make([]byte, width*height*4)
	tileImage.ReadPixels(sourcePixels)

	// Create destination pixel buffer
	destPixels := make([]byte, width*height*4)

	// Process pixels in memory (much faster than individual At()/Set() calls)
	for i := 0; i < len(sourcePixels); i += 4 {
		r, g, b, a := sourcePixels[i], sourcePixels[i+1], sourcePixels[i+2], sourcePixels[i+3]

		// Check if this pixel should be transparent (color 0 or fully transparent)
		if a == 0 || (r == 0 && g == 0 && b == 0 && a == 255) {
			// Skip setting transparent pixels - leave as 0
			continue
		}

		// Copy pixel to destination
		destPixels[i] = r   // Red
		destPixels[i+1] = g // Green
		destPixels[i+2] = b // Blue
		destPixels[i+3] = a // Alpha
	}

	// Upload all pixels to GPU in one operation
	tempImage.WritePixels(destPixels)

	// Cache the result
	spriteImageCache.Put(tileImage, tempImage)

	return tempImage
}

// setupDrawOptions creates and configures the drawing options for a sprite
func setupDrawOptions(fx, fy, destWidth, destHeight, scaleW, scaleH float64, flipX, flipY bool) *ebiten.DrawImageOptions {
	// Create drawing options
	opts := &ebiten.DrawImageOptions{}

	// Apply scaling
	if scaleW != 1.0 || scaleH != 1.0 {
		opts.GeoM.Scale(scaleW, scaleH)
	}

	// Apply flipping if needed
	if flipX {
		// For X flip: Scale by -1 on X axis, then translate to compensate
		opts.GeoM.Scale(-1, 1)
		opts.GeoM.Translate(destWidth, 0)
	}

	if flipY {
		// For Y flip: Scale by -1 on Y axis, then translate to compensate
		opts.GeoM.Scale(1, -1)
		opts.GeoM.Translate(0, destHeight)
	}

	// Apply final position
	opts.GeoM.Translate(fx, fy)

	// Ensure nearest-neighbor filtering for pixel-perfect rendering
	opts.Filter = ebiten.FilterNearest

	return opts
}

// getSpriteImage returns the *ebiten.Image for a given sprite ID.
// It first tries to find a sprite with a matching ID.
// If not found, it tries to use the spriteID as an index into the spritesheet.
// Returns nil if the sprite cannot be found.
func getSpriteImage(spriteID int) *ebiten.Image {
	allSprites := getCurrentSprites() // Get sprites from engine
	if allSprites == nil {
		// This can happen if sprites haven't been loaded yet.
		// Attempt to load them, similar to Spr/Sspr.
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: GetSpriteImage failed to load spritesheet: %v", err)
			return nil
		}
		currentSprites = loaded // Store for future calls within this package
		allSprites = currentSprites
		if allSprites == nil { // Still nil after attempt
			log.Println("Warning: GetSpriteImage called when currentSprites is nil and load failed")
			return nil
		}
	}

	// Use the same logic as findSpriteByID for consistency
	foundSpriteInfo := findSpriteByID(spriteID)
	if foundSpriteInfo != nil && foundSpriteInfo.Image != nil {
		return foundSpriteInfo.Image
	}

	// Optionally, log if a sprite is truly not found, but be mindful of performance if called often.
	// log.Printf("Debug: GetSpriteImage could not find sprite with ID or index: %d", spriteID)
	return nil
}

// parseSsprOptions parses the optional arguments for the Sspr function
func parseSsprOptions(options []any, sourceWidth, sourceHeight int) (destWidth, destHeight float64, flipX, flipY bool) {
	// Default values
	destWidth = float64(sourceWidth)
	destHeight = float64(sourceHeight)
	flipX = false
	flipY = false

	// Helper function for logging argument errors
	argError := func(pos int, expected string, val interface{}) {
		log.Printf("Warning: Sspr() optional arg %d: expected %s, got %T (%v)", pos+1, expected, val, val)
	}

	// Process optional dw parameter
	if len(options) >= 1 && options[0] != nil {
		dwVal, ok := options[0].(float64)
		if !ok {
			if dwInt, intOk := options[0].(int); intOk {
				dwVal = float64(dwInt)
				ok = true
			}
		}
		if !ok {
			argError(0, "float64 or int (destination width)", options[0])
		} else {
			destWidth = dwVal
		}
	}

	// Process optional dh parameter
	if len(options) >= 2 && options[1] != nil {
		dhVal, ok := options[1].(float64)
		if !ok {
			if dhInt, intOk := options[1].(int); intOk {
				dhVal = float64(dhInt)
				ok = true
			}
		}
		if !ok {
			argError(1, "float64 or int (destination height)", options[1])
		} else {
			destHeight = dhVal
		}
	}

	// Process optional flip_x parameter
	if len(options) >= 3 && options[2] != nil {
		flipXVal, ok := options[2].(bool)
		if !ok {
			argError(2, "bool (flip_x)", options[2])
		} else {
			flipX = flipXVal
		}
	}

	// Process optional flip_y parameter
	if len(options) >= 4 && options[3] != nil {
		flipYVal, ok := options[3].(bool)
		if !ok {
			argError(3, "bool (flip_y)", options[3])
		} else {
			flipY = flipYVal
		}
	}

	if len(options) > 4 {
		log.Printf("Warning: Sspr() called with too many arguments (%d), expected max 10 (sx, sy, sw, sh, dx, dy, dw, dh, flip_x, flip_y).", len(options)+6)
	}

	return destWidth, destHeight, flipX, flipY
}

// createSpriteSourceImage creates a temporary image from the specified region of the spritesheet
func createSpriteSourceImage(sourceX, sourceY, sourceWidth, sourceHeight int) *ebiten.Image {
	// Create a temporary image for the source region with transparency
	sourceImage := ebiten.NewImage(sourceWidth, sourceHeight)

	// Clear the image with transparent color
	sourceImage.Fill(colorRGBATransparent)

	// Create pixel buffer for batch operations
	pixels := make([]byte, sourceWidth*sourceHeight*4)

	// Process all pixels in batch
	for y := 0; y < sourceHeight; y++ {
		for x := 0; x < sourceWidth; x++ {
			// Get the color at this position on the spritesheet
			colorIndex := Sget(sourceX+x, sourceY+y)

			// Skip transparent pixels based on the palette transparency settings
			if colorIndex >= 0 && colorIndex < len(paletteTransparency) && paletteTransparency[colorIndex] {
				// Skip this pixel, leaving it transparent
				continue
			}

			if colorIndex >= 0 && colorIndex < len(pico8Palette) {
				// Set the pixel in the buffer
				offset := (y*sourceWidth + x) * 4
				r, g, b, a := pico8Palette[colorIndex].RGBA()
				pixels[offset] = uint8(r >> 8)   // Red
				pixels[offset+1] = uint8(g >> 8) // Green
				pixels[offset+2] = uint8(b >> 8) // Blue
				pixels[offset+3] = uint8(a >> 8) // Alpha
			}
		}
	}

	// Upload all pixels to GPU in one operation
	sourceImage.WritePixels(pixels)

	return sourceImage
}

// Sspr draws a sprite from the spritesheet with custom dimensions and optional stretching and flipping.
// Mimics PICO-8's sspr(sx, sy, sw, sh, dx, dy, [dw, dh], [flip_x], [flip_y]) function.
//
// sx: sprite sheet x position (in pixels)
// sy: sprite sheet y position (in pixels)
// sw: sprite width (in pixels)
// sh: sprite height (in pixels)
// dx: how far from the left of the screen to draw the sprite
// dy: how far from the top of the screen to draw the sprite
// dw: (optional) how many pixels wide to draw the sprite (default same as sw)
// dh: (optional) how many pixels tall to draw the sprite (default same as sh)
// flip_x: (optional) boolean, if true draw the sprite flipped horizontally (default false)
// flip_y: (optional) boolean, if true draw the sprite flipped vertically (default false)
//
// Example:
//
//	// Draw a 16x16 sprite from position (8,8) on the spritesheet to position (10,20) on the screen
//	Sspr(8, 8, 16, 16, 10, 20)
//
//	// Draw a 6x5 sprite from position (8,8) on the spritesheet to position (10,20) on the screen
//	Sspr(8, 8, 6, 5, 10, 20)
//
//	// Draw a 16x16 sprite from the spritesheet, stretched to 32x32 on the screen
//	Sspr(8, 8, 16, 16, 10, 20, 32, 32)
//
//	// Draw a 16x16 sprite, flipped horizontally
//	Sspr(8, 8, 16, 16, 10, 20, 16, 16, true, false)
func Sspr[SX Number, SY Number, SW Number, SH Number, DX Number, DY Number](sx SX, sy SY, sw SW, sh SH, dx DX, dy DY, options ...any) {
	// Convert generic types to required types
	sourceX := int(sx)      // Source X on spritesheet
	sourceY := int(sy)      // Source Y on spritesheet
	sourceWidth := int(sw)  // Source width on spritesheet
	sourceHeight := int(sh) // Source height on spritesheet
	destX := float64(dx)
	destY := float64(dy)

	// Use internal package variables set by engine.Draw
	if currentScreen == nil {
		log.Println("Warning: Sspr() called before screen was ready.")
		return
	}

	// --- Lazy Loading Logic ---
	if currentSprites == nil {
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Failed to load spritesheet for Sspr(): %v", err)
			return
		}
		currentSprites = loaded
	}

	// Parse optional arguments
	destWidth, destHeight, flipX, flipY := parseSsprOptions(options, sourceWidth, sourceHeight)

	// Validate source rectangle is within spritesheet bounds
	if !validateSpriteSheetBounds(sourceX, sourceY, sourceWidth, sourceHeight) {
		log.Printf("Warning: Sspr() source rectangle (%d,%d,%d,%d) is outside spritesheet bounds (0,0,%d,%d)",
			sourceX, sourceY, sourceWidth, sourceHeight, spritesheetWidth, spritesheetHeight)
		// Continue anyway, Ebiten will handle clipping
	}

	// Clamp dimensions to be non-negative
	destWidth = math.Max(0, destWidth)
	destHeight = math.Max(0, destHeight)
	if destWidth == 0 || destHeight == 0 {
		return // Don't draw if scaled to zero size
	}

	// Create a temporary image for the source region
	sourceImage := createSpriteSourceImage(sourceX, sourceY, sourceWidth, sourceHeight)

	// Set up drawing options
	op := &ebiten.DrawImageOptions{}

	// Apply camera offset to the intended top-left drawing position (dx, dy)
	screenDrawX, screenDrawY := applyCameraOffset(destX, destY)

	// Apply scaling to match the destination dimensions
	scaleX := destWidth / float64(sourceWidth)
	scaleY := destHeight / float64(sourceHeight)

	// Temporary variables for final translation, considering flips
	finalTranslateX := screenDrawX
	finalTranslateY := screenDrawY

	// Apply flip transformations if needed
	if flipX {
		scaleX *= -1.0
		finalTranslateX += destWidth // Adjust translation for horizontal flip
	}
	if flipY {
		scaleY *= -1.0
		finalTranslateY += destHeight // Adjust translation for vertical flip
	}

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(finalTranslateX, finalTranslateY) // Use camera-adjusted and flip-adjusted coordinates

	// Draw the image to the screen
	currentScreen.DrawImage(sourceImage, op)
}
