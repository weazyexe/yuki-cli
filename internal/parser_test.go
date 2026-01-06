package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FileType
	}{
		// Standard extensions
		{"SRT lowercase", "video.srt", FileTypeSRT},
		{"SRT uppercase", "VIDEO.SRT", FileTypeSRT},
		{"SRT mixed case", "Video.Srt", FileTypeSRT},
		{"VTT lowercase", "captions.vtt", FileTypeVTT},
		{"VTT uppercase", "CAPTIONS.VTT", FileTypeVTT},
		{"TXT lowercase", "transcript.txt", FileTypeTXT},
		{"TXT uppercase", "TRANSCRIPT.TXT", FileTypeTXT},

		// Multiple dots
		{"SRT with language code", "video.en.srt", FileTypeSRT},
		{"VTT with language code", "video.ru.vtt", FileTypeVTT},

		// Paths
		{"SRT with path", "/path/to/video.srt", FileTypeSRT},
		{"VTT with path", "C:\\Users\\video.vtt", FileTypeVTT},

		// Unknown/unsupported
		{"MP3 file", "audio.mp3", FileTypeUnknown},
		{"MP4 file", "video.mp4", FileTypeUnknown},
		{"PDF file", "document.pdf", FileTypeUnknown},
		{"No extension", "filename", FileTypeUnknown},
		{"Empty string", "", FileTypeUnknown},
		{"Dot only", ".", FileTypeUnknown},
		{"Hidden file without ext", ".hidden", FileTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFileType(tt.input)
			if got != tt.expected {
				t.Errorf("DetectFileType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseSRT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name: "standard SRT",
			input: `1
00:00:01,000 --> 00:00:04,000
Hello world

2
00:00:05,000 --> 00:00:08,000
Second line`,
			expected: "Hello world Second line",
		},
		{
			name: "SRT with HTML tags",
			input: `1
00:00:01,000 --> 00:00:04,000
<b>Bold</b> and <i>italic</i>`,
			expected: "Bold and italic",
		},
		{
			name: "SRT with font tags",
			input: `1
00:00:01,000 --> 00:00:04,000
<font color="red">Colored text</font>`,
			expected: "Colored text",
		},
		{
			name: "SRT with extra blank lines",
			input: `1
00:00:01,000 --> 00:00:04,000
First


2
00:00:05,000 --> 00:00:08,000
Second`,
			expected: "First Second",
		},
		{
			name: "SRT with multiline cue",
			input: `1
00:00:01,000 --> 00:00:04,000
Line one
Line two`,
			expected: "Line one Line two",
		},
		{
			name: "SRT with numbers in text",
			input: `1
00:00:01,000 --> 00:00:04,000
Chapter 10 of 20`,
			expected: "Chapter 10 of 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSRT(tt.input)
			if got != tt.expected {
				t.Errorf("parseSRT() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseVTT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name: "standard VTT",
			input: `WEBVTT

00:00:01.000 --> 00:00:04.000
Hello world

00:00:05.000 --> 00:00:08.000
Second line`,
			expected: "Hello world Second line",
		},
		{
			name: "VTT without header",
			input: `00:00:01.000 --> 00:00:04.000
Hello world`,
			expected: "Hello world",
		},
		{
			name: "VTT with NOTE block",
			input: `WEBVTT

NOTE This is a comment
that spans multiple lines

00:00:01.000 --> 00:00:04.000
Actual text`,
			expected: "Actual text",
		},
		{
			name: "VTT with STYLE block",
			input: `WEBVTT

STYLE
::cue {
  color: white;
}

00:00:01.000 --> 00:00:04.000
Styled text`,
			expected: "Styled text",
		},
		{
			name: "VTT with cue settings",
			input: `WEBVTT

00:00:01.000 --> 00:00:04.000 align:center position:50%
Centered text`,
			expected: "Centered text",
		},
		{
			name: "VTT with HTML tags",
			input: `WEBVTT

00:00:01.000 --> 00:00:04.000
<b>Bold</b> and <i>italic</i>`,
			expected: "Bold and italic",
		},
		{
			name: "VTT with cue identifier",
			input: `WEBVTT

cue-1
00:00:01.000 --> 00:00:04.000
First cue

cue-2
00:00:05.000 --> 00:00:08.000
Second cue`,
			expected: "First cue Second cue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVTT(tt.input)
			if got != tt.expected {
				t.Errorf("parseVTT() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	srtContent := `1
00:00:01,000 --> 00:00:04,000
Hello from SRT`
	srtFile := filepath.Join(tempDir, "test.srt")
	if err := os.WriteFile(srtFile, []byte(srtContent), 0644); err != nil {
		t.Fatalf("failed to create SRT file: %v", err)
	}

	vttContent := `WEBVTT

00:00:01.000 --> 00:00:04.000
Hello from VTT`
	vttFile := filepath.Join(tempDir, "test.vtt")
	if err := os.WriteFile(vttFile, []byte(vttContent), 0644); err != nil {
		t.Fatalf("failed to create VTT file: %v", err)
	}

	txtContent := "  Hello from TXT  "
	txtFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte(txtContent), 0644); err != nil {
		t.Fatalf("failed to create TXT file: %v", err)
	}

	emptyFile := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	unsupportedFile := filepath.Join(tempDir, "test.mp3")
	if err := os.WriteFile(unsupportedFile, []byte("fake mp3"), 0644); err != nil {
		t.Fatalf("failed to create unsupported file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		expected    string
		expectError bool
	}{
		{
			name:        "SRT file",
			filePath:    srtFile,
			expected:    "Hello from SRT",
			expectError: false,
		},
		{
			name:        "VTT file",
			filePath:    vttFile,
			expected:    "Hello from VTT",
			expectError: false,
		},
		{
			name:        "TXT file",
			filePath:    txtFile,
			expected:    "Hello from TXT",
			expectError: false,
		},
		{
			name:        "empty file",
			filePath:    emptyFile,
			expected:    "",
			expectError: true,
		},
		{
			name:        "non-existent file",
			filePath:    filepath.Join(tempDir, "nonexistent.txt"),
			expected:    "",
			expectError: true,
		},
		{
			name:        "unsupported file type",
			filePath:    unsupportedFile,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(tt.filePath)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseFile(%q) expected error, got nil", tt.filePath)
				}
			} else {
				if err != nil {
					t.Errorf("ParseFile(%q) unexpected error: %v", tt.filePath, err)
				}
				if got != tt.expected {
					t.Errorf("ParseFile(%q) = %q, want %q", tt.filePath, got, tt.expected)
				}
			}
		})
	}
}
