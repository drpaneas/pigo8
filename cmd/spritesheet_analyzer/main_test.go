package main

import (
	"testing"
)

func TestAnalyzeSprite(t *testing.T) {
	tests := []struct {
		name         string
		sprite       SpriteData
		expectEmpty  bool
		expectPixels int
		expectUsed   bool
	}{
		{
			name: "empty sprite - all zeros",
			sprite: SpriteData{
				ID:   1,
				Used: true,
				Pixels: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
			},
			expectEmpty:  true,
			expectPixels: 0,
			expectUsed:   true,
		},
		{
			name: "non-empty sprite",
			sprite: SpriteData{
				ID:   2,
				Used: true,
				Pixels: [][]int{
					{0, 1, 0, 0},
					{0, 0, 2, 0},
				},
			},
			expectEmpty:  false,
			expectPixels: 2,
			expectUsed:   true,
		},
		{
			name: "unused non-empty sprite",
			sprite: SpriteData{
				ID:   3,
				Used: false,
				Pixels: [][]int{
					{5, 5, 5, 5},
					{5, 5, 5, 5},
				},
			},
			expectEmpty:  false,
			expectPixels: 8,
			expectUsed:   false,
		},
		{
			name: "sprite with nil pixels",
			sprite: SpriteData{
				ID:     4,
				Used:   true,
				Pixels: nil,
			},
			expectEmpty:  true,
			expectPixels: 0,
			expectUsed:   true,
		},
		{
			name: "sprite with empty rows",
			sprite: SpriteData{
				ID:     5,
				Used:   true,
				Pixels: [][]int{},
			},
			expectEmpty:  true,
			expectPixels: 0,
			expectUsed:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeSprite(tt.sprite)

			if result.ID != tt.sprite.ID {
				t.Errorf("ID = %d, expected %d", result.ID, tt.sprite.ID)
			}
			if result.IsEmpty != tt.expectEmpty {
				t.Errorf("IsEmpty = %v, expected %v", result.IsEmpty, tt.expectEmpty)
			}
			if result.PixelCount != tt.expectPixels {
				t.Errorf("PixelCount = %d, expected %d", result.PixelCount, tt.expectPixels)
			}
			if result.Used != tt.expectUsed {
				t.Errorf("Used = %v, expected %v", result.Used, tt.expectUsed)
			}
			// Non-empty sprites should have a hash
			if !tt.expectEmpty && result.Hash == "" {
				t.Error("Expected hash for non-empty sprite, got empty string")
			}
			// Empty sprites should not have a hash
			if tt.expectEmpty && result.Hash != "" {
				t.Error("Expected empty hash for empty sprite, got:", result.Hash)
			}
		})
	}
}

func TestGenerateSpriteHash(t *testing.T) {
	// Test that identical pixels produce identical hashes
	pixels1 := [][]int{{1, 2}, {3, 4}}
	pixels2 := [][]int{{1, 2}, {3, 4}}
	pixels3 := [][]int{{1, 2}, {3, 5}} // Different

	hash1 := generateSpriteHash(pixels1)
	hash2 := generateSpriteHash(pixels2)
	hash3 := generateSpriteHash(pixels3)

	if hash1 != hash2 {
		t.Errorf("Identical pixels should produce identical hashes, got %s vs %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("Different pixels should produce different hashes, got identical %s", hash1)
	}

	// Test hash format (MD5 produces 32 hex characters)
	if len(hash1) != 32 {
		t.Errorf("Expected 32 character MD5 hash, got length %d", len(hash1))
	}
}

func TestIsEmptySprite(t *testing.T) {
	tests := []struct {
		name   string
		sprite SpriteData
		expect bool
	}{
		{
			name: "all zeros - empty",
			sprite: SpriteData{
				Pixels: [][]int{{0, 0}, {0, 0}},
			},
			expect: true,
		},
		{
			name: "has non-zero pixel - not empty",
			sprite: SpriteData{
				Pixels: [][]int{{0, 1}, {0, 0}},
			},
			expect: false,
		},
		{
			name: "nil pixels - empty",
			sprite: SpriteData{
				Pixels: nil,
			},
			expect: true,
		},
		{
			name: "empty pixel array - empty",
			sprite: SpriteData{
				Pixels: [][]int{},
			},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmptySprite(tt.sprite)
			if result != tt.expect {
				t.Errorf("isEmptySprite() = %v, expected %v", result, tt.expect)
			}
		})
	}
}

func TestHashConsistency(t *testing.T) {
	// Test that the hash is deterministic across multiple calls
	pixels := [][]int{{10, 20, 30}, {40, 50, 60}, {70, 80, 90}}

	hashes := make([]string, 10)
	for i := 0; i < 10; i++ {
		hashes[i] = generateSpriteHash(pixels)
	}

	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Hash inconsistency on iteration %d: got %s, expected %s", i, hashes[i], hashes[0])
		}
	}
}

func TestSpriteAnalysisStruct(t *testing.T) {
	// Test that SpriteAnalysis struct is properly initialized
	analysis := SpriteAnalysis{
		ID:         42,
		IsEmpty:    false,
		Hash:       "testhash",
		Used:       true,
		PixelCount: 100,
	}

	if analysis.ID != 42 {
		t.Errorf("ID = %d, expected 42", analysis.ID)
	}
	if analysis.IsEmpty != false {
		t.Error("IsEmpty should be false")
	}
	if analysis.Hash != "testhash" {
		t.Errorf("Hash = %s, expected testhash", analysis.Hash)
	}
	if analysis.Used != true {
		t.Error("Used should be true")
	}
	if analysis.PixelCount != 100 {
		t.Errorf("PixelCount = %d, expected 100", analysis.PixelCount)
	}
}
