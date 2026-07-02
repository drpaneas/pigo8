package main

// Job kinds accepted for CaptureJob.Kind.
const (
	jobKindStatic = "static"
	jobKindGIF    = "gif"
)

// InputStep represents one simulated user action: either holding a set of
// keys for a duration, or scrolling the mouse wheel a number of times at a
// point on the canvas. A step is a wheel step when WheelTicks != 0; Keys is
// ignored in that case, and vice versa.
type InputStep struct {
	Keys   []string // key names, e.g. "ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", "KeyW"/"KeyA"/"KeyS"/"KeyD", "KeyX"
	HoldMs int

	// Wheel step fields. WheelX/WheelY are fractions (0.0-1.0) of the
	// canvas's width/height, resolved to actual pixel coordinates at
	// capture time - this keeps the manifest independent of the exact
	// canvas size a job happens to render at. They must land inside the
	// game's actual interactive area (e.g. not a UI panel below the
	// canvas), or the game may silently ignore the wheel event.
	WheelTicks  int
	WheelDeltaY float64
	WheelX      float64
	WheelY      float64
}

// CaptureJob describes one example to capture and where its output goes.
type CaptureJob struct {
	Name       string // used for logging and the output filename (without extension)
	ExampleDir string // relative to the repo root, e.g. "examples/hello_world"
	Kind       string // "static" or "gif"
	Inputs     []InputStep
	CaptureMs  int // total capture window in milliseconds
	SampleMs   int // interval between frame samples in milliseconds

	// RootModule is true when ExampleDir is already part of the repo's
	// root module (e.g. cmd/editor) rather than a standalone example under
	// examples/* that needs its own go.mod plus a replace directive back
	// to the root module. Root-module jobs are built directly, skipping
	// that setup.
	RootModule bool
}

// editorDir is the repo-relative path to the editor tool, which is part of
// the root module rather than a standalone example under examples/.
const editorDir = "cmd/editor"

// manifest lists every example to capture for the documentation site.
// Each entry maps to a concrete existing example under examples/.
var manifest = []CaptureJob{
	{
		Name:       "hello-world",
		ExampleDir: "examples/hello_world",
		Kind:       jobKindStatic,
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "animation",
		ExampleDir: "examples/animation",
		Kind:       jobKindGIF,
		CaptureMs:  3000,
		SampleMs:   100,
	},
	{
		Name:       "spritesheet",
		ExampleDir: "examples/spritesheet",
		Kind:       jobKindStatic,
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "map",
		ExampleDir: "examples/map",
		Kind:       jobKindStatic,
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "map-layers",
		ExampleDir: "examples/map_layers",
		Kind:       jobKindStatic,
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		// camera_example2 is a static illustration (empty Update(), no input
		// handling) that always draws the same two-camera-offset scene, so a
		// GIF would just repeat identical frames - a static screenshot is the
		// honest representation of what this example actually shows.
		Name:       "camera",
		ExampleDir: "examples/camera/camera_example2",
		Kind:       jobKindStatic,
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		// palette's example is also static (empty Update()), so like camera
		// above, a still screenshot is more honest than a repeating GIF.
		Name:       "palette",
		ExampleDir: "examples/palette",
		Kind:       jobKindStatic,
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		// Player starts at (10,10) moving 1px/frame (~30px/sec at 30 TPS).
		// Move right then down toward the labyrinth (drawn starting at
		// x=30,y=30) rather than up, which would immediately leave the
		// visible 128x128 canvas.
		Name:       "color-collision",
		ExampleDir: "examples/colorCollision",
		Kind:       jobKindGIF,
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{keyArrowRight}, HoldMs: 700},
			{Keys: []string{keyArrowDown}, HoldMs: 700},
		},
	},
	{
		Name:       "map-collision",
		ExampleDir: "examples/gameboy",
		Kind:       jobKindGIF,
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{keyArrowRight}, HoldMs: 1500},
		},
	},
	{
		Name:       "pong",
		ExampleDir: "examples/pong",
		Kind:       jobKindGIF,
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{keyArrowUp}, HoldMs: 600},
			{Keys: []string{keyArrowDown}, HoldMs: 1200},
			{Keys: []string{keyArrowUp}, HoldMs: 600},
		},
	},
	{
		// cmd/editor is part of the repo's root module (not a standalone
		// example under examples/*), so it's built directly rather than
		// via the go.mod/replace-directive setup EnsureExampleModule does
		// for examples/*.
		Name:       "editor-sprite",
		ExampleDir: editorDir,
		Kind:       jobKindStatic,
		RootModule: true,
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		// Switches into Map Editor mode (X) before capturing, so the still
		// shows the map view rather than the default Sprite Editor.
		Name:       "editor-map",
		ExampleDir: editorDir,
		Kind:       jobKindStatic,
		RootModule: true,
		CaptureMs:  500,
		SampleMs:   500,
		Inputs: []InputStep{
			{Keys: []string{keyX}, HoldMs: 200},
		},
	},
	{
		// Demonstrates panning the Map Editor's camera with WASD: switch
		// into Map Editor mode, then pan right then down.
		Name:       "editor-pan",
		ExampleDir: editorDir,
		Kind:       jobKindGIF,
		RootModule: true,
		CaptureMs:  2200,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{keyX}, HoldMs: 200},
			{Keys: []string{keyD}, HoldMs: 900},
			{Keys: []string{keyS}, HoldMs: 900},
		},
	},
	{
		// Demonstrates zooming the Map Editor's camera in then out with the
		// mouse wheel. WheelX/WheelY (fractions of canvas size) target a
		// point safely inside the map viewport itself, not the info strip
		// below it - zoomAt() silently no-ops for points outside the map's
		// drawable area.
		Name:       "editor-zoom",
		ExampleDir: editorDir,
		Kind:       jobKindGIF,
		RootModule: true,
		CaptureMs:  2600,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{keyX}, HoldMs: 200},
			{WheelTicks: 6, WheelDeltaY: 120, WheelX: 0.5, WheelY: 0.33},
			{WheelTicks: 10, WheelDeltaY: -120, WheelX: 0.5, WheelY: 0.33},
		},
	},
}
