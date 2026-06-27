package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")

	// Test non-existent file
	if fileExists(tmpFile) {
		t.Error("fileExists returned true for non-existent file")
	}

	// Create the file
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !fileExists(tmpFile) {
		t.Error("fileExists returned false for existing file")
	}

	// Test directory
	if !fileExists(tmpDir) {
		t.Error("fileExists returned false for existing directory")
	}
}

func TestIsValidAudioFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "valid WAV file",
			content:  createPCM16WAV(22050, []int16{0, 1000, -1000, 0}),
			expected: true,
		},
		{
			name:     "valid WAV file with extra chunk",
			content:  createPCM16WAVWithExtraChunk(22050, "LIST", []byte{1, 2, 3, 4}, []int16{0, 1000, -1000, 0}),
			expected: true,
		},
		{
			name:     "invalid - missing RIFF marker",
			content:  []byte("XXXX" + string(make([]byte, 4)) + "WAVE" + string(make([]byte, 32))),
			expected: false,
		},
		{
			name:     "invalid - missing WAVE marker",
			content:  []byte("RIFF" + string(make([]byte, 4)) + "XXXX" + string(make([]byte, 32))),
			expected: false,
		},
		{
			name:     "invalid - zero data size",
			content:  createPCM16WAV(22050, nil),
			expected: false,
		},
		{
			name:     "invalid - too short",
			content:  []byte("RIFF"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(tmpDir, tt.name+".wav")
			if err := os.WriteFile(tmpFile, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := isValidAudioFile(tmpFile)
			if result != tt.expected {
				t.Errorf("isValidAudioFile() = %v, expected %v", result, tt.expected)
			}
		})
	}

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		if isValidAudioFile(filepath.Join(tmpDir, "nonexistent.wav")) {
			t.Error("isValidAudioFile returned true for non-existent file")
		}
	})
}

func createPCM16WAV(sampleRate int, samples []int16) []byte {
	return createPCM16WAVWithExtraChunk(sampleRate, "", nil, samples)
}

func createPCM16WAVWithExtraChunk(sampleRate int, chunkID string, chunkData []byte, samples []int16) []byte {
	var buf bytes.Buffer
	const (
		numChannels   = uint16(1)
		bitsPerSample = uint16(16)
		audioFormat   = uint16(1)
	)

	blockAlign := numChannels * bitsPerSample / 8
	byteRate := uint32(sampleRate) * uint32(blockAlign)
	dataSize := uint32(len(samples)) * uint32(blockAlign)
	extraChunkSize := uint32(0)
	if chunkID != "" {
		extraChunkSize = 8 + uint32(len(chunkData))
	}
	riffSize := uint32(36) + dataSize + extraChunkSize

	writeString := func(s string) {
		_, _ = buf.WriteString(s)
	}
	writeValue := func(v any) {
		_ = binary.Write(&buf, binary.LittleEndian, v)
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
	if chunkID != "" {
		writeString(chunkID)
		writeValue(uint32(len(chunkData)))
		_, _ = buf.Write(chunkData)
	}
	writeString("data")
	writeValue(dataSize)
	for _, sample := range samples {
		writeValue(sample)
	}

	return buf.Bytes()
}

func TestVerboseFlag(t *testing.T) {
	// Test that verbose variable exists and can be set
	originalVerbose := verbose
	defer func() { verbose = originalVerbose }()

	verbose = true
	if !verbose {
		t.Error("verbose flag not set to true")
	}

	verbose = false
	if verbose {
		t.Error("verbose flag not set to false")
	}
}
