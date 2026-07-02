# Math Functions

## Flr - Floor

Round down to the nearest integer:

```go
p8.Flr(1.9)   // Returns 1
p8.Flr(-1.1)  // Returns -2
p8.Flr(5)     // Returns 5
```

### Common Uses

```go
// Convert pixel position to tile position
tileX := p8.Flr(g.playerX / 8)
tileY := p8.Flr(g.playerY / 8)

// Grid snapping
snappedX := p8.Flr(g.x/8) * 8
```

## Rnd - Random Number

Generate random integers:

```go
p8.Rnd(10)    // Returns 0-9
p8.Rnd(6)     // Dice roll (0-5, add 1 for 1-6)
p8.Rnd(128)   // Random screen position
```

### Examples

```go
// Random position
x := p8.Rnd(128)
y := p8.Rnd(128)

// Random color
color := p8.Rnd(16)

// 50% chance
if p8.Rnd(2) == 0 {
    // Do something
}

// 1 in 100 chance
if p8.Rnd(100) == 0 {
    g.spawnEnemy()
}
```

## Sqrt - Square Root

```go
p8.Sqrt(16)   // Returns 4.0
p8.Sqrt(2)    // Returns ~1.414
p8.Sqrt(-1)   // Returns 0.0 (not NaN)
```

### Distance Calculation

```go
func distance(x1, y1, x2, y2 float64) float64 {
    dx := x2 - x1
    dy := y2 - y1
    return p8.Sqrt(dx*dx + dy*dy)
}
```

## Sign - Get Sign

Returns -1 for negative numbers, 1 for zero or positive:

```go
p8.Sign(5.0)    // Returns 1.0
p8.Sign(-3.0)   // Returns -1.0
p8.Sign(0.0)    // Returns 1.0 (not 0!)
```

> **Note:** `Sign` takes a `float64` parameter and returns `float64`. Zero is treated as positive, returning 1.

### Use Case: Direction

```go
// Move toward target
if g.x != targetX {
    g.x += p8.Sign(float64(targetX - g.x))
}
```

## Time - Elapsed Time

Returns seconds since game started:

```go
elapsed := p8.Time()  // e.g., 10.5 seconds
```

### Use Cases

```go
// Animation timing
frame := int(p8.Time() * 10) % 4  // Cycle through 4 frames

// Spawn enemies every 5 seconds
if int(p8.Time()) % 5 == 0 && g.lastSpawn != int(p8.Time()) {
    g.spawnEnemy()
    g.lastSpawn = int(p8.Time())
}
```

## T - Time Alias

```go
p8.T()  // Same as p8.Time()
```

## See Also

- [Math](../reference/api/09-math.md)

