// Package main basic hello world example using clear screen and print statements
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import p8 "github.com/drpaneas/pigo8"

// No need to define flags locally anymore as they're provided by the pigo8 package

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
