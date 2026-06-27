package pigo8

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"testing/fstest"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

func resetSharedAudioDataForTest() {
	sharedAudioStore.mu.Lock()
	sharedAudioStore.files = make(map[int][]byte)
	sharedAudioStore.loaded = false
	sharedAudioStore.mu.Unlock()
	setCustomResources(nil)
	if engine := musicEngines.get(audioFormatPCM); engine != nil {
		engine.stop(-1)
	}
	if engine := musicEngines.get(audioFormatF32); engine != nil {
		engine.stop(-1)
	}
}

func createTestWAV(t *testing.T, sampleRate int, samples []int16) []byte {
	t.Helper()

	var buf bytes.Buffer
	const (
		numChannels   = uint16(1)
		bitsPerSample = uint16(16)
		audioFormat   = uint16(1)
	)

	blockAlign := numChannels * bitsPerSample / 8
	byteRate := uint32(sampleRate) * uint32(blockAlign)
	dataSize := uint32(len(samples)) * uint32(blockAlign)
	riffSize := uint32(36) + dataSize

	writeString := func(s string) {
		t.Helper()
		if _, err := buf.WriteString(s); err != nil {
			t.Fatalf("write string %q: %v", s, err)
		}
	}
	writeValue := func(v any) {
		t.Helper()
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			t.Fatalf("write value %T: %v", v, err)
		}
	}

	writeString("RIFF")
	writeValue(riffSize)
	writeString("WAVE")
	writeString("fmt ")
	writeValue(uint32(16))
	writeValue(audioFormat)
	writeValue(numChannels)
	writeValue(uint32(sampleRate))
	writeValue(byteRate)
	writeValue(blockAlign)
	writeValue(bitsPerSample)
	writeString("data")
	writeValue(dataSize)
	for _, sample := range samples {
		writeValue(sample)
	}

	return buf.Bytes()
}

// TestMusicFunctions smoke-tests the public music APIs.
func TestMusicFunctions(t *testing.T) {
	// These tests just verify that the functions don't panic
	// They don't test actual audio playback since that would require
	// a more complex setup with audio hardware

	t.Run("Music function", func(_ *testing.T) {
		Music(1)
		Music(2, true)
		MusicLoop(3)
		Music(-1)
		MusicF32(1)
		MusicF32(2, true)
		MusicLoopF32(3)
		MusicF32(-1)
	})

	t.Run("StopMusic function", func(_ *testing.T) {
		StopMusic(1)
		StopMusic(-1)
		StopMusicF32(1)
		StopMusicF32(-1)
	})

	t.Run("F32 availability", func(t *testing.T) {
		if !IsAudioF32Available() {
			t.Fatal("expected F32 audio support to be available")
		}
	})
}

func TestSharedAudioDataHonorsExplicitAudioPaths(t *testing.T) {
	resetSharedAudioDataForTest()
	t.Cleanup(resetSharedAudioDataForTest)

	resources := &embeddedResources{
		FS: fstest.MapFS{
			"music0.wav": &fstest.MapFile{Data: createTestWAV(t, 22050, []int16{0, 100, 200, 300})},
			"music1.wav": &fstest.MapFile{Data: createTestWAV(t, 22050, []int16{300, 200, 100, 0})},
		},
		AudioPaths: []string{"music1.wav"},
	}
	setCustomResources(resources)

	reloadSharedMusicData()

	if got := sharedAudioFileCount(); got != 1 {
		t.Fatalf("sharedAudioFileCount() = %d, want 1", got)
	}
	if _, ok := getSharedMusicData(0); ok {
		t.Fatal("music0.wav should not have been loaded when it was not explicitly registered")
	}
	if _, ok := getSharedMusicData(1); !ok {
		t.Fatal("music1.wav should have been loaded")
	}
}

func TestSharedAudioDataReloadsAfterLateRegistration(t *testing.T) {
	resetSharedAudioDataForTest()
	t.Cleanup(resetSharedAudioDataForTest)

	ensureSharedMusicDataLoaded()
	if got := sharedAudioFileCount(); got != 0 {
		t.Fatalf("sharedAudioFileCount() before registration = %d, want 0", got)
	}

	setCustomResources(&embeddedResources{
		FS: fstest.MapFS{
			"music7.wav": &fstest.MapFile{Data: createTestWAV(t, 22050, []int16{0, 1000, -1000, 0})},
		},
		AudioPaths: []string{"music7.wav"},
	})

	reloadSharedMusicData()

	if got := sharedAudioFileCount(); got != 1 {
		t.Fatalf("sharedAudioFileCount() after registration = %d, want 1", got)
	}
	if _, ok := getSharedMusicData(7); !ok {
		t.Fatal("music7.wav should have been available after reloading shared audio data")
	}
}

func TestPrepareF32PlaybackResamplesToSharedRate(t *testing.T) {
	audioData := createTestWAV(t, 22050, []int16{0, 1000, -1000, 0})

	decoded, err := wav.DecodeF32(bytes.NewReader(audioData))
	if err != nil {
		t.Fatalf("DecodeF32() error = %v", err)
	}

	prepared, err := prepareF32Playback(audioData)
	if err != nil {
		t.Fatalf("prepareF32Playback() error = %v", err)
	}

	var expectedStream io.ReadSeeker = decoded
	if decoded.SampleRate() != sampleRate {
		expectedStream = audio.ResampleF32(decoded, decoded.Length(), decoded.SampleRate(), sampleRate)
	}

	wantLength, err := playbackStreamLength(expectedStream)
	if err != nil {
		t.Fatalf("playbackStreamLength(expectedStream) error = %v", err)
	}
	if prepared.length != wantLength {
		t.Fatalf("prepareF32Playback() length = %d, want %d", prepared.length, wantLength)
	}

	gotLength, err := playbackStreamLength(prepared.stream)
	if err != nil {
		t.Fatalf("playbackStreamLength(prepared.stream) error = %v", err)
	}
	if gotLength != prepared.length {
		t.Fatalf("playbackStreamLength(prepared.stream) = %d, want %d", gotLength, prepared.length)
	}
}
