package pigo8

import (
	"bytes"
	_ "embed"
	"fmt"

	// "fmt" // Not needed for this version

	"image/color"
	_ "image/png" // Keep in case other PNGs are loaded
	"log"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2" // Re-add text/v2
)

//go:embed pico-8.ttf
var pico8FontTTF []byte

var (
	// drawPaletteMap stores mappings for the draw palette: drawPaletteMap[originalColor] = mappedColor
	drawPaletteMap []int

	// colorToIndexMap provides O(1) lookup from color to palette index
	colorToIndexMap      map[color.Color]int
	colorToIndexMapMutex sync.RWMutex

	// originalPico8Palette holds the an immutable copy of the standard 16 PICO-8 colors.
	// This is used as a reference to check if the current palette is the default.
	originalPico8Palette = []color.Color{
		color.RGBA{R: 0, G: 0, B: 0, A: 255},       // 0 black
		color.RGBA{R: 29, G: 43, B: 83, A: 255},    // 1 dark-blue
		color.RGBA{R: 126, G: 37, B: 83, A: 255},   // 2 dark-purple
		color.RGBA{R: 0, G: 135, B: 81, A: 255},    // 3 dark-green
		color.RGBA{R: 171, G: 82, B: 54, A: 255},   // 4 brown
		color.RGBA{R: 95, G: 87, B: 79, A: 255},    // 5 dark-gray
		color.RGBA{R: 194, G: 195, B: 199, A: 255}, // 6 light-gray
		color.RGBA{R: 255, G: 241, B: 232, A: 255}, // 7 white
		color.RGBA{R: 255, G: 0, B: 77, A: 255},    // 8 red
		color.RGBA{R: 255, G: 163, B: 0, A: 255},   // 9 orange
		color.RGBA{R: 255, G: 236, B: 39, A: 255},  // 10 yellow
		color.RGBA{R: 0, G: 228, B: 54, A: 255},    // 11 green
		color.RGBA{R: 41, G: 173, B: 255, A: 255},  // 12 blue
		color.RGBA{R: 131, G: 118, B: 156, A: 255}, // 13 indigo
		color.RGBA{R: 255, G: 119, B: 168, A: 255}, // 14 pink
		color.RGBA{R: 255, G: 204, B: 170, A: 255}, // 15 peach
	}

	// pico8Palette defines the standard 16 PICO-8 colors.
	pico8Palette = []color.Color{
		color.RGBA{R: 0, G: 0, B: 0, A: 255},       // 0 black
		color.RGBA{R: 29, G: 43, B: 83, A: 255},    // 1 dark-blue
		color.RGBA{R: 126, G: 37, B: 83, A: 255},   // 2 dark-purple
		color.RGBA{R: 0, G: 135, B: 81, A: 255},    // 3 dark-green
		color.RGBA{R: 171, G: 82, B: 54, A: 255},   // 4 brown
		color.RGBA{R: 95, G: 87, B: 79, A: 255},    // 5 dark-gray
		color.RGBA{R: 194, G: 195, B: 199, A: 255}, // 6 light-gray
		color.RGBA{R: 255, G: 241, B: 232, A: 255}, // 7 white
		color.RGBA{R: 255, G: 0, B: 77, A: 255},    // 8 red
		color.RGBA{R: 255, G: 163, B: 0, A: 255},   // 9 orange
		color.RGBA{R: 255, G: 236, B: 39, A: 255},  // 10 yellow
		color.RGBA{R: 0, G: 228, B: 54, A: 255},    // 11 green
		color.RGBA{R: 41, G: 173, B: 255, A: 255},  // 12 blue
		color.RGBA{R: 131, G: 118, B: 156, A: 255}, // 13 indigo
		color.RGBA{R: 255, G: 119, B: 168, A: 255}, // 14 pink
		color.RGBA{R: 255, G: 204, B: 170, A: 255}, // 15 peach
	}

	// paletteTransparency defines which colors in the Pico8Palette should be treated as transparent
	// By default, only color 0 (black) is transparent
	paletteTransparency = []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}

	// pico8FaceSource is the loaded source for the PICO-8 TTF font.
	pico8FaceSource *text.GoTextFaceSource

	// defaultFontSize is the default size used for the Print function.
	// PICO-8 font is typically 6px high.
	defaultFontSize = 6.0

	// These variables hold the internal state for the cursor used by Print.
	// They were moved from main.go.
	cursorX     int
	cursorY     int
	cursorColor = 7 // Default to white (PICO-8 color 7)

	// Pixel buffer system for batch operations
	pixelBuffer       []byte
	pixelBufferWidth  int
	pixelBufferHeight int
	bufferDirty       bool
	pixelBufferMutex  sync.Mutex

	// Sprite modification batching
	spriteModifications = make(map[*ebiten.Image][]pixelMod)
	spriteModMutex      sync.Mutex

	// Screen pixel cache for batch reading operations
	screenPixelCache       []byte
	screenPixelCacheWidth  int
	screenPixelCacheHeight int
	screenCacheValid       bool
	screenCacheMutex       sync.RWMutex
)

