// Package main basic sprite editor
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	p8 "github.com/drpaneas/pigo8"
	"github.com/spf13/afero"
)

const (
	// Sprite dimensions
	spriteSize = 8 // Size of each sprite in pixels

	// Grid sizes
	defaultGridSize = 1 // 8x8 grid (1 sprite)
	mediumGridSize  = 2 // 16x16 grid (4 sprites)
	largeGridSize   = 4 // 32x32 grid (16 sprites)

	// Editor map dimensions
	editorMapWidth  = 320
	editorMapHeight = 320

	// Default colors
	defaultColor     = 1 // Default color
	transparentColor = 0 // Transparent color

	// UI constants
	paletteColumns = 8 // Number of columns in the palette display
	numFlags       = 8 // Number of sprite flags

)

type myGame struct {
	currentColor  int
	currentSprite int
	hoverX        int       // X coordinate of the pixel being hovered over (-1 if none)
	hoverY        int       // Y coordinate of the pixel being hovered over (-1 if none)
	gridSize      int       // Size of the working grid (1=8x8, 2=16x16, 4=32x32, 8=64x64)
	lastWheelTime int64     // Last time the mouse wheel was scrolled or keyboard was used (for debouncing)
	mapMode       bool      // Whether we are in map mode
	copiedSprite  [8][8]int // Buffer for copied sprite data

	// Undo/Redo state
	undoStack      []string      // Stack of saved state filenames for undo
	redoStack      []string      // Stack of saved state filenames for redo
	fs             afero.Fs      // Virtual filesystem for state snapshots
	stateMutex     sync.Mutex    // Mutex for thread-safe access to state
	lastSaveTime   time.Time     // Last time a state was saved
	saveCooldown   time.Duration // Minimum time between saves
	undoInProgress bool          // Flag to prevent re-entrant undo/redo operations

	// Key state tracking
	lastUndoTime int64 // Last time undo was triggered
	lastRedoTime int64 // Last time redo was triggered
	keyCooldown  int64 // Minimum time between undo/redo actions in milliseconds

	// Map editor state
	mapViewport   mapViewport
	mapLastMouseX int
	mapLastMouseY int
	mapData       [editorMapHeight][editorMapWidth]int
}

func (g *myGame) Init() {
	initSquareColors()

	// Initialize virtual filesystem and undo/redo stacks
	g.fs = afero.NewMemMapFs()
	g.undoStack = make([]string, 0)
	g.redoStack = make([]string, 0)
	g.saveCooldown = 100 * time.Millisecond

	// Initialize sprite flags to false
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			for flag := range numFlags {
				spriteFlags[row][col][flag] = false
			}
		}
	}

	// Initialize spritesheet (will also load from file if available)
	initSpritesheet()

	// Try to load map data from map.json. If it's missing (or the
	// environment has no writable filesystem, e.g. running as WASM in a
	// browser), fall back to an empty in-memory map instead of exiting -
	// a save failure here shouldn't prevent the editor from starting.
	if err := g.loadMapData(); err != nil {
		fmt.Println("No map.json found, starting with empty map")
		if err := g.saveMapData(); err != nil {
			fmt.Printf("Warning: could not save initial map.json, continuing with an in-memory, unsaved map: %v\n", err)
		} else {
			fmt.Println("Map saved to map.json")
		}
	}

	g.currentColor = defaultColor // Default color (usually red in PICO-8 palette)
	g.currentSprite = 1           // Default to first non-transparent sprite (sprite 0 is reserved)
	g.hoverX = -1                 // No hover initially
	g.hoverY = -1                 // No hover initially
	g.gridSize = defaultGridSize  // Start with 8x8 grid (1 sprite)
	g.mapViewport = defaultMapViewport()
	g.mapViewport.clamp(len(g.mapData[0]), len(g.mapData), g.mapScreenLayout().Canvas)

	// Ensure grid size is never less than 1
	if g.gridSize < defaultGridSize {
		g.gridSize = defaultGridSize
	}
	g.lastWheelTime = 0 // Initialize wheel time

	// Initialize the drawing canvas with the default sprite (1)
	g.updateDrawingCanvas()

	// Save the initial state to the undo stack
	if err := g.saveState(); err != nil {
		log.Printf("Failed to save initial state: %v", err)
	}
}

var (
	width           = 48  // Increased to accommodate the larger spritesheet and more space
	height          = 27  // Increased to accommodate the taller spritesheet
	mapViewWidth    = 128 // Default map viewport width in pixels (16 sprites)
	mapViewHeight   = 128 // Default map viewport height in pixels (16 sprites)
	unit            = 8
	spriteCellSize  = 8  // Size of each sprite cell in the spritesheet
	spriteSheetCols = 32 // Number of columns in the spritesheet
	spriteSheetRows = 24 // Number of rows in the spritesheet

	// Position of the spritesheet grid
	spritesheetStartX = 120 // Position spritesheet (adjusted 10px to the left)
)

var squareColors [64][64]int      // Up to 64x64 grid to store square colors
var spritesheet [24][32][8][8]int // 24x32 grid of 8x8 sprites
var spriteFlags [24][32][8]bool   // Flags for each sprite [row][col][flag0-7]

func main() {
	// Store initial default map view dimensions (pixels) from global vars
	initialDefaultMapViewWidthPx := mapViewWidth
	initialDefaultMapViewHeightPx := mapViewHeight

	// Calculate UI overhead in tiles based on initial global defaults for editor and map view
	// global 'width' and 'height' are editor total tiles (e.g., 48, 27)
	// global 'unit' is pixels per tile (e.g., 8)
	uiWidthOverheadInTiles := width - (initialDefaultMapViewWidthPx / unit)
	uiHeightOverheadInTiles := height - (initialDefaultMapViewHeightPx / unit)

	// Parse command line flags
	// Default values for flags are the initialDefaultMapViewWidth/Height
	widthFlag := flag.Int("w", initialDefaultMapViewWidthPx, "map viewport width in pixels")
	heightFlag := flag.Int("h", initialDefaultMapViewHeightPx, "map viewport height in pixels")
	flag.Parse()

	// Update global map viewport dimensions (pixels) from flags
	mapViewWidth = *widthFlag
	mapViewHeight = *heightFlag

	// Ensure new map viewport dimensions are multiples of 'unit' (sprite size)
	mapViewWidth = (mapViewWidth / unit) * unit
	mapViewHeight = (mapViewHeight / unit) * unit

	// Recalculate global editor window dimensions (tiles: global 'width', 'height')
	// based on the new map view dimensions (now in global mapViewWidth/Height) and the UI overhead
	width = (mapViewWidth / unit) + uiWidthOverheadInTiles
	height = (mapViewHeight / unit) + uiHeightOverheadInTiles

	// Initialize spritesheet if it doesn't exist
	if err := initPico8Spritesheet(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing spritesheet: %v\n", err)
		os.Exit(1)
	}

	if width > 256 || height > 256 {
		log.Printf("Editor window size is too large: %dx%d tiles.\n", width, height)
		os.Exit(1)
	}

	settings := p8.NewSettings()
	// These use the recalculated global 'width' and 'height'
	settings.ScreenWidth = width * unit
	settings.ScreenHeight = height * unit
	settings.ScaleFactor = 3
	p8.InsertGame(&myGame{})
	p8.PlayGameWith(settings)
}
