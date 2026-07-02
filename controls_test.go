package pigo8

import (
	"slices"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

// Note: Unit testing Btn and Btnp fully is difficult due to dependencies
// on Ebitengine's input handling and connected gamepads. These tests
// primarily focus on playerIndex validation.

func TestBtnPlayerIndexValidation(t *testing.T) {
	// We don't need to mock actual gamepad state for this validation
	// The functions should return false early if playerIndex is invalid.

	button := O // Arbitrary button

	t.Run("Valid player indices (0-7)", func(t *testing.T) {
		for i := 0; i <= 7; i++ {
			// We expect false because no gamepad is connected/mocked,
			// but the function shouldn't fail due to the index itself.
			assert.False(t, Btn(button, i), "Expected false for valid player index %d (no gamepad)", i)
		}
	})

	t.Run("Invalid player index (< 0)", func(t *testing.T) {
		assert.False(t, Btn(button, -1), "Expected false for player index -1")
	})

	t.Run("Invalid player index (> 7)", func(t *testing.T) {
		assert.False(t, Btn(button, 8), "Expected false for player index 8")
		assert.False(t, Btn(button, 99), "Expected false for player index 99")
	})

	t.Run("Default player index (0)", func(t *testing.T) {
		assert.False(t, Btn(button), "Expected false for default player index 0 (no gamepad)")
	})
}

// withConnectedGamepads temporarily sets connectedGamepadIDsSorted (and the
// backing map, so updateConnectedGamepads doesn't clobber it if called
// during the test) to ids, restoring the previous state on cleanup.
func withConnectedGamepads(t *testing.T, ids ...ebiten.GamepadID) {
	t.Helper()

	prevSorted := connectedGamepadIDsSorted
	prevMap := connectedGamepadIDs

	connectedGamepadIDsSorted = append([]ebiten.GamepadID(nil), ids...)
	slices.Sort(connectedGamepadIDsSorted)
	connectedGamepadIDs = make(map[ebiten.GamepadID]struct{}, len(ids))
	for _, id := range ids {
		connectedGamepadIDs[id] = struct{}{}
	}

	t.Cleanup(func() {
		connectedGamepadIDsSorted = prevSorted
		connectedGamepadIDs = prevMap
	})
}

func TestGamepadForPlayer(t *testing.T) {
	t.Run("no gamepads connected", func(t *testing.T) {
		withConnectedGamepads(t)
		_, ok := gamepadForPlayer(0)
		assert.False(t, ok, "expected no gamepad for player 0 with none connected")
	})

	t.Run("assigns gamepads in ascending ID order", func(t *testing.T) {
		// Deliberately out of order to confirm sorting, not insertion order, wins.
		withConnectedGamepads(t, 5, 2, 9)

		id0, ok0 := gamepadForPlayer(0)
		assert.True(t, ok0)
		assert.Equal(t, ebiten.GamepadID(2), id0, "player 0 should get the lowest gamepad ID")

		id1, ok1 := gamepadForPlayer(1)
		assert.True(t, ok1)
		assert.Equal(t, ebiten.GamepadID(5), id1, "player 1 should get the second-lowest gamepad ID")

		id2, ok2 := gamepadForPlayer(2)
		assert.True(t, ok2)
		assert.Equal(t, ebiten.GamepadID(9), id2, "player 2 should get the third gamepad ID")
	})

	t.Run("player index beyond connected gamepad count", func(t *testing.T) {
		withConnectedGamepads(t, 1)
		_, ok := gamepadForPlayer(1)
		assert.False(t, ok, "expected no gamepad for player 1 with only one connected")
	})

	t.Run("negative player index", func(t *testing.T) {
		withConnectedGamepads(t, 1)
		_, ok := gamepadForPlayer(-1)
		assert.False(t, ok, "expected no gamepad for a negative player index")
	})
}

func TestCheckButtonStateKeyboardAndMouseScopedToPlayerZero(t *testing.T) {
	withConnectedGamepads(t) // no gamepads, isolates keyboard/mouse behavior

	// Keyboard-backed button (O maps to the Z key): checkButtonState can't
	// simulate a real key press, but it must not attribute keyboard input to
	// any player other than 0 - i.e. player 1+ must short-circuit to the
	// (empty) gamepad path rather than ever consulting the keyboard.
	assert.False(t, checkButtonState(O, 1), "player 1 must not read keyboard input")
	assert.False(t, checkButtonState(O, 7), "player 7 must not read keyboard input")

	// Mouse buttons are shared across all player indices, including ones
	// with no assigned gamepad.
	for player := 0; player < maxLocalPlayers; player++ {
		assert.False(t, checkButtonState(ButtonMouseLeft, player),
			"mouse button check should run (and return false, no click) for player %d", player)
	}
}

func TestBtnpPlayerIndexValidation(t *testing.T) {
	// Similar logic to TestBtnPlayerIndexValidation
	button := X // Arbitrary button

	t.Run("Valid player indices (0-7)", func(t *testing.T) {
		for i := 0; i <= 7; i++ {
			assert.False(t, Btnp(button, i), "Expected false for valid player index %d (no gamepad)", i)
		}
	})

	t.Run("Invalid player index (< 0)", func(t *testing.T) {
		assert.False(t, Btnp(button, -1), "Expected false for player index -1")
	})

	t.Run("Invalid player index (> 7)", func(t *testing.T) {
		assert.False(t, Btnp(button, 8), "Expected false for player index 8")
		assert.False(t, Btnp(button, 99), "Expected false for player index 99")
	})

	t.Run("Default player index (0)", func(t *testing.T) {
		assert.False(t, Btnp(button), "Expected false for default player index 0 (no gamepad)")
	})
}
