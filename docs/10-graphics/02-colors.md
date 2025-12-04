# Colors and Palette

## The Default Palette

PIGO8 starts with the classic PICO-8 palette:

| Index | Name | RGB | Description |
|-------|------|-----|-------------|
| 0 | Black | (0, 0, 0) | Used for transparency |
| 1 | Dark Blue | (29, 43, 83) | |
| 2 | Dark Purple | (126, 37, 83) | |
| 3 | Dark Green | (0, 135, 81) | |
| 4 | Brown | (171, 82, 54) | |
| 5 | Dark Gray | (95, 87, 79) | |
| 6 | Light Gray | (194, 195, 199) | |
| 7 | White | (255, 241, 232) | |
| 8 | Red | (255, 0, 77) | |
| 9 | Orange | (255, 163, 0) | |
| 10 | Yellow | (255, 236, 39) | |
| 11 | Green | (0, 228, 54) | |
| 12 | Blue | (41, 173, 255) | |
| 13 | Indigo | (131, 118, 156) | |
| 14 | Pink | (255, 119, 168) | |
| 15 | Peach | (255, 204, 170) | |

## Transparency

By default, **color 0 (black) is transparent** when drawing sprites. This means:

- Pixels with color 0 in your sprites won't be drawn
- You can layer sprites on top of each other
- Change transparency settings with `Palt()`

## Custom Palettes (palette.hex)

Load palettes from sites like [Lospec](https://lospec.com/palette-list):

1. Download a palette in HEX format
2. Save as `palette.hex` in your project:

```
c60021
e70000
e76121
e7a263
e7c384
```

3. PIGO8 automatically loads it on startup

**Note**: Color 0 is always transparent, and color 1 becomes white. Your hex colors start at index 2.

## Palette Functions

### SetPaletteColor

Change a single color at runtime:

```go
import "image/color"

func (g *game) Init() {
    // Make color 8 a custom blue
    p8.SetPaletteColor(8, color.RGBA{0, 100, 200, 255})
}
```

### GetPaletteColor

Read the current color at an index:

```go
col := p8.GetPaletteColor(8)  // Returns color.Color
```

### SetPalette

Replace the entire palette:

```go
import "image/color"

func (g *game) Init() {
    grayscale := []color.Color{
        color.RGBA{0, 0, 0, 255},       // Black
        color.RGBA{85, 85, 85, 255},    // Dark gray
        color.RGBA{170, 170, 170, 255}, // Light gray
        color.RGBA{255, 255, 255, 255}, // White
    }
    p8.SetPalette(grayscale)
}
```

### GetPaletteSize

Get the number of colors:

```go
count := p8.GetPaletteSize()  // Returns 16 for default palette
```

