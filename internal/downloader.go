package internal

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// DownloadAudio downloads audio from YouTube video using yt-dlp
func DownloadAudio(url, outputDir string) (string, error) {
	outputTemplate := filepath.Join(outputDir, "audio.%(ext)s")

	spinner := NewSpinner("Downloading")

	cmd := exec.Command("yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"-o", outputTemplate,
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		spinner.StopWithError()
		return "", fmt.Errorf("yt-dlp error: %w\nOutput: %s", err, string(output))
	}

	spinner.Stop()

	audioPath := filepath.Join(outputDir, "audio.mp3")
	return audioPath, nil
}

// CheckDownloadDependencies verifies that yt-dlp is available
func CheckDownloadDependencies() error {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return fmt.Errorf("yt-dlp not found in PATH. Please install it first")
	}
	return nil
}

// CheckTranscribeDependencies verifies that mlx_whisper is available
func CheckTranscribeDependencies() error {
	_, err := exec.LookPath("mlx_whisper")
	if err != nil {
		return fmt.Errorf("mlx_whisper not found in PATH. Please install it first")
	}
	return nil
}

// CheckYouTubeDependencies verifies that all YouTube processing tools are available
func CheckYouTubeDependencies() error {
	if err := CheckDownloadDependencies(); err != nil {
		return err
	}
	if err := CheckTranscribeDependencies(); err != nil {
		return err
	}
	return nil
}
