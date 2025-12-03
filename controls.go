// Package pigo8 package provides a set of functions to handle input for the PICO-8 fantasy console.
package pigo8

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PICO-8 Button Index Constants
const (
	// Directional buttons (keyboard and gamepad)
	LEFT = iota
	RIGHT
	UP
	DOWN

	// Face buttons (keyboard and gamepad)
	O // Often mapped to A/Cross on standard controllers
	X // Often mapped to B/Circle on standard controllers

	ButtonStart
	ButtonSelect

	// Mouse buttons
	ButtonMouseLeft
	ButtonMouseRight
	ButtonMouseMiddle // Mouse wheel press
	ButtonMouseWheelUp
	ButtonMouseWheelDown

	// Gamepad-specific buttons (for direct mapping)
	ButtonJoypadUp
	ButtonJoypadDown
	ButtonJoypadLeft
	ButtonJoypadRight
	ButtonJoyA
	ButtonJoypadB
	ButtonJoypadX
	ButtonJoypadY

	// Shoulder buttons and triggers
	ButtonJoypadL1 // Left shoulder button
	ButtonJoypadR1 // Right shoulder button
	ButtonJoypadL2 // Left trigger (analog)
	ButtonJoypadR2 // Right trigger (analog)

	// Stick clicks
	ButtonJoypadL3 // Left stick click
	ButtonJoypadR3 // Right stick click

	// Additional Steam Deck back buttons
	ButtonJoypadL4 // Back left button 1 (Steam Deck)
	ButtonJoypadR4 // Back right button 1 (Steam Deck)
	ButtonJoypadL5 // Back left button 2 (Steam Deck)
	ButtonJoypadR5 // Back right button 2 (Steam Deck)

	// Alias for pause button (same as START)
	ButtonPause = ButtonStart
)

// Steam Deck detection caching
var (
	isSteamDeckResult bool
	steamDeckOnce     sync.Once
)

// isSteamDeck checks if the game is running on a Steam Deck by checking the hostname.
// The result is cached after the first call to avoid repeated shell command execution.
func isSteamDeck() bool {
	steamDeckOnce.Do(func() {
		cmd := exec.Command("uname", "--nodename")
		output, err := cmd.Output()
		if err != nil {
			// Command failed, likely not a Linux system or uname not available
			isSteamDeckResult = false
			return
		}
		// Trim whitespace and check if the output is exactly "steamdeck"
		isSteamDeckResult = strings.TrimSpace(string(output)) == "steamdeck"
	})
	return isSteamDeckResult
}

// pico8ButtonToStandard maps PICO-8 button indices to Ebitengine Standard Gamepad Buttons.
var pico8ButtonToStandard map[int]ebiten.StandardGamepadButton