// pixelMod represents a pending pixel modification
type pixelMod struct {
	x, y  int
	color color.Color
}

func init() {
	// Initialize the font face source from embedded TTF.
	drawPaletteMap = make([]int, len(pico8Palette)) // Initialize before use
	resetDrawPaletteMapInternal()

	// Initialize color lookup map
	buildColorToIndexMap()

	s, err := text.NewGoTextFaceSource(bytes.NewReader(pico8FontTTF))
	if err != nil {
		log.Fatalf("Failed to create font face source from pico-8.ttf: %v", err)
	}
	pico8FaceSource = s
}

// buildColorToIndexMap creates a map for O(1) color to index lookups
// Thread-safe: acquires write lock on colorToIndexMapMutex
func buildColorToIndexMap() {
	colorToIndexMapMutex.Lock()
	defer colorToIndexMapMutex.Unlock()

	colorToIndexMap = make(map[color.Color]int, len(pico8Palette))
	for i, paletteColor := range pico8Palette {
		colorToIndexMap[paletteColor] = i
	}
}

// Cls clears the current drawing screen with a specified PICO-8 color index.
// Uses the internal `currentScreen` variable set by the engine.
// If no colorIndex is provided, it defaults to 0 (Black).
func Cls(colorIndex ...int) {
	if currentScreen == nil {
		log.Println("Warning: Cls() called before screen was ready.")
		return
	}
	idx := 0 // Default to black (index 0)
	if len(colorIndex) > 0 {
		idx = colorIndex[0]
	}

	if idx < 0 || idx >= len(pico8Palette) {
		log.Printf("Warning: Cls() called with invalid color index %d. Defaulting to 0.", idx)
		idx = 0
	}
	currentScreen.Fill(pico8Palette[idx])

	// Clear the pixel buffer since we're clearing the screen
	clearPixelBuffer()

	// Reset the global print cursor position
	cursorX = 0
	cursorY = 0
}

// ClsRGBA clears the current drawing screen with a specified RGBA color.
// Uses the internal `currentScreen` variable set by the engine.
// This provides more flexibility than the standard Cls function by allowing
// any RGBA color to be used, not just those in the PICO-8 palette.
//
// Example:
//
//	ClsRGBA(color.RGBA{R: 100, G: 150, B: 200, A: 255}) // Clear with a custom blue color
//	ClsRGBA(color.RGBA{}) // Clear with transparent black (all zeros)
func ClsRGBA(clr color.RGBA) {
	if currentScreen == nil {
		log.Println("Warning: ClsRGBA() called before screen was ready.")
		return
	}

	// Fill the screen with the provided RGBA color
	currentScreen.Fill(clr)

	// Clear the pixel buffer since we're clearing the screen
	clearPixelBuffer()

	// Reset the global print cursor position
	cursorX = 0
	cursorY = 0
}

