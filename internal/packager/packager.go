package packager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Packager struct {
	bin string
}

func New() *Packager {
	bin := "shaka-packager"
	if b := os.Getenv("SHAKA_PACKAGER_BIN"); b != "" {
		bin = b
	}
	return &Packager{bin: bin}
}

type EncryptRequest struct {
	InputPath string
	OutputDir string
	KeyIDHex  string
	KeyHex    string
}

func (p *Packager) Encrypt(req EncryptRequest) error {
	if err := os.MkdirAll(req.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	audioOut := filepath.Join(req.OutputDir, "audio.mp4")
	videoOut := filepath.Join(req.OutputDir, "video.mp4")
	mpdOut := filepath.Join(req.OutputDir, "manifest.mpd")

	args := []string{
		fmt.Sprintf("in=%s,stream=audio,output=%s", req.InputPath, audioOut),
		fmt.Sprintf("in=%s,stream=video,output=%s", req.InputPath, videoOut),
		"--mpd_output", mpdOut,
		"--enable_raw_key_encryption",
		"--keys", fmt.Sprintf("key_id=%s:key=%s", req.KeyIDHex, req.KeyHex),
		"--clear_lead", "0",
		"--protection_scheme", "cenc",
	}

	cmd := exec.Command(p.bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shaka-packager: %w", err)
	}
	return nil
}
