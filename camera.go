package pigo8

// Camera state
var (
	// cameraX is the current camera X offset
	cameraX float64
	// cameraY is the current camera Y offset
	cameraY float64
)

// Camera sets the camera position, offsetting all subsequent drawing operations.
// If called with no arguments, resets the camera to (0,0).
// This function mimics PICO-8's camera(x, y) function.
//
// The camera offsets all drawing operations (Shapes, Print, Sprites, and Maps).
//
// Args:
//   - x: (optional) horizontal offset amount
//   - y: (optional) vertical offset amount
//
// Example:
//
//	// Set camera to position (64, 32)
//	Camera(64, 32)
//
//	// Reset camera to (0, 0)
//	Camera()
//
//	// Lock UI elements in place
//	Camera() // Reset camera
//	Print("SCORE: 1000", 2, 2) // Draw UI
//	Camera(playerX-64, playerY-64) // Set camera to follow player
//	Map() // Draw scrolling map
func Camera(args ...any) {
	// Reset camera if no arguments
	if len(args) == 0 {
		cameraX = 0
		cameraY = 0
		return
	}

	// Handle one argument (x only)
	if len(args) == 1 {
		if x, ok := convertToFloat64(args[0]); ok {
			cameraX = x
			cameraY = 0 // Set cameraY to 0 as per PICO-8 behavior
		}
		return
	}

	// Handle two arguments (x and y)
	if len(args) >= 2 {
		if x, ok := convertToFloat64(args[0]); ok {
			cameraX = x
		}
		if y, ok := convertToFloat64(args[1]); ok {
			cameraY = y
		}
	}
}

// CameraLerp smoothly moves the camera toward (targetX, targetY) instead of
// snapping to it instantly. Call it every frame from Update() with the same
// target you'd otherwise pass to Camera() - the camera then eases toward
// that point rather than jumping there in one frame, which is what most
// games mean by "smooth camera movement" (Camera(x, y) itself has no
// performance cost to optimize; it's a single offset applied at draw time).
//
// factor controls how quickly the camera catches up to the target each call,
// in the range (0, 1]: smaller values (e.g. 0.05) trail further behind for a
// softer, floatier follow; larger values (e.g. 0.5) catch up almost
// immediately. A factor of 1 is equivalent to calling Camera(targetX,
// targetY) directly. Values outside (0, 1] are clamped.
//
// Example:
//
//	// In Update(), follow the player with a soft lag instead of a hard lock:
//	p8.CameraLerp(playerX-64, playerY-64, 0.1)
func CameraLerp[X Number, Y Number](targetX X, targetY Y, factor float64) {
	switch {
	case factor <= 0:
		return // no movement
	case factor > 1:
		factor = 1
	}

	tx, ty := float64(targetX), float64(targetY)
	cameraX += (tx - cameraX) * factor
	cameraY += (ty - cameraY) * factor
}

// convertToFloat64 attempts to convert a value to float64
func convertToFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// applyCameraOffset applies the current camera offset to the given coordinates
// and returns the transformed coordinates
func applyCameraOffset(x, y float64) (float64, float64) {
	return x - cameraX, y - cameraY
}
