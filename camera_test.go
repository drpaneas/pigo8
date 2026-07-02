package pigo8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetCameraState resets the package-level camera offset to (0, 0) after a
// test, so tests don't leak state into each other.
func resetCameraState(t *testing.T) {
	t.Helper()
	originalX, originalY := cameraX, cameraY
	t.Cleanup(func() {
		cameraX, cameraY = originalX, originalY
	})
}

func TestCameraNoArgsResets(t *testing.T) {
	resetCameraState(t)
	cameraX, cameraY = 64, 32

	Camera()

	assert.Equal(t, 0.0, cameraX)
	assert.Equal(t, 0.0, cameraY)
}

func TestCameraOneArgSetsXAndResetsY(t *testing.T) {
	resetCameraState(t)
	cameraX, cameraY = 10, 20

	Camera(64)

	assert.Equal(t, 64.0, cameraX, "x should be set to the given value")
	assert.Equal(t, 0.0, cameraY, "y should reset to 0 when only x is given, matching PICO-8")
}

func TestCameraTwoArgsSetsBoth(t *testing.T) {
	resetCameraState(t)

	Camera(64, 32)

	assert.Equal(t, 64.0, cameraX)
	assert.Equal(t, 32.0, cameraY)
}

func TestCameraAcceptsMixedNumericTypes(t *testing.T) {
	resetCameraState(t)

	Camera(int32(10), float32(20.5))

	assert.Equal(t, 10.0, cameraX)
	assert.Equal(t, 20.5, cameraY)
}

func TestApplyCameraOffset(t *testing.T) {
	resetCameraState(t)

	t.Run("zero offset is a no-op", func(t *testing.T) {
		cameraX, cameraY = 0, 0
		x, y := applyCameraOffset(10, 20)
		assert.Equal(t, 10.0, x)
		assert.Equal(t, 20.0, y)
	})

	t.Run("subtracts the current camera offset", func(t *testing.T) {
		cameraX, cameraY = 5, 8
		x, y := applyCameraOffset(10, 20)
		assert.Equal(t, 5.0, x)
		assert.Equal(t, 12.0, y)
	})

	t.Run("matches Camera() end to end", func(t *testing.T) {
		Camera(64, 32)
		x, y := applyCameraOffset(64, 32)
		assert.Equal(t, 0.0, x, "a point at the camera position should map to screen origin")
		assert.Equal(t, 0.0, y)
	})
}