// Pget returns the PICO-8 color index (0-15) of the pixel at coordinates (x, y)
// on the current drawing screen.
// Uses the internal `currentScreen` variable.
//
// If the coordinates (x, y) are outside the screen bounds, it returns 0 (black).
// If the color at (x, y) does not exactly match any color in the Pico8Palette,
// it returns 0.
//
// Example:
//
//	func DrawExample() {
//	    // Set pixel at (10, 20) to red (index 8)
//	    Pset(10, 20, 8)
//
//	    // Get the color index of the pixel we just set
//	    pixelColorIndex := Pget(10, 20)
//
//	    // Print the retrieved color index (will print 8)
//	    // (Requires fmt package)
//	    // fmt.Printf("Color index at (10, 20): %d\n", pixelColorIndex)
//	}
func Pget(x, y int) int {
	if currentScreen == nil {
		log.Println("Warning: Pget() called before screen was ready.")
		return 0
	}
	bounds := currentScreen.Bounds()
	// Check if coordinates are outside the image bounds
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return 0 // PICO-8 pget returns 0 for out-of-bounds
	}

	// Try to get pixel from cache first (batch reading optimization)
	screenCacheMutex.RLock()
	if screenCacheValid && screenPixelCache != nil &&
		x < screenPixelCacheWidth && y < screenPixelCacheHeight {
		offset := (y*screenPixelCacheWidth + x) * 4
		if offset+3 < len(screenPixelCache) {
			r := screenPixelCache[offset]
			g := screenPixelCache[offset+1]
			b := screenPixelCache[offset+2]
			a := screenPixelCache[offset+3]
			screenCacheMutex.RUnlock()

			// Create color from RGBA values
			pixelColor := color.RGBA{r, g, b, a}

			// Use the colorToIndexMap for O(1) lookup (thread-safe)
			colorToIndexMapMutex.RLock()
			index, ok := colorToIndexMap[pixelColor]
			colorToIndexMapMutex.RUnlock()
			if ok {
				return index
			}
			// Cache lookup failed - fall through to fallback path
			// instead of returning 0 immediately (handles edge cases
			// like cache/palette mismatch or color space issues)
			goto fallback
		}
	}
	screenCacheMutex.RUnlock()

fallback:
	// Fallback to individual pixel read if cache is not available
	pixelColor := currentScreen.At(x, y)

	// Normalize to color.RGBA for consistent map lookup
	// currentScreen.At() may return different color types (NRGBA, Gray, etc.)
	// but colorToIndexMap is keyed by color.RGBA
	r, g, b, a := pixelColor.RGBA()
	normalizedColor := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}

	// Use the colorToIndexMap for O(1) lookup (thread-safe)
	colorToIndexMapMutex.RLock()
	index, ok := colorToIndexMap[normalizedColor]
	colorToIndexMapMutex.RUnlock()
	if ok {
		return index
	}

	// If no exact match is found in the palette, return 0
	return 0
}

// Pset draws a single pixel at coordinates (x, y) on the current drawing screen
// using the specified PICO-8 color index or the current cursorColor.
// Uses the internal `currentScreen` variable.
//
// The color is specified by its index (0-15) in the standard Pico8Palette.
// If no colorIndex is provided, the current cursorColor is used.
//
// If the coordinates (x, y) are outside the screen bounds, the function does nothing.
// If an invalid colorIndex is provided (e.g., < 0 or > 15), a warning is logged,
// and the function does nothing.
//
// Example:
//
//	Cursor(0, 0, 8) // Set current color to red
//	Pset(10, 20) // Draws a red pixel at (10, 20)
//	Pset(50, 50, 12) // Draws a blue pixel at (50, 50), color overrides cursorColor
func Pset(x, y int, colorIndex ...int) {
	// Check if screen is ready
	if currentScreen == nil {
		log.Println("Warning: Pset() called before screen was ready.")
		return
	}

	// Initialize pixel buffer if needed
	if pixelBuffer == nil {
		initPixelBuffer(GetScreenWidth(), GetScreenHeight())
	}

	// Determine original color to use
	originalColor := cursorColor // Default to current cursor color
	if len(colorIndex) > 0 {
		originalColor = colorIndex[0]
		// Validate originalColor against the DrawPaletteMap size
		if originalColor < 0 || originalColor >= len(drawPaletteMap) {
			log.Printf("Warning: Pset() called with invalid original color index %d. Draw palette map has %d entries. Ignoring.", originalColor, len(drawPaletteMap))
			return
		}
	} else {
		// Ensure cursorColor is also a valid index for DrawPaletteMap, if DrawPaletteMap is initialized
		if len(drawPaletteMap) > 0 && (cursorColor < 0 || cursorColor >= len(drawPaletteMap)) {
			log.Printf("Warning: Pset() using cursorColor %d which is out of bounds for DrawPaletteMap (size %d). Ignoring.", cursorColor, len(drawPaletteMap))
			return
		} else if len(drawPaletteMap) == 0 {
			// This case should ideally not happen if init and SetPalette are working correctly.
			log.Println("Warning: Pset() called when drawPaletteMap is uninitialized. Ignoring.")
			return
		}
	}

	// Get the mapped color from the draw palette
	// Ensure DrawPaletteMap is initialized and originalColor is a valid index
	if len(drawPaletteMap) == 0 || originalColor < 0 || originalColor >= len(drawPaletteMap) {
		log.Printf("Warning: Pset() DrawPaletteMap not ready or originalColor %d invalid for map size %d. Ignoring.", originalColor, len(drawPaletteMap))
		return
	}
	mappedColor := drawPaletteMap[originalColor]

	// Validate mappedColor against the actual Pico8Palette size
	if mappedColor < 0 || mappedColor >= len(pico8Palette) {
		log.Printf("Warning: Pset() mapped color index %d is out of bounds for Pico8Palette (size %d). Original color was %d. Ignoring.", mappedColor, len(pico8Palette), originalColor)
		return
	}

	// Apply camera offset
	fx, fy := applyCameraOffset(float64(x), float64(y))
	x, y = int(fx), int(fy)

	// Check bounds
	if x < 0 || x >= GetScreenWidth() || y < 0 || y >= GetScreenHeight() {
		return // Silently ignore out-of-bounds pixels
	}

	// Check if the mapped color is transparent
	// Ensure PaletteTransparency is initialized and mappedColor is a valid index
	if len(paletteTransparency) == 0 || mappedColor < 0 || mappedColor >= len(paletteTransparency) {
		log.Printf("Warning: Pset() PaletteTransparency not ready or mappedColor %d invalid for transparency map size %d. Ignoring.", mappedColor, len(paletteTransparency))
		return
	}
	if paletteTransparency[mappedColor] {
		// Don't draw transparent pixels
		return
	}

	// Get the actual color.Color struct for the mapped color
	pixelColor := pico8Palette[mappedColor]

	// Set pixel in buffer instead of immediate GPU upload
	setPixelInBuffer(x, y, pixelColor)
}

