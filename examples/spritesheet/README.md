# Spritesheet Demo

A demonstration of PIGO8's sprite rendering capabilities.

## 🎮 Play Online

**[▶️ Play Spritesheet Demo in your browser](https://drpaneas.github.io/pigo8/spritesheet/)**

## 📖 Description

This example demonstrates how to work with spritesheets in PIGO8:
- Drawing individual sprites with `p8.Spr()`
- Drawing sprite regions with `p8.Sspr()` (sub-sprite)
- Positioning sprites on screen

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/spritesheet
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

import p8 "github.com/drpaneas/pigo8"

type myGame struct{}

func (m *myGame) Init() {}

func (m *myGame) Update() {}

func (m *myGame) Draw() {
	p8.Cls(0)
	p8.Spr(2, 20, 22)
	p8.Spr(3, 28, 22)
	p8.Spr(34, 20, 30)
	p8.Spr(35, 28, 30)

	p8.Sspr(16, 0, 16, 16, 50, 50)
	p8.Sspr(64, 56, 32, 32, 80, 80)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

