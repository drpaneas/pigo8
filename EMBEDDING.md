# Resource Embedding in PIGO8

## Quick Start Guide

**To create a portable PIGO8 game that works anywhere:**

1. Add this line at the top of your main.go file:

   ```go
   //go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
   ```

2. Run these commands before distributing your game:

   ```bash
   go generate
   go build
   ```

That's it! Your game binary will now include all necessary resources and work correctly even when moved to a different directory.

## What This Does

PIGO8 uses these resource files for games:

- `map.json` - Contains the game map data
- `spritesheet.json` - Contains sprite definitions and pixel data
- `palette.hex` - Optional custom palette
- `music*.wav` - Optional audio files such as `music0.wav`, `music1.wav`, and so on

The `go generate` command automatically creates an `embed.go` file that embeds these resources into your binary, registers them with PIGO8, and makes your game fully portable.

## How Resource Loading Works

PIGO8 uses a smart resource loading system with the following priority order:

1. Files in the current directory (highest priority)
2. Common subdirectories: `assets/`, `resources/`, `data/`, `static/`
3. Embedded resources registered via `RegisterEmbeddedResources`
4. Default embedded resources in the PIGO8 library (lowest priority)

This approach gives you the best of both worlds:

- During development: Edit local files for quick iteration
- For distribution: Embed resources for portability

## Detailed Usage Guide

### Automatic Embedding with go:generate (Recommended)

PIGO8 provides a tool that automatically generates the necessary embedding code for your game. This is the recommended approach for distributing your game.

1. Add this line at the top of your main.go file:

   ```go
   //go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
   ```

2. Run the generate command to create the embed.go file:

   ```bash
   go generate
   ```

3. Build your game normally:

   ```bash
   go build
   ```

The generated `embed.go` file will embed your resources into the binary. Your game will now work correctly even when moved to a different directory.

### Manual Embedding (Alternative)

If you prefer to manually control the embedding process, you can create an `embed.go` file in your project:

```go
package main

import (
 "embed"
 
 p8 "github.com/drpaneas/pigo8"
)

// Embed the game-specific resources
//
//go:embed map.json spritesheet.json music0.wav
var resources embed.FS

func init() {
 // Register the embedded resources with PIGO8
 p8.RegisterEmbeddedResources(resources, "spritesheet.json", "map.json", "music0.wav")
}
```

Adjust the `go:embed` directive to include only the files you have. Audio files passed to `RegisterEmbeddedResources` should still follow the `music%d.wav` naming convention because playback uses numeric IDs.

## Resource Loading Priority

PIGO8 uses the following priority order when looking for resources:

1. Files in the current directory (highest priority)
2. Custom embedded resources registered via `RegisterEmbeddedResources`
3. Default embedded resources in the PIGO8 library (lowest priority)

This allows you to:

- Develop with local files for quick iteration
- Distribute with embedded resources for portability
- Always have fallback resources from the library
