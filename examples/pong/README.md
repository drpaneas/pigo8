# Pong

A simple Pong clone built with [PIGO8](https://github.com/drpaneas/pigo8), ported from NerdyTeachers' PICO-8 tutorial.

## 🎮 Play Online

**[▶️ Play Pong in your browser](https://drpaneas.github.io/pigo8/pong/)**

## 📖 Description

This project reimplements "Bite-Size Games #5: Pong" by **NerdyTeachers** (original PICO-8 source at <https://nerdyteachers.com/PICO-8/Bitesize_Games/5>) in Go, showcasing how to build a small arcade-style game with the PIGO8 engine.

### Controls

- **Arrow Keys**: Move paddle up/down
- **Z**: Start game

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/pong
go run .
```

## 📝 Source Code

```go
// See main.go for full implementation
// Key features demonstrated:
// - Ball physics and collision detection
// - AI paddle opponent
// - Score tracking
// - Sound effects with p8.Music()
```

