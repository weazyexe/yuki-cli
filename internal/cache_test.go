package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		// Positive cases
		{
			name:        "standard watch URL",
			input:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected:    "dQw4w9WgXcQ",
			expectError: false,
		},
		{
			name:        "watch URL without www",
			input:       "https://youtube.com/watch?v=dQw4w9WgXcQ",
			expected:    "dQw4w9WgXcQ",
			expectError: false,
		},
		{
			name:        "short youtu.be URL",
			input:       "https://youtu.be/dQw4w9WgXcQ",
			expected:    "dQw4w9WgXcQ",
			expectError: false,
		},
		{
			name:        "shorts URL",
			input:       "https://www.youtube.com/shorts/dQw4w9WgXcQ",
			expected:    "dQw4w9WgXcQ",
			expectError: false,
		},
		{
			name:        "watch URL with extra params",
			input:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			expected:    "dQw4w9WgXcQ",
			expectError: false,
		},
		{
			name:        "watch URL with time param first",
			input:       "https://www.youtube.com/watch?t=120&v=dQw4w9WgXcQ",
			expected:    "dQw4w9WgXcQ",
			expectError: false,
		},
		{
			name:        "video ID with underscore",
			input:       "https://www.youtube.com/watch?v=abc_123-XYZ",
			expected:    "abc_123-XYZ",
			expectError: false,
		},

		// Negative cases
		{
			name:        "empty string",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "watch URL without video ID",
			input:       "https://www.youtube.com/watch",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid URL",
			input:       "not-a-url",
			expected:    "",
			expectError: true,
		},
		{
			name:        "vimeo URL",
			input:       "https://vimeo.com/123456789",
			expected:    "",
			expectError: true,
		},
		{
			name:        "short video ID (10 chars)",
			input:       "https://www.youtube.com/watch?v=dQw4w9WgXc",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractVideoID(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("ExtractVideoID(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ExtractVideoID(%q) unexpected error: %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("ExtractVideoID(%q) = %q, want %q", tt.input, got, tt.expected)
				}
			}
		})
	}
}

func TestCachePaths(t *testing.T) {
	// Create a cache with a known base directory
	tempDir := t.TempDir()
	cache := &Cache{baseDir: tempDir}

	videoID := "testVideoID"

	// Test AudioPath
	expectedAudioPath := filepath.Join(tempDir, "audio", videoID+".mp3")
	if got := cache.AudioPath(videoID); got != expectedAudioPath {
		t.Errorf("AudioPath(%q) = %q, want %q", videoID, got, expectedAudioPath)
	}

	// Test TranscriptPath
	expectedTranscriptPath := filepath.Join(tempDir, "transcripts", videoID+".txt")
	if got := cache.TranscriptPath(videoID); got != expectedTranscriptPath {
		t.Errorf("TranscriptPath(%q) = %q, want %q", videoID, got, expectedTranscriptPath)
	}

	// Test BaseDir
	if got := cache.BaseDir(); got != tempDir {
		t.Errorf("BaseDir() = %q, want %q", got, tempDir)
	}
}

func TestCacheHasAudioAndTranscript(t *testing.T) {
	tempDir := t.TempDir()

	// Create cache directories
	audioDir := filepath.Join(tempDir, "audio")
	transcriptDir := filepath.Join(tempDir, "transcripts")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		t.Fatalf("failed to create audio dir: %v", err)
	}
	if err := os.MkdirAll(transcriptDir, 0755); err != nil {
		t.Fatalf("failed to create transcripts dir: %v", err)
	}

	cache := &Cache{baseDir: tempDir}
	videoID := "testVideo"

	// Initially, neither should exist
	if cache.HasAudio(videoID) {
		t.Error("HasAudio() should return false for non-existent file")
	}
	if cache.HasTranscript(videoID) {
		t.Error("HasTranscript() should return false for non-existent file")
	}

	// Create audio file
	audioPath := filepath.Join(audioDir, videoID+".mp3")
	if err := os.WriteFile(audioPath, []byte("fake audio"), 0644); err != nil {
		t.Fatalf("failed to create audio file: %v", err)
	}

	if !cache.HasAudio(videoID) {
		t.Error("HasAudio() should return true after creating file")
	}
	if cache.HasTranscript(videoID) {
		t.Error("HasTranscript() should still return false")
	}

	// Create transcript file
	transcriptPath := filepath.Join(transcriptDir, videoID+".txt")
	if err := os.WriteFile(transcriptPath, []byte("test transcript"), 0644); err != nil {
		t.Fatalf("failed to create transcript file: %v", err)
	}

	if !cache.HasTranscript(videoID) {
		t.Error("HasTranscript() should return true after creating file")
	}
}

