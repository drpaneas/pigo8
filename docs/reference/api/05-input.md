# Input

## `Btn(buttonIndex int, playerIndex ...int) bool`

Reports whether a button is currently held down, via keyboard (player 0 only), gamepad, mouse,
or gamepad axes. Mimics PICO-8's `btn()`.

**Parameters:**

| Name | Type | Description |
|------|------|--------------|
| `buttonIndex` | `int` | PICO-8 button index (0-15) - see the button constants (`LEFT`, `RIGHT`, `UP`, `DOWN`, `O`, `X`, `ButtonStart`, `ButtonSelect`, and mouse/gamepad-specific constants). |
| `playerIndex` | `int` (optional) | Local player slot (0-7). Defaults to 0. Player 0 reads keyboard, mouse, and its assigned gamepad; players 1-7 read only their assigned gamepad. Gamepads are assigned to players by ascending gamepad ID among currently connected gamepads. An out-of-range value (< 0 or > 7) returns `false`. |

**Example:**
```go
if p8.Btn(p8.LEFT) { g.x-- }
if p8.Btn(p8.RIGHT) { g.x++ }

// Two local players, each with their own gamepad:
if p8.Btn(p8.LEFT, 0) { player1.x-- }
if p8.Btn(p8.LEFT, 1) { player2.x-- }
```

**See also:** [Keyboard](../../40-input/01-keyboard.md), [Gamepad](../../40-input/02-gamepad.md)

## `Btnp(buttonIndex int, playerIndex ...int) bool`

Reports whether a button was **just pressed** this frame (transitioned from up to down, no
auto-repeat). Mimics PICO-8's `btnp()`. Keyboard input only applies to player 0; mouse input
applies to all player indices; gamepad input is scoped to each player's assigned gamepad (see
`Btn`'s `playerIndex` parameter above for the assignment rule).

**Example:**
```go
if p8.Btnp(p8.X) {
	// jump
}

// Check if the 'Start' button was just pressed on player 1's gamepad
if p8.Btnp(p8.ButtonStart, 1) {
	// pause for player 1
}
```

**See also:** [Keyboard](../../40-input/01-keyboard.md)

## `GetMouseXY() (int, int)`

Returns the current mouse X and Y coordinates. Mimics PICO-8's `mouse()`.

**Example:**
```go
mx, my := p8.GetMouseXY()
p8.Circ(mx, my, 4, 8)
```

**See also:** [Mouse](../../40-input/03-mouse.md)
