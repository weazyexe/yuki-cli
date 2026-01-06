package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidYouTubeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Positive cases
		{"standard watch URL", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"watch URL without www", "https://youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"watch URL with http", "http://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"short youtu.be URL", "https://youtu.be/dQw4w9WgXcQ", true},
		{"shorts URL", "https://www.youtube.com/shorts/dQw4w9WgXcQ", true},
		{"shorts URL without www", "https://youtube.com/shorts/dQw4w9WgXcQ", true},
		{"watch URL with extra params", "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", true},
		{"watch URL with time param", "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=120", true},
		{"video ID with underscore", "https://www.youtube.com/watch?v=abc_123-XYZ", true},
		{"video ID with hyphen", "https://www.youtube.com/watch?v=abc-123_XYZ", true},

		// Negative cases
		{"empty string", "", false},
		{"plain text", "not-a-url", false},
		{"watch URL without video ID", "https://www.youtube.com/watch", false},
		{"watch URL with empty v param", "https://www.youtube.com/watch?v=", false},
		{"youtu.be without video ID", "https://youtu.be/", false},
		{"vimeo URL", "https://vimeo.com/123456789", false},
		{"other domain", "https://example.com/watch?v=dQw4w9WgXcQ", false},
		{"youtube channel URL", "https://www.youtube.com/channel/UCxxxx", false},
		{"youtube playlist URL", "https://www.youtube.com/playlist?list=PLxxxx", false},
		{"ftp protocol", "ftp://www.youtube.com/watch?v=dQw4w9WgXcQ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidYouTubeURL(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidYouTubeURL(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDetectInputType(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a test directory
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		// YouTube URLs
		{"YouTube watch URL", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", InputTypeYouTube},
		{"YouTube short URL", "https://youtu.be/dQw4w9WgXcQ", InputTypeYouTube},
		{"YouTube shorts URL", "https://www.youtube.com/shorts/dQw4w9WgXcQ", InputTypeYouTube},

		// Files
		{"existing file", testFile, InputTypeFile},

		// Unknown
		{"empty string", "", InputTypeUnknown},
		{"non-existent file", "/path/to/nonexistent/file.txt", InputTypeUnknown},
		{"invalid URL", "not-a-url", InputTypeUnknown},
		{"directory path", testDir, InputTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectInputType(tt.input)
			if got != tt.expected {
				t.Errorf("DetectInputType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDetectInputType_PreferYouTubeOverFile(t *testing.T) {
	// Create a file that looks like a YouTube URL (edge case)
	tempDir := t.TempDir()
	weirdFile := filepath.Join(tempDir, "https:__www.youtube.com_watch")
	if err := os.WriteFile(weirdFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// YouTube URL should be detected as YouTube, not as file
	url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	got := DetectInputType(url)
	if got != InputTypeYouTube {
		t.Errorf("DetectInputType(%q) = %v, want InputTypeYouTube", url, got)
	}
}