const (
	// CharWidthApproximation approximates character width for PICO-8 font for measurement.
	CharWidthApproximation = 4.0
)

// Cursor sets the implicit print cursor position (x, y) and optionally the default draw color.
// It mimics the PICO-8 CURSOR(x, y, [color]) function.
// Calling Cursor() with no arguments resets the cursor position to (0, 0) but leaves the color unchanged.
//
// Args:
//   - args: Optional arguments interpreted as [x, y] or [x, y, colorIndex].
//   - If len(args) == 0: Resets cursor position to (0, 0).
//   - If len(args) == 2: Sets cursor position to (args[0], args[1]).
//   - If len(args) >= 3: Sets cursor position to (args[0], args[1]) and sets currentDrawColor to args[2].
//
// Example:
//
//	Cursor(10, 20)     // Set cursor to (10, 20)
//	Cursor(30, 40, 5) // Set cursor to (30, 40) and draw color to 5 (dark gray)
//	Cursor()          // Reset cursor position to (0, 0)
func Cursor(args ...int) {
	switch len(args) {
	case 0:
		cursorX = 0
		cursorY = 0
	case 2:
		cursorX = args[0]
		cursorY = args[1]
	case 3:
		cursorX = args[0]
		cursorY = args[1]
		// Validate and set color
		col := args[2]
		if col < 0 || col >= len(pico8Palette) {
			log.Printf("Warning: Cursor() called with invalid color index %d. Color not changed.", col)
		} else {
			// Update both color variables to keep them in sync
			cursorColor = col
			currentDrawColor = col
		}
	default:
		log.Printf("Warning: Cursor() called with invalid number of arguments (%d). Expected 0, 2, or 3.", len(args))
	}
}

