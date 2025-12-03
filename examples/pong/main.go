// Package main is a simple Pong clone using go-pigo8
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

// Court boundaries
const (
	courtLeft   = 0
	courtRight  = 127
	courtTop    = 10
	courtBottom = 127
	centerX     = (courtRight + courtLeft) / 2
	centerY     = (courtBottom + courtTop) / 2
	lineLen     = 4
)

// Paddle represents a player or computer paddle
type Paddle struct {
	x, y, width, height, speed float64
	color                      int
}

// Ball holds position, velocity, and rendering info
type Ball struct {
	x, y, size           float64
	dx, dy, speed, boost float64
	color                int
}

// Game encapsulates all game state
type Game struct {
	player        Paddle
	computer      Paddle
	ball          Ball
	playerScore   int
	computerScore int
	Scored        string
}

// Init initializes the game state with default paddle and ball positions
func (g *Game) Init() {
	difficulty := 1.5
	paddleHeight := 10.0 // 10
	paddleWidth := 2.0   // 2
	g.player = Paddle{courtLeft + paddleWidth*2, centerY + paddleHeight/2, paddleWidth, paddleHeight, 1.0 * difficulty, 12}
	g.computer = Paddle{courtRight - paddleWidth*3, centerY + paddleHeight/2, paddleWidth, paddleHeight, 0.75 * difficulty, 8}
	ballDy := float64(p8.Flr(p8.Rnd(2))) - 0.5
	g.ball = Ball{x: centerX, y: centerY, size: 2, color: 7, dx: 1.0 * difficulty, dy: ballDy, speed: 1.0 * difficulty, boost: 0.05 * difficulty}

	if p8.Restart {
		g.playerScore = 0
		g.computerScore = 0
		g.Scored = ""
		p8.Restart = false
	}

	// sound
	switch g.Scored {
	case "Player":
		p8.Music(3)
	case "Computer":
		p8.Music(4)
	default:
		p8.Music(5)
	}
}

// Update handles game logic each frame including input, AI, collisions and scoring
func (g *Game) Update() {

	// Player input
	if p8.Btn(p8.UP) && g.player.y > courtTop+1 {
		g.player.y -= g.player.speed
	}
	if p8.Btn(p8.DOWN) && g.player.y+g.player.height < courtBottom-1 {
		g.player.y += g.player.speed
	}

	// Simple AI: track ball when it's moving toward computer
	mid := g.computer.y + g.computer.height/2
	if g.ball.dx > 0 {
		if mid > g.ball.y && g.computer.y > courtTop+1 {
			g.computer.y -= g.computer.speed
		}
		if mid < g.ball.y && g.computer.y+g.computer.height < courtBottom-1 {
			g.computer.y += g.computer.speed
		}
	} else {
		// return to center
		if mid > ((centerY + g.player.height/2) + g.player.height) {
			g.computer.y -= g.computer.speed
		}
		if mid < ((centerY + g.player.height/2) - g.player.height) {
			g.computer.y += g.computer.speed
		}
	}

	// Collisions
	// 1. Ball vs paddles
	if collide(g.ball, g.computer) {
		g.ball.dx = -(g.ball.dx + g.ball.boost)
		p8.Music(0)
	}
	if collide(g.ball, g.player) {
		// adjust dy if player changes paddle angle
		if p8.Btn(p8.UP) || p8.Btn(p8.DOWN) {
			g.ball.dy += p8.Sign(g.ball.dy) * g.ball.boost * 2
		}
		g.ball.dx = -(g.ball.dx - g.ball.boost)
		p8.Music(1)
	}

	// 2. Ball vs top/bottom
	if g.ball.y <= courtTop+1 || g.ball.y+g.ball.size >= courtBottom-1 {
		g.ball.dy = -g.ball.dy
		p8.Music(2)
	}

	// 3. Ball vs Walls (aka scoring)
	if g.ball.x > courtRight {
		g.playerScore++
		g.Scored = "Player"
		g.Init()
	}
	if g.ball.x < courtLeft {
		g.computerScore++
		g.Scored = "Computer"
		g.Init()
	}

	// Move ball
	g.ball.x += g.ball.dx
	g.ball.y += g.ball.dy
}

// Draw renders the game elements to the screen each frame
func (g *Game) Draw() {
	p8.Cls(0)

	// Court outline
	p8.Rect(courtLeft, courtTop, courtRight, courtBottom, 5)

	// Center dashed line
	for y := courtTop; y < courtBottom; y += lineLen * 2 {
		p8.Line(centerX, float64(y), centerX, float64(y+lineLen), 5)
	}

	// Ball and paddles
	p8.Rectfill(g.ball.x, g.ball.y, g.ball.x+g.ball.size, g.ball.y+g.ball.size, g.ball.color)
	p8.Rectfill(g.player.x, g.player.y, g.player.x+g.player.width, g.player.y+g.player.height, g.player.color)
	p8.Rectfill(g.computer.x, g.computer.y, g.computer.x+g.computer.width, g.computer.y+g.computer.height, g.computer.color)

	// Scores
	p8.Print(g.playerScore, centerX/2, 2, 12)
	p8.Print(g.computerScore, centerX+centerX/2, 2, 8)
}

// collide checks axis-aligned collision between ball and paddle
func collide(b Ball, p Paddle) bool {
	return b.x+b.size >= p.x && b.x <= p.x+p.width &&
		b.y+b.size >= p.y && b.y <= p.y+p.height
}

func main() {
	settings := p8.NewSettings()
	settings.TargetFPS = 60
	settings.WindowTitle = "PIGO-8 Pong"
	p8.InsertGame(&Game{})
	p8.PlayGameWith(settings)
}