// initButtonMappings initializes the button and axis mappings based on the current platform
func initButtonMappings() {
	// Initialize with common button mappings
	pico8ButtonToStandard = map[int]ebiten.StandardGamepadButton{
		// Face buttons (Xbox-style: A=bottom, B=right, X=left, Y=top)
		O: ebiten.StandardGamepadButtonRightLeft,   // Left button (X on Xbox, Square on PlayStation)
		X: ebiten.StandardGamepadButtonRightBottom, // Bottom button (A on Xbox, X on PlayStation)

		// Menu buttons
		ButtonStart:  ebiten.StandardGamepadButtonCenterRight, // Start button
		ButtonSelect: ebiten.StandardGamepadButtonCenterLeft,  // Select/Back button

		// Gamepad-specific buttons
		ButtonJoypadUp:    ebiten.StandardGamepadButtonLeftTop,
		ButtonJoypadDown:  ebiten.StandardGamepadButtonLeftBottom,
		ButtonJoypadLeft:  ebiten.StandardGamepadButtonLeftLeft,
		ButtonJoypadRight: ebiten.StandardGamepadButtonLeftRight,
		ButtonJoyA:        ebiten.StandardGamepadButtonRightBottom, // A button (bottom face button)
		ButtonJoypadB:     ebiten.StandardGamepadButtonRightRight,  // B button (right face button)
		ButtonJoypadX:     ebiten.StandardGamepadButtonRightLeft,   // X button (left face button)
		ButtonJoypadY:     ebiten.StandardGamepadButtonRightTop,    // Y button (top face button)
	}

	// Set platform-specific overrides if needed
	if isSteamDeck() {
		// Steam Deck specific button mappings
		pico8ButtonToStandard = map[int]ebiten.StandardGamepadButton{
			// Face buttons (A,B,X,Y) - Steam Deck uses XBox layout
			X: ebiten.StandardGamepadButtonRightLeft,   // X button (Left)
			O: ebiten.StandardGamepadButtonRightBottom, // A button (Bottom)

			ButtonJoyA:    ebiten.StandardGamepadButtonRightBottom, // A button (bottom)
			ButtonJoypadB: ebiten.StandardGamepadButtonRightRight,  // B button (right)
			ButtonJoypadX: ebiten.StandardGamepadButtonRightLeft,   // X button (left)
			ButtonJoypadY: ebiten.StandardGamepadButtonRightTop,    // Y button (top)

			// D-pad directions
			UP:                ebiten.StandardGamepadButtonLeftTop,
			DOWN:              ebiten.StandardGamepadButtonLeftBottom,
			LEFT:              ebiten.StandardGamepadButtonLeftLeft,
			RIGHT:             ebiten.StandardGamepadButtonLeftRight,
			ButtonJoypadUp:    ebiten.StandardGamepadButtonLeftTop,
			ButtonJoypadDown:  ebiten.StandardGamepadButtonLeftBottom,
			ButtonJoypadLeft:  ebiten.StandardGamepadButtonLeftLeft,
			ButtonJoypadRight: ebiten.StandardGamepadButtonLeftRight,

			// Shoulder buttons
			ButtonJoypadL1: ebiten.StandardGamepadButtonFrontTopLeft,     // L1
			ButtonJoypadR1: ebiten.StandardGamepadButtonFrontTopRight,    // R1
			ButtonJoypadL2: ebiten.StandardGamepadButtonFrontBottomLeft,  // L2 (also analog)
			ButtonJoypadR2: ebiten.StandardGamepadButtonFrontBottomRight, // R2 (also analog)

			// Stick clicks
			ButtonJoypadL3: ebiten.StandardGamepadButtonLeftStick,  // Left stick click
			ButtonJoypadR3: ebiten.StandardGamepadButtonRightStick, // Right stick click

			// Menu buttons
			ButtonStart:  ebiten.StandardGamepadButtonCenterRight, // Menu button (right)
			ButtonSelect: ebiten.StandardGamepadButtonCenterLeft,  // View button (left)

			// Steam/Quick Access button is not mappable through standard gamepad API
		}

		// Map Steam Deck touchpad clicks
		// These are mapped to mouse buttons for compatibility
		pico8ButtonToStandard[ButtonMouseLeft] = ebiten.StandardGamepadButtonFrontBottomLeft   // Map to L2
		pico8ButtonToStandard[ButtonMouseRight] = ebiten.StandardGamepadButtonFrontBottomRight // Map to R2

		// Map back buttons (L4/L5, R4/R5 on Steam Deck)
		// These are mapped to L1/R1 for now since they're not standard
		pico8ButtonToStandard[ButtonJoypadL4] = ebiten.StandardGamepadButtonFrontTopLeft  // L4 -> L1
		pico8ButtonToStandard[ButtonJoypadR4] = ebiten.StandardGamepadButtonFrontTopRight // R4 -> R1
		pico8ButtonToStandard[ButtonJoypadL5] = ebiten.StandardGamepadButtonLeftStick     // L5 -> Left stick click
		pico8ButtonToStandard[ButtonJoypadR5] = ebiten.StandardGamepadButtonRightStick    // R5 -> Right stick click

		// Note: For full Steam Deck back button support, you might want to use
		// SDL's game controller API directly or a Steam Input wrapper
	}
}

// init initializes the button mappings when the package is imported
func init() {
	initButtonMappings()
}

// pico8ButtonToKeyboardP0 maps PICO-8 button indices to default keyboard keys for Player 0.
// Updated for better Steam Deck keyboard/on-screen keyboard support
var pico8ButtonToKeyboardP0 = map[int]ebiten.Key{
	// Arrow keys for direction
	LEFT:  ebiten.KeyLeft,
	RIGHT: ebiten.KeyRight,
	UP:    ebiten.KeyUp,
	DOWN:  ebiten.KeyDown,

	// Face buttons (mapped to common game keys)
	O: ebiten.KeyZ, // PICO-8 O button ('Z' key)
	X: ebiten.KeyX, // PICO-8 X button ('X' key)

	// Menu buttons
	ButtonStart:  ebiten.KeyEnter, // Start/Confirm
	ButtonSelect: ebiten.KeyTab,   // Select/Back

	// Additional Steam Deck specific mappings
	// These are useful for Steam Deck's on-screen keyboard
	// You can add more mappings as needed
}

