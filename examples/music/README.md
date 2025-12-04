# Music Demo

A demonstration of audio playback built with [PIGO8](https://github.com/drpaneas/pigo8).

## 🎮 Play Online

**[▶️ Play Music Demo in your browser](https://drpaneas.github.io/pigo8/music/)**

## 📖 Description

This example demonstrates audio functionality in PIGO8:
- Playing music files with `p8.Music()`
- Stopping music with `p8.Music(-1)` or `p8.StopMusic()`
- Exclusive playback mode (stops other tracks when starting)
- Loading multiple audio files (music0.wav through music63.wav)

### Controls

- **Up Arrow**: Play music 3
- **Down Arrow**: Play music 4
- **Left Arrow**: Play music 5
- **Right Arrow**: Play music 6 (exclusive mode)
- **Up + Down**: Stop all music

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/music
go generate  # Generate embedded assets
go run .
```

## 📝 Source Code

```go
package main

//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .

import (
	p8 "github.com/drpaneas/pigo8"
)

type Game struct{}

func (g *Game) Init() {}

func (g *Game) Update() {
	if p8.Btn(p8.UP) {
		p8.Music(3)
	}
	if p8.Btn(p8.DOWN) {
		p8.Music(4)
	}
	if p8.Btn(p8.LEFT) {
		p8.Music(5)
	}
	if p8.Btn(p8.RIGHT) {
		p8.Music(6, true) // Exclusive - stops other music
	}
	if p8.Btn(p8.UP) && p8.Btn(p8.DOWN) {
		p8.Music(-1) // Stop all music
	}
}

func (g *Game) Draw() {
	p8.Cls(1)
	p8.Print("Music Example", 30, 10, 7)
	p8.Print("Up: music 3", 10, 35, 7)
	p8.Print("Down: music 4", 10, 45, 7)
	p8.Print("Left: music 5", 10, 55, 7)
	p8.Print("Right: music 6 (exclusive)", 10, 65, 7)
	p8.Print("Up+Down: stop all", 10, 75, 7)
}

func main() {
	p8.InsertGame(&Game{})
	settings := p8.NewSettings()
	settings.WindowTitle = "PIGO8 Music Example"
	p8.PlayGameWith(settings)
}
```

