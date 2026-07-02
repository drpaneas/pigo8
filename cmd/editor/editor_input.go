package main

import (
	"log"
	"time"

	p8 "github.com/drpaneas/pigo8"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Refactored Update method with reduced cyclomatic complexity and integrated save-on-toggle logic
func (g *myGame) Update() {
	g.toggleMapMode()
	g.handleUndoRedo() // Handle undo/redo in both modes
	if g.mapMode {
		g.handleMapMode()
	} else {
		g.handleEditorMode()
	}
}

// toggleMapMode flips mapMode and saves data on entry/exit
func (g *myGame) toggleMapMode() {
	if p8.Btnp(p8.X) {
		// Save current state before switching modes to ensure all changes are captured
		if err := g.saveState(); err != nil {
			log.Printf("Error saving state before mode switch: %v", err)
		}

		g.mapMode = !g.mapMode
		if g.mapMode {
			// When entering map mode, save the current spritesheet first
			if err := saveSpritesheet(); err != nil {
				log.Printf("Error saving spritesheet before entering map mode: %v", err)
				// Decide if this is a fatal error or if we can proceed
				// For now, we'll log and continue, but PIGO-8 might not have the latest sprites
			}
			// Then, instruct PIGO-8 to reload this spritesheet
			if err := p8.LoadSpritesheet("spritesheet.json"); err != nil {
				log.Printf("Error loading spritesheet into PIGO-8: %v", err)
				// Similar to above, log and continue for now
			}
		} else {
			if err := g.saveMapData(); err != nil {
				log.Printf("Error saving map (changes not persisted to disk): %v", err)
				// Log and continue, consistent with the spritesheet save/load
				// handling above - a failed save shouldn't crash the editor.
			}
		}
	}
}

func (g *myGame) eraseAt(x, y int) {
	if g.inBounds(x, y) {
		// Only save if this cell is non-zero (actual change)
		if g.mapData[y][x] != 0 {
			log.Printf("Erasing at (%d,%d) - was %d", x, y, g.mapData[y][x])
			p8.Mset(x, y, 0)
			g.mapData[y][x] = 0
			g.saveCurrentStateIfNeeded()
			log.Printf("After erase - map[%d][%d] = %d, p8.Mget = %d",
				y, x, g.mapData[y][x], p8.Mget(x, y))
		}
	}
}

func (g *myGame) inBounds(x, y int) bool {
	mapWidth, mapHeight := g.mapDataBounds()
	return x >= 0 && x < mapWidth && y >= 0 && y < mapHeight
}

func (g *myGame) placeGridSprites(x, y int) {
	w, h := g.gridSize, g.gridSize
	if w < 1 {
		w, h = 1, 1
	}
	base := g.currentSprite
	changed := false

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			tx, ty := x+dx, y+dy
			if !g.inBounds(tx, ty) {
				log.Printf("  Out of bounds: (%d,%d)", tx, ty)
				continue
			}
			row := base/spriteSheetCols + dy
			col := base%spriteSheetCols + dx
			if row < spriteSheetRows && col < spriteSheetCols {
				idx := row*spriteSheetCols + col
				// Only mark as changed if we're actually changing the value
				if g.mapData[ty][tx] != idx {
					p8.Mset(tx, ty, idx)
					g.mapData[ty][tx] = idx
					changed = true
				}
			} else {
				log.Printf("  Invalid sprite position: row=%d, col=%d (max %d,%d)",
					row, col, spriteSheetRows-1, spriteSheetCols-1)
			}
		}
	}

	// Only save state if something was actually changed
	if changed {
		g.saveCurrentStateIfNeeded()
	}
}

// canTriggerAction checks if enough time has passed since the last action
func (g *myGame) canTriggerAction(lastActionTime *int64) bool {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-*lastActionTime < 200 { // 200ms cooldown
		return false
	}
	*lastActionTime = now
	return true
}

