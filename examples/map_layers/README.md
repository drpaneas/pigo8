# Map Layers Demo

A demonstration of working with multiple map layers using sprite flags.

## 🎮 Play Online

**[▶️ Play Map Layers Demo in your browser](https://drpaneas.github.io/pigo8/map_layers/)**

## 📖 Description

This example demonstrates how to work with layered maps in PIGO8:
- Using sprite flags to define layers
- Rendering specific layers with `p8.Map()` layer parameter
- Combining multiple layers for parallax or depth effects

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
cd examples/map_layers
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
	p8.Cls(1)
	layers := 4 + 6
	p8.Map(0, 0, 0, 0, 16, 16, layers)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

