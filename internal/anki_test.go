package internal

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestFieldChecksum(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple word", "hello"},
		{"word with spaces", "hello world"},
		{"unicode", "привет"},
		{"mixed", "Hello мир 123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic and returns consistent results
			result1 := fieldChecksum(tt.input)
			result2 := fieldChecksum(tt.input)

			if result1 != result2 {
				t.Errorf("fieldChecksum(%q) not deterministic: got %d and %d", tt.input, result1, result2)
			}
		})
	}

	// Test that different inputs produce different checksums
	t.Run("different inputs different checksums", func(t *testing.T) {
		checksum1 := fieldChecksum("hello")
		checksum2 := fieldChecksum("world")

		if checksum1 == checksum2 {
			t.Errorf("fieldChecksum() should produce different results for different inputs")
		}
	})

	// Test empty string
	t.Run("empty string returns zero", func(t *testing.T) {
		if got := fieldChecksum(""); got != 0 {
			t.Errorf("fieldChecksum(\"\") = %d, want 0", got)
		}
	})
}

func TestGenerateAPKG(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_deck.apkg")

	items := []VocabularyItem{
		{
			Word:       "hello",
			Definition: "привет",
			IPA:        "həˈloʊ",
			ExampleEN:  "Hello, world!",
			ExampleRU:  "Привет, мир!",
		},
		{
			Word:       "world",
			Definition: "мир",
			IPA:        "wɜːrld",
			ExampleEN:  "The world is beautiful.",
			ExampleRU:  "Мир прекрасен.",
		},
	}

	// Generate APKG
	err := GenerateAPKG(items, outputPath, "Test Deck")
	if err != nil {
		t.Fatalf("GenerateAPKG() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	// Verify it's a valid ZIP file
	zipReader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("output is not a valid ZIP file: %v", err)
	}
	defer zipReader.Close()

	// Check for required files
	requiredFiles := map[string]bool{
		"collection.anki2": false,
		"media":            false,
	}

	for _, file := range zipReader.File {
		if _, ok := requiredFiles[file.Name]; ok {
			requiredFiles[file.Name] = true
		}
	}

	for name, found := range requiredFiles {
		if !found {
			t.Errorf("missing required file in APKG: %s", name)
		}
	}
}

func TestGenerateAPKG_EmptyItems(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "empty_deck.apkg")

	// Generate APKG with empty items
	err := GenerateAPKG([]VocabularyItem{}, outputPath, "Empty Deck")
	if err != nil {
		t.Fatalf("GenerateAPKG() with empty items failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("output file was not created for empty deck")
	}
}

func TestGenerateAPKG_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "special_deck.apkg")

	items := []VocabularyItem{
		{
			Word:       "<script>alert('xss')</script>",
			Definition: "тест & спецсимволы < > \"",
			IPA:        "test'ing",
			ExampleEN:  "Test with \"quotes\" & <tags>",
			ExampleRU:  "Тест с 'кавычками' и <тегами>",
		},
	}

	// Should not fail with special characters (they should be escaped)
	err := GenerateAPKG(items, outputPath, "Special <Deck>")
	if err != nil {
		t.Fatalf("GenerateAPKG() with special characters failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}
}

func TestGenerateAPKG_InvalidPath(t *testing.T) {
	items := []VocabularyItem{
		{
			Word:       "test",
			Definition: "тест",
			IPA:        "test",
			ExampleEN:  "Test",
			ExampleRU:  "Тест",
		},
	}

	// Try to write to a non-existent directory
	err := GenerateAPKG(items, "/nonexistent/directory/deck.apkg", "Test")
	if err == nil {
		t.Error("GenerateAPKG() should fail with invalid path")
	}
}

func TestGenerateAPKG_UnicodeContent(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "unicode_deck.apkg")

	items := []VocabularyItem{
		{
			Word:       "日本語",
			Definition: "Japanese language",
			IPA:        "nihongo",
			ExampleEN:  "日本語を勉強しています",
			ExampleRU:  "Я изучаю японский язык",
		},
		{
			Word:       "emoji",
			Definition: "смайлик 😀",
			IPA:        "iˈmoʊdʒi",
			ExampleEN:  "I love emojis! 🎉",
			ExampleRU:  "Я люблю эмодзи! 🎉",
		},
	}

	err := GenerateAPKG(items, outputPath, "Unicode Deck 日本語")
	if err != nil {
		t.Fatalf("GenerateAPKG() with unicode content failed: %v", err)
	}

	// Verify file exists and is valid
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	zipReader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("output is not a valid ZIP file: %v", err)
	}
	zipReader.Close()
}
