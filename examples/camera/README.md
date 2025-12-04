# Camera Examples

A series of examples demonstrating camera functionality built with [PIGO8](https://github.com/drpaneas/pigo8), matching the [PICO-8 Camera Guide](https://nerdyteachers.com/PICO-8/Guide/CAMERA) 1:1.

## 🎮 Play Online

- **[▶️ Example 1: Basic Camera](https://drpaneas.github.io/pigo8/camera_example1/)** - No offset
- **[▶️ Example 2: Retroactive Effect](https://drpaneas.github.io/pigo8/camera_example2/)** - Camera affects previous elements
- **[▶️ Example 3: Locked UI](https://drpaneas.github.io/pigo8/camera_example3/)** - Two cameras for UI overlay

## 📖 Description

These examples demonstrate camera functionality in PIGO8:
- Using `p8.Camera(x, y)` to offset all drawing operations
- Understanding retroactive camera effects
- Creating fixed UI overlays with camera locking
- Following players in scrolling games

### Example 1: Basic Camera (No Offset)

Demonstrates basic drawing with default camera position (0,0):

```go
func (g *Game) Draw() {
    p8.Cls()
    p8.Rectfill(0, 0, 127, 127, 2) // Background
    p8.Rect(0, 0, 127, 127, 8)     // Outline
    p8.Print("camera(0,0)", 2, 2)  // Label
}
```

### Example 2: Retroactive Camera Effect

Shows how camera offset affects **previously drawn elements**:

```go
func (g *Game) Draw() {
    p8.Cls()
    p8.Rectfill(0, 0, 127, 127, 2) // Draw first
    p8.Print("camera(0,0)", 2, 2)
    
    p8.Camera(63, 63)              // NOW offset - affects above!
    p8.Rect(63, 63, 190, 190, 11)  // New elements
}
```

### Example 3: Locked UI Elements

Using two camera calls to create fixed UI overlays:

```go
func (g *Game) Draw() {
    p8.Cls()
    p8.Camera()                    // LOCKS following elements
    p8.Rectfill(0, 0, 127, 127, 2) // Locked in place
    p8.Print("SCORE: 100", 2, 2)   // UI doesn't move
    
    p8.Camera(63, 63)              // Only affects new elements
    p8.Rect(63, 63, 190, 190, 11)  // These are offset
}
```

## ⚙️ Requirements

- Go 1.24+
- PIGO8 library
  ```sh
  go get github.com/drpaneas/pigo8
  ```

## 🚀 Run Locally

```bash
# Example 1
cd examples/camera/camera_example1
go run .

# Example 2
cd examples/camera/camera_example2
go run .

# Example 3
cd examples/camera/camera_example3
go run .
```

## Camera Behavior Summary

### Core Functions
- `Camera()` - Resets camera to (0,0)
- `Camera(x, y)` - Sets camera offset

### Key Behaviors
1. **Retroactive Effect:** Camera affects previously drawn elements unless locked
2. **Locking:** Calling `Camera()` locks previous elements in place
3. **UI Pattern:** Use two camera calls for fixed UI overlays

### Common Use Case

```go
// Game world follows player
Camera(playerX - 64, playerY - 64)
Map() // Scrolling world

// Fixed UI overlay
Camera() // Reset and lock
Print("SCORE: " + score, 2, 2)
Print("HEALTH: " + health, 2, 10)
```
