package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/weazyexe/yuki-cli/internal"
)

var (
	count        int
	output       string
	level        string
	apiURL       string
	apiKey       string
	model        string
	noReview     bool
	noCache      bool
	clearCache   bool
	refreshCache bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yuki [flags] <youtube-url|file>",
		Short: "Convert YouTube videos or subtitle files to Anki flashcard decks",
		Long:  "CLI utility that extracts vocabulary from YouTube videos or subtitle/text files and creates Anki decks",
		Args:  cobra.MaximumNArgs(1),
		RunE:  run,
	}

	rootCmd.Flags().IntVarP(&count, "count", "n", 20, "Number of words to extract")
	rootCmd.Flags().StringVarP(&output, "output", "o", "deck.apkg", "Output file path")
	rootCmd.Flags().StringVarP(&level, "level", "l", "B1", "Language level: A2, B1, B2")
	rootCmd.Flags().StringVar(&apiURL, "api-url", "http://localhost:11434/v1", "OpenAI-compatible API URL")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "API key (or env: OPENAI_API_KEY)")
	rootCmd.Flags().StringVar(&model, "model", "gpt-4o-mini", "LLM model name")
	rootCmd.Flags().BoolVar(&noReview, "no-review", false, "Skip interactive review, add all words")
	rootCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable cache for this run")
	rootCmd.Flags().BoolVar(&clearCache, "clear-cache", false, "Clear cache and exit")
	rootCmd.Flags().BoolVar(&refreshCache, "refresh", false, "Re-download and re-transcribe (ignore cache)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Handle --clear-cache
	if clearCache {
		cache, err := internal.NewCache()
		if err != nil {
			return fmt.Errorf("failed to initialize cache: %w", err)
		}
		if err := cache.Clear(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Println("Cache cleared successfully")
		return nil
	}

	// Require input for normal operation
	if len(args) == 0 {
		return fmt.Errorf("YouTube URL or file path required")
	}
	input := args[0]

	// Detect input type early to provide better error messages
	inputType := internal.DetectInputType(input)
	if inputType == internal.InputTypeUnknown {
		return fmt.Errorf("input must be a valid YouTube URL or existing file: %s", input)
	}

	// Start timing
	startTime := time.Now()

	// Validate level
	validLevels := map[string]bool{"A2": true, "B1": true, "B2": true}
	if !validLevels[level] {
		return fmt.Errorf("invalid level: %s (must be A2, B1, or B2)", level)
	}

	// Get API key from flag or environment
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("API key required: use --api-key flag or set OPENAI_API_KEY environment variable")
	}

	// Get transcript based on input type
	var transcript string
	var err error

	switch inputType {
	case internal.InputTypeYouTube:
		transcript, err = processYouTube(input)
	case internal.InputTypeFile:
		transcript, err = processFile(input)
	}

	if err != nil {
		return err
	}

	// Extract vocabulary
	llmClient := internal.NewLLMClient(apiURL, apiKey, model)
	vocabulary, err := llmClient.ExtractVocabulary(transcript, count, level)
	if err != nil {
		return fmt.Errorf("vocabulary extraction failed: %w", err)
	}

	// Print total time before review
	totalTime := time.Since(startTime)
	fmt.Printf("\nExtracted %d words\n", len(vocabulary))
	fmt.Printf("Total time: %s\n", internal.FormatDuration(totalTime))

	// Interactive review
	if !noReview {
		vocabulary = internal.ReviewVocabulary(vocabulary)
		if len(vocabulary) == 0 {
			fmt.Println("No words selected. Exiting.")
			return nil
		}
	}

	fmt.Println("\nGenerating Anki deck...")
	deckName := filepath.Base(output)
	deckName = deckName[:len(deckName)-len(filepath.Ext(deckName))]
	if err := internal.GenerateAPKG(vocabulary, output, deckName); err != nil {
		return fmt.Errorf("APKG generation failed: %w", err)
	}

	fmt.Printf("Deck saved to: %s\n", output)
	return nil
}

// processYouTube handles YouTube URL input with download and transcription
func processYouTube(url string) (string, error) {
	// Check external dependencies
	if err := internal.CheckYouTubeDependencies(); err != nil {
		return "", err
	}

	// Extract video ID for caching
	videoID, err := internal.ExtractVideoID(url)
	if err != nil {
		return "", fmt.Errorf("failed to extract video ID: %w", err)
	}

	// Initialize cache (unless disabled)
	var cache *internal.Cache
	useCache := !noCache
	if useCache {
		cache, err = internal.NewCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not initialize cache: %v\n", err)
			useCache = false
		}
	}

	var audioPath string
	var transcript string

	// Check cache for transcript (most valuable to cache)
	if useCache && !refreshCache && cache.HasTranscript(videoID) {
		fmt.Printf("Using cached transcript for %s\n", videoID)
		transcript, err = cache.GetTranscript(videoID)
		if err != nil {
			return "", fmt.Errorf("failed to read cached transcript: %w", err)
		}
		return transcript, nil
	}

	// Need to download and/or transcribe

	// Check cache for audio
	if useCache && !refreshCache && cache.HasAudio(videoID) {
		fmt.Printf("Using cached audio for %s\n", videoID)
		audioPath = cache.AudioPath(videoID)
	} else {
		// Download audio
		tempDir, err := os.MkdirTemp("", "yuki-*")
		if err != nil {
			return "", fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		audioPath, err = internal.DownloadAudio(url, tempDir)
		if err != nil {
			return "", fmt.Errorf("download failed: %w", err)
		}

		// Cache the audio
		if useCache {
			if err := cache.SaveAudio(videoID, audioPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not cache audio: %v\n", err)
			} else {
				audioPath = cache.AudioPath(videoID)
			}
		}
	}

	// Transcribe
	tempDir, err := os.MkdirTemp("", "yuki-transcribe-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	transcript, err = internal.Transcribe(audioPath, tempDir)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	// Cache the transcript
	if useCache {
		if err := cache.SaveTranscript(videoID, transcript); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not cache transcript: %v\n", err)
		}
	}

	return transcript, nil
}

// processFile handles file input (SRT, VTT, TXT)
func processFile(filePath string) (string, error) {
	// Validate file exists and is not a directory
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot access file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	fmt.Printf("Parsing file: %s\n", filePath)

	transcript, err := internal.ParseFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	return transcript, nil
}
