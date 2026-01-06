package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileType represents the type of input file
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeSRT
	FileTypeVTT
	FileTypeTXT
)

// DetectFileType determines the file type from extension
func DetectFileType(filePath string) FileType {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".srt":
		return FileTypeSRT
	case ".vtt":
		return FileTypeVTT
	case ".txt":
		return FileTypeTXT
	default:
		return FileTypeUnknown
	}
}

// ParseFile reads a subtitle or text file and returns clean text
func ParseFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if len(content) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	text := string(content)
	fileType := DetectFileType(filePath)

	switch fileType {
	case FileTypeSRT:
		return parseSRT(text), nil
	case FileTypeVTT:
		return parseVTT(text), nil
	case FileTypeTXT:
		return strings.TrimSpace(text), nil
	default:
		return "", fmt.Errorf("unsupported file type: %s (supported: .srt, .vtt, .txt)", filepath.Ext(filePath))
	}
}

// SRT timestamp pattern: 00:00:01,000 --> 00:00:04,000
var srtTimestampRe = regexp.MustCompile(`\d{2}:\d{2}:\d{2},\d{3}\s*-->\s*\d{2}:\d{2}:\d{2},\d{3}`)
var srtIndexRe = regexp.MustCompile(`^\d+$`)
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// parseSRT parses SubRip format, removing timestamps and formatting
func parseSRT(content string) string {
	lines := strings.Split(content, "\n")
	var textLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip sequence numbers
		if srtIndexRe.MatchString(line) {
			continue
		}

		// Skip timestamp lines
		if srtTimestampRe.MatchString(line) {
			continue
		}

		// Remove HTML tags
		line = htmlTagRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)

		if line != "" {
			textLines = append(textLines, line)
		}
	}

	return strings.Join(textLines, " ")
}

// VTT timestamp pattern: 00:00:01.000 --> 00:00:04.000
var vttTimestampRe = regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{3}\s*-->\s*\d{2}:\d{2}:\d{2}\.\d{3}`)
var vttHeaderRe = regexp.MustCompile(`^(WEBVTT|NOTE|STYLE|REGION)`)
var vttCueSettingsRe = regexp.MustCompile(`\s+(align|position|line|size|vertical):[\w%]+`)

// parseVTT parses WebVTT format, removing timestamps, headers, and formatting
func parseVTT(content string) string {
	lines := strings.Split(content, "\n")
	var textLines []string
	skipUntilEmpty := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip header sections (NOTE, STYLE, REGION blocks)
		if vttHeaderRe.MatchString(line) {
			if line == "WEBVTT" {
				continue
			}
			// Start skipping until empty line for NOTE/STYLE/REGION blocks
			skipUntilEmpty = true
			continue
		}

		if skipUntilEmpty {
			if line == "" {
				skipUntilEmpty = false
			}
			continue
		}

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip timestamp lines (with optional cue settings)
		if vttTimestampRe.MatchString(line) {
			continue
		}

		// Skip cue identifiers (lines before timestamps, usually short alphanumeric)
		if !strings.Contains(line, " ") && len(line) < 30 && !strings.ContainsAny(line, ".,!?;:") {
			continue
		}

		// Remove cue settings from text lines
		line = vttCueSettingsRe.ReplaceAllString(line, "")

		// Remove HTML tags
		line = htmlTagRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)

		if line != "" {
			textLines = append(textLines, line)
		}
	}

	return strings.Join(textLines, " ")
}
