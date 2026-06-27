# Resource Embedding

Embed game assets into your binary for portable distribution.

## Quick Setup

Add to the top of `main.go`:

```go
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
```

Then build:

```bash
go generate
go build
```

Your executable now includes all assets.

## What Gets Embedded

The embedgen tool detects and embeds:

- `spritesheet.json` - Sprite data
- `map.json` - Map data
- `palette.hex` - Custom palette
- `music*.wav` - Audio files

## Manual Embedding

For custom control, create `embed.go`:

```go
package main

import (
    "embed"
    p8 "github.com/drpaneas/pigo8"
)

//go:embed spritesheet.json map.json music0.wav music1.wav
var resources embed.FS

func init() {
    p8.RegisterEmbeddedResources(
        resources,
        "spritesheet.json",
        "map.json",
        "music0.wav",
        "music1.wav",
    )
}
```

Audio files passed to `RegisterEmbeddedResources` should still use the `music%d.wav` naming convention because the playback APIs address them by numeric ID.

## Loading Priority

PIGO8 loads resources in this order:

1. **Current directory** (highest priority)
2. **Common subdirectories**: `assets/`, `resources/`, `data/`, `static/`
3. **Embedded resources**
4. **Library defaults** (lowest priority)

This allows local files to override embedded resources during development.

## Custom Palettes

Create `palette.hex` with hex colors:

```
c60021
e70000
e76121
e7a263
```

Color 0 is automatically transparent, colors shift to start at index 2.

## Complete Example

Project structure:

```
my-game/
├── main.go
├── embed.go          # Auto-generated
├── spritesheet.json
├── map.json
├── palette.hex
└── music0.wav
```

`main.go`:

```go
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import p8 "github.com/drpaneas/pigo8"

type game struct{}

func (g *game) Init()   {}
func (g *game) Update() {}
func (g *game) Draw()   { p8.Cls(1) }

func main() {
    p8.InsertGame(&game{})
    p8.Play()
}
```

Build and distribute:

```bash
go generate
go build -o my-game
./my-game  # Works anywhere!
```

