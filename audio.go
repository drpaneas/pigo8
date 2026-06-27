package pigo8

// MusicOptions contains optional parameters for Music playback
type MusicOptions struct {
	Exclusive bool // If true, stops all other audio before playing
	Loop      bool // If true, the audio will loop indefinitely
}

// Music plays the audio file with the given ID.
// If n is -1, it stops all currently playing audio.
// If n is a valid audio ID, it plays that audio file.
// If exclusive is true, it stops all other audio files before playing.
func Music(n int, exclusive ...bool) {
	shouldBeExclusive := len(exclusive) > 0 && exclusive[0]
	MusicWithOptions(n, MusicOptions{Exclusive: shouldBeExclusive})
}

// MusicWithOptions plays the audio file with the given ID using typed options.
// If n is -1, it stops all currently playing audio.
func MusicWithOptions(n int, opts MusicOptions) {
	if n == -1 {
		StopMusic(-1)
		return
	}
	getAudioPlayer().play(n, opts)
}

// MusicLoop is a convenience function that plays the audio file with the given ID in a loop.
// This is equivalent to MusicWithOptions(n, MusicOptions{Loop: true})
func MusicLoop(n int, exclusive ...bool) {
	shouldBeExclusive := len(exclusive) > 0 && exclusive[0]
	MusicWithOptions(n, MusicOptions{Exclusive: shouldBeExclusive, Loop: true})
}

// StopMusic stops the audio file with the given ID
// If id is -1, it stops all audio files
func StopMusic(id int) {
	getAudioPlayer().stop(id)
}
