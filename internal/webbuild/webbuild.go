// Package webbuild provides shared logic for compiling PIGO8 games to
// WebAssembly and generating a browser-playable page, used by both
// cmd/webexport and cmd/docshots.
package webbuild

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// pigo8ModulePath is the module path of this repository, used as the target
// of the replace directive written into example go.mod files by
// EnsureExampleModule.
const pigo8ModulePath = "github.com/drpaneas/pigo8"

// ensureCompleteMarker is written into gameDir once EnsureExampleModule has
// fully completed (init + replace directive + tidy) for that directory. See
// EnsureExampleModule's doc comment for why this is used instead of
// inferring completion from go.mod or go.sum existing.
const ensureCompleteMarker = ".pigo8-example-module-ready"

// BuildWASM compiles the Go program in gameDir to a WebAssembly binary at outputPath.
func BuildWASM(gameDir, outputPath string) error {
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", outputPath, ".")
	cmd.Dir = gameDir
	cmd.Env = buildWASMEnv(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// buildWASMEnv returns a copy of env with GOOS/GOARCH forced to js/wasm,
// removing any pre-existing GOOS/GOARCH entries to avoid duplicates.
func buildWASMEnv(env []string) []string {
	filtered := make([]string, 0, len(env)+2)
	for _, e := range env {
		if !strings.HasPrefix(e, "GOOS=") && !strings.HasPrefix(e, "GOARCH=") {
			filtered = append(filtered, e)
		}
	}
	return append(filtered, "GOOS=js", "GOARCH=wasm")
}

// CopyWASMExec copies wasm_exec.js from the active Go installation into outputDir.
func CopyWASMExec(outputDir string) error {
	goroot, err := getGoRoot()
	if err != nil {
		return fmt.Errorf("failed to determine GOROOT: %w", err)
	}

	possiblePaths := []string{
		filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"),  // Go 1.24+
		filepath.Join(goroot, "misc", "wasm", "wasm_exec.js"), // Go 1.23 and earlier
	}

	var wasmExecSrc string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			wasmExecSrc = path
			break
		}
	}
	if wasmExecSrc == "" {
		return fmt.Errorf("wasm_exec.js not found in GOROOT (%s). Tried: %v", goroot, possiblePaths)
	}

	return copyFile(wasmExecSrc, filepath.Join(outputDir, "wasm_exec.js"))
}

// getGoRoot returns the Go root directory using 'go env GOROOT'.
// This is more portable than runtime.GOROOT() which was deprecated in Go 1.24.
func getGoRoot() (string, error) {
	cmd := exec.Command("go", "env", "GOROOT")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// EnsureExampleModule makes sure gameDir has its own go.mod pointing at the
// local pigo8 checkout via a replace directive, matching how CI builds each
// example. Most examples' go.mod/go.sum are gitignored and generated fresh
// by this function; a few (the camera_example* directories) commit a real
// go.mod instead, which this function must leave alone rather than fail on.
//
// gameDir and repoRoot must both be absolute paths, since filepath.Rel
// requires consistently based paths to compute a correct relative path.
//
// Completion is signaled by an explicit marker file
// (gameDir/.pigo8-example-module-ready), not by go.mod or go.sum existing:
//   - go.mod alone isn't a reliable signal because it's created by the very
//     first step (go mod init); treating its mere existence as "done" would
//     let a partial failure (e.g. go mod tidy failing afterward) look like
//     success on a retry.
//   - go.sum isn't reliable either: go mod tidy only writes one when the
//     module has at least one non-replaced external dependency, so a target
//     with none (everything resolved through local replace directives)
//     would never produce a go.sum, and this function would never converge.
//
// If go.mod already exists (either a real committed one, or left over from
// an earlier partial failure) but the marker is missing, this function
// skips `go mod init` (which would fail on an existing go.mod), but still
// checks whether the replace directive is present and appends it if not -
// covering a partial failure that happened between `go mod init` and
// writing the replace directive - before (re-)running `go mod tidy`, which
// is always safe to repeat. This lets a genuinely partial prior run
// self-heal on retry instead of just failing loudly, while leaving
// already-correct committed go.mod files untouched beyond confirming
// they're tidy.
func EnsureExampleModule(gameDir, repoRoot, modulePath string) error {
	markerPath := filepath.Join(gameDir, ensureCompleteMarker)
	if _, err := os.Stat(markerPath); err == nil {
		return nil // already fully set up
	}

	goModPath := filepath.Join(gameDir, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		initCmd := exec.Command("go", "mod", "init", modulePath)
		initCmd.Dir = gameDir
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
		if err := initCmd.Run(); err != nil {
			return fmt.Errorf("go mod init in %s: %w", gameDir, err)
		}
	}

	if err := ensureReplaceDirective(gameDir, repoRoot, goModPath); err != nil {
		return err
	}

	if err := runGoModTidy(gameDir); err != nil {
		return err
	}

	if err := os.WriteFile(markerPath, []byte("marks that EnsureExampleModule completed successfully for this directory; safe to delete, it will be regenerated\n"), 0o644); err != nil {
		return fmt.Errorf("writing completion marker %s: %w", markerPath, err)
	}
	return nil
}

// ensureReplaceDirective appends a `replace` directive pointing
// pigo8ModulePath at the local repoRoot checkout to the go.mod at goModPath,
// unless one is already present.
func ensureReplaceDirective(gameDir, repoRoot, goModPath string) error {
	existing, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", goModPath, err)
	}
	if strings.Contains(string(existing), "replace "+pigo8ModulePath) {
		return nil // already present (real committed go.mod, or a prior successful run)
	}

	relPath, err := filepath.Rel(gameDir, repoRoot)
	if err != nil {
		return fmt.Errorf("computing relative path from %s to %s: %w", gameDir, repoRoot, err)
	}

	f, err := os.OpenFile(goModPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening %s to append replace directive: %w", goModPath, err)
	}
	if _, err := fmt.Fprintf(f, "\nreplace %s => %s\n", pigo8ModulePath, filepath.ToSlash(relPath)); err != nil {
		_ = f.Close()
		return fmt.Errorf("writing replace directive to %s: %w", goModPath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", goModPath, err)
	}
	return nil
}

// runGoModTidy runs `go mod tidy` in gameDir. Safe to call repeatedly.
func runGoModTidy(gameDir string) error {
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = gameDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy in %s: %w", gameDir, err)
	}
	return nil
}