// Print draws the given value (converted to a string using fmt.Sprintf) onto the current drawing screen.
// Uses the internal `currentScreen` variable.
// It mimics the PICO-8 PRINT(str, [x, y], [color]) function, including implicit cursor tracking.
// It returns the X and Y coordinates of the pixel immediately following the printed string.
//
// Args:
//   - s: The value to print. It will be converted to a string using fmt.Sprintf("%v", s).
//   - args: Optional arguments interpreted based on PICO-8 logic:
//   - If len(args) == 0: Prints at current cursor (cursorX, cursorY) with current cursorColor.
//   - If len(args) == 1: Prints at current cursor (cursorX, cursorY) with color args[0] (overrides cursorColor).
//   - If len(args) == 2: Prints starting at (args[0], args[1]) with current cursorColor.
//   - If len(args) >= 3: Prints starting at (args[0], args[1]) with color args[2] (overrides cursorColor).
//
// Returns:
//   - int: The X coordinate after the string (drawX + stringWidth).
//   - int: The Y coordinate after the string (drawY + fontHeight).
//
// Limitations:
//   - Returned Width: Width calculation is an approximation (char_count * 4).
//
// Example:
//
//	// Assume cursor starts at (0, 0), color is 7 (white)
//	Cursor(0, 0, 6) // Set current color to light gray
//	_, _ = Print("1 HELLO")         // Draws "1 HELLO" at (0,0) in light gray, cursor moves to (0, 6).
//	_, _ = Print(2.718, 8)          // Draws "2.718" at (0,6) in red, cursor moves to (0, 12).
//	_, _ = Print("3 AT", 20, 20)     // Draws "3 AT" at (20,20) in light gray, cursor moves to (20, 26).
//	endX, endY := Print("4 DONE")    // Draws "4 DONE" at (20, 26) in light gray, cursor moves to (20, 32).
//	_, _ = Print(true)              // Draws "true" at current cursor with current color.
func Print(s any, args ...int) (int, int) {
	str := fmt.Sprintf("%v", s)

	// Check if screen is ready
	if currentScreen == nil {
		log.Println("Warning: Print() called before screen was ready.")

		// Calculate return values based on arguments without changing cursor state
		posX, posY := cursorX, cursorY
		if len(args) >= 2 {
			posX, posY = args[0], args[1]
		}

		// Approximate measurement for return value
		advance := float64(len([]rune(str))) * CharWidthApproximation
		endX := int(math.Ceil(float64(posX) + advance))
		endY := posY + int(defaultFontSize)

		return endX, endY
	}

	// Parse arguments
	posX, posY, col := cursorX, cursorY, cursorColor // Use global cursorColor as default

	// If a new position is provided, override posX and posY.
	if len(args) >= 2 {
		posX, posY = args[0], args[1]
	}
	// If a color is provided (in len(args)==1 for color only,
	// or len(args)==3 for position and color), use the last argument.
	if len(args) == 1 || len(args) == 3 {
		col = args[len(args)-1]
		// Update both color variables to keep them in sync
		if col >= 0 && col < len(pico8Palette) {
			cursorColor = col
			currentDrawColor = col
		}
	}

	// Validate the color index.
	if col < 0 || col >= len(pico8Palette) {
		log.Printf("Warning: Print() called with invalid color index %d. Defaulting to cursorColor (%d).", col, cursorColor)
		col = cursorColor // Default to current cursorColor if invalid index given
	}

	// Apply camera offset
	fx, fy := applyCameraOffset(float64(posX), float64(posY))
	drawX, drawY := int(fx), int(fy)

	// --- Prepare for Drawing ---
	face := &text.GoTextFace{
		Source: pico8FaceSource,
		Size:   defaultFontSize,
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(drawX), float64(drawY))
	op.ColorScale.ScaleWithColor(pico8Palette[col])

	// --- Approximate Measurement for Return Value ---
	advance := float64(len([]rune(str))) * CharWidthApproximation
	endX := int(math.Ceil(float64(posX) + advance))
	endY := posY + int(defaultFontSize)

	// --- Draw ---
	text.Draw(currentScreen, str, face, op)

	// --- Update Cursor Position ---
	// If a position was explicitly provided, use that; otherwise, keep the current cursorX.
	if len(args) >= 2 {
		cursorX = args[0]
	} else {
		cursorX = posX
	}
	cursorY = endY

	return endX, endY
}

// resetDrawPaletteMapInternal resets the draw palette map so each color maps to itself.
func resetDrawPaletteMapInternal() {
	if len(drawPaletteMap) == 0 {
		// This case should ideally be prevented by proper initialization in init() and SetPalette()
		// However, as a safeguard, if Pico8Palette is available, try to initialize based on it.
		if len(pico8Palette) > 0 {
			drawPaletteMap = make([]int, len(pico8Palette))
		} else {
			// Fallback to a default size if Pico8Palette is also not ready (e.g., 16)
			drawPaletteMap = make([]int, 16)
			log.Println("Warning: resetDrawPaletteMapInternal called when DrawPaletteMap and Pico8Palette were uninitialized. Defaulting to 16 colors.")
		}
	}
	for i := 0; i < len(drawPaletteMap); i++ {
		drawPaletteMap[i] = i
	}
}

