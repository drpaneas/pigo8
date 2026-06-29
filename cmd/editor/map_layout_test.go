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
	assert.Equal(t, mapRect{X: 58, Y: 174, W: 200, H: 20}, layout.Coordinates)
	assert.Equal(t, mapRect{X: 270, Y: 168, W: 96, H: 32}, layout.MiniMap)
}

func TestMiniMapViewportRect(t *testing.T) {
	layout := newMapModeLayout(384, 216)
	vp := mapViewport{CameraX: 32, CameraY: 16, Zoom: 1}

	box := miniMapViewportRect(layout.MiniMap, layout.Canvas, vp, 128, 128)

	assert.Equal(t, mapRect{X: 294, Y: 172, W: 34, H: 4}, box)
}

func TestMiniMapViewportRectProductionMap(t *testing.T) {
	layout := newMapModeLayout(384, 216)
	vp := mapViewport{CameraX: 64, CameraY: 80, Zoom: 4}

	box := miniMapViewportRect(layout.MiniMap, layout.Canvas, vp, 320, 320)

	assert.Equal(t, mapRect{X: 289, Y: 176, W: 3, H: 1}, box)
}
