# Hello World

The simplest PIGO8 example - a basic hello world demonstration.

## 🎮 Play Online

**[▶️ Play Hello World in your browser](https://drpaneas.github.io/pigo8/hello_world/)**

## 📖 Description

This is the minimal PIGO8 example showing:
- Basic game structure with Init, Update, and Draw methods
- Screen clearing with `p8.Cls()`
- Text rendering with `p8.Print()`

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/hello_world
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
	p8.Cls(1)
	p8.Print("hello, world!", 40, 60)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

