# Animation Example

A sprite animation demonstration built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Animation Demo in your browser](https://drpaneas.github.io/pigo8/animation/)**

## 📖 Description

This example demonstrates how to create animated sprites in PIGO8. It shows:
- Frame-based sprite animation
- Multiple animated entities
- Variable animation speeds
- Entity movement patterns

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/animation
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

// Entity represents a game entity with animation capabilities
type Entity struct {
	sprite, x, y, timing, speed, first, last float64
}

// NewEntity creates a new animated entity with the given parameters
func NewEntity(sprite, x, y, timing, speed, first, last float64) Entity {
	return Entity{
		sprite: sprite,
		timing: timing,
		first:  first,
		last:   last,
		x:      x,
		y:      y,
		speed:  speed,
	}
}

// Animate updates the sprite index based on timing to create animation
func (ae *Entity) Animate() {
	ae.sprite += ae.timing
	if ae.sprite >= ae.last {
		ae.sprite = ae.first
	}
}

// Move changes the entity's position by the given offset
func (ae *Entity) Move(offset float64) {
	ae.x += offset
	if ae.x > 128 {
		ae.x = -8
	}
}

// Draw renders the entity to the screen
func (ae *Entity) Draw() {
	p8.Spr(ae.sprite, ae.x, ae.y)
}

var player Entity
var enemies = []Entity{}
var items = []Entity{}

type myGame struct{}

func (m *myGame) Init() {
	player = NewEntity(1, -8, 59, 0.25, 1, 1, 5)
	enemy1 := NewEntity(5, -20, 5, 0.1, 1.25, 5, 9)
	enemy2 := NewEntity(9, -14, 30, 0.2, 0.4, 9, 13)
	enemy3 := NewEntity(13, -11, 90, 0.4, 0.75, 13, 17)
	enemies = append(enemies, enemy1, enemy2, enemy3)
	item1 := NewEntity(48, 30, 110, 0.3, 48, 50, 56)
	item2 := NewEntity(56, 60, 110, 0.25, 54, 56, 60)
	item3 := NewEntity(60, 90, 110, 0.15, 4, 60, 64)
	items = append(items, item1, item2, item3)
}

func (m *myGame) Update() {
	player.Animate()
	player.Move(player.speed)

	for i := range enemies {
		enemies[i].Animate()
		enemies[i].Move(enemies[i].speed)
	}

	for i := range items {
		items[i].Animate()
	}
}

func (m *myGame) Draw() {
	p8.Cls(1)
	player.Draw()

	for _, enemy := range enemies {
		enemy.Draw()
	}

	for _, item := range items {
		item.Draw()
	}
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

