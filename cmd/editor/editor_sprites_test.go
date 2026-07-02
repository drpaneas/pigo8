package main

import (
	"io"
	"log"
	"os"
	"testing"
)

// TestSpritesheetFileReadableMissingFile ensures that when spritesheet.json
// does not exist, the check reports false without erroring - this is the
// path that used to rely on os.Stat + os.IsNotExist, which is unreliable on
// GOOS=js (WASM) because os.Stat there returns a generic "not implemented"
// error instead of a proper ErrNotExist. Using os.ReadFile instead means
// "missing" and "unreadable" are both treated the same way: fall back to
// creating a default spritesheet, regardless of platform.
func TestSpritesheetFileReadableMissingFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if spritesheetFileReadable() {
		t.Fatal("expected spritesheetFileReadable to return false when spritesheet.json does not exist")
	}
}

// TestSpritesheetFileReadableExistingFile ensures that an existing, readable
// spritesheet.json is correctly detected so initPico8Spritesheet does not
// overwrite it.
func TestSpritesheetFileReadableExistingFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.WriteFile("spritesheet.json", []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to write spritesheet.json: %v", err)
	}

	if !spritesheetFileReadable() {
		t.Fatal("expected spritesheetFileReadable to return true when spritesheet.json exists and is readable")
	}
}

// TestInitPico8SpritesheetCreatesDefaultWhenMissing ensures initPico8Spritesheet
// does not return an error just because spritesheet.json is missing - this
// used to fail on GOOS=js due to the os.Stat/os.IsNotExist mismatch, causing
// the editor to exit before starting its game loop in a browser.
func TestInitPico8SpritesheetCreatesDefaultWhenMissing(t *testing.T) {
	t.Chdir(t.TempDir())

	// initPico8Spritesheet populates every sprite cell via p8.Sset, which
	// logs a warning for each cell beyond what the (tiny, test-only)
	// in-memory spritesheet currently holds. Silence that expected noise.
	defer log.SetOutput(log.Writer())
	log.SetOutput(io.Discard)

	if err := initPico8Spritesheet(); err != nil {
		t.Fatalf("expected no error when spritesheet.json is missing, got: %v", err)
	}

	if !spritesheetFileReadable() {
		t.Fatal("expected initPico8Spritesheet to create a readable spritesheet.json")
	}
}
