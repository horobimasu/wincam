package tests_test

import (
	"os"
	"testing"

	"github.com/horobimasu/wincam"
)

func TestCaptureWebcam(t *testing.T) {
	bytes, err := wincam.CaptureWebcam()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("length of capture: %d", bytes.Len())
}

func TestCaptureWebcamSaveToDisk(t *testing.T) {
	const SAVE_PATH = "$USERPROFILE\\Desktop\\wincam_test.png"

	bytes, err := wincam.CaptureWebcam()
	if err != nil {
		t.Errorf("failed to capture: %v", err)
		return
	}

	savePath := os.ExpandEnv(SAVE_PATH)

	err = os.WriteFile(savePath, bytes.Bytes(), 0666)
	if err != nil {
		t.Errorf("failed to save capture: %v", err)
		return
	}

	t.Logf("saved to: %s", savePath)
}
