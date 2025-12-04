# Graphics Overview

PIGO8's graphics system is designed for simplicity: a fixed-size screen, a limited color palette, and immediate-mode drawing.

## Coordinate System

The screen uses a standard 2D coordinate system:

- **(0, 0)** is the top-left corner
- **X** increases to the right
- **Y** increases downward
- Default size is 128×128 pixels

```
(0,0) ────────────────► X (127,0)
  │
  │
  │
  │
  ▼
  Y
(0,127)                 (127,127)
```

## Drawing Order

PIGO8 uses immediate-mode rendering. Each frame:

1. Call `Cls()` to clear the screen
2. Draw background elements first
3. Draw foreground elements on top
4. Draw UI elements last

Later draws appear on top of earlier draws—like painting on a canvas.

```go
func (g *game) Draw() {
    p8.Cls(0)                    // 1. Clear (black background)
    p8.Map()                     // 2. Background tiles
    p8.Spr(1, g.playerX, g.playerY) // 3. Player sprite
    p8.Print(g.score, 2, 2, 7)   // 4. UI on top
}
```

## The 16-Color Palette

PIGO8 uses a 16-color palette by default (the PICO-8 palette):

| Index | Color | Hex | Index | Color | Hex |
|-------|-------|-----|-------|-------|-----|
| 0 | Black | `#000000` | 8 | Red | `#FF004D` |
| 1 | Dark Blue | `#1D2B53` | 9 | Orange | `#FFA300` |
| 2 | Dark Purple | `#7E2553` | 10 | Yellow | `#FFEC27` |
| 3 | Dark Green | `#008751` | 11 | Green | `#00E436` |
| 4 | Brown | `#AB5236` | 12 | Blue | `#29ADFF` |
| 5 | Dark Gray | `#5F574F` | 13 | Indigo | `#83769C` |
| 6 | Light Gray | `#C2C3C7` | 14 | Pink | `#FF77A8` |
| 7 | White | `#FFF1E8` | 15 | Peach | `#FFCCAA` |

Use color indices (0-15) in all drawing functions.

## Graphics Functions Summary

| Function | Purpose |
|----------|---------|
| `Cls(color)` | Clear screen |
| `Pset(x, y, color)` | Draw single pixel |
| `Pget(x, y)` | Read pixel color |
| `Rect()`, `Rectfill()` | Draw rectangles |
| `Circ()`, `Circfill()` | Draw circles |
| `Line()` | Draw lines |
| `Print()` | Draw text |
| `Spr()` | Draw sprites |
| `Sspr()` | Draw sprite regions |
| `Map()` | Draw tile maps |

