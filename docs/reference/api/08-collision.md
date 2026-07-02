# Collision

## `ColorCollision[X, Y Number](x X, y Y, color int) bool`

Reports whether the pixel at `(x, y)` matches `color`. Returns `false` if the coordinates are
out of screen bounds (0 to `GetScreenWidth()-1` / `GetScreenHeight()-1`, 0-127 by default) or the
color index is invalid.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `Number` | Screen pixel coordinates to check. |
| `color` | `int` | PICO-8 color index (0-15) to compare against. |

**Example:**
```go
if p8.ColorCollision(player.x, player.y, 3) {
	// touching a wall (color 3)
}
```

**See also:** [Color Collision](../../70-collision/01-color-collision.md)

## `MapCollision[X, Y Number](x X, y Y, flag int, size ...int) bool`

Reports whether a rectangular area starting at `(x, y)` overlaps any map tile whose sprite has
`flag` set. Internally resolves overlapping tiles via `Mget`/`Fget`.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `Number` | Top-left pixel coordinates of the area to check. |
| `flag` | `int` | Sprite flag number (0-7) to check for on underlying tiles. |
| `size` | `int` (optional, variadic) | `[]`: 8x8 area. `[s]`: `s`x`s` area. `[w, h]`: `w`x`h` area. |

**Example:**
```go
if p8.MapCollision(player.x, player.y, 0) {
	// collision with default 8x8 area, checking flag 0
}
playerWidth, playerHeight := 14, 15
if p8.MapCollision(player.x, player.y, 0, playerWidth, playerHeight) {
	// collision with a custom-sized area, checking flag 0
}
```

**See also:** [Map Collision](../../70-collision/02-map-collision.md)
