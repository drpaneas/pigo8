package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/png"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// Key names as they appear in the manifest's InputStep.Keys lists. The CDP
// Code/Key fields use the same DOM key names, so these double as both the
// map lookup key and the values dispatched to the browser.
const (
	keyArrowLeft  = "ArrowLeft"
	keyArrowUp    = "ArrowUp"
	keyArrowRight = "ArrowRight"
	keyArrowDown  = "ArrowDown"
)

// arrowKeyCodes maps the key names used in the manifest to the JS keyCode /
// CDP key values needed to synthesize a real keydown/keyup pair for Ebiten's
// keyboard listener (Ebiten reads raw keydown/keyup events, not typed text).
var arrowKeyCodes = map[string]struct {
	Code           string
	Key            string
	WindowsVirtual int64
	NativeVirtual  int64
}{
	keyArrowLeft:  {keyArrowLeft, keyArrowLeft, 37, 37},
	keyArrowUp:    {keyArrowUp, keyArrowUp, 38, 38},
	keyArrowRight: {keyArrowRight, keyArrowRight, 39, 39},
	keyArrowDown:  {keyArrowDown, keyArrowDown, 40, 40},
}

// pressKeys dispatches keydown events for the given keys. Returns an error
// for any key name not present in arrowKeyCodes.
//
// The dispatch itself must run inside a chromedp.Run/ActionFunc call: the
// generated CDP params' Do(ctx) method requires a context carrying the
// executor that chromedp's action-running machinery injects, which a bare
// context (e.g. the one from chromedp.NewContext) does not have on its own.
func pressKeys(ctx context.Context, keys []string) error {
	var downs []input.DispatchKeyEventParams
	for _, k := range keys {
		code, ok := arrowKeyCodes[k]
		if !ok {
			return fmt.Errorf("unsupported key %q in manifest input step", k)
		}
		downs = append(downs, *input.DispatchKeyEvent(input.KeyDown).
			WithCode(code.Code).
			WithKey(code.Key).
			WithWindowsVirtualKeyCode(code.WindowsVirtual).
			WithNativeVirtualKeyCode(code.NativeVirtual))
	}
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		for i := range downs {
			if err := downs[i].Do(ctx); err != nil {
				return err
			}
		}
		return nil
	}))
}

// releaseKeys dispatches keyup events for the given keys. Returns an error
// for any key name not present in arrowKeyCodes. See pressKeys for why the
// dispatch runs inside a chromedp.Run/ActionFunc call.
func releaseKeys(ctx context.Context, keys []string) error {
	var ups []input.DispatchKeyEventParams
	for _, k := range keys {
		code, ok := arrowKeyCodes[k]
		if !ok {
			return fmt.Errorf("unsupported key %q in manifest input step", k)
		}
		ups = append(ups, *input.DispatchKeyEvent(input.KeyUp).
			WithCode(code.Code).
			WithKey(code.Key).
			WithWindowsVirtualKeyCode(code.WindowsVirtual).
			WithNativeVirtualKeyCode(code.NativeVirtual))
	}
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		for i := range ups {
			if err := ups[i].Do(ctx); err != nil {
				return err
			}
		}
		return nil
	}))
}