// connectedGamepadIDs stores the currently connected gamepad IDs.
// Use a map for efficient add/remove operations.
var connectedGamepadIDs = make(map[ebiten.GamepadID]struct{})

// gamepadIDsBuf is a temporary buffer reused by UpdateConnectedGamepads.
var gamepadIDsBuf []ebiten.GamepadID

// updateConnectedGamepads refreshes the list of connected gamepad IDs.
// Call this function once per frame in your game's Update method.
func updateConnectedGamepads() {
	// Check for newly connected gamepads
	gamepadIDsBuf = inpututil.AppendJustConnectedGamepadIDs(gamepadIDsBuf[:0])
	for _, id := range gamepadIDsBuf {
		connectedGamepadIDs[id] = struct{}{}
	}

	// Check for disconnected gamepads
	for id := range connectedGamepadIDs {
		if inpututil.IsGamepadJustDisconnected(id) {
			delete(connectedGamepadIDs, id)
		}
	}
}

// isMouseButton checks if the given buttonIndex corresponds to a mouse button or wheel.
func isMouseButton(buttonIndex int) bool {
	switch buttonIndex {
	case ButtonMouseLeft, ButtonMouseRight, ButtonMouseMiddle, ButtonMouseWheelUp, ButtonMouseWheelDown:
		return true
	default:
		return false
	}
}

// handleMouseInput checks if the specified PICO-8 mouse button/wheel is currently active.
func handleMouseInput(buttonIndex int) bool {
	switch buttonIndex {
	case ButtonMouseLeft:
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	case ButtonMouseRight:
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	case ButtonMouseMiddle:
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	case ButtonMouseWheelUp:
		_, wheelY := ebiten.Wheel()
		return wheelY < 0
	case ButtonMouseWheelDown:
		_, wheelY := ebiten.Wheel()
		return wheelY > 0
	}
	return false
}

// handleKeyboardInput checks if the specified PICO-8 button is pressed on the keyboard for Player 0.
func handleKeyboardInput(buttonIndex int) bool {
	if key, ok := pico8ButtonToKeyboardP0[buttonIndex]; ok {
		return ebiten.IsKeyPressed(key)
	}
	return false
}

// isDirectionButton checks if the buttonIndex corresponds to a directional PICO-8 button.
func isDirectionButton(buttonIndex int) bool {
	switch buttonIndex {
	case LEFT, RIGHT, UP, DOWN, ButtonJoypadLeft, ButtonJoypadRight, ButtonJoypadUp, ButtonJoypadDown:
		return true
	default:
		return false
	}
}

// handleGamepadDirectionalInput checks for directional inputs (D-pad and analog stick) on the gamepad.
func handleGamepadDirectionalInput(buttonIndex int, gamepadID ebiten.GamepadID) bool {
	axisThreshold := 0.5
	switch buttonIndex {
	case LEFT, ButtonJoypadLeft:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftLeft) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickHorizontal) < -axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftLeft)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickHorizontal)) < -axisThreshold
	case RIGHT, ButtonJoypadRight:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftRight) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickHorizontal) > axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftRight)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickHorizontal)) > axisThreshold
	case UP, ButtonJoypadUp:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftTop) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickVertical) < -axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftTop)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickVertical)) < -axisThreshold
	case DOWN, ButtonJoypadDown:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftBottom) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickVertical) > axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftBottom)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickVertical)) > axisThreshold
	}
	return false
}

// handleGamepadStandardButtonInput checks for standard PICO-8 button presses on the gamepad.
func handleGamepadStandardButtonInput(buttonIndex int, gamepadID ebiten.GamepadID) bool {
	if standardButton, ok := pico8ButtonToStandard[buttonIndex]; ok {
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			return ebiten.IsStandardGamepadButtonPressed(gamepadID, standardButton)
		}
		switch buttonIndex {
		case LEFT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftLeft))
		case RIGHT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftRight))
		case UP:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftTop))
		case DOWN:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftBottom))
		case O:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton1)
		case X:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton0)
		case ButtonStart:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton9)
		case ButtonSelect:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton8)
		default:
			if btn, found := pico8ButtonToStandard[buttonIndex]; found {
				return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(btn))
			}
		}
	}
	return false
}

