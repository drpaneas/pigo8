//go:build js

// Package pigo8 provides platform-specific input handling for web/WASM platforms.
package pigo8

import (
	"sync"
	"syscall/js"
)

// Virtual button state for web platform
var (
	virtualButtonStates = make(map[int]bool)
	virtualButtonMutex  sync.RWMutex

	// virtualButtonFunc stores the js.Func to prevent memory leaks.
	// js.FuncOf allocates memory that should be released when no longer needed.
	virtualButtonFunc js.Func
)

// isSteamDeck always returns false on web platforms.
// Steam Deck detection requires exec.Command which is not available in WASM.
func isSteamDeck() bool {
	return false
}

// SetVirtualButton sets the state of a virtual button from JavaScript.
// This is called by the web UI when the user presses/releases virtual buttons.
func SetVirtualButton(buttonIndex int, pressed bool) {
	virtualButtonMutex.Lock()
	defer virtualButtonMutex.Unlock()
	virtualButtonStates[buttonIndex] = pressed
}

// getVirtualButtonState returns the current state of a virtual button.
// Used by checkButtonState to include virtual button input on web.
func getVirtualButtonState(buttonIndex int) bool {
	virtualButtonMutex.RLock()
	defer virtualButtonMutex.RUnlock()
	return virtualButtonStates[buttonIndex]
}

// vibrateOnPress triggers a short haptic vibration on mobile devices.
// This provides tactile feedback when pressing virtual buttons.
func vibrateOnPress() {
	navigator := js.Global().Get("navigator")
	if !navigator.Truthy() {
		return
	}
	vibrate := navigator.Get("vibrate")
	if !vibrate.Truthy() {
		return
	}
	// 10ms pulse for subtle haptic feedback
	navigator.Call("vibrate", 10)
}

// setVirtualButtonJS is the JavaScript callback wrapper for SetVirtualButton.
// It's exposed to JavaScript as window.pigo8SetButton(buttonIndex, pressed).
func setVirtualButtonJS(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return nil
	}
	buttonIndex := args[0].Int()
	pressed := args[1].Bool()
	SetVirtualButton(buttonIndex, pressed)

	// Haptic feedback on button press (mobile devices)
	if pressed {
		vibrateOnPress()
	}
	return nil
}

// initVirtualButtons registers the JavaScript callback for virtual button input.
// This is called from engine_js.go init().
// The js.Func is stored in virtualButtonFunc to prevent memory leaks.
func initVirtualButtons() {
	virtualButtonFunc = js.FuncOf(setVirtualButtonJS)
	js.Global().Set("pigo8SetButton", virtualButtonFunc)
}

// CleanupVirtualButtons releases the JavaScript callback memory.
// Call this if the game is ever unloaded to prevent memory leaks.
// For long-running games this is typically not needed as the function lives forever.
func CleanupVirtualButtons() {
	if virtualButtonFunc.Truthy() {
		virtualButtonFunc.Release()
	}
}
