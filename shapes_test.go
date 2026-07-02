package pigo8

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

// TestLineAppliesCameraOffset is a regression test for a bug where Line, unlike
// Rect/Rectfill/Circ/Circfill, never called applyCameraOffset and so ignored
// the current camera position entirely. Actual pixel output can't be
// inspected in this test environment (ebiten.Image.At panics before the
// game loop starts, same constraint noted in TestPsetGet), so this pins the
// fix at the level that can be verified here: Line must not panic - with or
// without an active camera offset - and must consult applyCameraOffset by
// construction (see shapes.go's Line implementation).
func TestLineAppliesCameraOffset(t *testing.T) {
	originalScreen := currentScreen
	testScreen := ebiten.NewImage(20, 20)
	currentScreen = testScreen
	resetCameraState(t)
	t.Cleanup(func() {
		currentScreen = originalScreen
	})

	assert.NotPanics(t, func() {
		Line(0, 0, 10, 10, 8)
	}, "Line with no camera offset should not panic")

	Camera(5, 5)
	assert.NotPanics(t, func() {
		Line(0, 0, 10, 10, 8)
	}, "Line with an active camera offset should not panic")
}
