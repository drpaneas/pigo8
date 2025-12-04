package pigo8

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Note: The global currentDrawColor is defined in engine.go and set by the Color() function

// warningsSeen tracks which warning messages have already been shown
var warningsSeen = make(map[string]bool)

// logWarningOnce logs a warning message only once, even if called multiple times with the same message
func logWarningOnce(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !warningsSeen[msg] {
		log.Print(msg)
		warningsSeen[msg] = true
	}
}

// parseColorArg extracts a color index from optional arguments.
// It handles int, float64, and float32 types, validating against the palette.
// The function updates currentDrawColor and cursorColor to match PICO-8 behavior.
// Returns the color index to use.
//
// funcName: name of the calling function for warning messages (e.g., "Rect/Rectfill", "Line", "Circ/Circfill")
// updateCursor: whether to also update cursorColor (true for rect/circ, false for line)
// options: the variadic options slice passed to the calling function
func parseColorArg(funcName string, updateCursor bool, options []interface{}) int {
	drawColorIndex := currentDrawColor

	if len(options) >= 1 {
		switch v := options[0].(type) {
		case int:
			if v >= 0 && v < len(pico8Palette) {
				drawColorIndex = v
				currentDrawColor = v
				if updateCursor {
					cursorColor = v
				}
			} else {
				logWarningOnce("Warning: %s optional color %d out of range (0-15). Using current color %d.", funcName, v, drawColorIndex)
			}
		case float64:
			intVal := int(v)
			if intVal >= 0 && intVal < len(pico8Palette) {
				drawColorIndex = intVal
				currentDrawColor = intVal
				if updateCursor {
					cursorColor = intVal
				}
			} else {
				logWarningOnce("Warning: %s optional color %d out of range (0-15). Using current color %d.", funcName, intVal, drawColorIndex)
			}
		case float32:
			intVal := int(v)
			if intVal >= 0 && intVal < len(pico8Palette) {
				drawColorIndex = intVal
				currentDrawColor = intVal
				if updateCursor {
					cursorColor = intVal
				}
			} else {
				logWarningOnce("Warning: %s optional color %d out of range (0-15). Using current color %d.", funcName, intVal, drawColorIndex)
			}
		default:
			logWarningOnce("Warning: %s optional color argument expected numeric type, got %T. Using current color %d.", funcName, options[0], drawColorIndex)
		}
	}

	return drawColorIndex
}

// parseRectArgs parses common arguments for Rect and Rectfill.
// It returns the calculated top-left corner (x, y), dimensions (w, h),
// the PICO-8 color index to use, and whether parsing was successful.
func parseRectArgs(x1, y1, x2, y2 float64, options []interface{}) (float32, float32, float32, float32, int, bool) {
	drawColorIndex := parseColorArg("Rect/Rectfill", true, options)

	if len(options) > 1 {
		logWarningOnce("Warning: Rect/Rectfill called with too many arguments (%d), expected max 5.", len(options)+4)
	}

	// Calculate top-left corner (x, y) and dimensions (width, height)
	// PICO-8 rect/rectfill is inclusive, so add 1 to the difference.
	rectX := math.Min(x1, x2)
	rectY := math.Min(y1, y2)
	rectW := math.Abs(x2-x1) + 1
	rectH := math.Abs(y2-y1) + 1

	return float32(rectX), float32(rectY), float32(rectW), float32(rectH), drawColorIndex, true
}

