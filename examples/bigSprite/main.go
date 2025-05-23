// Package main basic hello world example using clear screen and print statements
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
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
