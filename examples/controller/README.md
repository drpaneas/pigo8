# Controller Demo

A visual input tester built with [PIGO8](https://github.com/drpaneas/pigo8) that displays an NES-style controller and highlights buttons when pressed.

## 🎮 Play Online

**[▶️ Play Controller Demo in your browser](https://drpaneas.github.io/pigo8/controller/)**

## 📖 Description

This example demonstrates input handling in PIGO8. It shows:
- Drawing shapes (rectangles, circles) to create a controller UI
- Using `p8.Btn()` to detect button presses
- Visual feedback when buttons are pressed
- PICO-8 button mapping (0-5 for directions and face buttons)

### Controls

| Button | Key | Index |
|--------|-----|-------|
| Left   | Arrow Left | 0 |
| Right  | Arrow Right | 1 |
| Up     | Arrow Up | 2 |
| Down   | Arrow Down | 3 |
| O      | Z | 4 |
| X      | X | 5 |

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/controller
go run .
```

## 📝 Source Code

```go
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type Game struct{}

func (g *Game) Init()   {}
func (g *Game) Update() {}

func (g *Game) Draw() {
	p8.Cls(1)
	p8.Print("controller", 40, 20, 10)

	// Draw controller body
	p8.Rectfill(10, 50, 117, 90, 7)
	p8.Rect(10, 50, 117, 90, 0)

	// D-pad with button detection
	p8.Rectfill(20, 65, 27, 73, 0)
	p8.Print("0", 22, 66, 12)
	if p8.Btn(0) { // p8.LEFT
		p8.Print("0", 22, 66, 4) // Highlight when pressed
	}

	// ... more buttons ...
}

func main() {
	p8.InsertGame(&Game{})
	p8.Play()
}
```