// Rect draws an outline rectangle using two corner points.
// Mimics PICO-8's rect(x1, y1, x2, y2, [color]) function.
//
// x1, y1, x2, y2: Coordinates of two opposing corners (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Rect[X1 Number, Y1 Number, X2 Number, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Rect() called before screen was ready.")
		return
	}

	fx1, fy1, fx2, fy2 := float64(x1), float64(y1), float64(x2), float64(y2)

	// Apply camera offset
	fx1, fy1 = applyCameraOffset(fx1, fy1)
	fx2, fy2 = applyCameraOffset(fx2, fy2)

	// Round to nearest integer for pixel-perfect rendering
	fx1 = math.Round(fx1)
	fy1 = math.Round(fy1)
	fx2 = math.Round(fx2)
	fy2 = math.Round(fy2)

	rectX, rectY, rectW, rectH, drawColorIndex, ok := parseRectArgs(fx1, fy1, fx2, fy2, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(pico8Palette) {
		actualColor = pico8Palette[drawColorIndex]
	} else {
		actualColor = pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Rect(). Defaulting to black.", drawColorIndex)
	}

	// Draw outline by drawing four 1-pixel thick filled rectangles
	topY := rectY
	bottomY := rectY + rectH - 1
	leftX := rectX
	rightX := rectX + rectW - 1

	// Top horizontal line
	vector.DrawFilledRect(currentScreen, leftX, topY, rectW, 1, actualColor, false)
	// Bottom horizontal line
	vector.DrawFilledRect(currentScreen, leftX, bottomY, rectW, 1, actualColor, false)
	// Left vertical line (height adjusted to avoid drawing corners twice)
	vector.DrawFilledRect(currentScreen, leftX, topY+1, 1, rectH-2, actualColor, false)
	// Right vertical line (height adjusted to avoid drawing corners twice)
	vector.DrawFilledRect(currentScreen, rightX, topY+1, 1, rectH-2, actualColor, false)

	// Mark shadow buffer as dirty so Pget() will sync from GPU
	MarkShadowBufferDirtyFromSprite()

	/* // Original StrokeRect implementation - might clip at edges
	strokeWidth := float32(1.0) // PICO-8 rect outline is 1 pixel thick
	vector.StrokeRect(
		currentScreen,
		rectX,
		rectY,
		rectW,
		rectH,
		strokeWidth,
		actualColor,
		false, // No anti-aliasing
	)
	*/
}

// Rectfill draws a filled rectangle using two corner points.
// Mimics PICO-8's rectfill(x1, y1, x2, y2, [color]) function.
//
// x1, y1, x2, y2: Coordinates of two opposing corners (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Rectfill[X1 Number, Y1 Number, X2 Number, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Rectfill() called before screen was ready.")
		return
	}

	fx1, fy1, fx2, fy2 := float64(x1), float64(y1), float64(x2), float64(y2)

	// Apply camera offset
	fx1, fy1 = applyCameraOffset(fx1, fy1)
	fx2, fy2 = applyCameraOffset(fx2, fy2)

	// Round to nearest integer for pixel-perfect rendering
	fx1 = math.Round(fx1)
	fy1 = math.Round(fy1)
	fx2 = math.Round(fx2)
	fy2 = math.Round(fy2)

	rectX, rectY, rectW, rectH, drawColorIndex, ok := parseRectArgs(fx1, fy1, fx2, fy2, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	originalColorIndex := drawColorIndex

	// Validate originalColorIndex against DrawPaletteMap bounds
	if len(drawPaletteMap) == 0 || originalColorIndex < 0 || originalColorIndex >= len(drawPaletteMap) {
		log.Printf("Warning: Rectfill() DrawPaletteMap not ready or originalColorIndex %d invalid for map size %d. Ignoring.", originalColorIndex, len(drawPaletteMap))
		return
	}
	mappedColorIndex := drawPaletteMap[originalColorIndex]

	// Validate mappedColorIndex against Pico8Palette and PaletteTransparency bounds
	if mappedColorIndex < 0 || mappedColorIndex >= len(pico8Palette) || mappedColorIndex >= len(paletteTransparency) {
		log.Printf("Warning: Rectfill() mappedColorIndex %d is out of bounds for Pico8Palette (size %d) or PaletteTransparency (size %d). Original: %d. Ignoring.", mappedColorIndex, len(pico8Palette), len(paletteTransparency), originalColorIndex)
		return
	}

	// Check if the mapped color is transparent
	if paletteTransparency[mappedColorIndex] {
		return // Don't draw transparent rectangles
	}

	// Get the actual color from the palette using the mapped index
	actualColor := pico8Palette[mappedColorIndex]

	// Draw filled rectangle using Ebitengine vector graphics
	vector.DrawFilledRect(
		currentScreen,
		rectX,
		rectY,
		rectW,
		rectH,
		actualColor,
		false, // No anti-aliasing
	)

	// Mark shadow buffer as dirty so Pget() will sync from GPU
	MarkShadowBufferDirtyFromSprite()
}

// parseLineArgs parses common arguments for Line function.
// It returns the PICO-8 color index to use and whether parsing was successful.
func parseLineArgs(options []interface{}) (int, bool) {
	drawColorIndex := parseColorArg("Line", false, options)

	if len(options) > 1 {
		logWarningOnce("Warning: Line called with too many arguments (%d), expected max 5.", len(options)+4)
	}

	return drawColorIndex, true
}

