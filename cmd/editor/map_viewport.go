package main

import "math"

const (
	mapBaseTileSize = 8.0
	mapMinZoom      = 0.5
	mapMaxZoom      = 4.0
	mapZoomStep     = 1.25
)

type mapRect struct {
	X int
	Y int
	W int
	H int
}

type mapViewport struct {
	CameraX float64
	CameraY float64
	Zoom    float64
}

var _ = mapViewport{}.worldToScreen

func defaultMapViewport() mapViewport {
	return mapViewport{Zoom: 1}
}

func (vp mapViewport) tileSize() float64 {
	zoom := vp.Zoom
	if zoom == 0 {
		zoom = defaultMapViewport().Zoom
	}

	zoom = min(max(zoom, mapMinZoom), mapMaxZoom)
	return mapBaseTileSize * zoom
}

func (vp mapViewport) visibleTileSpan(canvas mapRect) (float64, float64) {
	tileSize := vp.tileSize()
	if tileSize == 0 {
		return 0, 0
	}

	return float64(canvas.W) / tileSize, float64(canvas.H) / tileSize
}

func (vp *mapViewport) clamp(mapWidth, mapHeight int, canvas mapRect) {
	spanX, spanY := vp.visibleTileSpan(canvas)
	_, _, maxTileX, maxTileY := vp.visibleTileBounds(canvas, mapWidth, mapHeight)
	maxX := max(0.0, float64(maxTileX)-spanX)
	maxY := max(0.0, float64(maxTileY)-spanY)

	vp.CameraX = min(max(vp.CameraX, 0), maxX)
	vp.CameraY = min(max(vp.CameraY, 0), maxY)
}

func (vp *mapViewport) pan(dx, dy float64, canvas mapRect, mapWidth, mapHeight int) {
	vp.CameraX += dx
	vp.CameraY += dy
	vp.clamp(mapWidth, mapHeight, canvas)
}

func (vp *mapViewport) zoomAt(multiplier float64, screenX, screenY int, canvas mapRect, mapWidth, mapHeight int) {
	worldX, worldY, ok := vp.screenToWorld(screenX, screenY, canvas)
	if !ok {
		return
	}

	if multiplier <= 0 {
		return
	}

	zoom := vp.Zoom
	if zoom == 0 {
		zoom = defaultMapViewport().Zoom
	}

	vp.Zoom = min(max(zoom*multiplier, mapMinZoom), mapMaxZoom)
	tileSize := vp.tileSize()
	vp.CameraX = worldX - float64(screenX-canvas.X)/tileSize
	vp.CameraY = worldY - float64(screenY-canvas.Y)/tileSize
	vp.clamp(mapWidth, mapHeight, canvas)
}

func (vp mapViewport) screenToWorld(screenX, screenY int, canvas mapRect) (float64, float64, bool) {
	if screenX < canvas.X || screenX >= canvas.X+canvas.W || screenY < canvas.Y || screenY >= canvas.Y+canvas.H {
		return 0, 0, false
	}

	tileSize := vp.tileSize()
	worldX := vp.CameraX + float64(screenX-canvas.X)/tileSize
	worldY := vp.CameraY + float64(screenY-canvas.Y)/tileSize
	return worldX, worldY, true
}

func (vp mapViewport) screenToTile(screenX, screenY int, canvas mapRect) (int, int, bool) {
	worldX, worldY, ok := vp.screenToWorld(screenX, screenY, canvas)
	if !ok {
		return 0, 0, false
	}

	return int(math.Floor(worldX)), int(math.Floor(worldY)), true
}

func (vp mapViewport) worldToScreen(tileX, tileY int, canvas mapRect) (float64, float64) {
	tileSize := vp.tileSize()
	screenX := float64(canvas.X) + (float64(tileX)-vp.CameraX)*tileSize
	screenY := float64(canvas.Y) + (float64(tileY)-vp.CameraY)*tileSize
	return screenX, screenY
}

func (vp mapViewport) visibleTileBounds(canvas mapRect, mapWidth, mapHeight int) (int, int, int, int) {
	spanX, spanY := vp.visibleTileSpan(canvas)

	minX := max(0, int(math.Floor(vp.CameraX)))
	minY := max(0, int(math.Floor(vp.CameraY)))
	maxX := min(mapWidth, int(math.Ceil(vp.CameraX+spanX)))
	maxY := min(mapHeight, int(math.Ceil(vp.CameraY+spanY)))

	return minX, minY, maxX, maxY
}
