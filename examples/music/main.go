// Package main provides a music example for the PIGO8 engine
package main

//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .

import (
	"log"

	p8 "github.com/drpaneas/pigo8"
)

// Game implements the PIGO8 Cartridge interface
type Game struct{}

// Init is called once at the start of the game
func (g *Game) Init() {
	log.Println("Music example initialized")
}

// Update is called every frame for game logic
func (g *Game) Update() {
	if p8.Btnp(p8.UP) {
		log.Println("Playing music 3")
		p8.Music(3)
	}

	if p8.Btnp(p8.DOWN) {
		log.Println("Playing music 4")
		p8.Music(4)
	}

	if p8.Btnp(p8.LEFT) {
		log.Println("Playing music 5")
		p8.Music(5)
	}

	if p8.Btnp(p8.RIGHT) {
		log.Println("Playing music 6 exclusively")
		p8.Music(6, true)
	}

	if p8.Btnp(p8.O) {
		log.Println("Looping music 0 exclusively")
		p8.MusicLoop(0, true)
	}

	if p8.Btnp(p8.X) {
		log.Println("Stopping music 0")
		p8.StopMusic(0)
	}

	if p8.Btnp(p8.ButtonStart) {
		log.Println("Stopping all music")
		p8.StopMusic(-1)
	}
}

// Draw is called every frame for rendering
func (g *Game) Draw() {
	// Clear the screen with dark blue
	p8.Cls(1)

	// Draw instructions
	p8.Print("Music Example", 30, 10, 7)
	p8.Print("Up to play music 3", 10, 35, 7)
	p8.Print("Down to play music 4", 10, 45, 7)
	p8.Print("Left to play music 5", 10, 55, 7)
	p8.Print("Right to play music 6 exclusively", 10, 65, 7)
	p8.Print("O to loop music 0 exclusively", 10, 75, 7)
	p8.Print("X to stop music 0", 10, 85, 7)
	p8.Print("Start to stop all music", 10, 95, 7)

}

func main() {
	// Create and insert our game
	game := &Game{}
	p8.InsertGame(game)

	// Configure the PIGO8 console
	settings := p8.NewSettings()
	settings.WindowTitle = "PIGO8 Music Example"
	settings.ScaleFactor = 4

	// Start the game
	p8.PlayGameWith(settings)
}
