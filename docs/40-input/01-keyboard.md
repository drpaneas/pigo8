# Keyboard Input

## Button Constants

PIGO8 uses PICO-8 style button indices:

| Constant | Key | Index |
|----------|-----|-------|
| `p8.LEFT` | Arrow Left | 0 |
| `p8.RIGHT` | Arrow Right | 1 |
| `p8.UP` | Arrow Up | 2 |
| `p8.DOWN` | Arrow Down | 3 |
| `p8.O` | Z | 4 |
| `p8.X` | X | 5 |
| `p8.ButtonStart` | Enter | 6 |
| `p8.ButtonSelect` | Tab | 7 |

## Btn - Is Button Held?

```go
if p8.Btn(p8.LEFT) {
    // Left arrow is currently pressed
    g.playerX--
}
```

Returns `true` every frame the button is held down.

> **Note:** The optional player index parameter (e.g., `Btn(p8.LEFT, 1)`) is currently not used. All input is treated as player 0.

### Continuous Movement

```go
func (g *game) Update() {
    if p8.Btn(p8.LEFT)  { g.x-- }
    if p8.Btn(p8.RIGHT) { g.x++ }
    if p8.Btn(p8.UP)    { g.y-- }
    if p8.Btn(p8.DOWN)  { g.y++ }
}
```

## Btnp - Was Button Just Pressed?

```go
if p8.Btnp(p8.X) {
    // X button was just pressed this frame
    g.jump()
}
```

Returns `true` only on the first frame the button is pressed.

### Menu Navigation

```go
func (g *game) Update() {
    if p8.Btnp(p8.UP) {
        g.menuSelection--
    }
    if p8.Btnp(p8.DOWN) {
        g.menuSelection++
    }
    if p8.Btnp(p8.X) {
        g.selectMenuItem()
    }
}
```

## Common Patterns

### Movement with Speed

```go
const speed = 2

func (g *game) Update() {
    if p8.Btn(p8.LEFT)  { g.x -= speed }
    if p8.Btn(p8.RIGHT) { g.x += speed }
    if p8.Btn(p8.UP)    { g.y -= speed }
    if p8.Btn(p8.DOWN)  { g.y += speed }
}
```

### Jump with Btnp

```go
func (g *game) Update() {
    // Ground check
    if g.onGround && p8.Btnp(p8.O) {
        g.velocityY = -5
        g.onGround = false
    }
    
    // Apply gravity
    g.velocityY += 0.2
    g.y += g.velocityY
}
```

### Fire Rate Limiting

```go
func (g *game) Update() {
    if g.fireCooldown > 0 {
        g.fireCooldown--
    }
    
    if p8.Btn(p8.X) && g.fireCooldown == 0 {
        g.shoot()
        g.fireCooldown = 10  // 10 frame cooldown
    }
}
```

