package webbuild

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// setupFakeTargetModule builds a tiny, self-contained module in a temp
// directory that declares itself under pigo8ModulePath (so
// EnsureExampleModule's hardcoded replace directive targets it) and depends
// on a small third-party package that is already present in the local
// module cache. This lets go mod tidy produce a real go.sum entry without
// ever touching the network, keeping the test hermetic and offline-safe.
func setupFakeTargetModule(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	goMod := fmt.Sprintf("module %s\n\ngo 1.25.0\n\nrequire github.com/stretchr/testify v1.11.1\n", pigo8ModulePath)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("writing fake target go.mod: %v", err)
	}

	// Depends on a real, already-cached third-party package so "go mod tidy"
	// has something to checksum into go.sum.
	src := `package pigo8fake

import "github.com/stretchr/testify/assert"

// Greeting is a trivial exported symbol used only to exercise
// EnsureExampleModule's replace-directive wiring; it is not part of the
// real pigo8 API.
var Greeting = assert.ObjectsAreEqual("hello", "hello")
`
	if err := os.WriteFile(filepath.Join(dir, "lib.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("writing fake target lib.go: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = dir
	tidyCmd.Env = append(os.Environ(), "GOPROXY=off")
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy for fake target module: %v\n%s", err, out)
	}

	return dir
}

// setupFakeGameDir creates a temp directory with a trivial main package that
// imports the fake target module set up by setupFakeTargetModule.
func setupFakeGameDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	src := `package main

import (
	"fmt"

	pigo8fake "` + pigo8ModulePath + `"
)

func main() {
	fmt.Println(pigo8fake.Greeting)
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("writing fake game main.go: %v", err)
	}
	return dir
}

func TestEnsureExampleModule(t *testing.T) {
	// Force fully offline module resolution so this test never depends on
	// network access, regardless of the ambient environment.
	t.Setenv("GOPROXY", "off")

	repoRoot := setupFakeTargetModule(t)
	gameDir := setupFakeGameDir(t)

	if err := EnsureExampleModule(gameDir, repoRoot, "example.com/mygame"); err != nil {
		t.Fatalf("EnsureExampleModule: %v", err)
	}

	goModPath := filepath.Join(gameDir, "go.mod")
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	goMod := string(goModBytes)

	relPath, err := filepath.Rel(gameDir, repoRoot)
	if err != nil {
		t.Fatalf("filepath.Rel: %v", err)
	}
	wantReplace := fmt.Sprintf("replace %s => %s", pigo8ModulePath, filepath.ToSlash(relPath))
	if !strings.Contains(goMod, wantReplace) {
		t.Errorf("go.mod missing expected replace directive %q, got:\n%s", wantReplace, goMod)
	}

	if _, err := os.Stat(filepath.Join(gameDir, "go.sum")); err != nil {
		t.Errorf("expected go.sum to exist after EnsureExampleModule: %v", err)
	}

	t.Run("idempotent", func(t *testing.T) {
		before, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("reading go.mod before second call: %v", err)
		}

		if err := EnsureExampleModule(gameDir, repoRoot, "example.com/mygame"); err != nil {
			t.Fatalf("second EnsureExampleModule call: %v", err)
		}

		after, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("reading go.mod after second call: %v", err)
		}
		if string(before) != string(after) {
			t.Errorf("go.mod changed on second (already-set-up) call, want a no-op:\nbefore:\n%s\nafter:\n%s", before, after)
		}
	})
}

func TestGenerateHTML(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "index.html")
	const title = "My Test Game"

	if err := GenerateHTML(outputPath, title); err != nil {
		t.Fatalf("GenerateHTML: %v", err)
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading generated HTML: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("generated HTML file is empty")
	}
	if !strings.Contains(string(got), title) {
		t.Errorf("generated HTML does not contain title %q", title)
	}
}
