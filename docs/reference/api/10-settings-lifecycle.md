# Settings & Lifecycle

## `Cartridge` interface

```go
type Cartridge interface {
	Init()   // Called once at the start.
	Update() // Called every frame for logic.
	Draw()   // Called every frame for drawing.
}
```

Your game struct must implement this interface. See [The Game Loop](../../03-game-loop.md).

## `InsertGame(cartridge Cartridge)`

Registers your game implementation with PIGO8. Must be called before `Play()`/`PlayGameWith()`.

## `Play()`

Runs the console with default settings, using the cartridge registered via `InsertGame`.
Equivalent to `PlayGameWith(NewSettings())`.

## `PlayGameWith(settings *Settings)`

Runs the console with custom settings.

## `NewSettings() *Settings`

Creates a `Settings` object populated with defaults.

## `Settings` struct

| Field | Type | Default | Description |
|-------|------|---------|--------------|
| `ScaleFactor` | `int` | 4 | Integer window scaling factor. |
| `WindowTitle` | `string` | "PIGO-8 Game" | Window title bar text. |
| `TargetFPS` | `int` | 30 | Target ticks per second. |
| `ScreenWidth` | `int` | 128 | Logical screen width. |
| `ScreenHeight` | `int` | 128 | Logical screen height. |
| `Multiplayer` | `bool` | false | Enable multiplayer networking. |
| `Fullscreen` | `bool` | false | Start in fullscreen. |
| `ColorSpace` | `ebiten.ColorSpace` | default | Rendering color space. |
| `DisableHiDPI` | `bool` | true | Disable HiDPI scaling. |
| `SpriteImageCacheSize` | `int` | 256 | Max cached transparent sprite images. |
| `SpritePixelCacheSize` | `int` | 256 | Max cached sprite pixel data entries. |
| `SsprCacheSize` | `int` | 128 | Max cached `Sspr` source regions. |
| `MapCacheEnabled` | `bool` | true | Enable map tile caching. |
| `EnableFrameStats` | `bool` | false | Enable frame-level performance statistics. |

**Example:**
```go
settings := p8.NewSettings()
settings.WindowTitle = "My Game"
settings.ScaleFactor = 4
settings.TargetFPS = 60
settings.ScreenWidth = 160
settings.ScreenHeight = 144

p8.InsertGame(&game{})
p8.PlayGameWith(settings)
```

**See also:** [Settings](../../04-settings.md), [The Game Loop](../../03-game-loop.md)
