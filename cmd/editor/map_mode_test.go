package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisibleSceneBoundaryWorldsReturnsVisibleSceneIntervals(t *testing.T) {
	assert.Equal(t, []int{16, 32}, visibleSceneBoundaryWorlds(12.5, 20, 320))
	assert.Equal(t, []int{16, 32}, visibleSceneBoundaryWorlds(12.5, 20.5, 320))
	assert.Equal(t, []int{304}, visibleSceneBoundaryWorlds(300.25, 19.75, 320))
}

func TestVisibleSceneBoundaryLinesStayWorldAnchoredWhenZoomed(t *testing.T) {
	canvas := mapRect{X: 10, Y: 20, W: 160, H: 120}
	vp := mapViewport{CameraX: 14.25, CameraY: 30.5, Zoom: 4}

	vertical, horizontal := visibleSceneBoundaryLines(vp, canvas, 320, 320)

	assert.Equal(t, []sceneBoundaryLine{{World: 16, Screen: 66}}, vertical)
	assert.Equal(t, []sceneBoundaryLine{{World: 32, Screen: 68}}, horizontal)
}

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
