# Sprites

## `Spr[SN, X, Y Number](spriteNumber SN, x X, y Y, options ...any)`

Draws sprite `spriteNumber` (and optionally a fractional block of surrounding sprites) at
`(x, y)`.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `spriteNumber` | `Number` | Index of the top-left sprite to draw. |
| `x`, `y` | `Number` | Screen coordinates for the top-left corner. |
| `w`, `h` (via `options`) | `float64` or `int` (optional) | Width/height multiplier in sprites (default 1.0 each). |
| `flipX`, `flipY` (via `options`) | `bool` (optional) | Flip horizontally/vertically (default false). |

**Example:**
```go
p8.Spr(1, 10, 20)                   // draw sprite 1 at (10, 20)
p8.Spr(1, 50, 20, 1.5, 1.0)         // draw sprite 1 plus half of sprite 2
p8.Spr(0, 90, 20, 1.5, 1.5, true)   // 1.5x1.5 sprite block, flipped horizontally
```

**See also:** [Drawing Sprites](../../20-sprites/01-drawing-sprites.md)

## `Sspr[SX, SY, SW, SH, DX, DY Number](sx SX, sy SY, sw SW, sh SH, dx DX, dy DY, options ...any)`

Draws an arbitrary rectangular region of the spritesheet, with optional stretching and flipping.
Mimics PICO-8's `sspr(sx, sy, sw, sh, dx, dy, [dw, dh], [flip_x], [flip_y])`.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `sx`, `sy` | `Number` | Source position on the spritesheet, in pixels. |
| `sw`, `sh` | `Number` | Source width/height, in pixels. |
| `dx`, `dy` | `Number` | Destination position on screen, in pixels. |
| `dw`, `dh` (via `options`) | `Number` (optional) | Destination width/height (defaults to `sw`/`sh`). |
| `flipX`, `flipY` (via `options`) | `bool` (optional) | Flip horizontally/vertically (default false). |

**Example:**
```go
p8.Sspr(8, 8, 16, 16, 10, 20)             // copy a 16x16 region as-is
p8.Sspr(8, 8, 16, 16, 10, 20, 32, 32)     // stretched to 32x32 on screen
p8.Sspr(8, 8, 16, 16, 10, 20, 16, 16, true, false) // flipped horizontally
```

**See also:** [Spritesheet Regions](../../20-sprites/02-spritesheet-regions.md)

## `Sget[X, Y Number](x X, y Y) int`

Returns the PICO-8 color index (0-15) of the pixel at `(x, y)` on the spritesheet. Returns 0 if
out of bounds.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `Number` | Pixel coordinates on the spritesheet. |

**Example:**
```go
pixelColor := p8.Sget(10, 20)
```

## `Sset[X, Y Number](x X, y Y, colorIndex ...int)`

Sets the color of a pixel at `(x, y)` on the spritesheet. Uses the current draw color if
`colorIndex` is omitted.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `Number` | Pixel coordinates on the spritesheet. |
| `colorIndex` | `int` (optional, variadic) | PICO-8 color index (0-15). Defaults to the current draw color. |

**Example:**
```go
p8.Sset(10, 0, 8)  // red pixel at (10, 0) on the spritesheet
p8.Color(12)
p8.Sset(16, 0)     // blue pixel, using the current draw color
```

## `Fget(spriteNum int, flag ...int) (bitfield int, isSet bool)`

Returns a sprite's flag status. With no `flag` argument, returns the full 8-bit `bitfield`
(0-255). With a `flag` argument (0-7), `isSet` reports whether that specific flag is set.

**Example:**
```go
_, isSet := p8.Fget(1, 0)   // is flag 0 set on sprite 1?
allFlags, _ := p8.Fget(2)   // full bitfield for sprite 2
```

**See also:** [Sprite Flags](../../20-sprites/03-sprite-flags.md)

## `Fset(spriteNum int, flagOrValue interface{}, value ...interface{})`

Sets a sprite's flag status. If `flagOrValue` is a flag number (0-7), `value` sets that flag on
or off. If no `value` is given, `flagOrValue` is treated as a boolean or bitfield applied to all
flags.

**Example:**
```go
p8.Fset(1, 0, true)  // set flag 0 on sprite 1
p8.Fset(2, false)    // clear all flags on sprite 2
p8.Fset(2, 170)      // set flags 1,3,5,7 via bitfield (170 = 2+8+32+128)
```

**See also:** [Sprite Flags](../../20-sprites/03-sprite-flags.md)
