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

## Local Multiplayer

`Btn`/`Btnp` accept an optional player index (0-7) as their second argument for local
multiplayer with multiple gamepads on the same machine:

```go
if p8.Btn(p8.LEFT, 0) { player0.x-- }  // player index 0's input
if p8.Btn(p8.LEFT, 1) { player1.x-- }  // player index 1's gamepad
```

Gamepads are assigned to players by ascending gamepad ID among the currently connected
gamepads - plugging in two controllers gives you players 0 and 1, in whatever order the OS
reports them (not necessarily physical plug-in order). With only one gamepad connected, it's
assigned to player 0 alongside the keyboard (both work simultaneously for player 0) - there's
no player 1 gamepad until a second one is connected. Mouse input is shared across all player
indices, since there's only one mouse. An out-of-range player index (negative, or greater than
7) returns `false` rather than erroring.

```go
func (g *game) Update() {
    for player := range g.players {
        if p8.Btn(p8.LEFT, player)  { g.players[player].x-- }
        if p8.Btn(p8.RIGHT, player) { g.players[player].x++ }
    }
}
```

