package pigo8

import (
	"fmt"
	"hash/fnv"
	"image/color"
	"os/exec"
	"strings"
	"testing"
)

// =============================================================================
// Benchmark 1: isSteamDeck caching
// =============================================================================

// BenchmarkIsSteamDeckUncached measures the cost of running the shell command
// every time (the old implementation).
func BenchmarkIsSteamDeckUncached(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Simulate the old uncached behavior
		cmd := exec.Command("uname", "--nodename")
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		_ = strings.TrimSpace(string(output)) == "steamdeck"
	}
}

// BenchmarkIsSteamDeckCached measures the cost of returning a cached result
// (the new implementation).
func BenchmarkIsSteamDeckCached(b *testing.B) {
	b.ReportAllocs()
	// First call initializes the cache
	_ = isSteamDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isSteamDeck()
	}
}

// =============================================================================
// Benchmark 2: Sprite hash generation (string vs uint64)
// =============================================================================

// createTestPixels creates a test pixel grid for benchmarking
func createTestPixels(width, height int) [][]int {
	pixels := make([][]int, height)
	for y := 0; y < height; y++ {
		pixels[y] = make([]int, width)
		for x := 0; x < width; x++ {
			pixels[y][x] = (x + y) % 16 // Simulate PICO-8 color indices
		}
	}
	return pixels
}

// generateSpriteHashString is the old string-based hash generation
func generateSpriteHashString(pixels [][]int, flags FlagsData) string {
	hasher := fnv.New64a()

	for _, row := range pixels {
		for _, pixel := range row {
			pixelBytes := [4]byte{
				byte(pixel),
				byte(pixel >> 8),
				byte(pixel >> 16),
				byte(pixel >> 24),
			}
			hasher.Write(pixelBytes[:])
		}
	}

	flagBytes := [4]byte{
		byte(flags.Bitfield),
		byte(flags.Bitfield >> 8),
		byte(flags.Bitfield >> 16),
		byte(flags.Bitfield >> 24),
	}
	hasher.Write(flagBytes[:])

	for _, flag := range flags.Individual {
		if flag {
			hasher.Write([]byte{1})
		} else {
			hasher.Write([]byte{0})
		}
	}

	return fmt.Sprintf("%x", hasher.Sum64())
}

// generateSpriteHashUint64 is the new uint64-based hash generation
func generateSpriteHashUint64(pixels [][]int, flags FlagsData) uint64 {
	hasher := fnv.New64a()

	for _, row := range pixels {
		for _, pixel := range row {
			pixelBytes := [4]byte{
				byte(pixel),
				byte(pixel >> 8),
				byte(pixel >> 16),
				byte(pixel >> 24),
			}
			hasher.Write(pixelBytes[:])
		}
	}

	flagBytes := [4]byte{
		byte(flags.Bitfield),
		byte(flags.Bitfield >> 8),
		byte(flags.Bitfield >> 16),
		byte(flags.Bitfield >> 24),
	}
	hasher.Write(flagBytes[:])

	for _, flag := range flags.Individual {
		if flag {
			hasher.Write([]byte{1})
		} else {
			hasher.Write([]byte{0})
		}
	}

	return hasher.Sum64()
}

// BenchmarkSpriteHashString measures the cost of string-based hashing (old implementation)
func BenchmarkSpriteHashString(b *testing.B) {
	pixels := createTestPixels(8, 8)
	flags := FlagsData{
		Bitfield:   5,
		Individual: []bool{true, false, true, false, false, false, false, false},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSpriteHashString(pixels, flags)
	}
}

// BenchmarkSpriteHashUint64 measures the cost of uint64-based hashing (new implementation)
func BenchmarkSpriteHashUint64(b *testing.B) {
	pixels := createTestPixels(8, 8)
	flags := FlagsData{
		Bitfield:   5,
		Individual: []bool{true, false, true, false, false, false, false, false},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSpriteHashUint64(pixels, flags)
	}
}

