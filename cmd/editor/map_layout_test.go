package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMapModeLayout(t *testing.T) {
	layout := newMapModeLayout(384, 216)

	assert.Equal(t, mapRect{X: 10, Y: 10, W: 364, H: 144}, layout.Canvas)
	assert.Equal(t, mapRect{X: 10, Y: 162, W: 364, H: 44}, layout.Strip)
	assert.Equal(t, mapRect{X: 18, Y: 170, W: 28, H: 28}, layout.SelectedTile)
	assert.Equal(t, mapRect{X: 58, Y: 174, W: 300, H: 20}, layout.Coordinates)
}

func TestNewMapModeLayoutUsesFullBottomStripForInfo(t *testing.T) {
	layout := newMapModeLayout(384, 216)

	assert.Equal(
		t,
		layout.Strip.X+layout.Strip.W-(mapModeGap*2),
		layout.Coordinates.X+layout.Coordinates.W,
	)
}
