package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSpriteNavigationCandidateNormalMove covers ordinary moves within the
// grid that don't touch a boundary or the reserved sprite 0.
func TestSpriteNavigationCandidateNormalMove(t *testing.T) {
	// Sprite 5 is at row 0, col 5 (spriteSheetCols == 32).
	assert.Equal(t, 6, spriteNavigationCandidate(5, 0, 1), "right")
	assert.Equal(t, 4, spriteNavigationCandidate(5, 0, -1), "left")
	assert.Equal(t, 5+spriteSheetCols, spriteNavigationCandidate(5, 1, 0), "down")
}

// TestSpriteNavigationCandidateRejectsSpriteZero is a regression test: sprite
// 0 is reserved (see Init) and must never become selectable via keyboard
// navigation, matching the mouse-click behavior in handleSpriteSelection
// (which already refuses `idx == 0`). Before this fix, navigating left from
// sprite 1, or up from sprite spriteSheetCols, landed on sprite 0.
func TestSpriteNavigationCandidateRejectsSpriteZero(t *testing.T) {
	t.Run("left from sprite 1 stays put instead of landing on 0", func(t *testing.T) {
		got := spriteNavigationCandidate(1, 0, -1)
		assert.Equal(t, 1, got, "should refuse to move onto reserved sprite 0")
	})

	t.Run("up from the first column of row 1 stays put instead of landing on 0", func(t *testing.T) {
		got := spriteNavigationCandidate(spriteSheetCols, -1, 0)
		assert.Equal(t, spriteSheetCols, got, "should refuse to move onto reserved sprite 0")
	})
}

// TestSpriteNavigationCandidateRejectsOutOfBounds confirms grid-edge moves
// are no-ops rather than wrapping or landing outside the spritesheet.
func TestSpriteNavigationCandidateRejectsOutOfBounds(t *testing.T) {
	t.Run("left from column 0 stays put", func(t *testing.T) {
		// Sprite 1 is row 0, col 1; use row 1's first column (spriteSheetCols) instead
		// so the "column 0" boundary is exercised without also hitting sprite 0.
		current := spriteSheetCols // row 1, col 0
		got := spriteNavigationCandidate(current, 0, -1)
		assert.Equal(t, current, got)
	})

	t.Run("right from the last column stays put", func(t *testing.T) {
		current := spriteSheetCols - 1 // row 0, last col
		got := spriteNavigationCandidate(current, 0, 1)
		assert.Equal(t, current, got)
	})

	t.Run("up from row 0 stays put", func(t *testing.T) {
		got := spriteNavigationCandidate(5, -1, 0)
		assert.Equal(t, 5, got)
	})

	t.Run("down from the last row stays put", func(t *testing.T) {
		current := (spriteSheetRows-1)*spriteSheetCols + 3
		got := spriteNavigationCandidate(current, 1, 0)
		assert.Equal(t, current, got)
	})
}
