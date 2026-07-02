# Screen & Drawing

## `Cls(colorIndex ...int)`

Clears the current drawing screen with a specified PICO-8 color index. If no `colorIndex` is
provided, it defaults to 0 (black).

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `colorIndex` | `int` (optional, variadic) | PICO-8 color index (0-15). Defaults to 0 (black). |

**Example:**
```go
p8.Cls()     // clear to black
p8.Cls(12)   // clear to blue
```

**See also:** [Screen](../../10-graphics/01-screen.md)

## `ClsRGBA(clr color.RGBA)`

Clears the current drawing screen with a specified RGBA color instead of a PICO-8 color index,
allowing any RGBA color to be used.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `clr` | `color.RGBA` | The RGBA color to clear the screen with. |

**Example:**
```go
p8.ClsRGBA(color.RGBA{R: 100, G: 150, B: 200, A: 255}) // custom blue
p8.ClsRGBA(color.RGBA{})                               // transparent black
```

**See also:** [Screen](../../10-graphics/01-screen.md)

## `Pset(x, y int, colorIndex ...int)`

Draws a single pixel at coordinates `(x, y)` using the specified PICO-8 color index, or the
current cursor color if omitted. `Pset` uses raw **screen** coordinates - unlike sprites and
shapes, it is not affected by the camera offset. Out-of-bounds coordinates or invalid color
indices are silently ignored (with a logged warning for invalid colors).

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `int` | Screen pixel coordinates. |
| `colorIndex` | `int` (optional, variadic) | PICO-8 color index (0-15). Defaults to the current cursor color. |

**Example:**
```go
p8.Cursor(0, 0, 8)   // set current color to red
p8.Pset(10, 20)      // draws a red pixel at (10, 20)
p8.Pset(50, 50, 12)  // draws a blue pixel at (50, 50), overriding the cursor color
```

**See also:** [Pixels](../../10-graphics/03-pixels.md)

## `Pget(x, y int) int`

Returns the PICO-8 color index (0-15) of the pixel at `(x, y)`. Returns 0 (black) if the
coordinates are out of bounds or the pixel doesn't exactly match a palette color.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `int` | Screen pixel coordinates to sample. |

**Example:**
```go
p8.Pset(10, 20, 8)
idx := p8.Pget(10, 20) // idx == 8
```

**See also:** [Pixels](../../10-graphics/03-pixels.md)

## `Line[X1, Y1, X2, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{})`

Draws a line between two points. Mimics PICO-8's `line(x1, y1, x2, y2, color)`.

**Note:** unlike `Rect`, `Circ`, and their filled variants, `Line` does **not** apply the
current camera offset - coordinates always map directly to screen pixels.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x1`, `y1` | `Number` | Starting point coordinates. |
| `x2`, `y2` | `Number` | Ending point coordinates. |
| `color` (via `options`) | `int` (optional) | PICO-8 color index (0-15). Defaults to the current draw color (7, white). |

**Example:**
```go
p8.Line(0, 0, 127, 127, 8) // diagonal red line across the screen
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Rect[X1, Y1, X2, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{})`

Draws an outline rectangle using two opposing corner points. Mimics PICO-8's `rect(x1, y1, x2, y2, color)`.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x1`, `y1`, `x2`, `y2` | `Number` | Coordinates of two opposing corners. |
| `color` (via `options`) | `int` (optional) | PICO-8 color index (0-15). Defaults to the current draw color. |

**Example:**
```go
p8.Rect(10, 10, 40, 30, 11) // green outlined rectangle
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Rectfill[X1, Y1, X2, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{})`

Draws a filled rectangle using two opposing corner points. Mimics PICO-8's `rectfill(x1, y1, x2, y2, color)`.

**Parameters:** Same as `Rect`.

**Example:**
```go
p8.Rectfill(10, 10, 40, 30, 11) // filled green rectangle
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Circ[X, Y, R Number](x X, y Y, radius R, options ...interface{})`

Draws an outline circle. Mimics PICO-8's `circ(x, y, radius, color)`.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `Number` | Center point coordinates. |
| `radius` | `Number` | Circle radius. |
| `color` (via `options`) | `int` (optional) | PICO-8 color index (0-15). Defaults to the current draw color. |

**Example:**
```go
p8.Circ(64, 64, 20, 9) // orange outlined circle at the center
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Circfill[X, Y, R Number](x X, y Y, radius R, options ...interface{})`

Draws a filled circle. Mimics PICO-8's `circfill(x, y, radius, color)`.

**Parameters:** Same as `Circ`.

**Example:**
```go
p8.Circfill(64, 64, 20, 9) // filled orange circle
```

**See also:** [Shapes](../../10-graphics/04-shapes.md)

## `Print(s any, args ...int) (int, int)`

Draws `s` (converted via `fmt.Sprintf("%v", s)`) onto the screen at the implicit print cursor,
mimicking PICO-8's `print(str, [x, y], color)`, including cursor tracking. Returns the `(x, y)`
coordinates immediately after the printed string.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `s` | `any` | Value to print (formatted with `%v`). |
| `args` | `int` (optional, variadic) | `[]`: use cursor position/color. `[color]`: cursor position, given color. `[x, y]`: given position, cursor color. `[x, y, color]`: given position and color. |

**Returns:** the X and Y screen coordinates immediately after the printed text.

**Example:**
```go
p8.Cursor(0, 0, 6)
p8.Print("1 HELLO")            // prints at (0,0) in light gray, cursor moves to (0,6)
p8.Print("2 AT", 20, 20)       // prints at (20,20) in light gray
endX, endY := p8.Print("3 DONE") // prints at (20,26), returns next cursor position
```

**See also:** [Text](../../10-graphics/05-text.md)

## `Cursor(args ...int)`

Sets the implicit print cursor position and optionally the default draw color, mimicking
PICO-8's `cursor(x, y, color)`. Calling with no arguments resets the position to `(0, 0)` without
changing the color.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `args` | `int` (optional, variadic) | `[]`: reset to (0,0). `[x, y]`: set position. `[x, y, color]`: set position and draw color. |

**Example:**
```go
p8.Cursor(10, 20)     // move cursor to (10, 20)
p8.Cursor(30, 40, 5)  // move cursor and set draw color to 5
p8.Cursor()           // reset to (0, 0)
```

**See also:** [Text](../../10-graphics/05-text.md)

## `GetScreenWidth() int`

Returns the current logical screen width in pixels (128 by default, or the value configured via
`Settings.ScreenWidth`).

**Example:**
```go
w := p8.GetScreenWidth()
```

## `GetScreenHeight() int`

Returns the current logical screen height in pixels (128 by default, or the value configured via
`Settings.ScreenHeight`).

**Example:**
```go
h := p8.GetScreenHeight()
```

**See also:** [Screen](../../10-graphics/01-screen.md), [Settings](../../04-settings.md)
