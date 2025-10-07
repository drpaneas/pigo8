// Package main provides a spritesheet analyzer tool for analyzing sprite data and optimization.
package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// SpritesheetData represents the structure of a spritesheet.json file
type SpritesheetData struct {
	SpriteSheetColumns int          `json:"SpriteSheetColumns"`
	SpriteSheetRows    int          `json:"SpriteSheetRows"`
	SpriteSheetWidth   int          `json:"SpriteSheetWidth"`
	SpriteSheetHeight  int          `json:"SpriteSheetHeight"`
	Sprites            []SpriteData `json:"sprites"`
}

// SpriteData represents a single sprite
type SpriteData struct {
	ID     int      `json:"id"`
	X      int      `json:"x"`
	Y      int      `json:"y"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Used   bool     `json:"used"`
	Flags  FlagData `json:"flags"`
	Pixels [][]int  `json:"pixels"`
}

// FlagData represents sprite flags
type FlagData struct {
	Bitfield   int    `json:"bitfield"`
	Individual []bool `json:"individual"`
}

// SpriteAnalysis holds analysis results for a sprite
type SpriteAnalysis struct {
	ID         int
	IsEmpty    bool
	Hash       string
	Used       bool
	PixelCount int
}

// AnalysisResult holds the complete analysis
type AnalysisResult struct {
	TotalSprites        int
	TheoreticalTotal    int // Full sheet capacity (rows × cols)
	EmptySprites        []int
	UnusedSprites       []int
	DuplicateGroups     map[string][]int
	UniqueSprites       []int
	UniqueLoaded        int // Count of unique sprites actually loaded
	DuplicatesMapped    int // Count of duplicates mapped to existing sprites
	EmptiesMappedTo0    int // Count of empty sprites mapped to ID 0
	MemorySavings       string
	OptimizationSummary string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <spritesheet.json>")
		os.Exit(1)
	}

	filename := os.Args[1]
	result := analyzeSpritesheet(filename)
	printAnalysis(result, filename)

	// Optionally create optimized version
	if len(os.Args) > 2 && os.Args[2] == "--optimize" {
		createOptimizedSpritesheet(filename, result)
	}
}

func analyzeSpritesheet(filename string) AnalysisResult {
	// Read and parse spritesheet
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", filename, err)
	}

	var sheet SpritesheetData
	if err := json.Unmarshal(data, &sheet); err != nil {
		log.Fatalf("Failed to parse %s: %v", filename, err)
	}

	fmt.Printf("Analyzing %s...\n", filename)
	fmt.Printf("Spritesheet size: %dx%d (%d sprites)\n",
		sheet.SpriteSheetColumns, sheet.SpriteSheetRows, len(sheet.Sprites))

	// Analyze each sprite
	analyses := make([]SpriteAnalysis, len(sheet.Sprites))
	hashToSprites := make(map[string][]int)

	for i, sprite := range sheet.Sprites {
		analysis := analyzeSprite(sprite)
		analyses[i] = analysis

		// Group by hash for duplicate detection
		if !analysis.IsEmpty {
			hashToSprites[analysis.Hash] = append(hashToSprites[analysis.Hash], sprite.ID)
		}
	}

	// Compile results with proper counters
	result := AnalysisResult{
		TotalSprites:     len(sheet.Sprites),
		TheoreticalTotal: sheet.SpriteSheetRows * sheet.SpriteSheetColumns,
		DuplicateGroups:  make(map[string][]int),
	}

	// If no sheet dimensions, estimate from sprite count
	if result.TheoreticalTotal <= 0 {
		result.TheoreticalTotal = len(sheet.Sprites) // Conservative estimate
	}

	for _, analysis := range analyses {
		if analysis.IsEmpty {
			result.EmptySprites = append(result.EmptySprites, analysis.ID)
			result.EmptiesMappedTo0++
		} else if !analysis.Used {
			result.UnusedSprites = append(result.UnusedSprites, analysis.ID)
		}
	}

	// Find duplicates (groups with more than 1 sprite) and count properly
	for hash, spriteIDs := range hashToSprites {
		if len(spriteIDs) > 1 {
			result.DuplicateGroups[hash] = spriteIDs
			// Count duplicates: all but the first one are duplicates
			result.DuplicatesMapped += len(spriteIDs) - 1
			result.UniqueLoaded++ // One representative sprite
		} else {
			result.UniqueSprites = append(result.UniqueSprites, spriteIDs[0])
			result.UniqueLoaded++
		}
	}

	// Calculate accurate memory savings against theoretical total
	theoreticalReduction := result.TheoreticalTotal - result.UniqueLoaded
	theoreticalPercent := float64(theoreticalReduction) / float64(result.TheoreticalTotal) * 100

	// Calculate savings against actual sprites in file
	actualOptimizedSize := result.UniqueLoaded
	actualReduction := len(sheet.Sprites) - actualOptimizedSize
	actualPercent := float64(actualReduction) / float64(len(sheet.Sprites)) * 100

	result.MemorySavings = fmt.Sprintf("%d → %d sprites (%.1f%% reduction from file)",
		len(sheet.Sprites), actualOptimizedSize, actualPercent)

	result.OptimizationSummary = fmt.Sprintf("Theoretical: %d → %d (%.1f%% reduction), Unique: %d, Duplicates: %d, Empties: %d",
		result.TheoreticalTotal, result.UniqueLoaded, theoreticalPercent,
		result.UniqueLoaded, result.DuplicatesMapped, result.EmptiesMappedTo0)

	return result
}

func analyzeSprite(sprite SpriteData) SpriteAnalysis {
	analysis := SpriteAnalysis{
		ID:   sprite.ID,
		Used: sprite.Used,
	}

	// Check if sprite is empty (all pixels are 0)
	isEmpty := true
	pixelCount := 0

	for _, row := range sprite.Pixels {
		for _, pixel := range row {
			if pixel != 0 {
				isEmpty = false
				pixelCount++
			}
		}
	}

	analysis.IsEmpty = isEmpty
	analysis.PixelCount = pixelCount

	// Generate hash for duplicate detection (only for non-empty sprites)
	if !isEmpty {
		analysis.Hash = generateSpriteHash(sprite.Pixels)
	}

	return analysis
}

func generateSpriteHash(pixels [][]int) string {
	// Use efficient byte operations instead of string concatenation to avoid GC pressure
	hasher := md5.New()

	// Write pixel data directly as bytes to avoid string allocation
	for _, row := range pixels {
		for _, pixel := range row {
			// Write each pixel as 4 bytes (int32) using binary representation
			b := make([]byte, 4)
			b[0] = byte(pixel)
			b[1] = byte(pixel >> 8)
			b[2] = byte(pixel >> 16)
			b[3] = byte(pixel >> 24)
			hasher.Write(b)
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func printAnalysis(result AnalysisResult, filename string) {
	fmt.Printf("\n=== SPRITESHEET ANALYSIS: %s ===\n", filepath.Base(filename))
	fmt.Printf("Total sprites in file: %d\n", result.TotalSprites)
	fmt.Printf("Theoretical sheet capacity: %d\n", result.TheoreticalTotal)
	fmt.Printf("Empty sprites: %d\n", len(result.EmptySprites))
	fmt.Printf("Unused sprites: %d\n", len(result.UnusedSprites))
	fmt.Printf("Duplicate groups: %d\n", len(result.DuplicateGroups))
	fmt.Printf("Unique sprites loaded: %d\n", result.UniqueLoaded)
	fmt.Printf("Duplicates mapped: %d\n", result.DuplicatesMapped)
	fmt.Printf("Empties mapped to 0: %d\n", result.EmptiesMappedTo0)
	fmt.Printf("Memory savings: %s\n", result.MemorySavings)
	fmt.Printf("Optimization summary: %s\n", result.OptimizationSummary)

	if len(result.EmptySprites) > 0 {
		fmt.Printf("\nEmpty sprites (first 20): ")
		for i, id := range result.EmptySprites {
			if i >= 20 {
				fmt.Printf("... and %d more", len(result.EmptySprites)-20)
				break
			}
			fmt.Printf("%d ", id)
		}
		fmt.Println()
	}

	if len(result.DuplicateGroups) > 0 {
		fmt.Printf("\nDuplicate groups (first 10):\n")
		count := 0
		for hash, spriteIDs := range result.DuplicateGroups {
			if count >= 10 {
				fmt.Printf("... and %d more groups\n", len(result.DuplicateGroups)-10)
				break
			}
			fmt.Printf("  Hash %s: sprites %v\n", hash[:8], spriteIDs)
			count++
		}
	}
}

func createOptimizedSpritesheet(originalFile string, result AnalysisResult) {
	fmt.Printf("\nCreating optimized spritesheet...\n")

	// Read original file
	data, err := os.ReadFile(originalFile)
	if err != nil {
		log.Fatalf("Failed to read original file: %v", err)
	}

	var sheet SpritesheetData
	if err := json.Unmarshal(data, &sheet); err != nil {
		log.Fatalf("Failed to parse original file: %v", err)
	}

	// Create optimized sprite list
	var optimizedSprites []SpriteData
	spriteMap := make(map[int]SpriteData)

	// Add original sprites to map
	for _, sprite := range sheet.Sprites {
		spriteMap[sprite.ID] = sprite
	}

	// Add unique sprites
	for _, id := range result.UniqueSprites {
		if sprite, exists := spriteMap[id]; exists && sprite.Used && !isEmptySprite(sprite) {
			optimizedSprites = append(optimizedSprites, sprite)
		}
	}

	// Add one representative from each duplicate group
	for _, spriteIDs := range result.DuplicateGroups {
		if len(spriteIDs) > 0 {
			id := spriteIDs[0] // Take the first one as representative
			if sprite, exists := spriteMap[id]; exists && sprite.Used && !isEmptySprite(sprite) {
				optimizedSprites = append(optimizedSprites, sprite)
			}
		}
	}

	// Sort by ID
	sort.Slice(optimizedSprites, func(i, j int) bool {
		return optimizedSprites[i].ID < optimizedSprites[j].ID
	})

	// Create new spritesheet
	optimizedSheet := SpritesheetData{
		SpriteSheetColumns: sheet.SpriteSheetColumns,
		SpriteSheetRows:    sheet.SpriteSheetRows,
		SpriteSheetWidth:   sheet.SpriteSheetWidth,
		SpriteSheetHeight:  sheet.SpriteSheetHeight,
		Sprites:            optimizedSprites,
	}

	// Write optimized file
	optimizedFile := fmt.Sprintf("%s_optimized.json",
		originalFile[:len(originalFile)-5]) // Remove .json extension

	optimizedData, err := json.MarshalIndent(optimizedSheet, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal optimized data: %v", err)
	}

	if err := os.WriteFile(optimizedFile, optimizedData, 0644); err != nil {
		log.Fatalf("Failed to write optimized file: %v", err)
	}

	fmt.Printf("Optimized spritesheet saved as: %s\n", optimizedFile)
	fmt.Printf("Original size: %.1f KB\n", float64(len(data))/1024)
	fmt.Printf("Optimized size: %.1f KB\n", float64(len(optimizedData))/1024)
	fmt.Printf("File size reduction: %.1f%%\n",
		float64(len(data)-len(optimizedData))/float64(len(data))*100)
}

func isEmptySprite(sprite SpriteData) bool {
	for _, row := range sprite.Pixels {
		for _, pixel := range row {
			if pixel != 0 {
				return false
			}
		}
	}
	return true
}
