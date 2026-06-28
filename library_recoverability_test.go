package pigo8

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

const helperOK = "helper-ok"

func TestSprWithInvalidSpritesheetDoesNotExitProcess(t *testing.T) {
	if os.Getenv("PIGO8_HELPER_PROCESS") == "spr-invalid" {
		tempDir, err := os.MkdirTemp("", "pigo8-spr-invalid-*")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Logf("failed to remove temp dir: %v", err)
			}
		}()

		if err := os.WriteFile(filepath.Join(tempDir, "spritesheet.json"), []byte("{invalid json"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		currentScreen = ebiten.NewImage(8, 8)
		currentSprites = nil
		Spr(1, 0, 0)
		fmt.Println(helperOK)
		return
	}

	output, err := runRecoverabilityHelper(t, "spr-invalid")
	if err != nil {
		t.Fatalf("helper process failed: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte(helperOK)) {
		t.Fatalf("expected helper confirmation, got:\n%s", output)
	}
}

func TestPauseQuitDoesNotExitProcess(t *testing.T) {
	if os.Getenv("PIGO8_HELPER_PROCESS") == "pause-quit" {
		originalPauseConfirmPressed := pauseConfirmPressed
		pauseConfirmPressed = func() bool { return true }
		defer func() {
			pauseConfirmPressed = originalPauseConfirmPressed
		}()

		g := &game{
			initialized:     true,
			firstFrameDrawn: true,
			paused:          true,
			pauseSelected:   EngPauseOptionExit,
		}
		err := g.Update()
		fmt.Printf("update-error:%v\n", err)
		fmt.Println(helperOK)
		return
	}

	output, err := runRecoverabilityHelper(t, "pause-quit")
	if err != nil {
		t.Fatalf("helper process failed: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte(helperOK)) {
		t.Fatalf("expected helper confirmation, got:\n%s", output)
	}
	if !bytes.Contains(output, []byte("update-error:pigo8: quit requested")) {
		t.Fatalf("expected quit request error, got:\n%s", output)
	}
}

func runRecoverabilityHelper(t *testing.T, mode string) ([]byte, error) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), "PIGO8_HELPER_PROCESS="+mode)
	return cmd.CombinedOutput()
}
