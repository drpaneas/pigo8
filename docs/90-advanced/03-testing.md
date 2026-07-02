# Testing Your Game

PIGO8 games are plain Go structs implementing the `Cartridge` interface, so standard Go testing
applies directly to your game logic - the trick is keeping logic that can be tested (movement,
collision, state transitions) separate from code that can only run inside the game loop
(`Draw()` calls, which require an active screen).

## Testing Logic Decoupled from `Update`/`Draw`

Structure gameplay logic as plain methods that don't call PIGO8 drawing functions, so they can
be tested without running the game loop:

```go
type Player struct {
	X, Y  float64
	Score int
}

// Move is pure logic - no p8 calls - so it's directly testable.
func (p *Player) Move(dx, dy float64) {
	p.X += dx
	p.Y += dy
	if p.X < 0 {
		p.X = 0
	}
}
```

```go
func TestPlayerMove(t *testing.T) {
	p := &Player{X: 5, Y: 5}
	p.Move(-10, 0)
	if p.X != 0 {
		t.Errorf("expected X to clamp at 0, got %v", p.X)
	}
}
```

Your `Cartridge.Update()` method then becomes a thin adapter that reads input and calls these
tested methods:

```go
func (g *game) Update() {
	var dx, dy float64
	if p8.Btn(p8.LEFT) {
		dx = -1
	}
	if p8.Btn(p8.RIGHT) {
		dx = 1
	}
	g.player.Move(dx, dy)
}
```

## Testing Collision Logic

Collision helpers like `p8.MapCollision` and `p8.ColorCollision` need an initialized screen/map,
so wrap your collision *decisions* in testable functions instead of testing the PIGO8 calls
themselves:

```go
// canMoveTo is pure logic given a collision check result - test this directly.
func canMoveTo(hitWall bool, isJumping bool) bool {
	return !hitWall || isJumping
}

func TestCanMoveTo(t *testing.T) {
	if canMoveTo(true, false) {
		t.Error("should not be able to move into a wall while not jumping")
	}
	if !canMoveTo(true, true) {
		t.Error("should be able to move through a wall while jumping")
	}
}
```

## Testing State Transitions

Model game states as an enum and test transitions independently of rendering:

```go
type State int

const (
	StateMenu State = iota
	StatePlaying
	StateGameOver
)

func nextState(current State, playerHealth int, startPressed bool) State {
	switch current {
	case StateMenu:
		if startPressed {
			return StatePlaying
		}
	case StatePlaying:
		if playerHealth <= 0 {
			return StateGameOver
		}
	}
	return current
}

func TestNextState_GameOverOnZeroHealth(t *testing.T) {
	got := nextState(StatePlaying, 0, false)
	if got != StateGameOver {
		t.Errorf("expected StateGameOver, got %v", got)
	}
}
```

## Running Tests

Standard Go tooling works as expected:

```bash
go test ./...
go test -run TestPlayerMove -v
go test -cover ./...
```

## What Not to Unit Test

Don't try to unit test `Draw()` itself, or functions that call drawing primitives (`Spr`,
`Rect`, `Cls`, etc.) directly - they require an active Ebitengine screen and are better verified
visually by running the game (`go run .`) or via the [Web Export](02-web-export.md) for quick
browser checks.