func TestCacheSaveAndGetTranscript(t *testing.T) {
	tempDir := t.TempDir()

	// Create cache directories
	transcriptDir := filepath.Join(tempDir, "transcripts")
	if err := os.MkdirAll(transcriptDir, 0755); err != nil {
		t.Fatalf("failed to create transcripts dir: %v", err)
	}

	cache := &Cache{baseDir: tempDir}
	videoID := "testVideo"
	transcript := "This is a test transcript content."

	// Save transcript
	if err := cache.SaveTranscript(videoID, transcript); err != nil {
		t.Fatalf("SaveTranscript() failed: %v", err)
	}

	// Verify file exists
	if !cache.HasTranscript(videoID) {
		t.Error("HasTranscript() should return true after SaveTranscript()")
	}

	// Get transcript
	got, err := cache.GetTranscript(videoID)
	if err != nil {
		t.Fatalf("GetTranscript() failed: %v", err)
	}
	if got != transcript {
		t.Errorf("GetTranscript() = %q, want %q", got, transcript)
	}
}

func TestCacheSaveAudio(t *testing.T) {
	tempDir := t.TempDir()

	// Create cache directories
	audioDir := filepath.Join(tempDir, "audio")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		t.Fatalf("failed to create audio dir: %v", err)
	}

	cache := &Cache{baseDir: tempDir}
	videoID := "testVideo"

	// Create a source audio file
	sourceFile := filepath.Join(tempDir, "source.mp3")
	audioContent := []byte("fake audio content")
	if err := os.WriteFile(sourceFile, audioContent, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Save audio
	if err := cache.SaveAudio(videoID, sourceFile); err != nil {
		t.Fatalf("SaveAudio() failed: %v", err)
	}

	// Verify file exists
	if !cache.HasAudio(videoID) {
		t.Error("HasAudio() should return true after SaveAudio()")
	}

	// Verify content
	cachedContent, err := os.ReadFile(cache.AudioPath(videoID))
	if err != nil {
		t.Fatalf("failed to read cached audio: %v", err)
	}
	if string(cachedContent) != string(audioContent) {
		t.Error("cached audio content doesn't match source")
	}
}

func TestCacheClear(t *testing.T) {
	tempDir := t.TempDir()

	// Create cache structure with files
	audioDir := filepath.Join(tempDir, "audio")
	transcriptDir := filepath.Join(tempDir, "transcripts")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		t.Fatalf("failed to create audio dir: %v", err)
	}
	if err := os.MkdirAll(transcriptDir, 0755); err != nil {
		t.Fatalf("failed to create transcripts dir: %v", err)
	}

	// Create some files
	if err := os.WriteFile(filepath.Join(audioDir, "test.mp3"), []byte("audio"), 0644); err != nil {
		t.Fatalf("failed to create audio file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(transcriptDir, "test.txt"), []byte("transcript"), 0644); err != nil {
		t.Fatalf("failed to create transcript file: %v", err)
	}

	cache := &Cache{baseDir: tempDir}

	// Clear cache
	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	// Verify directory is removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("cache directory should be removed after Clear()")
	}
}

func TestGetTranscript_NonExistent(t *testing.T) {
	tempDir := t.TempDir()

	// Create transcripts directory
	transcriptDir := filepath.Join(tempDir, "transcripts")
	if err := os.MkdirAll(transcriptDir, 0755); err != nil {
		t.Fatalf("failed to create transcripts dir: %v", err)
	}

	cache := &Cache{baseDir: tempDir}

	_, err := cache.GetTranscript("nonexistent")
	if err == nil {
		t.Error("GetTranscript() should return error for non-existent file")
	}
}