// Pal mimics PICO-8's pal(c0, c1, p) function.
// It configures draw palette mappings. When a color c0 is requested for drawing,
// it will instead use the color c1.
// - pal(): Resets all draw palette mappings to default (color i draws as i).
// - pal(c0, c1): Maps c0 to c1 for future drawing operations (assumes draw palette, p=0).
// - pal(c0, c1, p):
//   - If p=0: Maps c0 to c1 for the draw palette.
//   - If p=1: Screen palette (post-processing effect on the current screen display) is not implemented.
//     A warning will be logged, and the operation might default to p=0 or do nothing for p=1.
func Pal(args ...interface{}) {
	if len(drawPaletteMap) == 0 {
		log.Println("Warning: Pal() called before DrawPaletteMap was initialized. Attempting to initialize.")
		resetDrawPaletteMapInternal() // Attempt to initialize it
		if len(drawPaletteMap) == 0 { // Still zero after attempt
			log.Println("Error: Pal() could not initialize DrawPaletteMap. Aborting Pal operation.")
			return
		}
	}

	if len(args) == 0 { // pal()
		resetDrawPaletteMapInternal()
		return
	}

	var c0, c1, p int
	var ok bool

	// Parse c0
	c0, ok = args[0].(int)
	if !ok {
		if c0Float, floatOk := args[0].(float64); floatOk {
			c0 = int(c0Float)
		} else {
			log.Printf("Warning: Pal() expects integer for c0, got %T. Aborting.", args[0])
			return
		}
	}

	// Parse c1
	if len(args) >= 2 {
		c1, ok = args[1].(int)
		if !ok {
			if c1Float, floatOk := args[1].(float64); floatOk {
				c1 = int(c1Float)
			} else {
				log.Printf("Warning: Pal() expects integer for c1, got %T. Aborting.", args[1])
				return
			}
		}
	} else {
		log.Printf("Warning: Pal() called with too few arguments. Expected pal(c0, c1, [p]). Aborting.")
		return
	}

	// Parse p (palette group)
	p = 0 // Default to draw palette
	if len(args) >= 3 {
		p, ok = args[2].(int)
		if !ok {
			if pFloat, floatOk := args[2].(float64); floatOk {
				p = int(pFloat)
			} else {
				log.Printf("Warning: Pal() expects integer for p, got %T. Defaulting to p=0.", args[2])
				p = 0
			}
		}
	}

	// Validate palette indices
	paletteSize := len(drawPaletteMap)
	if c0 < 0 || c0 >= paletteSize {
		log.Printf("Warning: Pal() c0 index %d out of bounds [0, %d). Aborting.", c0, paletteSize)
		return
	}
	if c1 < 0 || c1 >= paletteSize {
		log.Printf("Warning: Pal() c1 index %d out of bounds [0, %d). Aborting.", c1, paletteSize)
		return
	}

	switch p {
	case 0: // Draw palette
		drawPaletteMap[c0] = c1
	case 1: // Screen palette
		log.Printf("Warning: Pal() with p=1 (screen palette) is not yet implemented.")
		// For now, screen palette calls do not modify the drawPaletteMap or screen.
	default:
		log.Printf("Warning: Pal() called with invalid palette group p=%d. Expected 0 or 1. Aborting.", p)
	}
}

// Palt sets the transparency for a specific color in the palette.
// When called with no arguments, it resets all colors to default transparency
// (only black is transparent).
//
// Args:
//   - color: A color number from the PICO-8 palette (0-15).
//   - transparent: true to make the color transparent, false to make it opaque.
//
// Example:
//
//	Spr(1, 10, 10)  // Draw sprite with default transparency (black is transparent)
//	Palt(8, true)   // Make red (color 8) transparent
//	Spr(1, 20, 10)  // Draw sprite with red transparent
//	Palt()          // Reset to default transparency
func Palt(args ...interface{}) {
	// If called with no arguments, reset to default transparency settings
	if len(args) == 0 {
		// Default: only black (color 0) is transparent
		for i := range paletteTransparency {
			paletteTransparency[i] = (i == 0)
		}
		return
	}

	// Extract color argument
	colorIndex, ok := args[0].(int)
	if !ok {
		// Try to convert from float64 if provided
		if colorFloat, floatOk := args[0].(float64); floatOk {
			colorIndex = int(colorFloat)
		} else {
			log.Printf("Warning: Palt() called with invalid color type: %T", args[0])
			return
		}
	}

	// Validate color index
	if colorIndex < 0 || colorIndex >= len(paletteTransparency) {
		log.Printf("Warning: Palt() called with out-of-range color index: %d", colorIndex)
		return
	}

	// Extract transparent argument
	if len(args) < 2 {
		log.Printf("Warning: Palt() called without transparency value")
		return
	}

	transparent, ok := args[1].(bool)
	if !ok {
		log.Printf("Warning: Palt() called with invalid transparency type: %T", args[1])
		return
	}

	// Set the transparency for the specified color
	paletteTransparency[colorIndex] = transparent
}

