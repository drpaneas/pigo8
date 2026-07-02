# Maps

## `Map(args ...any)`

Draws a rectangular region of the map to the screen.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `mx`, `my` | `int` (optional) | Map tile coordinates to start reading from (default 0, 0). |
| `sx`, `sy` | `int` (optional) | Screen pixel coordinates to draw at (default 0, 0). |
| `w`, `h` | `int` (optional) | Dimensions in tiles (default 16x16). |
| `layers` | `int` (optional) | Bitfield to filter drawn sprites by their flags (0 = draw all). |

**Example:**
```go
p8.Map()                          // draw the default 16x16 tile region at (0,0)
p8.Map(0, 0, 0, 0, 32, 32)        // draw a 32x32 tile region
p8.Map(0, 0, 0, 0, 32, 32, 1)     // only draw tiles whose sprite has flag 0 set
```

**See also:** [Drawing Maps](../../30-maps/01-drawing-maps.md)

## `Mget[C, R Number](column C, row R) int`

Returns the sprite number placed at the given map column/row.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `column`, `row` | `Number` | Map tile coordinates to read from. |

**Example:**
```go
sprite := p8.Mget(5, 3)
```

## `Mset[C, R, S Number](column C, row R, sprite S)`

Sets the sprite number at the given map column/row.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `column`, `row` | `Number` | Map tile coordinates to write to. |
| `sprite` | `Number` | Sprite number to place at that tile. |

**Example:**
```go
p8.Mset(5, 3, 1) // place sprite 1 at column 5, row 3
```

**See also:** [Map Data](../../30-maps/02-map-data.md)
