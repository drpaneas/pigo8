package pigo8

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/drpaneas/pigo8/network"
	"github.com/hajimehoshi/ebiten/v2"
)

// --- Pause Menu Constants ---

// Pause menu option constants
const (
	// Using different names to avoid conflicts with pause_menu.go
	EngPauseOptionContinue = iota
	EngPauseOptionReset
	EngPauseOptionExit
	EngPauseOptionCount // Used to track the number of options
)

// --- Settings ---

// Settings defines configurable parameters for the PIGO8 console.
type Settings struct {
	ScaleFactor  int               // Integer scaling factor for the window (Default: 4).
	WindowTitle  string            // Title displayed on the window bar (Default: "PIGO-8 Game").
	TargetFPS    int               // Target ticks per second (Default: 30).
	ScreenWidth  int               // Custom screen width (Default: 128 for PICO-8 compatibility).
	ScreenHeight int               // Custom screen height (Default: 128 for PICO-8 compatibility).
	Multiplayer  bool              // Enable multiplayer networking (Default: false).
	Fullscreen   bool              // Start the game in fullscreen mode (Default: false).
	ColorSpace   ebiten.ColorSpace // Color space for rendering (Default: ColorSpaceDefault).
	DisableHiDPI bool              // Disable HiDPI scaling (Default: false).
}

// NewSettings creates a new Settings object with default values.
func NewSettings() *Settings {
	return &Settings{
		ScaleFactor:  4,
		WindowTitle:  "PIGO-8 Game",
		TargetFPS:    30,
		ScreenWidth:  defaultViewportWidth,  // Default PICO-8 width
		ScreenHeight: defaultViewportHeight, // Default PICO-8 height
		Multiplayer:  false,                 // Networking disabled by default
		Fullscreen:   false,                 // Windowed mode by default
		ColorSpace:   ebiten.ColorSpaceDefault,
		DisableHiDPI: true, // Better performance for retro-style games
	}
}

// --- Configuration (Internal Constants) ---

const (
	// defaultViewportWidth is the default PICO-8 screen width.
	defaultViewportWidth = 128
	// defaultViewportHeight is the default PICO-8 screen height.
	defaultViewportHeight = 128
)

// These variables hold the current logical screen dimensions.
// They can be modified through Settings when calling PlayGameWith.
var (
	// screenWidth is the current screen width (unexported).
	screenWidth = defaultViewportWidth
	// screenHeight is the current screen height (unexported).
	screenHeight = defaultViewportHeight
)

// --- Internal State for Drawing ---

// These hold the current context for the user's draw callback.
// They are set by the engine's Draw method each frame.
var (
	currentScreen    *ebiten.Image // Internal: Current screen image
	currentSprites   []spriteInfo  // Internal: Loaded sprites
	currentSpritesMu sync.RWMutex  // Protects currentSprites from concurrent access
	currentDrawColor int           // Internal: Current draw color (0-15)
	elapsedTime      float64       // Internal: Time elapsed since game start (in seconds)
	timeIncrement    float64       // Internal: Amount to increment time each update
)

// --- Cartridge Definition and Loading ---

// Cartridge defines the functions a user game must implement.
// This is the user's "game cartridge" code.
type Cartridge interface {
	Init()   // Called once at the start.
	Update() // Called every frame for logic.
	Draw()   // Called every frame for drawing.
}

// Internal default implementation.
type emptyCartridge struct{}

func (d *emptyCartridge) Init()   {}
func (d *emptyCartridge) Update() {}
func (d *emptyCartridge) Draw()   {}

// loadedCartridge holds the registered implementation provided by the user.
var loadedCartridge Cartridge = &emptyCartridge{}

// InsertGame sets the user's game implementation (the cartridge).
// Pass an instance of a struct that implements the Cartridge interface.
func InsertGame(cartridge Cartridge) {
	if cartridge == nil {
		loadedCartridge = &emptyCartridge{} // Reset to default if nil
	} else {
		loadedCartridge = cartridge
	}
}

// CurrentCartridge returns the currently loaded cartridge.
// This is useful for accessing the game state from network callbacks.
func CurrentCartridge() Cartridge {
	return loadedCartridge
}

// --- Internal Ebiten Game Implementation ---

// game is the internal struct that satisfies ebiten.Game interface.
type game struct {
	initialized     bool
	firstFrameDrawn bool // Track if the first frame has been drawn

	// Pause menu state
	paused        bool
	pauseSelected int
}

// Layout implements ebiten.Game.
func (g *game) Layout(_, _ int) (int, int) {
	w := screenWidth
	h := screenHeight
	if w <= 0 {
		w = defaultViewportWidth // fallback to 128
	}
	if h <= 0 {
		h = defaultViewportHeight // fallback to 128
	}
	return w, h
}

// Restart is a flag that indicates if the game should be restarted
var Restart bool

