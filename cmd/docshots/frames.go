package main

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
)

// framesToGIF converts a sequence of RGBA frames into an animated GIF,
// quantizing each frame to a fixed 256-color palette and showing each frame
// for delayMs milliseconds before advancing (looping indefinitely).
func framesToGIF(frames []image.Image, delayMs int) (*gif.GIF, error) {
	out := &gif.GIF{LoopCount: 0}
	delayHundredths := delayMs / 10
	if delayHundredths < 1 {
		delayHundredths = 1
	}

	for _, frame := range frames {
		bounds := frame.Bounds()
		paletted := image.NewPaletted(bounds, palette.Plan9)
		draw.Draw(paletted, bounds, frame, bounds.Min, draw.Src)

		out.Image = append(out.Image, paletted)
		out.Delay = append(out.Delay, delayHundredths)
	}

	return out, nil
}

// bestFrame picks a representative single frame for a static screenshot -
// currently the last captured frame, since by then any startup/loading
// artifacts have settled.
func bestFrame(frames []image.Image) image.Image {
	if len(frames) == 0 {
		return nil
	}
	return frames[len(frames)-1]
}