// --- Palette Management Functions ---

// SetPalette replaces the current color palette with a new one.
// This also resizes the transparency array to match the new palette size,
// setting only the first color (index 0) as transparent by default.
//
// newPalette: Slice of color.Color values to use as the new palette.
//
// Example:
//
//	// Create a 4-color grayscale palette
//	grayscale := []color.Color{
//		color.RGBA{0, 0, 0, 255},       // Black
//		color.RGBA{85, 85, 85, 255},    // Dark Gray
//		color.RGBA{170, 170, 170, 255}, // Light Gray
//		color.RGBA{255, 255, 255, 255}, // White
//	}
//	SetPalette(grayscale)
func SetPalette(newPalette []color.Color) {
	if len(newPalette) == 0 {
		log.Println("Warning: Attempted to set empty palette. Ignoring.")
		return
	}

	// Replace the palette
	pico8Palette = newPalette

	// Resize transparency array to match
	oldTransparency := paletteTransparency
	paletteTransparency = make([]bool, len(newPalette))

	// Copy over existing transparency settings for indices that still exist
	for i := 0; i < len(paletteTransparency) && i < len(oldTransparency); i++ {
		paletteTransparency[i] = oldTransparency[i]
	}

	// Ensure at least the first color is transparent if we have any colors
	if len(paletteTransparency) > 0 {
		paletteTransparency[0] = true

		// Resize and reset draw palette map as well
		drawPaletteMap = make([]int, len(newPalette))
		resetDrawPaletteMapInternal()
	}

	// Rebuild colorToIndexMap to stay in sync with the new palette
	// This ensures Pget() returns correct color indices
	buildColorToIndexMap()
}

// Transparency is controlled using the Palt() function

// GetPaletteSize returns the current number of colors in the palette.
//
// Example:
//
//	// Get the current palette size
//	size := GetPaletteSize()
//	Print(fmt.Sprintf("Palette has %d colors", size), 10, 10, 7)
func GetPaletteSize() int {
	return len(pico8Palette)
}

// GetPaletteColor returns the color.Color at the specified index in the palette.
// Returns nil if the index is out of range.
//
// colorIndex: Index of the color to retrieve.
//
// Example:
//
//	// Get the color at index 3
//	color3 := GetPaletteColor(3)
func GetPaletteColor(colorIndex int) color.Color {
	if colorIndex >= 0 && colorIndex < len(pico8Palette) {
		return pico8Palette[colorIndex]
	}
	return nil
}

// SetPaletteColor replaces a single color in the palette at the specified index.
// If the index is out of range, the function does nothing.
//
// colorIndex: Index of the color to replace.
// newColor: The new color.Color to use.
//
// Example:
//
//	// Change color 7 (white) to a light blue
//	SetPaletteColor(7, color.RGBA{200, 220, 255, 255})
func SetPaletteColor(colorIndex int, newColor color.Color) {
	if colorIndex >= 0 && colorIndex < len(pico8Palette) {
		pico8Palette[colorIndex] = newColor
		// Rebuild colorToIndexMap to stay in sync with the modified palette
		// This ensures Pget() returns correct color indices
		buildColorToIndexMap()
	} else {
		log.Printf("Warning: Attempted to set color at out-of-range index %d. Palette has %d colors.",
			colorIndex, len(pico8Palette))
	}
}

// --- Transparency Functions ---

// No alpha transparency functions needed - using only binary transparency
// where colors are either fully visible or fully transparent

// IsDefaultPico8PaletteActive checks if the current p8.Pico8Palette is the standard PICO-8 palette.
// It compares the current palette's length and RGBA values against the original PICO-8 colors.
func IsDefaultPico8PaletteActive() bool {
	if len(pico8Palette) != len(originalPico8Palette) {
		return false
	}
	for i := range pico8Palette {
		// Compare RGBA values directly. ebiten.Color.RGBA() returns premultiplied alpha values.
		r1, g1, b1, a1 := pico8Palette[i].RGBA()
		r2, g2, b2, a2 := originalPico8Palette[i].RGBA()
		if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
			return false
		}
	}
	return true
}

