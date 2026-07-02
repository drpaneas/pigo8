package webbuild

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestBuildWASMEnv(t *testing.T) {
	env := buildWASMEnv([]string{"GOOS=darwin", "GOARCH=arm64", "FOO=bar"})
	want := map[string]bool{"GOOS=js": true, "GOARCH=wasm": true, "FOO=bar": true}
	got := map[string]bool{}
	for _, e := range env {
		got[e] = true
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing expected env entry %q in %v", k, env)
		}
	}
	for _, e := range env {
		if e == "GOOS=darwin" || e == "GOARCH=arm64" {
			t.Errorf("stale GOOS/GOARCH entry %q should have been filtered", e)
		}
	}
}
