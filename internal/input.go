package internal

import (
	"os"
	"regexp"
)

// InputType represents the source of input
type InputType int

const (
	InputTypeYouTube InputType = iota
	InputTypeFile
	InputTypeUnknown
)

// IsValidYouTubeURL checks if the given string is a valid YouTube URL
func IsValidYouTubeURL(url string) bool {
	patterns := []string{
		`^https?://(www\.)?youtube\.com/watch\?v=[\w-]+`,
		`^https?://youtu\.be/[\w-]+`,
		`^https?://(www\.)?youtube\.com/shorts/[\w-]+`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, url)
		if matched {
			return true
		}
	}
	return false
}

// DetectInputType determines if input is a YouTube URL or a file
func DetectInputType(input string) InputType {
	// Check YouTube URL patterns first
	if IsValidYouTubeURL(input) {
		return InputTypeYouTube
	}
	// Check if file exists
	info, err := os.Stat(input)
	if err == nil && !info.IsDir() {
		return InputTypeFile
	}
	return InputTypeUnknown
}
