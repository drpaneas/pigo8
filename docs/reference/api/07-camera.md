# Camera

## `Camera(args ...any)`

Sets the camera position, offsetting all subsequent drawing operations (shapes, `Print`,
sprites, and maps). With no arguments, resets the camera to `(0, 0)`. Mimics PICO-8's
`camera(x, y)`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `x`, `y` | `any` (optional) | Horizontal/vertical camera offset. Calling with only `x` sets the horizontal offset and resets `y` to 0 (matches PICO-8 behavior). |

**Example:**
```go
p8.Camera(64, 32) // set camera to (64, 32)
p8.Camera(64)     // set x to 64, y resets to 0
p8.Camera()       // reset to (0, 0)

// Lock UI in place while the world scrolls:
p8.Camera()
p8.Print("SCORE: 1000", 2, 2)
p8.Camera(playerX-64, playerY-64)
p8.Map()
```

**See also:** [Camera](../../60-camera/00-camera.md)
