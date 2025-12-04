# Settings

PIGO8 provides sensible defaults, but you can customize window size, frame rate, resolution, and more.

## Default Settings

```go
settings := p8.NewSettings()
// Returns:
// ScaleFactor:  4      (window is 4x the game resolution)
// WindowTitle:  "PIGO-8 Game"
// TargetFPS:    30
// ScreenWidth:  128    (PICO-8 default)
// ScreenHeight: 128
// Multiplayer:  false
// Fullscreen:   false
// DisableHiDPI: true   (better performance for pixel art)
```

## Using Custom Settings

```go
func main() {
    settings := p8.NewSettings()
    settings.WindowTitle = "My Awesome Game"
    settings.TargetFPS = 60
    settings.ScaleFactor = 6
    
    p8.InsertGame(&game{})
    p8.PlayGameWith(settings)
}
```

## Setting Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `ScaleFactor` | int | 4 | Window size multiplier (4 = 512×512 window for 128×128 game) |
| `WindowTitle` | string | "PIGO-8 Game" | Title shown in window bar |
| `TargetFPS` | int | 30 | Frames per second (30 or 60 recommended) |
| `ScreenWidth` | int | 128 | Logical game width in pixels |
| `ScreenHeight` | int | 128 | Logical game height in pixels |
| `Fullscreen` | bool | false | Start in fullscreen mode |
| `Multiplayer` | bool | false | Enable networking features |
| `DisableHiDPI` | bool | true | Disable HiDPI scaling (better for pixel art) |

## Custom Resolution

PIGO8 isn't limited to 128×128. Try these classic resolutions:

```go
// Game Boy (160×144)
settings.ScreenWidth = 160
settings.ScreenHeight = 144

// NES (256×240)
settings.ScreenWidth = 256
settings.ScreenHeight = 240

// Commodore 64 (320×200)
settings.ScreenWidth = 320
settings.ScreenHeight = 200
```

## Fullscreen Mode

```go
settings.Fullscreen = true
```

Press `Alt+Enter` (or `Cmd+Enter` on macOS) to toggle fullscreen at runtime.

## Reading Screen Dimensions

In your game code, get the current screen size:

```go
func (g *game) Draw() {
    width := p8.GetScreenWidth()   // Returns 128 by default
    height := p8.GetScreenHeight() // Returns 128 by default
    
    // Center something on screen
    centerX := width / 2
    centerY := height / 2
}
```

## Complete Example: Game Boy Style

```go
package main

import p8 "github.com/drpaneas/pigo8"

type game struct{}

func (g *game) Init()   {}
func (g *game) Update() {}

func (g *game) Draw() {
    p8.Cls(0)
    p8.Print("game boy!", 52, 70, 3)
}

func main() {
    settings := p8.NewSettings()
    settings.ScreenWidth = 160
    settings.ScreenHeight = 144
    settings.WindowTitle = "Game Boy Style"
    settings.ScaleFactor = 4
    
    p8.InsertGame(&game{})
    p8.PlayGameWith(settings)
}
```