// BenchmarkSpriteHashMapStringKey measures map operations with string keys
func BenchmarkSpriteHashMapStringKey(b *testing.B) {
	pixels := createTestPixels(8, 8)
	flags := FlagsData{
		Bitfield:   5,
		Individual: []bool{true, false, true, false, false, false, false, false},
	}

	// Pre-populate map with 256 entries (simulate spritesheet)
	stringMap := make(map[string]int, 256)
	for i := 0; i < 256; i++ {
		testPixels := createTestPixels(8, 8)
		testPixels[0][0] = i // Make each unique
		hash := generateSpriteHashString(testPixels, flags)
		stringMap[hash] = i
	}

	targetHash := generateSpriteHashString(pixels, flags)

	var sink int
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if v, ok := stringMap[targetHash]; ok {
			sink = v
		}
	}
	_ = sink // Prevent compiler optimization
}

// BenchmarkSpriteHashMapUint64Key measures map operations with uint64 keys
func BenchmarkSpriteHashMapUint64Key(b *testing.B) {
	pixels := createTestPixels(8, 8)
	flags := FlagsData{
		Bitfield:   5,
		Individual: []bool{true, false, true, false, false, false, false, false},
	}

	// Pre-populate map with 256 entries (simulate spritesheet)
	uint64Map := make(map[uint64]int, 256)
	for i := 0; i < 256; i++ {
		testPixels := createTestPixels(8, 8)
		testPixels[0][0] = i // Make each unique
		hash := generateSpriteHashUint64(testPixels, flags)
		uint64Map[hash] = i
	}

	targetHash := generateSpriteHashUint64(pixels, flags)

	var sink int
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if v, ok := uint64Map[targetHash]; ok {
			sink = v
		}
	}
	_ = sink // Prevent compiler optimization
}

// =============================================================================
// Benchmark 3: Color lookup (linear search vs map lookup)
// =============================================================================

// linearColorLookup performs a linear search through the palette (old implementation)
func linearColorLookup(targetColor color.RGBA, palette []color.Color) int {
	for i, c := range palette {
		r1, g1, b1, a1 := targetColor.RGBA()
		r2, g2, b2, a2 := c.RGBA()
		if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
			return i
		}
	}
	return 0
}

// mapColorLookup performs an O(1) map lookup (new implementation)
func mapColorLookup(targetColor color.RGBA, colorMap map[color.Color]int) int {
	if index, ok := colorMap[targetColor]; ok {
		return index
	}
	return 0
}

// BenchmarkColorLookupLinear measures linear search performance
func BenchmarkColorLookupLinear(b *testing.B) {
	// Create test palette (PICO-8 16 colors)
	palette := []color.Color{
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

	// Test with worst case: last color in palette
	targetColor := color.RGBA{R: 255, G: 204, B: 170, A: 255} // peach (index 15)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = linearColorLookup(targetColor, palette)
	}
}

// BenchmarkColorLookupMap measures map lookup performance
func BenchmarkColorLookupMap(b *testing.B) {
	// Create test palette (PICO-8 16 colors)
	palette := []color.Color{
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

	// Build the color map
	colorMap := make(map[color.Color]int, len(palette))
	for i, c := range palette {
		colorMap[c] = i
	}

	// Test with worst case for linear: last color in palette
	targetColor := color.RGBA{R: 255, G: 204, B: 170, A: 255} // peach (index 15)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapColorLookup(targetColor, colorMap)
	}
}

// BenchmarkColorLookupLinearFirstColor measures linear search for first color (best case)
func BenchmarkColorLookupLinearFirstColor(b *testing.B) {
	palette := []color.Color{
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

	// Test with best case: first color in palette
	targetColor := color.RGBA{R: 0, G: 0, B: 0, A: 255} // black (index 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = linearColorLookup(targetColor, palette)
	}
}

// BenchmarkColorLookupMapFirstColor measures map lookup for first color
func BenchmarkColorLookupMapFirstColor(b *testing.B) {
	palette := []color.Color{
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

	// Build the color map
	colorMap := make(map[color.Color]int, len(palette))
	for i, c := range palette {
		colorMap[c] = i
	}

	// Test with first color in palette
	targetColor := color.RGBA{R: 0, G: 0, B: 0, A: 255} // black (index 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapColorLookup(targetColor, colorMap)
	}
}