// hideControlsOverlay hides the web-export template's virtual controller UI
// (header bar, D-pad, action buttons) so captured screenshots/GIFs show only
// the game itself. It's a no-op if the elements aren't present.
func hideControlsOverlay(ctx context.Context) error {
	script := `(() => {
		const selectors = ['.header-bar', '.controls-overlay'];
		for (const s of selectors) {
			const el = document.querySelector(s);
			if (el) { el.style.display = 'none'; }
		}
	})()`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// focusCanvas gives keyboard focus to the game's <canvas> element. Ebiten
// attaches its keydown/keyup listeners directly to the canvas (not to
// document/window) and only focuses it itself in response to a real
// click/touch, so without this, synthetic CDP key events never reach the
// game's input handling even though dispatching them succeeds without
// error. Ebiten makes the canvas focusable via tabindex, so a plain
// .focus() call is enough.
func focusCanvas(ctx context.Context) error {
	script := `(() => {
		const c = document.querySelector('canvas');
		if (c) { c.focus(); }
	})()`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// captureCanvasPNG captures the current browser viewport (which the web
// export template sizes to exactly match the game canvas) as a PNG and
// decodes it into an image.Image.
//
// This deliberately does NOT use canvas.toDataURL(): Ebiten renders via
// WebGL, and WebGL canvases don't preserve their drawing buffer after the
// browser compositor consumes each frame unless the context requests
// preserveDrawingBuffer, so toDataURL() called from outside the render loop
// returns a blank/transparent image. A full CDP viewport screenshot captures
// the actual composited pixels instead, avoiding that problem entirely.
func captureCanvasPNG(ctx context.Context) (image.Image, error) {
	var buf []byte
	if err := chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf)); err != nil {
		return nil, fmt.Errorf("capturing viewport screenshot: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("decoding screenshot PNG data: %w", err)
	}
	return img, nil
}

// captureJob runs a full capture session for one manifest entry against a
// page already served at pageURL, returning the captured frames in order.
func captureJob(ctx context.Context, job CaptureJob, pageURL string) ([]image.Image, error) {
	tabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, 30*time.Second)
	defer timeoutCancel()

	if err := chromedp.Run(tabCtx,
		chromedp.Navigate(pageURL),
		chromedp.WaitVisible("canvas", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // let the WASM game finish Init()
	); err != nil {
		return nil, fmt.Errorf("loading page %s: %w", pageURL, err)
	}

	if err := hideControlsOverlay(tabCtx); err != nil {
		return nil, fmt.Errorf("hiding controls overlay: %w", err)
	}
	if err := focusCanvas(tabCtx); err != nil {
		return nil, fmt.Errorf("focusing canvas: %w", err)
	}

	var frames []image.Image
	elapsed := 0

	// For each input step: press the keys, sample frames throughout the hold
	// duration (so the GIF actually shows the action happening), then release.
	for i, step := range job.Inputs {
		if err := pressKeys(tabCtx, step.Keys); err != nil {
			return nil, fmt.Errorf("pressing keys for input step %d: %w", i, err)
		}

		stepElapsed := 0
		for stepElapsed < step.HoldMs {
			frame, err := captureFrame(tabCtx)
			if err != nil {
				return nil, fmt.Errorf("capturing frame during input step %d: %w", i, err)
			}
			frames = append(frames, frame)

			sleepMs := job.SampleMs
			if remaining := step.HoldMs - stepElapsed; remaining < sleepMs {
				sleepMs = remaining
			}
			time.Sleep(time.Duration(sleepMs) * time.Millisecond)
			stepElapsed += sleepMs
		}

		if err := releaseKeys(tabCtx, step.Keys); err != nil {
			return nil, fmt.Errorf("releasing keys for input step %d: %w", i, err)
		}
		elapsed += step.HoldMs
	}

	// Continue sampling for any remaining capture window after inputs are done
	// (covers input-less "static"/"gif" jobs entirely, and any tail time after
	// the last input step for jobs that have inputs).
	for elapsed < job.CaptureMs {
		frame, err := captureFrame(tabCtx)
		if err != nil {
			return nil, fmt.Errorf("capturing frame at %dms elapsed: %w", elapsed, err)
		}
		frames = append(frames, frame)

		time.Sleep(time.Duration(job.SampleMs) * time.Millisecond)
		elapsed += job.SampleMs
	}

	// Guarantee at least one frame even if CaptureMs was 0 or fully consumed
	// by input steps with no tail sampling time remaining.
	if len(frames) == 0 {
		frame, err := captureFrame(tabCtx)
		if err != nil {
			return nil, fmt.Errorf("capturing final frame: %w", err)
		}
		frames = append(frames, frame)
	}

	return frames, nil
}

func captureFrame(ctx context.Context) (image.Image, error) {
	return captureCanvasPNG(ctx)
}
