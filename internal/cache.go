package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const (
	cacheDirName     = "yuki"
	audioSubDir      = "audio"
	transcriptSubDir = "transcripts"
)

// Cache manages the yuki cache directory
type Cache struct {
	baseDir string
}

// NewCache creates a new cache manager
func NewCache() (*Cache, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	c := &Cache{baseDir: cacheDir}

	// Ensure directories exist
	dirs := []string{
		c.baseDir,
		filepath.Join(c.baseDir, audioSubDir),
		filepath.Join(c.baseDir, transcriptSubDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory %s: %w", dir, err)
		}
	}

	return c, nil
}

// getCacheDir returns the cache directory path
func getCacheDir() (string, error) {
	// Check XDG_CACHE_HOME first
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, cacheDirName), nil
	}

	// Fall back to ~/.cache
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	return filepath.Join(homeDir, ".cache", cacheDirName), nil
}

// ExtractVideoID extracts the video ID from a YouTube URL
func ExtractVideoID(url string) (string, error) {
	patterns := []string{
		`youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`,
		`youtu\.be/([a-zA-Z0-9_-]{11})`,
		`youtube\.com/shorts/([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not extract video ID from URL: %s", url)
}

// AudioPath returns the cache path for audio
func (c *Cache) AudioPath(videoID string) string {
	return filepath.Join(c.baseDir, audioSubDir, videoID+".mp3")
}

// TranscriptPath returns the cache path for transcript
func (c *Cache) TranscriptPath(videoID string) string {
	return filepath.Join(c.baseDir, transcriptSubDir, videoID+".txt")
}

// HasAudio checks if audio is cached
func (c *Cache) HasAudio(videoID string) bool {
	_, err := os.Stat(c.AudioPath(videoID))
	return err == nil
}

// HasTranscript checks if transcript is cached
func (c *Cache) HasTranscript(videoID string) bool {
	_, err := os.Stat(c.TranscriptPath(videoID))
	return err == nil
}

// GetTranscript retrieves cached transcript
func (c *Cache) GetTranscript(videoID string) (string, error) {
	content, err := os.ReadFile(c.TranscriptPath(videoID))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// SaveAudio copies audio file to cache
func (c *Cache) SaveAudio(videoID, sourcePath string) error {
	return copyFile(sourcePath, c.AudioPath(videoID))
}

// SaveTranscript saves transcript to cache
func (c *Cache) SaveTranscript(videoID, transcript string) error {
	return os.WriteFile(c.TranscriptPath(videoID), []byte(transcript), 0644)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

// Clear removes all cached data
func (c *Cache) Clear() error {
	return os.RemoveAll(c.baseDir)
}

// BaseDir returns the cache base directory path
func (c *Cache) BaseDir() string {
	return c.baseDir
}
