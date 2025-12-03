# Flashlight Demo

A demonstration of dynamic sprite pixel manipulation with flashing light effects.

## 🎮 Play Online

**[▶️ Play Flashlight Demo in your browser](https://drpaneas.github.io/pigo8/flashlight/)**

## 📖 Description

This example demonstrates dynamic sprite modification in PIGO8:
- Modifying sprite pixels at runtime with `p8.Sset()`
- Reading sprite pixel colors with `p8.Sget()`
- Creating flashing/blinking effects
- Frame-based timing for visual effects

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/flashlight
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type myGame struct {
	counter float64
	msg     string
}

func (m *myGame) Init() {
	// Initialize the counter
	m.counter = 0
}

func (m *myGame) Update() {
	// Increment the counter
	m.counter += 0.1

	// Create a flashing effect by alternating colors
	switch {
	case m.counter < 1:
		// Red light
		p8.Sset(12, 0, 8) // Red pixel at position (12,0) on the spritesheet
	case m.counter < 2:
		// Blue light
		p8.Sset(12, 0, 12) // Blue pixel at position (12,0) on the spritesheet
	default:
		// Reset counter
		m.counter = 0
	}

	c := p8.Sget(12, 0)
	if c == 8 {
		m.msg = "i"
	} else {
		m.msg = "ou"
	}
}

func (m *myGame) Draw() {
	// Clear screen with dark blue
	p8.Cls(3)

	// Draw sprite 1 (which contains our modified pixel)
	p8.Spr(1, 20, 30)

	p8.Print(m.msg, 5)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

