# Input

## `Btn(buttonIndex int, playerIndex ...int) bool`

Reports whether a button is currently held down, via keyboard (player 0 only), gamepad, mouse,
or gamepad axes. Mimics PICO-8's `btn()`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `buttonIndex` | `int` | PICO-8 button index (0-15) - see the button constants (`LEFT`, `RIGHT`, `UP`, `DOWN`, `O`, `X`, `ButtonStart`, `ButtonSelect`, and mouse/gamepad-specific constants). |
| `playerIndex` | `int` (optional) | Player index (0-7). Defaults to 0. |

**Example:**
```go
if p8.Btn(p8.LEFT) { g.x-- }
if p8.Btn(p8.RIGHT) { g.x++ }
```

**See also:** [Keyboard](../../40-input/01-keyboard.md), [Gamepad](../../40-input/02-gamepad.md)

## `Btnp(buttonIndex int, playerIndex ...int) bool`

Reports whether a button was **just pressed** this frame (transitioned from up to down, no
auto-repeat). Mimics PICO-8's `btnp()`. Keyboard input only applies to player 0; mouse input
applies to all player indices.

**Example:**
```go
if p8.Btnp(p8.X) {
	// jump
}
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
