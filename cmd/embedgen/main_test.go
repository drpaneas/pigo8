package main

import (
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
			content:  createValidWAVHeader(1000),
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
			content:  createValidWAVHeader(0),
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

// createValidWAVHeader creates a minimal valid WAV file header with the specified data size
func createValidWAVHeader(dataSize int) []byte {
	header := make([]byte, 44)

	// RIFF marker
	copy(header[0:4], "RIFF")

	// File size - 8 (little-endian)
	fileSize := 36 + dataSize
	header[4] = byte(fileSize)
	header[5] = byte(fileSize >> 8)
	header[6] = byte(fileSize >> 16)
	header[7] = byte(fileSize >> 24)

	// WAVE marker
	copy(header[8:12], "WAVE")

	// fmt subchunk marker
	copy(header[12:16], "fmt ")

	// fmt subchunk size (16 for PCM)
	header[16] = 16

	// Audio format (1 = PCM)
	header[20] = 1

	// Number of channels (1 = mono)
	header[22] = 1

	// Sample rate (44100)
	header[24] = 0x44
	header[25] = 0xAC

	// Byte rate (44100 * 1 * 16/8 = 88200)
	header[28] = 0x88
	header[29] = 0x58
	header[30] = 0x01

	// Block align (1 * 16/8 = 2)
	header[32] = 2

	// Bits per sample (16)
	header[34] = 16

	// data subchunk marker
	copy(header[36:40], "data")

	// Data size (little-endian)
	header[40] = byte(dataSize)
	header[41] = byte(dataSize >> 8)
	header[42] = byte(dataSize >> 16)
	header[43] = byte(dataSize >> 24)

	return header
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
