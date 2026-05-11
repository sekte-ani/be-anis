package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"be-anis/model"
	"be-anis/repository"
)

type RecommendService interface {
	Recommend(req model.RecommendRequest) (model.RecommendResponse, error)
}

type recommendService struct {
	recommendRepo repository.RecommendRepository
	groqRepo      repository.GroqRepository
	ollamaRepo    repository.OllamaRepository
	embeddingRepo repository.EmbeddingRepository
}

func NewRecommendService(
	recommendRepo repository.RecommendRepository,
	groqRepo repository.GroqRepository,
	ollamaRepo repository.OllamaRepository,
	embeddingRepo repository.EmbeddingRepository,
) RecommendService {
	return &recommendService{
		recommendRepo: recommendRepo,
		groqRepo:      groqRepo,
		ollamaRepo:    ollamaRepo,
		embeddingRepo: embeddingRepo,
	}
}

func (s *recommendService) Recommend(req model.RecommendRequest) (model.RecommendResponse, error) {
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return model.RecommendResponse{}, errors.New("input is required")
	}
	startedAt := time.Now()

	log.Printf(
		"[recommend][service] started input_chars=%d groq_model=%s ollama_model=%s",
		len([]rune(input)), repository.GroqModelName(), repository.OllamaModelName(),
	)

	var (
		wg          sync.WaitGroup
		errOnce     sync.Once
		firstErr    error
		mockRefs    []model.MockReference
		rekomendasi string
		sektor      string
	)

	setErr := func(err error) {
		if err == nil {
			return
		}
		errOnce.Do(func() {
			firstErr = err
		})
	}

	wg.Add(2)

	go func() {
		defer wg.Done()

		log.Printf("[recommend][service] classify_started")
		classification, err := s.groqRepo.ClassifyAndExtract(input)
		if err != nil {
			setErr(fmt.Errorf("failed to classify input: %w", err))
			return
		}

		keywords := classification.Keywords
		if strings.TrimSpace(keywords) == "" {
			keywords = input
		}
		sektor = classification.Sektor
		log.Printf("[recommend][service] classify_succeeded sektor=%q keywords=%q", sektor, keywords)

		log.Printf("[recommend][service] embedding_started")
		vector, err := s.embeddingRepo.Embed(keywords)
		if err != nil {
			setErr(fmt.Errorf("failed to get embedding: %w", err))
			return
		}
		log.Printf("[recommend][service] embedding_succeeded vector_dims=%d", len(vector))

		log.Printf("[recommend][service] search_mock_started limit=%d sektor=%q", 3, sektor)
		refs, err := s.recommendRepo.SearchMock(vector, 3, sektor)
		if err != nil {
			setErr(fmt.Errorf("failed to search similar mock: %w", err))
			return
		}
		mockRefs = refs
		log.Printf("[recommend][service] search_mock_succeeded mock_count=%d", len(mockRefs))
	}()

	go func() {
		defer wg.Done()
		log.Printf("[recommend][service] feature_generation_started")
		features, err := s.ollamaRepo.GenerateWebsiteFeatures(input)
		if err != nil {
			setErr(fmt.Errorf("failed to generate website features: %w", err))
			return
		}
		rekomendasi = features
		log.Printf("[recommend][service] feature_generation_succeeded fitur_chars=%d", len([]rune(rekomendasi)))
	}()

	wg.Wait()
	if firstErr != nil {
		log.Printf("[recommend][service] failed duration_ms=%d err=%v", time.Since(startedAt).Milliseconds(), firstErr)
		return model.RecommendResponse{}, firstErr
	}

	log.Printf(
		"[recommend][service] completed duration_ms=%d sektor=%q mock_count=%d fitur_chars=%d",
		time.Since(startedAt).Milliseconds(), sektor, len(mockRefs), len([]rune(rekomendasi)),
	)

	return model.RecommendResponse{
		RekomendasiFitur: rekomendasi,
		SektorTerdeteksi: sektor,
		MockReferences:   mockRefs,
	}, nil
}
