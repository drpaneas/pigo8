package webbuild

import (
	"fmt"
	"os"
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
// EnsureExampleModule's hardcoded replace directive targets it). It
// deliberately has zero external dependencies - anything beyond the
// standard library risks pulling in a transitive dependency that isn't
// already present in a given environment's module cache (a real,
// previously-hit failure: a third-party test dependency's own transitive
// dependency wasn't cached on a fresh CI runner, breaking `go mod tidy`
// even with GOPROXY=off). Zero dependencies keeps this test hermetic
// everywhere, with no network or cache-state assumptions at all.
func setupFakeTargetModule(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	goMod := fmt.Sprintf("module %s\n\ngo 1.25.0\n", pigo8ModulePath)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("writing fake target go.mod: %v", err)
	}

	src := `package pigo8fake

// Greeting is a trivial exported symbol used only to exercise
// EnsureExampleModule's replace-directive wiring; it is not part of the
// real pigo8 API.
var Greeting = "hello"
`
	if err := os.WriteFile(filepath.Join(dir, "lib.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("writing fake target lib.go: %v", err)
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

	if _, err := os.Stat(filepath.Join(gameDir, ensureCompleteMarker)); err != nil {
		t.Errorf("expected completion marker to exist after EnsureExampleModule: %v", err)
	}
	// The fake target module has zero external dependencies (see
	// setupFakeTargetModule), so go.sum is correctly NOT created - this is
	// exactly the case that motivated using an explicit marker file instead
	// of go.sum as the completion signal.
	if _, err := os.Stat(filepath.Join(gameDir, "go.sum")); err == nil {
		t.Error("expected no go.sum for a target module with zero external dependencies")
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

// TestEnsureExampleModule_PreExistingGoMod covers examples that commit a
// real go.mod (like camera_example*) rather than having one generated: no
// marker file will exist yet, but EnsureExampleModule must not try to run
// `go mod init` again (which would fail since go.mod already exists) - it
// should just tidy the existing module and write the marker.
func TestEnsureExampleModule_PreExistingGoMod(t *testing.T) {
	t.Setenv("GOPROXY", "off")

	repoRoot := setupFakeTargetModule(t)
	gameDir := setupFakeGameDir(t)

	relPath, err := filepath.Rel(gameDir, repoRoot)
	if err != nil {
		t.Fatalf("filepath.Rel: %v", err)
	}
	preExisting := fmt.Sprintf(
		"module example.com/mygame\n\ngo 1.25.0\n\nreplace %s => %s\n\nrequire %s v0.0.0-00010101000000-000000000000\n",
		pigo8ModulePath, filepath.ToSlash(relPath), pigo8ModulePath,
	)
	if err := os.WriteFile(filepath.Join(gameDir, "go.mod"), []byte(preExisting), 0o644); err != nil {
		t.Fatalf("writing pre-existing go.mod: %v", err)
	}

	if err := EnsureExampleModule(gameDir, repoRoot, "example.com/mygame"); err != nil {
		t.Fatalf("EnsureExampleModule with pre-existing go.mod: %v", err)
	}

	if _, err := os.Stat(filepath.Join(gameDir, ensureCompleteMarker)); err != nil {
		t.Errorf("expected completion marker to exist after EnsureExampleModule: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(gameDir, "go.mod"))
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	if !strings.Contains(string(got), "replace "+pigo8ModulePath) {
		t.Errorf("pre-existing replace directive was lost, got:\n%s", got)
	}
}

// TestEnsureExampleModule_SelfHealsPartialFailure simulates the exact bug
// this design fixes: a previous run got as far as `go mod init` (so go.mod
// exists) but failed before the marker was written. A retry must not
// silently report success without actually finishing setup - it must
// complete the remaining steps and write the marker.
func TestEnsureExampleModule_SelfHealsPartialFailure(t *testing.T) {
	t.Setenv("GOPROXY", "off")

	repoRoot := setupFakeTargetModule(t)
	gameDir := setupFakeGameDir(t)

	// Simulate a partial failure: go.mod exists (as if `go mod init` had
	// already run) but the replace directive and marker are both missing.
	partial := "module example.com/mygame\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(gameDir, "go.mod"), []byte(partial), 0o644); err != nil {
		t.Fatalf("writing partial go.mod: %v", err)
	}

	if _, err := os.Stat(filepath.Join(gameDir, ensureCompleteMarker)); err == nil {
		t.Fatal("test setup bug: marker should not exist yet")
	}

	if err := EnsureExampleModule(gameDir, repoRoot, "example.com/mygame"); err != nil {
		t.Fatalf("EnsureExampleModule did not self-heal from partial failure: %v", err)
	}

	if _, err := os.Stat(filepath.Join(gameDir, ensureCompleteMarker)); err != nil {
		t.Errorf("expected completion marker after self-healing: %v", err)
	}
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
