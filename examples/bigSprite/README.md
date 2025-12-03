# Big Sprite Demo

Demonstration of working with sprites larger than 8x8 pixels.

## 🎮 Play Online

**[▶️ Play Big Sprite Demo in your browser](https://drpaneas.github.io/pigo8/bigSprite/)**

## 📖 Description

This example demonstrates how to work with larger sprites in PIGO8:
- Using `p8.Sspr()` for drawing larger sprite regions
- Palette transparency with `p8.Palt()`
- Working with sprites from `.p8` cartridge files

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/bigSprite
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type myGame struct{}

func (m *myGame) Init() {
	p8.Palt(4, true)
}

func (m *myGame) Update() {
}

func (m *myGame) Draw() {
	p8.Cls(0)
	sx := 88
	sy := 8
	sw := 16
	sh := 16
	dx := 10
	dy := 10
	p8.Sspr(sx, sy, sw, sh, dx, dy)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

