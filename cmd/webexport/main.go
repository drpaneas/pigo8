// Package main provides a tool to export PIGO8 games to WebAssembly for web browsers.
// It compiles the game to WASM and generates an HTML page with a gameboy-style UI.
//
// Usage:
//
//	go run ./cmd/webexport -game ./examples/pong -o ./dist
//	go run ./cmd/webexport -game ./examples/pong -o ./dist -title "My Pong Game"
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Parse command line flags
	var (
		gameDir   string
		outputDir string
		title     string
		serve     bool
		port      int
	)

	flag.StringVar(&gameDir, "game", ".", "Directory containing the game to export")
	flag.StringVar(&outputDir, "o", "dist", "Output directory for the web build")
	flag.StringVar(&title, "title", "", "Game title (defaults to directory name)")
	flag.BoolVar(&serve, "serve", false, "Start a local HTTP server after building")
	flag.IntVar(&port, "port", 8080, "Port for the local HTTP server (used with -serve)")
	flag.Parse()

	// Resolve paths
	gameDir, err := filepath.Abs(gameDir)
	if err != nil {
		fmt.Printf("Error resolving game directory: %v\n", err)
		os.Exit(1)
	}

	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		fmt.Printf("Error resolving output directory: %v\n", err)
		os.Exit(1)
	}

	// Validate game directory exists
	if _, err := os.Stat(gameDir); os.IsNotExist(err) {
		fmt.Printf("Error: Game directory does not exist: %s\n", gameDir)
		os.Exit(1)
	}

	// Set default title if not provided
	if title == "" {
		title = filepath.Base(gameDir)
		// Capitalize first letter
		if len(title) > 0 {
			title = strings.ToUpper(title[:1]) + title[1:]
		}
	}

	fmt.Printf("PIGO-8 Web Export\n")
	fmt.Printf("=================\n")
	fmt.Printf("Game: %s\n", gameDir)
	fmt.Printf("Output: %s\n", outputDir)
	fmt.Printf("Title: %s\n\n", title)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Step 1: Build WASM
	fmt.Println("Building WebAssembly...")
	wasmPath := filepath.Join(outputDir, "game.wasm")
	if err := buildWASM(gameDir, wasmPath); err != nil {
		fmt.Printf("Error building WASM: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Built game.wasm")

	// Step 2: Copy wasm_exec.js
	fmt.Println("Copying WASM runtime...")
	if err := copyWASMExec(outputDir); err != nil {
		fmt.Printf("Error copying wasm_exec.js: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Copied wasm_exec.js")

	// Step 3: Generate HTML
	fmt.Println("Generating HTML...")
	htmlPath := filepath.Join(outputDir, "index.html")
	if err := generateHTML(htmlPath, title); err != nil {
		fmt.Printf("Error generating HTML: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated index.html")

	// Print success message
	fmt.Printf("\n✅ Web export complete!\n\n")
	fmt.Printf("Output files:\n")
	fmt.Printf("  %s/\n", outputDir)
	fmt.Printf("    ├── index.html    (game page with gameboy UI)\n")
	fmt.Printf("    ├── game.wasm     (compiled game)\n")
	fmt.Printf("    └── wasm_exec.js  (Go WASM runtime)\n\n")

	if serve {
		fmt.Printf("Starting local server at http://localhost:%d\n", port)
		fmt.Println("Press Ctrl+C to stop")
		serveHTTP(outputDir, port)
	} else {
		fmt.Printf("To test locally:\n")
		fmt.Printf("  cd %s && python3 -m http.server %d\n", outputDir, port)
		fmt.Printf("  Then open http://localhost:%d in your browser\n", port)
	}
}

// buildWASM compiles the game to WebAssembly
func buildWASM(gameDir, outputPath string) error {
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", outputPath, ".")
	cmd.Dir = gameDir
	cmd.Env = buildWASMEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildWASMEnv creates an environment for WASM compilation.
// It filters out any existing GOOS/GOARCH to avoid duplicate entries,
// then sets the required values for WebAssembly compilation.
func buildWASMEnv() []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env)+2)

	// Filter out existing GOOS and GOARCH to avoid duplicates
	for _, e := range env {
		if !strings.HasPrefix(e, "GOOS=") && !strings.HasPrefix(e, "GOARCH=") {
			filtered = append(filtered, e)
		}
	}

	// Add the required WASM environment variables
	filtered = append(filtered, "GOOS=js", "GOARCH=wasm")
	return filtered
}

// copyWASMExec copies wasm_exec.js from the Go installation
func copyWASMExec(outputDir string) error {
	// Find GOROOT using 'go env GOROOT' (more portable than runtime.GOROOT)
	goroot, err := getGoRoot()
	if err != nil {
		return fmt.Errorf("failed to determine GOROOT: %w", err)
	}

	// Try multiple possible locations as the path has changed between Go versions
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

	// Copy file
	wasmExecDst := filepath.Join(outputDir, "wasm_exec.js")
	return copyFile(wasmExecSrc, wasmExecDst)
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

// copyFile copies a file from src to dst
func copyFile(src, dst string) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		// Capture Close error only if no prior error occurred
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		// Capture Close error only if no prior error occurred
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// serveHTTP starts a simple HTTP server for testing
func serveHTTP(dir string, port int) {
	// Use Python's http.server for simplicity
	cmd := exec.Command("python3", "-m", "http.server", fmt.Sprintf("%d", port))
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
