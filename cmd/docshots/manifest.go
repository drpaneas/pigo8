package main

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
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "animation",
		ExampleDir: "examples/animation",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   100,
	},
	{
		Name:       "spritesheet",
		ExampleDir: "examples/spritesheet",
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "map",
		ExampleDir: "examples/map",
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "map-layers",
		ExampleDir: "examples/map_layers",
		Kind:       "static",
		CaptureMs:  500,
		SampleMs:   500,
	},
	{
		Name:       "camera",
		ExampleDir: "examples/camera/camera_example2",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowRight"}, HoldMs: 800},
			{Keys: []string{"ArrowDown"}, HoldMs: 800},
		},
	},
	{
		Name:       "palette",
		ExampleDir: "examples/palette",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
	},
	{
		Name:       "color-collision",
		ExampleDir: "examples/colorCollision",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowRight"}, HoldMs: 1000},
			{Keys: []string{"ArrowUp"}, HoldMs: 1000},
		},
	},
	{
		Name:       "map-collision",
		ExampleDir: "examples/gameboy",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowRight"}, HoldMs: 1500},
		},
	},
	{
		Name:       "pong",
		ExampleDir: "examples/pong",
		Kind:       "gif",
		CaptureMs:  3000,
		SampleMs:   150,
		Inputs: []InputStep{
			{Keys: []string{"ArrowUp"}, HoldMs: 600},
			{Keys: []string{"ArrowDown"}, HoldMs: 1200},
			{Keys: []string{"ArrowUp"}, HoldMs: 600},
		},
	},
}
