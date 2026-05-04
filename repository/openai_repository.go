package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	openAIChatURL = "https://api.openai.com/v1/chat/completions"
	openAIModel   = "gpt-4o-mini"

	visionPrompt = `Kamu adalah sistem yang menganalisis mockup website.
Lihat gambar ini dan generate keywords yang relevan
dengan bisnis/industri yang direpresentasikan.
Return hanya keywords, pisahkan dengan koma, maksimal 15 keyword.
Jangan tambahkan penjelasan apapun.`
)

type OpenAIRepository interface {
	GenerateKeywordsFromImage(imageURL string) (string, error)
}

type openAIRepository struct {
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIRepository(apiKey string) OpenAIRepository {
	return &openAIRepository{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
}

type openAIChatMessage struct {
	Role    string             `json:"role"`
	Content []openAIContentPart `json:"content"`
}

type openAIContentPart struct {
	Type     string             `json:"type"`
	Text     string             `json:"text,omitempty"`
	ImageURL *openAIImageURL    `json:"image_url,omitempty"`
}

type openAIImageURL struct {
	URL string `json:"url"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (r *openAIRepository) GenerateKeywordsFromImage(imageURL string) (string, error) {
	if r.apiKey == "" {
		return "", errors.New("OPENAI_API_KEY is not configured")
	}

	reqBody := openAIChatRequest{
		Model: openAIModel,
		Messages: []openAIChatMessage{
			{
				Role: "user",
				Content: []openAIContentPart{
					{Type: "text", Text: visionPrompt},
					{Type: "image_url", ImageURL: &openAIImageURL{URL: imageURL}},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, openAIChatURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build openai request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read openai response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("failed to decode openai response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("openai error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("openai returned no choices")
	}

	keywords := strings.TrimSpace(parsed.Choices[0].Message.Content)
	return keywords, nil
}