// handleUndoRedo manages the undo/redo key presses with proper debouncing
func (g *myGame) handleUndoRedo() {
	if g.undoInProgress { // Prevent re-entrant calls
		return
	}

	// Initialize key cooldown if not set
	if g.keyCooldown == 0 {
		g.keyCooldown = 200 // 200ms cooldown by default
	}

	// Check for Cmd+Z (Undo) or Cmd+Shift+Z (Redo)
	if ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				// Redo with Cmd+Shift+Z
				if g.canTriggerAction(&g.lastRedoTime) {
					g.undoInProgress = true
					defer func() { g.undoInProgress = false }()
					g.redo()
				}
			} else {
				// Undo with Cmd+Z
				if g.canTriggerAction(&g.lastUndoTime) {
					g.undoInProgress = true
					defer func() { g.undoInProgress = false }()
					g.undo()
				}
			}
		}
	}
}

// -------------------- Editor Mode --------------------
func (g *myGame) handleEditorMode() {
	mx, my := p8.GetMouseXY()
	g.toggleSpriteFlags(mx, my)
	g.handleDrawingGrid(mx, my)
	g.handleSpriteSelection(mx, my)
	g.handlePaletteSelection(mx, my)
	g.handleWheel()
	g.handleKeyboardNavigation()
	g.handleCopyPaste()

	// Handle undo/redo with proper debouncing
	g.handleUndoRedo()
}

func (g *myGame) toggleSpriteFlags(mx, my int) {
	// Use the same coordinates and size as in drawCheckboxes
	const checkboxSize = 8
	const offset = 15 // Match the offset from drawSelectionAndPalette

	// Get the same base coordinates used in drawCheckboxes
	baseX := 10
	baseY := 10 + 8*12 - 2 + offset // Add the offset to match drawCheckboxes position

	for i := 0; i < 8; i++ {
		checkboxX := baseX + i*checkboxSize*3/2
		checkboxY := baseY
		if mx >= checkboxX && mx < checkboxX+checkboxSize && my >= checkboxY && my < checkboxY+checkboxSize && p8.Btnp(p8.ButtonMouseLeft) {
			g.toggleFlagAtIndex(i)
		}
	}
}

func (g *myGame) toggleFlagAtIndex(i int) {
	g.saveCurrentStateIfNeeded()

	base := g.currentSprite
	r, c := base/spriteSheetCols, base%spriteSheetCols
	cur := spriteFlags[r][c][i]
	for dr := range g.safeGridSize() {
		for dc := range g.safeGridSize() {
			rr, cc := r+dr, c+dc
			if rr < spriteSheetRows && cc < spriteSheetCols {
				spriteFlags[rr][cc][i] = !cur
			}
		}
	}
	if err := saveSpritesheet(); err != nil {
		log.Printf("Error saving spritesheet after toggling flag: %v", err)
	}
}

func (g *myGame) safeGridSize() int {
	if g.gridSize < 1 {
		return 1
	}
	return g.gridSize
}

func (g *myGame) handleDrawingGrid(mx, my int) {
	const gx, gy, size = 10, 10, 8
	gridPx := size * g.gridSize
	cell := max(1, 96/gridPx)
	row := (my - gy) / cell
	col := (mx - gx) / cell

	if row < 0 || row >= gridPx || col < 0 || col >= gridPx {
		g.hoverX, g.hoverY = -1, -1
		return
	}
	g.updateHover(row, col)
	if p8.Btn(p8.ButtonMouseLeft) {
		g.drawAt(row, col, g.currentColor)
	} else if p8.Btn(p8.ButtonMouseRight) {
		g.drawAt(row, col, 0)
	}
}

func (g *myGame) updateHover(row, col int) {
	base := g.currentSprite
	r := base/spriteSheetCols + row/8
	c := base%spriteSheetCols + col/8
	pr, pc := row%8, col%8
	g.hoverX, g.hoverY = c*8+pc, r*8+pr
}

