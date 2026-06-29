package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapViewportTileSize(t *testing.T) {
	vp := mapViewport{Zoom: 2}
	assert.InDelta(t, 16.0, vp.tileSize(), 0.001)
}

func TestMapViewportScreenToTile(t *testing.T) {
	vp := mapViewport{CameraX: 12.5, CameraY: 7.25, Zoom: 1.5}
	canvas := mapRect{X: 10, Y: 10, W: 160, H: 120}

	tileX, tileY, ok := vp.screenToTile(34, 26, canvas)

	assert.True(t, ok)
	assert.Equal(t, 14, tileX)
	assert.Equal(t, 8, tileY)
}

func TestMapViewportClamp(t *testing.T) {
	vp := mapViewport{CameraX: 200, CameraY: 200, Zoom: 1}
	canvas := mapRect{X: 10, Y: 10, W: 160, H: 120}

	vp.clamp(128, 128, canvas)

	assert.InDelta(t, 108.0, vp.CameraX, 0.001)
	assert.InDelta(t, 113.0, vp.CameraY, 0.001)
}

func TestMapViewportPan(t *testing.T) {
	vp := mapViewport{CameraX: 10, CameraY: 20, Zoom: 1}
	canvas := mapRect{X: 10, Y: 10, W: 160, H: 120}

	vp.pan(2.5, -3.0, canvas, 128, 128)

	assert.InDelta(t, 12.5, vp.CameraX, 0.001)
	assert.InDelta(t, 17.0, vp.CameraY, 0.001)
}

func TestMapViewportZoomAtKeepsAnchorStable(t *testing.T) {
	vp := mapViewport{CameraX: 20, CameraY: 10, Zoom: 1}
	canvas := mapRect{X: 10, Y: 10, W: 160, H: 120}

	beforeWorldX, beforeWorldY, ok := vp.screenToWorld(90, 70, canvas)
	assert.True(t, ok)
	beforeX, beforeY, ok := vp.screenToTile(90, 70, canvas)
	assert.True(t, ok)

	vp.zoomAt(mapZoomStep, 90, 70, canvas, 128, 128)

	afterWorldX, afterWorldY, ok := vp.screenToWorld(90, 70, canvas)
	assert.True(t, ok)
	afterX, afterY, ok := vp.screenToTile(90, 70, canvas)
	assert.True(t, ok)
	assert.InDelta(t, beforeWorldX, afterWorldX, 0.000001)
	assert.InDelta(t, beforeWorldY, afterWorldY, 0.000001)
	assert.Equal(t, beforeX, afterX)
	assert.Equal(t, beforeY, afterY)
}
