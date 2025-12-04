# Installation

## Prerequisites

PIGO8 requires:

- **Go 1.21 or later** ([download](https://go.dev/dl/))
- A C compiler for CGO (required by the underlying Ebitengine):
  - **macOS**: Xcode Command Line Tools (`xcode-select --install`)
  - **Linux**: GCC (`sudo apt install gcc` or equivalent)
  - **Windows**: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MinGW-w64

## Installing PIGO8

Create a new Go module and add PIGO8:

```bash
mkdir my-game && cd my-game
go mod init my-game
go get github.com/drpaneas/pigo8
```

## Verifying Installation

Create a `main.go` file:

```go
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

Run it:

```bash
go run .
```

You should see a window with a dark blue background. Press `Enter` to open the pause menu, and select "Quit" to exit.

## Project Structure

A typical PIGO8 project looks like:

```
my-game/
├── main.go              # Game code
├── embed.go             # Resource embedding (auto-generated)
├── spritesheet.json     # Sprite definitions
├── map.json             # Tile map data
├── palette.hex          # Custom color palette (optional)
├── music1.wav           # Sound effects (optional)
└── go.mod
```

## Platform-Specific Notes

### macOS Apple Silicon

Works out of the box. No additional configuration needed.

### Linux

Install these dependencies for audio and graphics:

```bash
# Debian/Ubuntu
sudo apt install libasound2-dev libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install alsa-lib-devel mesa-libGL-devel xorg-x11-server-devel
```

### Windows

Use PowerShell or CMD. If you encounter CGO errors, ensure your compiler is in PATH.

### WebAssembly

See the [Web Export](90-advanced/02-web-export.md) guide for browser deployment.

