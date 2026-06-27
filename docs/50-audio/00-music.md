# Audio

PIGO8 plays embedded WAV files named `music0.wav`, `music1.wav`, and so on.

## Setting Up Audio

Add this line near the top of your `main.go`:

```go
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
```

Then run:

```bash
go generate
```

The generated `embed.go` registers your `music*.wav` files with PIGO8. Audio is loaded from registered embedded resources, so make sure you run `go generate` after adding or renaming audio files.

## Playback APIs

```go
p8.Music(n)              // Play audio file n
p8.Music(n, true)        // Play n exclusively (stops/rewinds other audio)
p8.MusicLoop(n)          // Play n in a loop
p8.MusicLoop(n, true)    // Loop n exclusively
p8.StopMusic(n)          // Stop a specific audio ID
p8.StopMusic(-1)         // Stop all audio
```

For advanced control, `p8.MusicWithOptions(...)` is also available.

## Important Behavior

- `p8.Music(n, true)` is exclusive playback. It does not enable looping.
- Use `p8.MusicLoop(...)` or `p8.MusicWithOptions(..., p8.MusicOptions{Loop: true})` for looping music.
- Calling `p8.Music(n)` again while the same ID is already playing does not restart it. Stop it first or use exclusive playback if you want to restart from the beginning.
- The float32 APIs `p8.MusicF32`, `p8.MusicLoopF32`, and `p8.StopMusicF32` exist for advanced use. The default `p8.Music` APIs are the simpler choice for most games.
- The legacy and float32 APIs keep separate playback state. Use one family consistently for a given track session instead of assuming `StopMusic(...)` will stop audio started with `MusicF32(...)`.

## Examples

```go
// Play a one-shot sound effect.
p8.Music(1)

// Start looping background music and stop any other tracks.
p8.MusicLoop(0, true)

// Stop that looping track later.
p8.StopMusic(0)
```

```go
func (g *game) Update() {
    if p8.Btnp(p8.O) {
        p8.Music(2) // Fire sound
    }

    if g.hitEnemy() {
        p8.Music(3) // Hit sound
    }
}

func (g *game) Init() {
    p8.MusicLoop(0, true)
}
```

## Audio Tips

### Short Sound Effects

Keep sound effects short for responsive gameplay.

### Background Music

For looping tracks, edit the WAV so the start and end join cleanly.

### Web Browsers

Browser autoplay policies require user interaction before audio starts. The web export handles this after the first button press.

