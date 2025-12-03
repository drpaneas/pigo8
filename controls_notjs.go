//go:build !js

// Package pigo8 provides platform-specific input handling for non-web platforms.
package pigo8

import (
	"os/exec"
	"strings"
	"sync"
)

// Steam Deck detection caching (non-web only)
var (
	isSteamDeckResult bool
	steamDeckOnce     sync.Once
)

// isSteamDeck checks if the game is running on a Steam Deck by checking the hostname.
// The result is cached after the first call to avoid repeated shell command execution.
// This function only works on non-web platforms (exec.Command is not available in WASM).
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

// getVirtualButtonState returns false on non-web platforms (no virtual buttons).
// The buttonIndex parameter is required to match the signature in controls_js.go.
func getVirtualButtonState(_ int) bool {
	return false
}