// initPixelBuffer initializes the pixel buffer for batch operations
func initPixelBuffer(width, height int) {
	pixelBufferMutex.Lock()
	defer pixelBufferMutex.Unlock()

	if pixelBufferWidth != width || pixelBufferHeight != height {
		pixelBufferWidth = width
		pixelBufferHeight = height
		pixelBuffer = make([]byte, width*height*4) // RGBA format
		bufferDirty = false
		log.Printf("Initialized pixel buffer: %dx%d (%d bytes)", width, height, len(pixelBuffer))
	}

	// Also initialize screen pixel cache for reading operations
	initScreenPixelCache(width, height)
}

// flushPixelBuffer uploads all pending pixel changes to the GPU
func flushPixelBuffer() {
	pixelBufferMutex.Lock()
	defer pixelBufferMutex.Unlock()

	if bufferDirty && currentScreen != nil && len(pixelBuffer) > 0 {
		currentScreen.WritePixels(pixelBuffer)
		bufferDirty = false

		// Update screen pixel cache after flushing
		updateScreenPixelCache()
	}
}

// setPixelInBuffer sets a pixel in the buffer without immediate GPU upload
func setPixelInBuffer(x, y int, clr color.Color) {
	if x < 0 || x >= pixelBufferWidth || y < 0 || y >= pixelBufferHeight {
		return
	}

	offset := (y*pixelBufferWidth + x) * 4
	r, g, b, a := clr.RGBA()
	pixelBuffer[offset] = uint8(r >> 8)   // Red
	pixelBuffer[offset+1] = uint8(g >> 8) // Green
	pixelBuffer[offset+2] = uint8(b >> 8) // Blue
	pixelBuffer[offset+3] = uint8(a >> 8) // Alpha
	bufferDirty = true
}

// clearPixelBuffer clears the pixel buffer and marks it as clean
func clearPixelBuffer() {
	pixelBufferMutex.Lock()
	defer pixelBufferMutex.Unlock()

	if len(pixelBuffer) > 0 {
		for i := range pixelBuffer {
			pixelBuffer[i] = 0
		}
		bufferDirty = false
	}

	// Also invalidate screen pixel cache
	invalidateScreenPixelCache()
}

// initScreenPixelCache initializes the screen pixel cache for batch reading operations
func initScreenPixelCache(width, height int) {
	screenCacheMutex.Lock()
	defer screenCacheMutex.Unlock()

	if screenPixelCacheWidth != width || screenPixelCacheHeight != height {
		screenPixelCacheWidth = width
		screenPixelCacheHeight = height
		screenPixelCache = make([]byte, width*height*4) // RGBA format
		screenCacheValid = false
		log.Printf("Initialized screen pixel cache: %dx%d (%d bytes)", width, height, len(screenPixelCache))
	}
}

// updateScreenPixelCache reads all pixels from the screen into the cache
func updateScreenPixelCache() {
	screenCacheMutex.Lock()
	defer screenCacheMutex.Unlock()

	if currentScreen == nil || screenPixelCache == nil {
		screenCacheValid = false
		return
	}

	// Read all pixels from GPU in one batch operation
	currentScreen.ReadPixels(screenPixelCache)
	screenCacheValid = true
}

// invalidateScreenPixelCache marks the screen pixel cache as invalid
func invalidateScreenPixelCache() {
	screenCacheMutex.Lock()
	defer screenCacheMutex.Unlock()
	screenCacheValid = false
}

// GetScreenPixelCacheStats returns statistics about the screen pixel cache
func GetScreenPixelCacheStats() (width, height int, valid bool, size int) {
	screenCacheMutex.RLock()
	defer screenCacheMutex.RUnlock()

	return screenPixelCacheWidth, screenPixelCacheHeight, screenCacheValid, len(screenPixelCache)
}

// ForceUpdateScreenPixelCache forces an update of the screen pixel cache
func ForceUpdateScreenPixelCache() {
	updateScreenPixelCache()
}

// setScreenSize updates the screen dimensions and reinitializes the pixel buffer
func setScreenSize(width, height int) {
	screenWidth = width
	screenHeight = height

	// Reinitialize pixel buffer with new dimensions
	initPixelBuffer(width, height)
}
