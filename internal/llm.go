package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// VocabularyItem represents a single vocabulary entry
type VocabularyItem struct {
	Word      string `json:"word"`
	Definition string `json:"definition"`
	IPA       string `json:"ipa"`
	ExampleEN string `json:"example_en"`
	ExampleRU string `json:"example_ru"`
}

// LLMClient handles communication with OpenAI-compatible API
type LLMClient struct {
	client *openai.Client
	model  string
}

// NewLLMClient creates a new LLM client
func NewLLMClient(apiURL, apiKey, model string) *LLMClient {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiURL

	return &LLMClient{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}
}

// ExtractVocabulary extracts vocabulary from transcript using LLM
func (c *LLMClient) ExtractVocabulary(transcript string, count int, level string) ([]VocabularyItem, error) {
	prompt := fmt.Sprintf(`Из транскрипта выбери %d слов/фраз уровня %s.

Для каждого верни JSON массив объектов:
{
  "word": "string (слово или фраза на английском)",
  "definition": "string (определение на русском)",
  "ipa": "string (фонетическая транскрипция)",
  "example_en": "string (пример предложения на английском)",
  "example_ru": "string (перевод примера на русский)"
}

Важно:
- Выбирай только слова/фразы уровня %s (не проще и не сложнее)
- Примеры должны быть из контекста транскрипта или похожие по смыслу
- Верни ТОЛЬКО JSON массив, без дополнительного текста

Транскрипт:
%s`, count, level, level, transcript)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
	})

	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := resp.Choices[0].Message.Content
	content = cleanJSONResponse(content)

	var items []VocabularyItem
	if err := json.Unmarshal([]byte(content), &items); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w\nResponse: %s", err, content)
	}

	return items, nil
}

// cleanJSONResponse removes markdown code blocks and extra whitespace
func cleanJSONResponse(s string) string {
	s = strings.TrimSpace(s)

	// Remove markdown code blocks
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}

	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}

	return strings.TrimSpace(s)
}
