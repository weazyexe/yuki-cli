package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Transcribe converts audio file to text using mlx_whisper
func Transcribe(audioPath, outputDir string) (string, error) {
	spinner := NewSpinner("Transcribing")

	cmd := exec.Command("mlx_whisper",
		audioPath,
		"--model", "mlx-community/whisper-medium-mlx",
		"--output-format", "txt",
		"--output-dir", outputDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		spinner.StopWithError()
		return "", fmt.Errorf("whisper error: %w\nOutput: %s", err, string(output))
	}

	spinner.Stop()

	// Whisper creates output file with same name as input but .txt extension
	baseName := filepath.Base(audioPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	txtPath := filepath.Join(outputDir, baseName+".txt")

	content, err := os.ReadFile(txtPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript: %w", err)
	}

	return string(content), nil
}
