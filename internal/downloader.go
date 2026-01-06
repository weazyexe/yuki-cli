package internal

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// DownloadAudio downloads audio from YouTube video using yt-dlp
func DownloadAudio(url, outputDir string) (string, error) {
	outputTemplate := filepath.Join(outputDir, "audio.%(ext)s")

	cmd := exec.Command("yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"-o", outputTemplate,
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp error: %w\nOutput: %s", err, string(output))
	}

	audioPath := filepath.Join(outputDir, "audio.mp3")
	return audioPath, nil
}

// CheckDependencies verifies that required external tools are available
func CheckDependencies() error {
	deps := []string{"yt-dlp", "mlx_whisper"}

	for _, dep := range deps {
		_, err := exec.LookPath(dep)
		if err != nil {
			return fmt.Errorf("%s not found in PATH. Please install it first", dep)
		}
	}

	return nil
}
