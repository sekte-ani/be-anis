package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	groqBaseURL = "https://api.groq.com/openai/v1"
	groqModel   = "llama-3.1-8b-instant"

	classifyPrompt = `Kamu adalah klasifier dan ekstraktor untuk sistem rekomendasi mockup website
software house A.N.I Tech. Input adalah deskripsi bisnis klien dalam Bahasa Indonesia.

Tugasmu mengembalikan JSON dengan dua field: "sektor" dan "keywords".

1. Field "sektor" — pilih SATU nilai dari daftar berikut, atau null jika input ambigu/umum:
   - "kuliner" : makanan, minuman, restoran, cafe, F&B, food delivery, bakery, katering
   - "perdagang" : retail, grosir, marketplace, toko serba ada (BUKAN toko makanan)
   - "kesehatan" : klinik, rumah sakit, apotek, kosmetik, salon, spa, gym, kecantikan
   - "pendidikan" : sekolah, kampus, kursus, bimbel, e-learning, training
   - "jasa" : konsultan, hukum, akuntan, fotografer, event organizer, jasa profesional
   - "pemerintah" : instansi pemerintah, NGO, organisasi sosial, layanan publik
   - "keuangan" : bank, fintech, koperasi, investasi, asuransi, pinjaman
   - "logistik" : pengiriman, kurir, ekspedisi, pergudangan, supply chain
   - "kreatif" : agency, studio desain, software house, web/app developer, media kreatif
   - "gaya_hidup" : fashion, hobi, travel, hiburan, komunitas, wellness lifestyle
   - "agrikultur" : pertanian, peternakan, perikanan, perkebunan
   - "otomotif" : bengkel, dealer mobil/motor, suku cadang, rental kendaraan

2. Field "keywords" — 5 sampai 15 keyword DOMAIN spesifik yang menggambarkan produk/jasa,
   dipisahkan koma. DILARANG pakai keyword generik IT seperti: online, website, platform,
   digital, sistem, aplikasi, e-commerce, fitur, customer, marketing, strategi.

Output HARUS JSON valid dengan bentuk:
{"sektor": "kuliner", "keywords": "kata1, kata2, kata3"}
atau bila tidak yakin:
{"sektor": null, "keywords": "kata1, kata2"}

Tidak ada teks lain di luar JSON.`

	featureRecommendSystemPrompt = `Kamu adalah konsultan IT dari software house Indonesia bernama A.N.I Tech.
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
	validSektor = map[string]struct{}{
		"kuliner":    {},
		"perdagang":  {},
		"kesehatan":  {},
		"pendidikan": {},
		"jasa":       {},
		"pemerintah": {},
		"keuangan":   {},
		"logistik":   {},
		"kreatif":    {},
		"gaya_hidup": {},
		"agrikultur": {},
		"otomotif":   {},
	}

	featureListLineRegex  = regexp.MustCompile(`^\s*(?:\d+[.)]|[-*])\s+(.+?)\s*$`)
	featurePlainLineRegex = regexp.MustCompile(`^\s*\*\*[^*]+?\*\*\s*[—:-]\s+.+$`)
	markdownFenceRegex    = regexp.MustCompile("(?s)^```(?:markdown|md)?\\s*(.*?)\\s*```$")
)

type GroqClassification struct {
	Sektor   string
	Keywords string
}

type GroqRepository interface {
	ClassifyAndExtract(input string) (GroqClassification, error)
	GenerateWebsiteFeatures(input string) (string, error)
}

type groqRepository struct {
	client *openai.Client
	apiKey string
}

func GroqModelName() string {
	return groqModel
}

func NewGroqRepository(apiKey string) GroqRepository {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = groqBaseURL

	return &groqRepository{
		client: openai.NewClientWithConfig(cfg),
		apiKey: apiKey,
	}
}

func (r *groqRepository) ClassifyAndExtract(input string) (GroqClassification, error) {
	if r.apiKey == "" {
		return GroqClassification{}, errors.New("GROQ_API_KEY is not configured")
	}
	startedAt := time.Now()
	log.Printf(
		"[recommend][groq] request_started model=%s base_url=%s input_chars=%d",
		groqModel, groqBaseURL, len([]rune(strings.TrimSpace(input))),
	)

	userPrompt := fmt.Sprintf(
		"Deskripsi bisnis klien: '%s'\nKlasifikasikan sektor dan ekstrak keyword domain. Output JSON.",
		input,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := r.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: groqModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: classifyPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: 0,
	})
	if err != nil {
		log.Printf(
			"[recommend][groq] request_failed model=%s duration_ms=%d err=%v",
			groqModel, time.Since(startedAt).Milliseconds(), err,
		)
		return GroqClassification{}, fmt.Errorf("groq request failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		log.Printf(
			"[recommend][groq] request_failed model=%s duration_ms=%d err=%q",
			groqModel, time.Since(startedAt).Milliseconds(), "groq returned no choices",
		)
		return GroqClassification{}, errors.New("groq returned no choices")
	}
	out, err := parseClassification(resp.Choices[0].Message.Content)
	if err != nil {
		log.Printf(
			"[recommend][groq] request_failed model=%s duration_ms=%d err=%v",
			groqModel, time.Since(startedAt).Milliseconds(), err,
		)
		return GroqClassification{}, err
	}
	log.Printf(
		"[recommend][groq] request_succeeded model=%s duration_ms=%d sektor=%q keywords=%q",
		groqModel, time.Since(startedAt).Milliseconds(), out.Sektor, out.Keywords,
	)
	return out, nil
}

func (r *groqRepository) GenerateWebsiteFeatures(input string) (string, error) {
	if r.apiKey == "" {
		return "", errors.New("GROQ_API_KEY is not configured")
	}
	startedAt := time.Now()
	log.Printf(
		"[recommend][groq_feature] request_started model=%s base_url=%s input_chars=%d",
		groqModel, groqBaseURL, len([]rune(strings.TrimSpace(input))),
	)

	userPrompt := fmt.Sprintf(
		"Klien saya ingin membuat website dengan deskripsi berikut:\n'%s'\nBerikan rekomendasi fitur website yang cocok untuk bisnis ini.",
		input,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	resp, err := r.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: groqModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: featureRecommendSystemPrompt,
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
			"[recommend][groq_feature] request_failed model=%s duration_ms=%d err=%v",
			groqModel, time.Since(startedAt).Milliseconds(), err,
		)
		return "", fmt.Errorf("groq request failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		log.Printf(
			"[recommend][groq_feature] request_failed model=%s duration_ms=%d err=%q",
			groqModel, time.Since(startedAt).Milliseconds(), "groq returned no choices",
		)
		return "", errors.New("groq returned no choices")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	normalized := normalizeFeatureMarkdown(raw)

	log.Printf(
		"[recommend][groq_feature] request_succeeded model=%s duration_ms=%d output_chars=%d",
		groqModel, time.Since(startedAt).Milliseconds(), len([]rune(normalized)),
	)
	return normalized, nil
}

func parseClassification(raw string) (GroqClassification, error) {
	var payload struct {
		Sektor   *string `json:"sektor"`
		Keywords string  `json:"keywords"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return GroqClassification{}, fmt.Errorf("failed to parse groq json: %w (raw=%s)", err, raw)
	}

	out := GroqClassification{Keywords: normalizeKeywords(payload.Keywords)}
	if payload.Sektor != nil {
		s := strings.ToLower(strings.TrimSpace(*payload.Sektor))
		if _, ok := validSektor[s]; ok {
			out.Sektor = s
		}
	}
	return out, nil
}

func normalizeKeywords(raw string) string {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return ""
	}
	cleaned = strings.ReplaceAll(cleaned, "\n", ",")
	cleaned = strings.ReplaceAll(cleaned, ";", ",")

	parts := strings.Split(cleaned, ",")
	seen := map[string]struct{}{}
	keywords := make([]string, 0, len(parts))
	for _, part := range parts {
		k := strings.TrimSpace(part)
		if k == "" {
			continue
		}
		key := strings.ToLower(k)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keywords = append(keywords, k)
	}
	if len(keywords) > 15 {
		keywords = keywords[:15]
	}
	return strings.Join(keywords, ", ")
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
