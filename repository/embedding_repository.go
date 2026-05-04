package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EmbeddingRepository interface {
	Embed(keywords string) ([]float64, error)
}

type embeddingRepository struct {
	baseURL    string
	httpClient *http.Client
}

func NewEmbeddingRepository(baseURL string) EmbeddingRepository {
	return &embeddingRepository{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type embedRequest struct {
	Keywords string `json:"keywords"`
}

type embedResponse struct {
	Vector []float64 `json:"vector"`
}

func (r *embeddingRepository) Embed(keywords string) ([]float64, error) {
	body, err := json.Marshal(embedRequest{Keywords: keywords})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed request: %w", err)
	}

	url := r.baseURL + "/embed"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build embed request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("embed request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embed response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embed service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed embedResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode embed response: %w", err)
	}
	return parsed.Vector, nil
}
