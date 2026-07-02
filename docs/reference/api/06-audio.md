# Audio

## `Music(n int, exclusive ...bool)`

Plays the audio file with ID `n`. If `n` is -1, stops all currently playing audio. If
`exclusive` is `true`, stops all other audio first.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `n` | `int` | Audio ID to play, or -1 to stop all currently playing audio. |
| `exclusive` | `bool` (optional, variadic) | If true, stops all other audio before playing. Defaults to false. |

**Example:**
```go
p8.Music(0)        // play audio 0 (mixes with anything already playing)
p8.Music(0, true)  // play audio 0, stopping everything else first
```

## `MusicLoop(n int, exclusive ...bool)`

Plays audio `n` in a loop. Equivalent to `MusicWithOptions(n, MusicOptions{Loop: true})`.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `n` | `int` | Audio ID to play in a loop. |
| `exclusive` | `bool` (optional, variadic) | If true, stops all other audio before playing. Defaults to false. |

## `StopMusic(id int)`

Stops the audio file with the given ID. If `id` is -1, stops all audio.

**Parameters:**
| Name | Type | Description |
|------|------|--------------|
| `id` | `int` | Audio ID to stop, or -1 to stop all audio. |

## `MusicWithOptions(n int, opts MusicOptions)` / `MusicOptions`

Plays audio `n` using explicit options instead of variadic booleans.

**`MusicOptions` fields:**
| Field | Type | Description |
|-------|------|--------------|
| `Exclusive` | `bool` | If true, stops all other audio before playing. |
| `Loop` | `bool` | If true, loops indefinitely. |

**Example:**
```go
p8.MusicWithOptions(0, p8.MusicOptions{Loop: true, Exclusive: true})
```

## 32-bit float variants: `MusicF32`, `MusicLoopF32`, `StopMusicF32`, `MusicF32WithOptions`

Identical behavior to their non-`F32` counterparts, but use 32-bit float audio, the recommended
approach for Ebitengine v2.8+ for better performance and easier audio processing.

**Example:**
```go
p8.MusicF32(0)
p8.MusicLoopF32(1)
p8.StopMusicF32(-1) // stop all
```

**See also:** [Music and Sound](../../50-audio/00-music.md)
