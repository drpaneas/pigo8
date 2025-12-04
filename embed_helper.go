package pigo8

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const none = "none"

// embeddedResources represents a set of embedded resources for a PIGO8 game
type embeddedResources struct {
	FS              fs.FS
	SpritesheetPath string
	MapPath         string
	PalettePath     string
	AudioPaths      []string // Paths to audio files
}

// customResources holds application-specific embedded resources
var customResources *embeddedResources

// autoDetectResourcesAttempted tracks whether we've already tried to auto-detect resources
var autoDetectResourcesAttempted bool

// RegisterEmbeddedResources allows applications to register their own embedded resources
// This should be called before any PIGO8 functions that might need these resources
func RegisterEmbeddedResources(resources fs.FS, spritesheetPath, mapPath string, audioPaths ...string) {
	// Check if one of the audioPaths is actually palette.hex
	palettePath := ""
	var filteredAudioPaths []string

	for _, path := range audioPaths {
		if filepath.Base(path) == "palette.hex" {
			palettePath = path
		} else {
			filteredAudioPaths = append(filteredAudioPaths, path)
		}
	}

	customResources = &embeddedResources{
		FS:              resources,
		SpritesheetPath: spritesheetPath,
		MapPath:         mapPath,
		PalettePath:     palettePath,
		AudioPaths:      filteredAudioPaths,
	}

	// Format spritesheet, map, and palette paths for logging, showing "none" if empty
	spritesheetDisplay := spritesheetPath
	if spritesheetDisplay == "" {
		spritesheetDisplay = none
	}

	mapDisplay := mapPath
	if mapDisplay == "" {
		mapDisplay = none
	}

	paletteDisplay := palettePath
	if paletteDisplay == "" {
		paletteDisplay = none
	}

	log.Printf("Registered custom embedded resources: spritesheet=%s, map=%s, palette=%s, audio files=%d", spritesheetDisplay, mapDisplay, paletteDisplay, len(filteredAudioPaths))

	// Automatically initialize audio if there are audio files
	if len(filteredAudioPaths) > 0 || hasEmbeddedAudioFiles(resources) {
		initAudioPlayer()
		log.Println("Audio system initialized automatically")
	}
}

// tryLoadEmbeddedFile attempts to load a file from embedded resources
// It first tries custom resources, then falls back to default resources
// Returns the file content and a boolean indicating success
func tryLoadEmbeddedFile(defaultPath string, isMap bool) ([]byte, bool) {
	// Auto-detect resources if not already done
	if !autoDetectResourcesAttempted {
		autoDetectResources()
	}

	// First try custom resources if registered
	if customResources != nil {
		var path string
		if isMap {
			path = customResources.MapPath
		} else {
			path = customResources.SpritesheetPath
		}

		data, err := fs.ReadFile(customResources.FS, path)
		if err == nil {
			fileType := "spritesheet"
			if isMap {
				fileType = "map"
			}
			log.Printf("Using custom embedded %s file: %s", fileType, path)
			return data, true
		}
	}

	// Fall back to default resources
	data, err := defaultResources.ReadFile(defaultPath)
	if err != nil {
		return nil, false
	}

	fileType := "spritesheet"
	if isMap {
		fileType = "map"
	}
	log.Printf("Using default embedded %s file: %s", fileType, defaultPath)
	return data, true
}

// autoDetectResources attempts to find common resource locations
// and automatically register them if found
func autoDetectResources() {
	// Only try once
	autoDetectResourcesAttempted = true

	// Don't override if already registered
	if customResources != nil {
		return
	}

	// Get information about the current build
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	// Check if we're running in a Go module
	hasModule := false
	for _, setting := range buildInfo.Settings {
		if setting.Key == "GOMOD" && setting.Value != "" {
			hasModule = true
			break
		}
	}

	// Log module detection status
	if hasModule {
		log.Printf("Running in a Go module, checking for resources")
	}

	// Check common resource locations
	resourceDirs := []string{
		".", // Current directory
		"assets",
		"resources",
		"data",
		"static",
	}

	// Try each location
	for _, dir := range resourceDirs {
		mapPath := filepath.Join(dir, "map.json")
		spritesheetPath := filepath.Join(dir, "spritesheet.json")

		mapExists := fileExists(mapPath)
		spritesheetExists := fileExists(spritesheetPath)

		// If either resource exists, try to create an embedded FS
		if mapExists || spritesheetExists {
			log.Printf("Found resources in directory: %s", dir)
			// We can't create an embed.FS at runtime, but we can note the locations
			// for future file operations
			break
		}
	}
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// hasEmbeddedAudioFiles checks if there are any music*.wav files in the embedded resources
func hasEmbeddedAudioFiles(resources fs.FS) bool {
	hasAudio := false

	// Walk through the embedded filesystem to find audio files
	_ = fs.WalkDir(resources, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue even if there's an error
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if the file is a music*.wav file
		if strings.HasPrefix(filepath.Base(path), "music") && strings.HasSuffix(path, ".wav") {
			hasAudio = true
			return fs.SkipAll // Stop walking once we find one audio file
		}

		return nil
	})

	return hasAudio
}

// initAudioPlayer initializes the audio player
func initAudioPlayer() {
	// Use the existing getAudioPlayer function to initialize the audio player
	// This ensures we're using the same singleton pattern
	ap := getAudioPlayer()

	// Verify that the audio player was initialized successfully
	if ap == nil {
		log.Println("Error: Failed to initialize audio player")
		return
	}

	// Check if any audio files were loaded
	if len(ap.musicData) == 0 {
		log.Println("Warning: Audio player initialized but no audio files were loaded")
	} else {
		log.Printf("Audio system initialized successfully with %d audio files", len(ap.musicData))
	}
}

// tryLoadEmbeddedMap attempts to load a map from embedded resources
func tryLoadEmbeddedMap() ([]byte, error) {
	data, ok := tryLoadEmbeddedFile(defaultMapPath, true)
	if !ok {
		return nil, fmt.Errorf("failed to load embedded map file")
	}
	return data, nil
}

// tryLoadEmbeddedSpritesheet attempts to load a spritesheet from embedded resources
func tryLoadEmbeddedSpritesheet() ([]byte, error) {
	data, ok := tryLoadEmbeddedFile(defaultSpritesheetPath, false)
	if !ok {
		return nil, fmt.Errorf("failed to load embedded spritesheet file")
	}
	return data, nil
}
