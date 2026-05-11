package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
)

var validSektor = map[string]struct{}{
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

type GroqClassification struct {
	Sektor   string
	Keywords string
}

type GroqRepository interface {
	ClassifyAndExtract(input string) (GroqClassification, error)
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
