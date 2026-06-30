package main

import (
	"fmt"
	"math"

	p8 "github.com/drpaneas/pigo8"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	sceneBoundaryTiles = 16
	sceneBoundaryColor = 8
)

type sceneBoundaryLine struct {
	World  int
	Screen int
}

func (g *myGame) mapScreenLayout() mapModeLayout {
	return newMapModeLayout(width*unit, height*unit)
}

func (g *myGame) mapDataBounds() (int, int) {
	return len(g.mapData[0]), len(g.mapData)
}

func (g *myGame) handleMapMode() {
	layout := g.mapScreenLayout()
	canvas := layout.Canvas
	mapWidth, mapHeight := g.mapDataBounds()

	g.mapViewport.clamp(mapWidth, mapHeight, canvas)

	panStep := 1.0 / max(g.mapViewport.Zoom, mapMinZoom)
	panX := 0.0
	panY := 0.0

	if p8.Btn(p8.LEFT) || ebiten.IsKeyPressed(ebiten.KeyA) {
		panX -= panStep
	}
	if p8.Btn(p8.RIGHT) || ebiten.IsKeyPressed(ebiten.KeyD) {
		panX += panStep
	}
	if p8.Btn(p8.UP) || ebiten.IsKeyPressed(ebiten.KeyW) {
		panY -= panStep
	}
	if p8.Btn(p8.DOWN) || ebiten.IsKeyPressed(ebiten.KeyS) {
		panY += panStep
	}
	if panX != 0 || panY != 0 {
		g.mapViewport.pan(panX, panY, canvas, mapWidth, mapHeight)
	}

	mx, my := p8.GetMouseXY()
	if p8.Btnp(p8.ButtonMouseMiddle) {
		g.mapLastMouseX = mx
		g.mapLastMouseY = my
	}
	if p8.Btn(p8.ButtonMouseMiddle) {
		tileSize := g.mapViewport.tileSize()
		if tileSize > 0 {
			g.mapViewport.pan(
				-float64(mx-g.mapLastMouseX)/tileSize,
				-float64(my-g.mapLastMouseY)/tileSize,
				canvas,
				mapWidth,
				mapHeight,
			)
		}
		g.mapLastMouseX = mx
		g.mapLastMouseY = my
	}

	switch {
	case p8.Btnp(p8.ButtonMouseWheelUp):
		g.mapViewport.zoomAt(mapZoomStep, mx, my, canvas, mapWidth, mapHeight)
	case p8.Btnp(p8.ButtonMouseWheelDown):
		g.mapViewport.zoomAt(1/mapZoomStep, mx, my, canvas, mapWidth, mapHeight)
	}

	g.mapViewport.clamp(mapWidth, mapHeight, canvas)
	g.placeOrEraseSprites(layout)
}

func (g *myGame) drawMapMode() {
	layout := g.mapScreenLayout()
	mx, my := p8.GetMouseXY()
	tileX, tileY, ok := g.mapTargetTile(layout, mx, my)

	p8.Rectfill(
		layout.Canvas.X,
		layout.Canvas.Y,
		layout.Canvas.X+layout.Canvas.W-1,
		layout.Canvas.Y+layout.Canvas.H-1,
		0,
	)
	g.drawMapTiles(layout)
	g.drawMapHover(layout, tileX, tileY, ok)
	g.drawSceneBoundaries(layout)
	g.drawMapBottomStrip(layout, tileX, tileY, ok)
}

func (g *myGame) drawMapTiles(layout mapModeLayout) {
	mapWidth, mapHeight := g.mapDataBounds()
	minX, minY, maxX, maxY := g.mapViewport.visibleTileBounds(layout.Canvas, mapWidth, mapHeight)
	tileScale := g.mapViewport.tileSize() / mapBaseTileSize

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			screenX, screenY := g.mapViewport.worldToScreen(x, y, layout.Canvas)
			p8.Spr(g.mapData[y][x], screenX, screenY, tileScale, tileScale)
		}
	}
}

func (g *myGame) drawMapHover(layout mapModeLayout, tileX, tileY int, ok bool) {
	if !ok {
		return
	}

	mapWidth, mapHeight := g.mapDataBounds()
	tileSize := g.mapViewport.tileSize()

	for dy := 0; dy < g.safeGridSize(); dy++ {
		for dx := 0; dx < g.safeGridSize(); dx++ {
			targetX := tileX + dx
			targetY := tileY + dy
			if targetX < 0 || targetX >= mapWidth || targetY < 0 || targetY >= mapHeight {
				continue
			}

			screenX, screenY := g.mapViewport.worldToScreen(targetX, targetY, layout.Canvas)
			p8.Rect(
				screenX,
				screenY,
				screenX+tileSize-1,
				screenY+tileSize-1,
				g.getUIElementColor(),
			)
		}
	}
}

func visibleSceneBoundaryWorlds(camera, span float64, worldLimit int) []int {
	if span <= 0 || worldLimit <= 0 {
		return nil
	}

	visibleEnd := min(float64(worldLimit), camera+span)
	firstWorld := int(math.Ceil(camera/sceneBoundaryTiles)) * sceneBoundaryTiles
	lines := make([]int, 0, int(span/sceneBoundaryTiles)+1)
	for world := firstWorld; float64(world) < visibleEnd; world += sceneBoundaryTiles {
		lines = append(lines, world)
	}

	return lines
}

