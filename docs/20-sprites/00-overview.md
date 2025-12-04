# Sprites Overview

Sprites are small images used for characters, objects, and tiles in your game. PIGO8 loads sprites from a `spritesheet.json` file.

## The Spritesheet

A spritesheet is a grid of small images (typically 8×8 pixels each). In PICO-8 style, the sheet is 128×128 pixels containing 256 sprites in a 16×16 grid.

```
Sprite IDs:
 0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15
16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31
32 33 34 35 ...
```

## spritesheet.json Format

```json
{
  "version": "1.0",
  "description": "Game sprites",
  "sprites": [
    {
      "id": 1,
      "x": 8,
      "y": 0,
      "width": 8,
      "height": 8,
      "flags": 0,
      "pixels": "0000000001111110011111100111111001111110011111100111111000000000"
    }
  ]
}
```

### Sprite Fields

| Field | Description |
|-------|-------------|
| `id` | Unique sprite number |
| `x`, `y` | Position on the spritesheet (pixels) |
| `width`, `height` | Dimensions (usually 8×8) |
| `flags` | Bitfield for collision/layer filtering (0-255) |
| `pixels` | Hex string of color indices |

## Creating Sprites

### PIGO8 Editor

Run the built-in editor:

```bash
go run github.com/drpaneas/pigo8/cmd/editor
```

### From PICO-8

If you have existing PICO-8 games:

```bash
go run github.com/drpaneas/parsepico -input game.p8 -output .
```

This extracts `spritesheet.json` and `map.json`.

### Manual Creation

Create sprites programmatically or use image editors and convert to the JSON format.

## Sprite Functions Summary

| Function | Purpose |
|----------|---------|
| `Spr(n, x, y, ...)` | Draw sprite n at position (x, y) |
| `Sspr(sx, sy, sw, sh, dx, dy, ...)` | Draw arbitrary spritesheet region |
| `Sget(x, y)` | Get pixel color from spritesheet |
| `Sset(x, y, c)` | Set pixel color on spritesheet |
| `Fget(n, f)` | Get sprite flag value |
| `Fset(n, f, v)` | Set sprite flag value |

