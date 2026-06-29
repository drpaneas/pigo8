package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapTargetTileUsesViewport(t *testing.T) {
	g := &myGame{
		mapViewport: mapViewport{CameraX: 12.5, CameraY: 7.25, Zoom: 1.5},
	}
	layout := mapModeLayout{
		Canvas: mapRect{X: 10, Y: 10, W: 160, H: 120},
	}

	tileX, tileY, ok := g.mapTargetTile(layout, 34, 26)

	assert.True(t, ok)
	assert.Equal(t, 14, tileX)
	assert.Equal(t, 8, tileY)
}

func TestMapTargetTileRejectsOutsideCanvas(t *testing.T) {
	g := &myGame{
		mapViewport: mapViewport{CameraX: 10, CameraY: 10, Zoom: 1},
	}
	layout := mapModeLayout{
		Canvas: mapRect{X: 10, Y: 10, W: 160, H: 120},
	}

	_, _, ok := g.mapTargetTile(layout, 9, 50)
	assert.False(t, ok)
}

func TestMapTargetTileUsesFullEditorMapBounds(t *testing.T) {
	g := &myGame{
		mapViewport: mapViewport{CameraX: 200.5, CameraY: 150.5, Zoom: 1},
	}
	layout := mapModeLayout{
		Canvas: mapRect{X: 10, Y: 10, W: 160, H: 120},
	}

	tileX, tileY, ok := g.mapTargetTile(layout, 10, 10)

	assert.True(t, ok)
	assert.Equal(t, 200, tileX)
	assert.Equal(t, 150, tileY)
}
