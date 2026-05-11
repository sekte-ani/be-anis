package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	ollamaModel = "llama3.2:1b"

	recommendFeatureSystemPrompt = `Kamu adalah konsultan IT dari software house Indonesia bernama A.N.I Tech.
Tugasmu adalah memberikan rekomendasi fitur website yang relevan, praktis,
dan sesuai kebutuhan bisnis klien.

Output HARUS dalam format Markdown agar bisa dirender oleh React Markdown.

Aturan output (WAJIB):
- Berikan tepat 5-7 rekomendasi fitur
- Gunakan ordered list Markdown, setiap item HARUS diawali "1. ", "2. ", dst
- Setiap item berformat: ` + "`1. **Nama Fitur** — penjelasan singkat satu kalimat.`" + `
  (nama fitur dibungkus ` + "`**...**`" + `, dipisah em dash " — " sebelum penjelasan)
- Setiap item berada di baris sendiri (akhiri dengan newline)
- Gunakan Bahasa Indonesia yang profesional namun mudah dipahami
- Fokus pada fitur yang benar-benar dibutuhkan bisnis tersebut
- Jangan tambahkan judul, heading, kalimat pembuka, atau kalimat penutup
- Jangan tambahkan penjelasan tambahan di luar list
- Jangan gunakan bullet "-" atau "*" sebagai list utama
- Jangan bungkus output dalam code block atau backticks`
)

var (
	featureListLineRegex  = regexp.MustCompile(`^\s*(?:\d+[.)]|[-*])\s+(.+?)\s*$`)
	featurePlainLineRegex = regexp.MustCompile(`^\s*\*\*[^*]+?\*\*\s*[—:-]\s+.+$`)
	markdownFenceRegex    = regexp.MustCompile("(?s)^```(?:markdown|md)?\\s*(.*?)\\s*```$")
)

type OllamaRepository interface {
	GenerateWebsiteFeatures(input string) (string, error)
}

type ollamaRepository struct {
	client *openai.Client
	url    string
}

func OllamaModelName() string {
	return ollamaModel
}

func NewOllamaRepository(ollamaURL string) OllamaRepository {
	baseURL := normalizeOllamaBaseURL(ollamaURL)
	cfg := openai.DefaultConfig("ollama")
	cfg.BaseURL = baseURL

	return &ollamaRepository{
		client: openai.NewClientWithConfig(cfg),
		url:    baseURL,
	}
}

func (r *ollamaRepository) GenerateWebsiteFeatures(input string) (string, error) {
	if r.url == "" {
		return "", errors.New("OLLAMA_URL is not configured")
	}
	startedAt := time.Now()
	log.Printf(
		"[recommend][ollama] request_started model=%s base_url=%s input_chars=%d",
		ollamaModel, r.url, len([]rune(strings.TrimSpace(input))),
	)

	userPrompt := fmt.Sprintf(
		"Klien saya ingin membuat website dengan deskripsi berikut:\n'%s'\nBerikan rekomendasi fitur website yang cocok untuk bisnis ini.",
		input,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	resp, err := r.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: ollamaModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: recommendFeatureSystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.2,
	})
	if err != nil {
		log.Printf(
			"[recommend][ollama] request_failed model=%s base_url=%s duration_ms=%d err=%v",
			ollamaModel, r.url, time.Since(startedAt).Milliseconds(), err,
		)
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		log.Printf(
			"[recommend][ollama] request_failed model=%s base_url=%s duration_ms=%d err=%q",
			ollamaModel, r.url, time.Since(startedAt).Milliseconds(), "ollama returned no choices",
		)
		return "", errors.New("ollama returned no choices")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	normalized := normalizeFeatureMarkdown(raw)

	log.Printf(
		"[recommend][ollama] request_succeeded model=%s base_url=%s duration_ms=%d output_chars=%d",
		ollamaModel, r.url, time.Since(startedAt).Milliseconds(), len([]rune(normalized)),
	)
	return normalized, nil
}

func normalizeOllamaBaseURL(ollamaURL string) string {
	base := strings.TrimSpace(strings.TrimRight(ollamaURL, "/"))
	if base == "" {
		return ""
	}
	if strings.HasSuffix(base, "/v1") {
		return base
	}
	return base + "/v1"
}

func normalizeFeatureMarkdown(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}

	if match := markdownFenceRegex.FindStringSubmatch(text); len(match) == 2 {
		text = strings.TrimSpace(match[1])
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	items := make([]string, 0, 7)
	current := ""

	flushCurrent := func() {
		current = strings.TrimSpace(current)
		if current == "" {
			return
		}
		items = append(items, current)
		current = ""
	}

	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			continue
		}

		if match := featureListLineRegex.FindStringSubmatch(clean); len(match) == 2 {
			flushCurrent()
			current = strings.TrimSpace(match[1])
			continue
		}

		if featurePlainLineRegex.MatchString(clean) {
			flushCurrent()
			current = clean
			continue
		}

		if current != "" {
			current = strings.TrimSpace(current + " " + clean)
		}
	}
	flushCurrent()

	if len(items) == 0 {
		return text
	}

	var out strings.Builder
	for i, item := range items {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(fmt.Sprintf("%d. %s", i+1, item))
	}
	return out.String()
}
