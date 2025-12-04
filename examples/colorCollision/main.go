// Package main provides a color collision detection example for the PIGO8 engine
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

// Player represents a game entity with collision capabilities
type Player struct {
	x, y, speed    float64
	collisionColor int
}

// Game implements the PIGO8 Cartridge interface
type Game struct {
	player Player
}

// Init is called once at the start
func (g *Game) Init() {
	g.player = Player{
		x:              10,
		y:              10,
		speed:          1,
		collisionColor: 10,
	}
}

// Update is called every frame for game logic
func (g *Game) Update() {
	beforeX := g.player.x
	beforeY := g.player.y

	if p8.Btn(p8.LEFT) {
		g.player.x -= g.player.speed
		if p8.ColorCollision(g.player.x, g.player.y, g.player.collisionColor) {
			g.player.x = beforeX
		}
	}
	if p8.Btn(p8.RIGHT) {
		g.player.x += g.player.speed
		if p8.ColorCollision(g.player.x, g.player.y, g.player.collisionColor) {
			g.player.x = beforeX
		}
	}
	if p8.Btn(p8.UP) {
		g.player.y -= g.player.speed
		if p8.ColorCollision(g.player.x, g.player.y, g.player.collisionColor) {
			g.player.y = beforeY
		}
	}
	if p8.Btn(p8.DOWN) {
		g.player.y += g.player.speed
		if p8.ColorCollision(g.player.x, g.player.y, g.player.collisionColor) {
			g.player.y = beforeY
		}
	}
}

// Draw is called every frame for rendering
func (g *Game) Draw() {
	p8.Cls()
	p8.Rectfill(g.player.x, g.player.y, g.player.x, g.player.y, 12)

	// Draw a labyrinth
	p8.Line(30, 30, 30, 100, g.player.collisionColor)
	p8.Line(30, 100, 100, 100, g.player.collisionColor)
	p8.Line(100, 100, 100, 30, g.player.collisionColor)
	p8.Line(100, 30, 35, 30, g.player.collisionColor)
	p8.Line(36, 30, 36, 60, g.player.collisionColor)
	p8.Line(36, 60, 75, 60, g.player.collisionColor)
	p8.Line(75, 60, 75, 70, g.player.collisionColor)
	p8.Line(75, 70, 35, 70, g.player.collisionColor)
	p8.Line(36, 70, 55, 80, g.player.collisionColor)
	p8.Line(55, 80, 90, 80, g.player.collisionColor)
}

func main() {
	// Create and insert our game
	game := &Game{}
	p8.InsertGame(game)

	// Start the game
	p8.Play()
}
