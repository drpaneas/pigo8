package main

const (
	mapModePadding     = 10
	mapModeGap         = 8
	mapModeStripHeight = 44
	mapPreviewSize     = 28
)

type mapModeLayout struct {
	Canvas       mapRect
	Strip        mapRect
	SelectedTile mapRect
	Coordinates  mapRect
}

func newMapModeLayout(screenWidth, screenHeight int) mapModeLayout {
	canvasHeight := screenHeight - (mapModePadding * 2) - mapModeGap - mapModeStripHeight
	canvas := mapRect{
		X: mapModePadding,
		Y: mapModePadding,
		W: screenWidth - (mapModePadding * 2),
		H: canvasHeight,
	}

	strip := mapRect{
		X: mapModePadding,
		Y: canvas.Y + canvas.H + mapModeGap,
		W: canvas.W,
		H: mapModeStripHeight,
	}

	selectedTile := mapRect{
		X: strip.X + mapModeGap,
		Y: strip.Y + mapModeGap,
		W: mapPreviewSize,
		H: mapPreviewSize,
	}

	coordinatesX := selectedTile.X + selectedTile.W + 12
	coordinates := mapRect{
		X: coordinatesX,
		Y: strip.Y + (strip.H-20)/2,
		W: strip.W - (coordinatesX - strip.X) - (mapModeGap * 2),
		H: 20,
	}

	return mapModeLayout{
		Canvas:       canvas,
		Strip:        strip,
		SelectedTile: selectedTile,
		Coordinates:  coordinates,
	}
}