func visibleSceneBoundaryLines(vp mapViewport, canvas mapRect, mapWidth, mapHeight int) ([]sceneBoundaryLine, []sceneBoundaryLine) {
	spanX, spanY := vp.visibleTileSpan(canvas)
	verticalWorlds := visibleSceneBoundaryWorlds(vp.CameraX, spanX, mapWidth)
	horizontalWorlds := visibleSceneBoundaryWorlds(vp.CameraY, spanY, mapHeight)

	vertical := make([]sceneBoundaryLine, 0, len(verticalWorlds))
	for _, world := range verticalWorlds {
		screenX, _ := vp.worldToScreen(world, 0, canvas)
		vertical = append(vertical, sceneBoundaryLine{
			World:  world,
			Screen: int(math.Round(screenX)),
		})
	}

	horizontal := make([]sceneBoundaryLine, 0, len(horizontalWorlds))
	for _, world := range horizontalWorlds {
		_, screenY := vp.worldToScreen(0, world, canvas)
		horizontal = append(horizontal, sceneBoundaryLine{
			World:  world,
			Screen: int(math.Round(screenY)),
		})
	}

	return vertical, horizontal
}

func (g *myGame) drawSceneBoundaries(layout mapModeLayout) {
	mapWidth, mapHeight := g.mapDataBounds()
	vertical, horizontal := visibleSceneBoundaryLines(g.mapViewport, layout.Canvas, mapWidth, mapHeight)
	canvasRight := layout.Canvas.X + layout.Canvas.W - 1
	canvasBottom := layout.Canvas.Y + layout.Canvas.H - 1

	for _, line := range vertical {
		if line.Screen < layout.Canvas.X || line.Screen > canvasRight {
			continue
		}
		p8.Line(line.Screen, layout.Canvas.Y, line.Screen, canvasBottom, sceneBoundaryColor)
	}

	for _, line := range horizontal {
		if line.Screen < layout.Canvas.Y || line.Screen > canvasBottom {
			continue
		}
		p8.Line(layout.Canvas.X, line.Screen, canvasRight, line.Screen, sceneBoundaryColor)
	}
}

func (g *myGame) drawMapBottomStrip(layout mapModeLayout, tileX, tileY int, ok bool) {
	stripRight := layout.Strip.X + layout.Strip.W - 1
	stripBottom := layout.Strip.Y + layout.Strip.H - 1
	uiColor := g.getUIElementColor()

	p8.Rectfill(layout.Strip.X, layout.Strip.Y, stripRight, stripBottom, 1)
	p8.Rect(layout.Strip.X, layout.Strip.Y, stripRight, stripBottom, uiColor)

	p8.Rectfill(
		layout.SelectedTile.X,
		layout.SelectedTile.Y,
		layout.SelectedTile.X+layout.SelectedTile.W-1,
		layout.SelectedTile.Y+layout.SelectedTile.H-1,
		0,
	)
	p8.Spr(
		g.currentSprite,
		layout.SelectedTile.X,
		layout.SelectedTile.Y,
		float64(layout.SelectedTile.W)/mapBaseTileSize,
		float64(layout.SelectedTile.H)/mapBaseTileSize,
	)
	p8.Rect(
		layout.SelectedTile.X,
		layout.SelectedTile.Y,
		layout.SelectedTile.X+layout.SelectedTile.W-1,
		layout.SelectedTile.Y+layout.SelectedTile.H-1,
		uiColor,
	)

	hoverText := "tile: -, - sprite: -"
	if ok {
		hoverText = fmt.Sprintf(
			"tile: %d,%d sprite: %d",
			tileX,
			tileY,
			g.mapData[tileY][tileX],
		)
	}
	p8.Print(hoverText, layout.Coordinates.X, layout.Coordinates.Y-6, uiColor)
	p8.Print(
		fmt.Sprintf("selected: %d", g.currentSprite),
		layout.Coordinates.X,
		layout.Coordinates.Y+4,
		uiColor,
	)
}

func (g *myGame) mapTargetTile(layout mapModeLayout, mouseX, mouseY int) (int, int, bool) {
	tileX, tileY, ok := g.mapViewport.screenToTile(mouseX, mouseY, layout.Canvas)
	if !ok {
		return 0, 0, false
	}

	mapWidth, mapHeight := g.mapDataBounds()
	if tileX < 0 || tileX >= mapWidth || tileY < 0 || tileY >= mapHeight {
		return 0, 0, false
	}

	return tileX, tileY, true
}

func (g *myGame) placeOrEraseSprites(layout mapModeLayout) {
	mx, my := p8.GetMouseXY()
	tileX, tileY, ok := g.mapTargetTile(layout, mx, my)
	if !ok {
		return
	}

	if p8.Btn(p8.ButtonMouseRight) {
		g.eraseAt(tileX, tileY)
		return
	}
	if p8.Btn(p8.ButtonMouseLeft) {
		g.placeGridSprites(tileX, tileY)
	}
}
