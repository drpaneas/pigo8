package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"strings"
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
	for i := range downs {
		if err := downs[i].Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// releaseKeys dispatches keyup events for the given keys. Returns an error
// for any key name not present in arrowKeyCodes.
func releaseKeys(ctx context.Context, keys []string) error {
	for _, k := range keys {
		code, ok := arrowKeyCodes[k]
		if !ok {
			return fmt.Errorf("unsupported key %q in manifest input step", k)
		}
		up := input.DispatchKeyEvent(input.KeyUp).
			WithCode(code.Code).
			WithKey(code.Key).
			WithWindowsVirtualKeyCode(code.WindowsVirtual).
			WithNativeVirtualKeyCode(code.NativeVirtual)
		if err := up.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// captureCanvasPNG evaluates JS in the page to grab the current <canvas>
// contents as a base64-encoded PNG data URL, then decodes it into an image.Image.
func captureCanvasPNG(ctx context.Context) (image.Image, error) {
	var dataURL string
	script := `(() => {
		const c = document.querySelector('canvas');
		if (!c) return '';
		return c.toDataURL('image/png');
	})()`
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &dataURL)); err != nil {
		return nil, fmt.Errorf("evaluating canvas capture script: %w", err)
	}
	if dataURL == "" {
		return nil, fmt.Errorf("no canvas element found on page")
	}

	const prefix = "data:image/png;base64,"
	encoded := strings.TrimPrefix(dataURL, prefix)
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decoding canvas base64 data: %w", err)
	}

	img, _, err := image.Decode(strings.NewReader(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("decoding canvas PNG data: %w", err)
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