func (g *myGame) handleSpriteSelection(mx, my int) {
	row := (my - 10) / spriteCellSize
	col := (mx - spritesheetStartX) / spriteCellSize
	if row >= 0 && row < spriteSheetRows && col >= 0 && col < spriteSheetCols && p8.Btnp(p8.ButtonMouseLeft) {
		idx := row*spriteSheetCols + col
		if idx > 0 {
			g.currentSprite = idx
		}
		g.updateDrawingCanvas()
	}
}

func (g *myGame) handlePaletteSelection(mx, my int) {
	const gx, gy = 10, 10 + 8*12 - 2 + 40
	row := (my - gy) / 12
	col := (mx - gx) / 12
	colors := p8.GetPaletteSize()
	if row >= 0 && col >= 0 && row*8+col < colors && p8.Btnp(p8.ButtonMouseLeft) {
		g.currentColor = row*8 + col
	}
}

func (g *myGame) handleWheel() {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-g.lastWheelTime <= 150 { // 150ms debounce for wheel and keyboard
		return
	}
	if p8.Btnp(p8.ButtonMouseWheelUp) && g.gridSize < 4 {
		g.gridSize = min(g.gridSize*2, 8)
		g.lastWheelTime = now
		g.updateDrawingCanvas()
	} else if p8.Btnp(p8.ButtonMouseWheelDown) && g.gridSize > 1 {
		g.gridSize = max(1, g.gridSize/2)
		g.lastWheelTime = now
		g.updateDrawingCanvas()
	}
}

// handleKeyboardNavigation handles keyboard arrow key navigation between sprites
// copySprite copies the current sprite data to the clipboard
func (g *myGame) copySprite() {
	r, c := g.currentSprite/spriteSheetCols, g.currentSprite%spriteSheetCols
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			g.copiedSprite[y][x] = spritesheet[r][c][y][x]
		}
	}
}

// pasteSprite pastes the copied sprite data to the current sprite
func (g *myGame) pasteSprite() {
	r, c := g.currentSprite/spriteSheetCols, g.currentSprite%spriteSheetCols
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			spritesheet[r][c][y][x] = g.copiedSprite[y][x]
		}
	}
	updateMapSprites(g.currentSprite)
	g.updateDrawingCanvas()
}

// handleCopyPaste handles keyboard shortcuts for copy and paste
func (g *myGame) handleCopyPaste() {
	// Check for CMD+C (Copy)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) && (ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl)) {
		g.copySprite()
	}

	// Check for CMD+V (Paste)
	if inpututil.IsKeyJustPressed(ebiten.KeyV) && (ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl)) {
		g.saveCurrentStateIfNeeded()
		g.pasteSprite()
	}
}

// spriteNavigationCandidate computes the sprite index reached by moving
// dRow/dCol grid steps from currentSprite, or returns currentSprite
// unchanged if that move would leave the spritesheet grid or land on sprite
// 0 (reserved - see Init - and never selectable, matching
// handleSpriteSelection's mouse-click behavior).
func spriteNavigationCandidate(currentSprite, dRow, dCol int) int {
	currentRow := currentSprite / spriteSheetCols
	currentCol := currentSprite % spriteSheetCols

	newRow := currentRow + dRow
	newCol := currentCol + dCol
	if newRow < 0 || newRow >= spriteSheetRows || newCol < 0 || newCol >= spriteSheetCols {
		return currentSprite
	}

	candidate := newRow*spriteSheetCols + newCol
	if candidate == 0 {
		return currentSprite
	}
	return candidate
}

func (g *myGame) handleKeyboardNavigation() {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-g.lastWheelTime <= 150 { // 150ms debounce for keyboard navigation
		return
	}

	var dRow, dCol int
	switch {
	case p8.Btnp(p8.LEFT):
		dCol = -1
	case p8.Btnp(p8.RIGHT):
		dCol = 1
	case p8.Btnp(p8.UP):
		dRow = -1
	case p8.Btnp(p8.DOWN):
		dRow = 1
	default:
		return
	}

	candidate := spriteNavigationCandidate(g.currentSprite, dRow, dCol)
	if candidate == g.currentSprite {
		return
	}

	g.currentSprite = candidate
	g.lastWheelTime = now
	g.updateDrawingCanvas()
}
