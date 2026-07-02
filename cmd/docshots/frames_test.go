package main

import (
	"image"
	"image/color"
	"testing"
)

func TestFramesToGIF(t *testing.T) {
	// Two 4x4 frames: solid red, then solid blue.
	red := image.NewRGBA(image.Rect(0, 0, 4, 4))
	blue := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			red.Set(x, y, color.RGBA{R: 255, A: 255})
			blue.Set(x, y, color.RGBA{B: 255, A: 255})
		}
	}

	g, err := framesToGIF([]image.Image{red, blue}, 100)
	if err != nil {
		t.Fatalf("framesToGIF: %v", err)
	}
	if len(g.Image) != 2 {
		t.Fatalf("expected 2 encoded frames, got %d", len(g.Image))
	}
	if len(g.Delay) != 2 || g.Delay[0] != 10 {
		// GIF delay is in 1/100ths of a second; 100ms == 10.
		t.Errorf("expected delay [10, 10], got %v", g.Delay)
	}
}

func TestBestFrame(t *testing.T) {
	frames := []image.Image{
		image.NewRGBA(image.Rect(0, 0, 2, 2)),
		image.NewRGBA(image.Rect(0, 0, 2, 2)),
		image.NewRGBA(image.Rect(0, 0, 2, 2)),
	}
	got := bestFrame(frames)
	if got == nil {
		t.Fatal("expected a non-nil frame")
	}
}