// Line draws a line between two points.
// Mimics PICO-8's line(x1, y1, x2, y2, [color]) function.
//
// x1, y1: Coordinates of the starting point (any Number type).
// x2, y2: Coordinates of the ending point (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Line[X1 Number, Y1 Number, X2 Number, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Line() called before screen was ready.")
		return
	}

	// Convert to float64 for calculations
	fx1, fy1, fx2, fy2 := float64(x1), float64(y1), float64(x2), float64(y2)

	// Round to nearest integer for pixel-perfect rendering
	fx1 = math.Round(fx1)
	fy1 = math.Round(fy1)
	fx2 = math.Round(fx2)
	fy2 = math.Round(fy2)

	// Parse optional color argument
	drawColorIndex, ok := parseLineArgs(options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(pico8Palette) {
		actualColor = pico8Palette[drawColorIndex]
	} else {
		actualColor = pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Line(). Defaulting to black.", drawColorIndex)
	}

	// Draw the line using Ebitengine's vector package
	vector.StrokeLine(
		currentScreen,
		float32(fx1),
		float32(fy1),
		float32(fx2),
		float32(fy2),
		1.0, // Line width of 1 pixel to match PICO-8
		actualColor,
		false, // No anti-aliasing to match PICO-8's pixel-perfect style
	)

	// Mark shadow buffer as dirty so Pget() will sync from GPU
	MarkShadowBufferDirtyFromSprite()
}

// parseCircArgs parses common arguments for Circ and Circfill.
// It returns the center coordinates (x, y), radius, the PICO-8 color index to use,
// and whether parsing was successful.
func parseCircArgs(x, y, radius float64, options []interface{}) (float32, float32, float32, int, bool) {
	drawColorIndex := parseColorArg("Circ/Circfill", true, options)

	if len(options) > 1 {
		logWarningOnce("Warning: Circ/Circfill called with too many arguments (%d), expected max 4.", len(options)+3)
	}

	return float32(x), float32(y), float32(radius), drawColorIndex, true
}

// Circ draws an outline circle.
// Mimics PICO-8's circ(x, y, radius, [color]) function.
//
// x, y: Coordinates of the center point (any Number type).
// radius: Radius of the circle (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Circ[X Number, Y Number, R Number](x X, y Y, radius R, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Circ() called before screen was ready.")
		return
	}

	fx, fy, fr := float64(x), float64(y), float64(radius)

	// Apply camera offset
	fx, fy = applyCameraOffset(fx, fy)

	// Round to nearest integer for pixel-perfect rendering
	fx = math.Round(fx)
	fy = math.Round(fy)

	circX, circY, circR, drawColorIndex, ok := parseCircArgs(fx, fy, fr, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(pico8Palette) {
		actualColor = pico8Palette[drawColorIndex]
	} else {
		actualColor = pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Circ(). Defaulting to black.", drawColorIndex)
	}

	// Draw the circle outline using Ebitengine vector graphics
	vector.StrokeCircle(
		currentScreen,
		circX,
		circY,
		circR,
		1.0, // Stroke width of 1 pixel to match PICO-8's style
		actualColor,
		false, // No anti-aliasing to match PICO-8's pixel-perfect style
	)

	// Mark shadow buffer as dirty so Pget() will sync from GPU
	MarkShadowBufferDirtyFromSprite()
}

// Circfill draws a filled circle.
// Mimics PICO-8's circfill(x, y, radius, [color]) function.
//
// x, y: Coordinates of the center point (any Number type).
// radius: Radius of the circle (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Circfill[X Number, Y Number, R Number](x X, y Y, radius R, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Circfill() called before screen was ready.")
		return
	}

	fx, fy, fr := float64(x), float64(y), float64(radius)

	// Apply camera offset
	fx, fy = applyCameraOffset(fx, fy)

	// Round to nearest integer for pixel-perfect rendering
	fx = math.Round(fx)
	fy = math.Round(fy)

	circX, circY, circR, drawColorIndex, ok := parseCircArgs(fx, fy, fr, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(pico8Palette) {
		actualColor = pico8Palette[drawColorIndex]
	} else {
		actualColor = pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Circfill(). Defaulting to black.", drawColorIndex)
	}

	// Draw the filled circle using Ebitengine vector graphics
	vector.DrawFilledCircle(
		currentScreen,
		circX,
		circY,
		circR,
		actualColor,
		false, // No anti-aliasing to match PICO-8's pixel-perfect style
	)

	// Mark shadow buffer as dirty so Pget() will sync from GPU
	MarkShadowBufferDirtyFromSprite()
}
