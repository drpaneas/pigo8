// Package main implements a tool that builds PIGO8 example games to
// WebAssembly, drives them headlessly via Chrome DevTools Protocol, and
// captures screenshots/GIFs used in the documentation site.
//
// Usage:
//
//	go run ./cmd/docshots                  # capture every job in the manifest
//	go run ./cmd/docshots -only hello-world # capture a single named job
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/drpaneas/pigo8/internal/webbuild"
)

func main() {
	only := flag.String("only", "", "Name of a single manifest entry to capture (default: all)")
	outDir := flag.String("out", "docs/img/generated", "Output directory for captured images")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getting working directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "creating output dir: %v\n", err)
		os.Exit(1)
	}

	jobs := manifest
	if *only != "" {
		jobs = filterJobs(manifest, *only)
		if len(jobs) == 0 {
			fmt.Fprintf(os.Stderr, "no manifest entry named %q\n", *only)
			os.Exit(1)
		}
	}

	for _, job := range jobs {
		if err := runJob(repoRoot, job, *outDir); err != nil {
			fmt.Fprintf(os.Stderr, "job %q failed: %v\n", job.Name, err)
			os.Exit(1)
		}
		fmt.Printf("captured %s\n", job.Name)
	}
}

func filterJobs(all []CaptureJob, name string) []CaptureJob {
	var out []CaptureJob
	for _, j := range all {
		if j.Name == name {
			out = append(out, j)
		}
	}
	return out
}

func runJob(repoRoot string, job CaptureJob, outDir string) error {
	gameDir := filepath.Join(repoRoot, job.ExampleDir)
	buildDir, err := os.MkdirTemp("", "docshots-build-*")
	if err != nil {
		return fmt.Errorf("creating temp build dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(buildDir) }()

	modulePath := "github.com/drpaneas/pigo8/examples/" + filepath.Base(job.ExampleDir)
	if err := webbuild.EnsureExampleModule(gameDir, repoRoot, modulePath); err != nil {
		return fmt.Errorf("ensuring example module: %w", err)
	}

	wasmPath := filepath.Join(buildDir, "game.wasm")
	if err := webbuild.BuildWASM(gameDir, wasmPath); err != nil {
		return fmt.Errorf("building wasm: %w", err)
	}
	if err := webbuild.CopyWASMExec(buildDir); err != nil {
		return fmt.Errorf("copying wasm_exec.js: %w", err)
	}
	if err := webbuild.GenerateHTML(filepath.Join(buildDir, "index.html"), job.Name); err != nil {
		return fmt.Errorf("generating html: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting listener: %w", err)
	}
	server := &http.Server{
		Handler:           http.FileServer(http.Dir(buildDir)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "docshots: http server error: %v\n", err)
		}
	}()
	defer func() { _ = server.Close() }()

	pageURL := fmt.Sprintf("http://%s/index.html", listener.Addr().String())

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), chromedp.DefaultExecAllocatorOptions[:]...)
	defer allocCancel()
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	frames, err := captureJob(browserCtx, job, pageURL)
	if err != nil {
		return fmt.Errorf("capturing: %w", err)
	}
	if len(frames) == 0 {
		return fmt.Errorf("no frames captured")
	}

	switch job.Kind {
	case jobKindStatic:
		return savePNG(bestFrame(frames), filepath.Join(outDir, job.Name+".png"))
	case jobKindGIF:
		g, err := framesToGIF(frames, job.SampleMs)
		if err != nil {
			return fmt.Errorf("assembling gif: %w", err)
		}
		return saveGIF(g, filepath.Join(outDir, job.Name+".gif"))
	default:
		return fmt.Errorf("unknown job kind %q", job.Kind)
	}
}

func savePNG(img image.Image, path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("closing %s: %w", path, closeErr)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encoding png to %s: %w", path, err)
	}
	return nil
}

func saveGIF(g *gif.GIF, path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("closing %s: %w", path, closeErr)
		}
	}()
	if err := gif.EncodeAll(f, g); err != nil {
		return fmt.Errorf("encoding gif to %s: %w", path, err)
	}
	return nil
}