// Btn checks if a specific PICO-8 button is currently held down via gamepad, keyboard (Player 0 only), mouse, or gamepad axes.
// Mimics the PICO-8 btn() function behavior (returns true while held).
//
// buttonIndex: The PICO-8 button index (0-15).
// playerIndex: Optional PICO-8 player index (0-7). Defaults to 0 (player 1) if omitted.
func Btn(buttonIndex int, _ ...int) bool {
	return getCachedButtonState(buttonIndex)
}

// Note: For "just pressed" behavior similar to PICO-8's btnp(), you would use
// inpututil functions.

// Btnp checks if a specific PICO-8 button was just pressed via gamepad, keyboard (Player 0 only), or mouse.
// Mimics the PICO-8 btnp() function behavior (without auto-repeat).
// It returns true only on the single frame the button transitions from up to down.
//
// buttonIndex: The PICO-8 button index (0-15).
// playerIndex: Optional PICO-8 player index (0-7). Defaults to 0 (player 1) if omitted.
//
//	Keyboard input is only checked for playerIndex 0.
//	Mouse input is available for all player indices.
//
// Usage:
//
//	Btnp(buttonIndex)
//	Btnp(buttonIndex, playerIndex)
//
// Example:
//
//	// Check if the 'X' button/key was just pressed for player 0
//	if Btnp(X) {
//		// Jump action
//	}
//
//	// Check if the right mouse button was just pressed
//	if Btnp(MOUSE_RIGHT) {
//		// Handle right click
//	}
//
//	// Check if the 'Start' button (gamepad only) was just pressed for player 1
//	if Btnp(START, 1) {
//		// Pause game for player 1
//	}
func Btnp(buttonIndex int, _ ...int) bool {
	// Check if button is pressed this frame but wasn't pressed last frame
	current := getCachedButtonState(buttonIndex)
	previous := getCachedButtonStatePrev(buttonIndex)
	return current && !previous
}

// Add input state caching
var (
	buttonStates     = make(map[int]bool) // buttonIndex -> isPressed
	buttonStatesPrev = make(map[int]bool) // previous frame button states
	inputCacheMutex  sync.RWMutex
	inputCacheValid  bool
)

// updateInputCache updates the cached button states
func updateInputCache() {
	inputCacheMutex.Lock()
	defer inputCacheMutex.Unlock()

	// Copy current states to previous
	for k, v := range buttonStates {
		buttonStatesPrev[k] = v
	}

	// Update current states for all buttons
	for buttonIndex := 0; buttonIndex <= ButtonJoypadR5; buttonIndex++ {
		buttonStates[buttonIndex] = checkButtonState(buttonIndex)
	}

	inputCacheValid = true
}

// checkButtonState checks the actual button state (uncached)
func checkButtonState(buttonIndex int) bool {
	// Handle mouse buttons
	if isMouseButton(buttonIndex) {
		return handleMouseInput(buttonIndex)
	}

	// Handle keyboard input
	if handleKeyboardInput(buttonIndex) {
		return true
	}

	// Handle gamepad input
	for gamepadID := range connectedGamepadIDs {
		if isDirectionButton(buttonIndex) {
			if handleGamepadDirectionalInput(buttonIndex, gamepadID) {
				return true
			}
		} else {
			if handleGamepadStandardButtonInput(buttonIndex, gamepadID) {
				return true
			}
		}
	}

	return false
}

// getCachedButtonState returns the cached button state
func getCachedButtonState(buttonIndex int) bool {
	inputCacheMutex.RLock()
	defer inputCacheMutex.RUnlock()

	if !inputCacheValid {
		return checkButtonState(buttonIndex)
	}

	return buttonStates[buttonIndex]
}

// getCachedButtonStatePrev returns the cached previous button state
func getCachedButtonStatePrev(buttonIndex int) bool {
	inputCacheMutex.RLock()
	defer inputCacheMutex.RUnlock()

	return buttonStatesPrev[buttonIndex]
}
