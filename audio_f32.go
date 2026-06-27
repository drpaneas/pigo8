package pigo8

// MusicF32 plays the audio file with the given ID using 32-bit float format.
// This is the new recommended approach for Ebitengine v2.8+ as it provides
// better performance and easier audio processing.
// If n is -1, it stops all currently playing audio.
// If n is a valid audio ID, it plays that audio file.
// If exclusive is true, it stops all other audio files before playing.
func MusicF32(n int, exclusive ...bool) {
	shouldBeExclusive := len(exclusive) > 0 && exclusive[0]
	MusicF32WithOptions(n, MusicOptions{Exclusive: shouldBeExclusive})
}

// MusicF32WithOptions plays the audio file with the given ID using 32-bit float format and typed options.
// If n is -1, it stops all currently playing audio.
func MusicF32WithOptions(n int, opts MusicOptions) {
	if n == -1 {
		StopMusicF32(-1)
		return
	}
	getAudioPlayerF32().play(n, opts)
}

// MusicLoopF32 is a convenience function that plays the audio file with the given ID in a loop
// using 32-bit float format.
// This is equivalent to MusicF32WithOptions(n, MusicOptions{Loop: true})
func MusicLoopF32(n int, exclusive ...bool) {
	shouldBeExclusive := len(exclusive) > 0 && exclusive[0]
	MusicF32WithOptions(n, MusicOptions{Exclusive: shouldBeExclusive, Loop: true})
}

// StopMusicF32 stops the audio file with the given ID using 32-bit float format
// If id is -1, it stops all audio files
func StopMusicF32(id int) {
	getAudioPlayerF32().stop(id)
}

// IsAudioF32Available returns true if 32-bit float audio is supported
// This can be used to check if the new audio features are available
func IsAudioF32Available() bool {
	// In Ebitengine v2.8+, 32-bit float audio is always available
	return true
}
