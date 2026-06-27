package pigo8

import (
	"bufio"
	"fmt"
	"image/color"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// loadPaletteFromHexFile attempts to load a palette from a palette.hex file
// in the current directory, common locations, or embedded resources.
// Returns true if a palette was successfully loaded, false otherwise.
func loadPaletteFromHexFile() bool {
	const paletteFilename = "palette.hex"

	// First try to load from the current directory
	data, err := os.ReadFile(paletteFilename)
	if err != nil {
		// Check common alternative locations
		commonLocations := []string{
			filepath.Join("assets", paletteFilename),
			filepath.Join("resources", paletteFilename),
			filepath.Join("data", paletteFilename),
			filepath.Join("static", paletteFilename),
		}

		for _, location := range commonLocations {
			data, err = os.ReadFile(location)
			if err == nil {
				log.Printf("Loaded palette from %s", location)
				break
			}
		}

		// If still not found, try embedded resources
		if err != nil {
			embeddedData, embErr := tryLoadEmbeddedPalette()
			if embErr != nil {
				// No embedded palette found
				return false
			}
			data = embeddedData
		}
	} else {
		log.Printf("Using palette file from current directory: %s", paletteFilename)
	}

	// Process the palette data
	colors, err := parseHexPalette(data)
	if err != nil {
		log.Printf("Error parsing palette file: %v", err)
		return false
	}

	// Create a new palette with a transparent color at index 0 and white at index 1
	reservedColors := 2
	newPalette := make([]color.Color, len(colors)+reservedColors) // +2 for transparent and white
	newPalette[0] = color.RGBA{0, 0, 0, 0}                        // Transparent black at index 0
	newPalette[1] = color.RGBA{255, 255, 255, 255}                // Opaque white at index 1

	// Add the colors from the file, shifted by reservedColors
	for i, c := range colors {
		newPalette[i+reservedColors] = c
	}

	// Set the new palette
	SetPalette(newPalette)

	// Ensure color 0 is transparent
	paletteTransparency[0] = true
	// By default, other colors including index 1 (white) will be opaque unless specified otherwise

	// Log when palette is loaded
	log.Printf("Custom palette loaded: %d colors (index 0 transparent, index 1 white)", len(newPalette))

	return true
}

// tryLoadEmbeddedPalette attempts to load a palette from embedded resources
func tryLoadEmbeddedPalette() ([]byte, error) {
	// Auto-detect resources if not already done
	if !autoDetectResourcesAttempted {
		autoDetectResources()
	}

	// Try custom resources if registered
	resources := getCustomResources()
	if resources != nil && resources.PalettePath != "" {
		data, err := fs.ReadFile(resources.FS, resources.PalettePath)
		if err == nil {
			log.Printf("Using embedded palette file: %s", resources.PalettePath)
			return data, nil
		}
	}

	// No embedded palette found
	return nil, fmt.Errorf("no embedded palette file found")
}

// parseHexPalette parses a byte slice containing hex color codes (one per line)
// and returns a slice of color.Color.
func parseHexPalette(data []byte) ([]color.Color, error) {
	var colors []color.Color

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Remove any # prefix if present
		line = strings.TrimPrefix(line, "#")

		// Parse the hex color
		c, err := hexToColor(line)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color '%s': %v", line, err)
		}

		colors = append(colors, c)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning palette file: %v", err)
	}

	return colors, nil
}

// hexToColor converts a hex string to a color.RGBA.
func hexToColor(hex string) (color.Color, error) {
	// Remove any # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Ensure the hex string is the correct length
	if len(hex) != 6 {
		return nil, fmt.Errorf("hex color must be 6 characters long")
	}

	// Parse the hex values
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid red component: %v", err)
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid green component: %v", err)
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid blue component: %v", err)
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255, // Full opacity
	}, nil
}
