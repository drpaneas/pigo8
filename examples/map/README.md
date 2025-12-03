# Map Demo

A demonstration of PIGO8's map functionality.

## 🎮 Play Online

**[▶️ Play Map Demo in your browser](https://drpaneas.github.io/pigo8/map/)**

## 📖 Description

This example demonstrates how to work with tile maps in PIGO8:
- Loading and rendering maps with `p8.Map()`
- Reading map tiles with `p8.Mget()`
- Modifying map tiles dynamically with `p8.Mset()`
- Frame-based timing for animations

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/map
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

import (
	"log"
	"strconv"

	p8 "github.com/drpaneas/pigo8"
)

type myGame struct {
	frame      int
	changed    bool
	origSprite int
}

func (m *myGame) Init() {
	// Store the original sprite at position (0,5)
	m.origSprite = p8.Mget(0, 5)
	log.Printf("Original sprite at (0,5): %d", m.origSprite)
}

func (m *myGame) Update() {
	// Every 60 frames (about 1 second), toggle the sprite
	m.frame++
	if m.frame >= 60 {
		m.frame = 0
		m.changed = !m.changed

		// Toggle between original sprite and sprite #67
		if m.changed {
			p8.Mset(0, 5, 67)
			log.Println("Changed sprite to 67")
		} else {
			p8.Mset(0, 5, m.origSprite)
			log.Printf("Restored original sprite %d", m.origSprite)
		}
	}
}

func (m *myGame) Draw() {
	p8.Cls(1)
	p8.Map()

	// Display current sprite number
	currentSprite := p8.Mget(0, 5)
	p8.Print("Sprite: "+strconv.Itoa(currentSprite), 5, 5, 7)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

