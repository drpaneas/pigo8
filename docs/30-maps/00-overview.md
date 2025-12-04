# Maps Overview

Maps are grids of tiles that form levels, backgrounds, and game worlds. Each cell references a sprite by ID.

## The Map Grid

In PICO-8 style, the default map is:

- **128 tiles wide × 128 tiles tall**
- Each tile is **8×8 pixels**
- Total: **1024×1024 pixels** (when fully rendered)

```
Map coordinates (tiles):
(0,0)  (1,0)  (2,0)  ...  (127,0)
(0,1)  (1,1)  (2,1)  ...  (127,1)
  ...
(0,127)              ...  (127,127)
```

## map.json Format

```json
{
  "version": "1.0",
  "description": "Level 1",
  "width": 128,
  "height": 128,
  "name": "world1",
  "cells": [
    {"x": 0, "y": 0, "sprite": 1},
    {"x": 1, "y": 0, "sprite": 1},
    {"x": 2, "y": 0, "sprite": 2}
  ]
}
```

The `cells` array is sparse—only non-zero tiles are listed.

## Map Functions Summary

| Function | Purpose |
|----------|---------|
| `Map(mx, my, sx, sy, w, h, layers)` | Draw map tiles to screen |
| `Mget(x, y)` | Get sprite ID at tile position |
| `Mset(x, y, sprite)` | Set sprite ID at tile position |

## Creating Maps

### PIGO8 Editor

```bash
go run github.com/drpaneas/pigo8/cmd/editor
```

### From PICO-8

```bash
go run github.com/drpaneas/parsepico -input game.p8 -output .
```

### Procedural Generation

```go
func (g *game) Init() {
    for y := 0; y < 16; y++ {
        for x := 0; x < 16; x++ {
            if p8.Rnd(10) < 2 {
                p8.Mset(x, y, 5)  // Random obstacles
            }
        }
    }
}
```

