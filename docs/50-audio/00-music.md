# Audio

PIGO8 plays WAV files for sound effects and music.

## Setting Up Audio

Name your audio files `music1.wav`, `music2.wav`, etc., and place them in your project directory.

## Music Function

```go
p8.Music(n)           // Play audio file n
p8.Music(n, true)     // Play exclusively (stops other audio)
p8.Music(-1)          // Stop all audio
```

### Examples

```go
// Play sound effect 1
p8.Music(1)

// Play background music, stop other sounds
p8.Music(0, true)

// Play jump sound
func (g *game) jump() {
    g.velocityY = -5
    p8.Music(2)  // Jump sound effect
}
```

## StopMusic Function

```go
p8.StopMusic(-1)  // Stop all audio
p8.StopMusic(1)   // Stop specific audio
```

## Audio Tips

### Short Sound Effects

Keep sound effects short (< 1 second) for responsive gameplay.

### Background Music

For looping music, ensure your WAV file loops cleanly.

### Web Browsers

Due to browser autoplay policies, audio starts only after user interaction. The web export handles this automatically.

## Example: Game with Sound

```go
func (g *game) Update() {
    // Player movement with footstep sound
    if p8.Btn(p8.LEFT) || p8.Btn(p8.RIGHT) {
        if g.frame % 20 == 0 {  // Every 20 frames
            p8.Music(4)  // Footstep sound
        }
    }
    
    // Collision sound
    if g.hitEnemy() {
        p8.Music(3)  // Hit sound
    }
}

func (g *game) Init() {
    p8.Music(0, true)  // Start background music
}
```

## Embedding Audio

For standalone builds, embed audio files:

```go
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
```

The embedgen tool automatically detects and embeds `music*.wav` files.

