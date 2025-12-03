//go:build js

// Package pigo8 provides web-specific initialization for WASM builds.
package pigo8

func init() {
	// Initialize virtual button JavaScript bridge
	// This exposes window.pigo8SetButton(buttonIndex, pressed) to JavaScript
	initVirtualButtons()
}

