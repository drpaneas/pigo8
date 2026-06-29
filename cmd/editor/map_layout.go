package main

const (
	mapModePadding     = 10
	mapModeGap         = 8
	mapModeStripHeight = 44
	mapMiniMapWidth    = 96
	mapMiniMapHeight   = 32
	mapPreviewSize     = 28
)

type mapModeLayout struct {
	Canvas       mapRect
	Strip        mapRect
	SelectedTile mapRect
	Coordinates  mapRect
	MiniMap      mapRect
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

	coordinates := mapRect{
		X: selectedTile.X + selectedTile.W + 12,
		Y: strip.Y + (strip.H-20)/2,
		W: 200,
		H: 20,
	}

	miniMap := mapRect{
		X: strip.X + strip.W - mapModeGap - mapMiniMapWidth,
		Y: strip.Y + (strip.H-mapMiniMapHeight)/2,
		W: mapMiniMapWidth,
		H: mapMiniMapHeight,
	}

	return mapModeLayout{
		Canvas:       canvas,
		Strip:        strip,
		SelectedTile: selectedTile,
		Coordinates:  coordinates,
		MiniMap:      miniMap,
	}
}

func miniMapViewportRect(miniMap, canvas mapRect, vp mapViewport, mapWidth, mapHeight int) mapRect {
	if mapWidth <= 0 || mapHeight <= 0 || canvas.W <= 0 || canvas.H <= 0 {
		return mapRect{}
	}

	spanX, spanY := vp.visibleTileSpan(canvas)
	scaledWidth := int(spanX * float64(miniMap.W) / float64(mapWidth))
	scaledHeight := int(spanY * float64(miniMap.H) / float64(mapHeight))
	if spanX > 0 && scaledWidth == 0 {
		scaledWidth = 1
	}
	if spanY > 0 && scaledHeight == 0 {
		scaledHeight = 1
	}

	return mapRect{
		X: miniMap.X + int(vp.CameraX*float64(miniMap.W)/float64(mapWidth)),
		Y: miniMap.Y + int(vp.CameraY*float64(miniMap.H)/float64(mapHeight)),
		W: scaledWidth,
		H: scaledHeight,
	}
}
