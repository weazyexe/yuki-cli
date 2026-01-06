package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/weazyexe/yt2anki/internal"
)

var (
	count    int
	output   string
	level    string
	apiURL   string
	apiKey   string
	model    string
	noReview bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yt2anki [flags] <youtube-url>",
		Short: "Convert YouTube videos to Anki flashcard decks",
		Long:  "CLI utility that extracts vocabulary from YouTube videos and creates Anki decks",
		Args:  cobra.ExactArgs(1),
		RunE:  run,
	}

	rootCmd.Flags().IntVarP(&count, "count", "n", 20, "Number of words to extract")
	rootCmd.Flags().StringVarP(&output, "output", "o", "deck.apkg", "Output file path")
	rootCmd.Flags().StringVarP(&level, "level", "l", "B1", "Language level: A2, B1, B2")
	rootCmd.Flags().StringVar(&apiURL, "api-url", "http://localhost:11434/v1", "OpenAI-compatible API URL")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "API key (or env: OPENAI_API_KEY)")
	rootCmd.Flags().StringVar(&model, "model", "gpt-4o-mini", "LLM model name")
	rootCmd.Flags().BoolVar(&noReview, "no-review", false, "Skip interactive review, add all words")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	url := args[0]

	// Validate YouTube URL
	if !isValidYouTubeURL(url) {
		return fmt.Errorf("invalid YouTube URL: %s", url)
	}

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

	// Check external dependencies
	if err := internal.CheckDependencies(); err != nil {
		return err
	}

	// Create temp directory for intermediate files
	tempDir, err := os.MkdirTemp("", "yt2anki-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Println("Downloading audio from YouTube...")
	audioPath, err := internal.DownloadAudio(url, tempDir)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	fmt.Printf("Audio saved to: %s\n", audioPath)

	fmt.Println("Transcribing audio...")
	transcript, err := internal.Transcribe(audioPath, tempDir)
	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	}
	fmt.Printf("Transcript length: %d characters\n", len(transcript))

	fmt.Println("Extracting vocabulary with LLM...")
	llmClient := internal.NewLLMClient(apiURL, apiKey, model)
	vocabulary, err := llmClient.ExtractVocabulary(transcript, count, level)
	if err != nil {
		return fmt.Errorf("vocabulary extraction failed: %w", err)
	}
	fmt.Printf("Extracted %d words\n", len(vocabulary))

	// Interactive review
	if !noReview {
		vocabulary = internal.ReviewVocabulary(vocabulary)
		if len(vocabulary) == 0 {
			fmt.Println("No words selected. Exiting.")
			return nil
		}
	}

	fmt.Println("Generating Anki deck...")
	deckName := filepath.Base(output)
	deckName = deckName[:len(deckName)-len(filepath.Ext(deckName))]
	if err := internal.GenerateAPKG(vocabulary, output, deckName); err != nil {
		return fmt.Errorf("APKG generation failed: %w", err)
	}

	fmt.Printf("Deck saved to: %s\n", output)
	return nil
}

func isValidYouTubeURL(url string) bool {
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
