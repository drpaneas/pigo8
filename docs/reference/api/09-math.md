# Math

## `Flr[T Number](a T) int`

Rounds down to the nearest whole integer. Mimics PICO-8's `flr()`.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `a` | `Number` | The number to round down. |

**Example:**
```go
p8.Flr(1.99)  // 1
p8.Flr(-5.3)  // -6
```

## `Rnd[T Number](a T) int`

Returns a random integer in `[0, floor(a))`. Returns 0 if `a` is zero or negative. Mimics
PICO-8's `flr(rnd(a))`. Uses Go's `math/rand`; unlike PICO-8, results are not deterministic
across runs unless you explicitly seed the global source.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `a` | `Number` | The exclusive upper bound for the random result. |

**Example:**
```go
p8.Rnd(5)    // 0, 1, 2, 3, or 4
p8.Rnd(5.9)  // 0-4 (floor(5.9) = 5)
```

## `Sqrt[T Number](a T) float64`

Returns the square root. Returns 0 for negative input (unlike `math.Sqrt`, which returns `NaN`),
matching PICO-8 behavior.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `a` | `Number` | The number to take the square root of. |

**Example:**
```go
p8.Sqrt(16) // 4.0
p8.Sqrt(-4) // 0.0
```

## `Sign(v float64) float64`

Returns the sign of `v`, with 0 treated as `+1`.

**Example:**
```go
p8.Sign(-3.5) // -1.0
p8.Sign(0)    // 1.0
```

## `Time() float64`

Returns the number of seconds elapsed since the game started, based on the number of `Update`
calls. Multiple calls within the same frame return the same value.

**Example:**
```go
frame := int(p8.Time() * 10) % 4
p8.Spr(1+frame, x, y) // simple 4-frame animation
```

**See also:** [Math Functions](../../80-math/00-math.md)
