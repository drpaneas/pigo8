# Sprite Flags

Each sprite has 8 flags (bits 0-7) that can be used for game logic—commonly for collision detection and layer filtering.

## Flag Numbers

Flags are numbered 0-7 and accessed by their number, not by constants:

```go
// Flag numbers (0-7) for use with Fget/Fset
// Flag 0: Often used for "solid/collision"
// Flag 1-7: Custom usage

// When used as a bitfield (for Map layers parameter):
// Flag 0 = 1, Flag 1 = 2, Flag 2 = 4, Flag 3 = 8
// Flag 4 = 16, Flag 5 = 32, Flag 6 = 64, Flag 7 = 128
```

## Fget - Get Flag

`Fget` returns **two values**: a bitfield and a boolean.

```go
bitfield, isSet := p8.Fget(spriteNumber, flagNumber)
```

### Get All Flags (Bitfield)

When checking all flags, use the first return value:

```go
allFlags, _ := p8.Fget(1)  // Returns bitfield (0-255)
```

### Check Specific Flag

When checking a specific flag, use the second return value:

```go
_, isSolid := p8.Fget(1, 0)  // Check if flag 0 is set on sprite 1
```

> **Important:** Always use both return values appropriately. When checking a specific flag, ignore the first value with `_`.

## Fset - Set Flag

### Set a Specific Flag

```go
p8.Fset(1, 0, true)   // Set flag 0 on sprite 1
p8.Fset(1, 0, false)  // Clear flag 0 on sprite 1
```

### Set All Flags at Once

```go
p8.Fset(1, false)  // Clear all flags (bitfield = 0)
p8.Fset(1, true)   // Set all flags (bitfield = 255)
p8.Fset(1, 170)    // Set flags 1, 3, 5, 7 (binary: 10101010)
```

## Common Use Cases

### Collision Detection

Mark solid tiles with flag 0:

```go
// In spritesheet.json, wall tiles have flags: 1
// In game code:
func (g *game) Update() {
    newX := g.playerX + g.dx
    
    // Check if destination tile is solid
    tileX := p8.Flr(newX / 8)
    tileY := p8.Flr(g.playerY / 8)
    spriteID := p8.Mget(tileX, tileY)
    
    _, isSolid := p8.Fget(spriteID, 0)
    if !isSolid {
        g.playerX = newX  // Allow movement
    }
}
```

### Layer Filtering

Use flags to control which sprites appear:

```go
func (g *game) Draw() {
    p8.Cls(0)
    
    // Draw background layer (flag 1)
    p8.Map(0, 0, 0, 0, 16, 16, 2)  // Only sprites with flag 1
    
    // Draw foreground layer (flag 2)
    p8.Map(0, 0, 0, 0, 16, 16, 4)  // Only sprites with flag 2
}
```

### Collectible Items

Mark collectibles with a flag:

```go
// Flag 1 = collectible
func (g *game) Update() {
    tileX := p8.Flr(g.playerX / 8)
    tileY := p8.Flr(g.playerY / 8)
    spriteID := p8.Mget(tileX, tileY)
    
    _, isCollectible := p8.Fget(spriteID, 1)
    if isCollectible {
        g.score++
        p8.Mset(tileX, tileY, 0)  // Remove collected item
    }
}
```

