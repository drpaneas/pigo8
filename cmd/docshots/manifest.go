package main

// Job kinds accepted for CaptureJob.Kind.
const (
	jobKindStatic = "static"
	jobKindGIF    = "gif"
)

// InputStep represents holding a set of keys for a duration before releasing them.
type InputStep struct {
	Keys   []string // key names, e.g. "ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"
	HoldMs int
}

// CaptureJob describes one example to capture and where its output goes.
type CaptureJob struct {
	Name       string // used for logging and the output filename (without extension)
	ExampleDir string // relative to the repo root, e.g. "examples/hello_world"
	Kind       string // "static" or "gif"
	Inputs     []InputStep
	CaptureMs  int // total capture window in milliseconds
	SampleMs   int // interval between frame samples in milliseconds
}

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
}
