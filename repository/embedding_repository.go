package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	startedAt := time.Now()
	body, err := json.Marshal(embedRequest{Keywords: keywords})
	if err != nil {
		log.Printf("[recommend][embedding] request_failed err=%v", err)
		return nil, fmt.Errorf("failed to marshal embed request: %w", err)
	}

	url := r.baseURL + "/embed"
	log.Printf(
		"[recommend][embedding] request_started endpoint=%s keywords_chars=%d",
		url, len([]rune(strings.TrimSpace(keywords))),
	)

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[recommend][embedding] request_failed endpoint=%s err=%v", url, err)
		return nil, fmt.Errorf("failed to build embed request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		log.Printf(
			"[recommend][embedding] request_failed endpoint=%s duration_ms=%d err=%v",
			url, time.Since(startedAt).Milliseconds(), err,
		)
		return nil, fmt.Errorf("embed request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(
			"[recommend][embedding] request_failed endpoint=%s duration_ms=%d err=%v",
			url, time.Since(startedAt).Milliseconds(), err,
		)
		return nil, fmt.Errorf("failed to read embed response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf(
			"[recommend][embedding] request_failed endpoint=%s status=%d duration_ms=%d",
			url, resp.StatusCode, time.Since(startedAt).Milliseconds(),
		)
		return nil, fmt.Errorf("embed service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed embedResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		log.Printf(
			"[recommend][embedding] request_failed endpoint=%s status=%d duration_ms=%d err=%v",
			url, resp.StatusCode, time.Since(startedAt).Milliseconds(), err,
		)
		return nil, fmt.Errorf("failed to decode embed response: %w", err)
	}
	log.Printf(
		"[recommend][embedding] request_succeeded endpoint=%s status=%d duration_ms=%d vector_dims=%d",
		url, resp.StatusCode, time.Since(startedAt).Milliseconds(), len(parsed.Vector),
	)
	return parsed.Vector, nil
}