// ResetGame fully resets the game state
func (g *game) ResetGame() {
	log.Println("Resetting game...")
	// Reset all game state
	g.initialized = false
	g.firstFrameDrawn = false
	g.paused = false
	g.pauseSelected = EngPauseOptionContinue
	// Reset any other engine state here if needed
	Restart = true
}

// Update implements ebiten.Game.
func (g *game) Update() error {
	// Increment frame counter and update cache tick (Optimization 1: avoids time.Now())
	// IncrementFrameCounter() is called unconditionally to ensure LRU cache works
	// even when frame stats are disabled
	SetFrameTick(IncrementFrameCounter())

	if !g.initialized {
		log.Println("Cartridge Initializing...")
		// Log initial memory usage
		logInitialMemory()
		loadedCartridge.Init()
		g.initialized = true
		// Don't call Update on the first frame, wait for Draw to be called first
		return nil
	}

	// Only call Update after the first frame has been drawn
	if g.firstFrameDrawn {
		updateConnectedGamepads()
		updateMouseState()
		updateInputCache() // Update input cache for this frame

		// Check for START button press to toggle pause menu
		if Btnp(ButtonStart) {
			// Toggle pause state
			g.paused = !g.paused
			if g.paused {
				// Reset selection to Continue when pausing
				g.pauseSelected = EngPauseOptionContinue
				fmt.Println("Game paused")
			} else {
				fmt.Println("Game unpaused")
			}
		}

		// Update pause menu or game logic based on pause state
		if g.paused {
			// Handle pause menu navigation
			// Navigate up (keyboard or gamepad)
			if Btnp(UP) || Btnp(ButtonJoypadUp) {
				g.pauseSelected--
				if g.pauseSelected < 0 {
					g.pauseSelected = EngPauseOptionCount - 1
				}
			}

			// Navigate down (keyboard or gamepad)
			if Btnp(DOWN) || Btnp(ButtonJoypadDown) {
				g.pauseSelected++
				if g.pauseSelected >= EngPauseOptionCount {
					g.pauseSelected = 0
				}
			}

			// Process selection with X button (keyboard) or A button (gamepad)
			if Btnp(X) || Btnp(ButtonJoyA) || Btnp(O) { // O is often the confirm button on some controllers
				switch g.pauseSelected {
				case EngPauseOptionContinue:
					// Continue the game (unpause)
					g.paused = false
				case EngPauseOptionReset:
					// Reset the game
					g.ResetGame()
					// The next frame will trigger initialization
				case EngPauseOptionExit:
					// Exit the game immediately
					fmt.Println("Exiting application...")
					os.Exit(0) // This should immediately terminate the program
				}
			}
		} else {
			// Only update game logic when not paused
			loadedCartridge.Update()
			// Update elapsed time
			elapsedTime += timeIncrement
		}
	}

	return nil
}

// Draw implements ebiten.Game.
func (g *game) Draw(screen *ebiten.Image) {
	// Set the current screen for drawing
	currentScreen = screen
	// Invalidate currentScreen after Draw completes to catch misuse
	// (e.g., goroutines trying to draw outside of the Draw() context)
	defer func() { currentScreen = nil }()

	// Initialize pixel buffer if needed
	if pixelBuffer == nil {
		initPixelBuffer(GetScreenWidth(), GetScreenHeight())
	}

	// Initialize screen pixel cache if needed
	if screenPixelCache == nil {
		initScreenPixelCache(GetScreenWidth(), GetScreenHeight())
	}

	// Clear the screen
	// screen.Clear()

	// Call the user's Draw function
	loadedCartridge.Draw()

	// Flush all pending pixel operations at the end of the frame
	flushPixelBuffer()
	flushSpriteModifications()

	// Draw pause menu on top if active
	if g.paused {
		// Calculate menu dimensions
		menuWidth := 80
		menuHeight := 40
		menuX := (screenWidth - menuWidth) / 2
		menuY := (screenHeight - menuHeight) / 2

		// Determine background color - use the darkest color in the palette
		darkColor := findDarkestColorIndex()
		lightColor := findLightestColorIndex()
		midColor := findMidToneColorIndex()

		// Draw menu background and border
		Rectfill(menuX, menuY, menuX+menuWidth, menuY+menuHeight, darkColor)
		Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight, lightColor)

		// Draw title
		Print("pause", menuX+(menuWidth-30)/2, menuY+6, midColor)

		// Draw menu options
		optionY := menuY + 15
		for i, option := range []string{"resume", "restart", "quit game"} {
			if i == g.pauseSelected {
				// Draw selection cursor
				Print(">", menuX+10, optionY, midColor)
			}
			Print(option, menuX+20, optionY, lightColor)
			optionY += 8
		}

		// Flush again after menu drawing
		flushPixelBuffer()
		flushSpriteModifications()
	}

	// Flush frame metrics (Optimization 8: batch atomic ops)
	FlushFrameMetrics()

	// Mark that the first frame has been drawn
	if !g.firstFrameDrawn {
		g.firstFrameDrawn = true
	}
}

