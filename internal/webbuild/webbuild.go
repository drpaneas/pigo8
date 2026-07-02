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
// example (examples' go.mod/go.sum are gitignored, not committed).
//
// gameDir and repoRoot must both be absolute paths, since filepath.Rel
// requires consistently based paths to compute a correct relative path.
//
// Completion is signaled by the presence of gameDir/go.sum rather than
// go.mod: go.mod is created by the first step (go mod init), so treating it
// as "already set up" would let a partial failure (e.g. go mod tidy failing
// after go.mod exists but before go.sum is written) look like success on any
// retry. go.sum is only ever produced by a successful go mod tidy, so its
// presence means every earlier step also completed successfully.
func EnsureExampleModule(gameDir, repoRoot, modulePath string) error {
	if _, err := os.Stat(filepath.Join(gameDir, "go.sum")); err == nil {
		return nil // already fully set up
	}

	relPath, err := filepath.Rel(gameDir, repoRoot)
	if err != nil {
		return fmt.Errorf("computing relative path from %s to %s: %w", gameDir, repoRoot, err)
	}

	initCmd := exec.Command("go", "mod", "init", modulePath)
	initCmd.Dir = gameDir
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("go mod init in %s: %w", gameDir, err)
	}

	goModPath := filepath.Join(gameDir, "go.mod")
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

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = gameDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy in %s: %w", gameDir, err)
	}
	return nil
}
