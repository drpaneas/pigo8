# Gamepad Support

PIGO8 automatically detects gamepads. The standard buttons work without any configuration.

## Button Mapping

| PIGO8 Constant | Xbox Controller | PlayStation | Steam Deck |
|---------------|-----------------|-------------|------------|
| `p8.O` | X | Square | X |
| `p8.X` | A | X | A |
| `p8.ButtonStart` | Menu | Options | Menu |
| `p8.ButtonSelect` | View | Share | View |

### Directional Input

D-pad and left analog stick both work for directional buttons:

```go
// Works with keyboard arrows, D-pad, and analog stick
if p8.Btn(p8.LEFT)  { g.x-- }
if p8.Btn(p8.RIGHT) { g.x++ }
```

## Additional Gamepad Buttons

For games needing more buttons:

| Constant | Description |
|----------|-------------|
| `p8.ButtonJoyA` | A button (Xbox layout) |
| `p8.ButtonJoypadB` | B button |
| `p8.ButtonJoypadX` | X button |
| `p8.ButtonJoypadY` | Y button |
| `p8.ButtonJoypadL1` | Left shoulder |
| `p8.ButtonJoypadR1` | Right shoulder |
| `p8.ButtonJoypadL2` | Left trigger |
| `p8.ButtonJoypadR2` | Right trigger |

## Steam Deck

PIGO8 includes specific support for Steam Deck:

- D-pad and analog stick work for movement
- Face buttons map to PIGO8 buttons
- Back buttons (L4, R4, L5, R5) are accessible

```go
// Steam Deck back buttons
if p8.Btn(p8.ButtonJoypadL4) {
    // Left back button
}
```

## Testing Gamepad

```go
func (g *game) Draw() {
    p8.Cls(0)
    
    // Show button states
    y := 10
    for i := 0; i <= 5; i++ {
        color := 5  // Gray
        if p8.Btn(i) {
            color = 11  // Green
        }
        p8.Print(i, 10, y, color)
        y += 8
    }
}
```