// --- Helper for User Code ---

// getCurrentSprites returns the currently loaded sprite data slice.
// Useful if user code needs to check len(sprites) or access flags.
func getCurrentSprites() []spriteInfo {
	return currentSprites
}

// CurrentScreen returns the current Ebiten screen image for direct drawing.
// This allows advanced users to use Ebiten's drawing functions directly.
func CurrentScreen() *ebiten.Image {
	return currentScreen
}

// --- Time Functions ---

// Time returns the number of seconds (as a decimal) since the game started.
// This is calculated by counting the number of times the Update method is called.
// Multiple calls to Time() in the same frame will return the same value.
func Time() float64 {
	return elapsedTime
}

// T is an alias for Time() for PICO-8 compatibility.
func T() float64 {
	return Time()
}

// --- Play Functions ---

// logInitialMemory logs the initial memory usage of the PIGO-8 console
func logInitialMemory() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("PIGO-8 Memory: %.2f MB", float64(m.Alloc)/1024/1024)
}

// PlayGameWith initializes Ebitengine, loads resources, and starts the main game loop.
// It uses the provided Settings and runs the Cartridge previously loaded via Insert.
func PlayGameWith(settings *Settings) {
	// Use default settings if nil is passed
	cfg := settings
	if cfg == nil {
		log.Println("Warning: pigo8.PlayGameWith called with nil Settings, using defaults.")
		cfg = NewSettings()
	}

	// Only initialize networking if multiplayer is enabled
	if cfg.Multiplayer {
		// Check if network is already initialized
		if network.IsNetworkInitialized() {
			log.Println("Network already initialized, skipping initialization")
		} else {
			// Check for network configuration from command line arguments
			networkConfig := network.ParseNetworkArgs()

			// Set game name from window title if not specified
			if networkConfig.GameName == "PIGO8 Game" {
				networkConfig.GameName = cfg.WindowTitle
			}

			// Initialize networking
			if err := network.InitNetwork(networkConfig); err != nil {
				log.Printf("Warning: Failed to initialize network: %v", err)
			}
			log.Println("Multiplayer networking enabled")
		}
		defer network.ShutdownNetwork()
	} else {
		log.Println("Multiplayer networking disabled")
	}

	// Reset time tracking variables
	elapsedTime = 0.0

	// Update logical screen dimensions if custom values are provided
	width := defaultViewportWidth
	height := defaultViewportHeight

	if cfg.ScreenWidth > 0 {
		width = cfg.ScreenWidth
	} else {
		cfg.ScreenWidth = defaultViewportWidth
	}

	if cfg.ScreenHeight > 0 {
		height = cfg.ScreenHeight
	} else {
		cfg.ScreenHeight = defaultViewportHeight
	}

	// Set screen size and initialize pixel buffer
	setScreenSize(width, height)

	// Try to load custom palette from palette.hex if it exists
	loadPaletteFromHexFile()

	// Configure Ebitengine window using Settings object
	winWidth := screenWidth * cfg.ScaleFactor
	winHeight := screenHeight * cfg.ScaleFactor
	if winWidth <= 0 || winHeight <= 0 {
		log.Printf("Warning: Calculated window size (%dx%d based on ScaleFactor %d) is non-positive. Using default %dx%d.", winWidth, winHeight, cfg.ScaleFactor, defaultViewportWidth, defaultViewportHeight)
		winWidth, winHeight = defaultViewportWidth, defaultViewportHeight
	}

	// Calculate time increment based on target FPS
	timeIncrement = 1.0 / float64(cfg.TargetFPS)

	internalGame := &game{
		initialized: false,
	}

	// Configure Ebitengine window using Settings object
	ebiten.SetWindowTitle(cfg.WindowTitle)
	ebiten.SetWindowSize(winWidth, winHeight)
	ebiten.SetTPS(cfg.TargetFPS)

	// Set fullscreen mode if enabled
	if cfg.Fullscreen {
		ebiten.SetFullscreen(true)
	}

	// Use RunGameOptions for new v2.8 features
	opts := &ebiten.RunGameOptions{
		ColorSpace:   cfg.ColorSpace,
		DisableHiDPI: cfg.DisableHiDPI,
	}

	log.Println("Booting PIGO8 console...")
	err := ebiten.RunGameWithOptions(internalGame, opts)
	if err != nil {
		log.Panicf("pico8.PlayGameWith: Ebitengine loop failed: %v", err)
	}
	log.Println("PIGO8 console shutdown.")
}

// Play initializes and runs the PIGO8 console with default settings.
// It runs the Cartridge previously loaded via Insert.
// Play is a convenience function that creates a game with default settings and plays it.
func Play() {
	PlayGameWith(NewSettings())
}

// GetScreenWidth returns the current logical screen width.
func GetScreenWidth() int {
	return screenWidth
}

// GetScreenHeight returns the current logical screen height.
func GetScreenHeight() int {
	return screenHeight
}
