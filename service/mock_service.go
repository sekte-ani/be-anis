package service

import (
	"fmt"
	"mime/multipart"
	"strings"

	"be-anis/model"
	"be-anis/repository"
)

type MockService interface {
	List(query model.ListMocksQuery) ([]model.Mock, *model.Paginator, error)
	Get(mockID string) (*model.Mock, error)
	Create(req model.CreateMockRequest) (*model.Mock, error)
	Update(mockID string, req model.UpdateMockRequest) (*model.Mock, error)
	Delete(mockID string) error
	GenerateKeywords(req model.GenerateKeywordsRequest) (model.GenerateKeywordsResponse, error)
	UploadImage(file *multipart.FileHeader) (string, error)
}

type mockService struct {
	mockRepo      repository.MockRepository
	openAIRepo    repository.OpenAIRepository
	embeddingRepo repository.EmbeddingRepository
	storageRepo   repository.StorageRepository
}

func NewMockService(
	mockRepo repository.MockRepository,
	openAIRepo repository.OpenAIRepository,
	embeddingRepo repository.EmbeddingRepository,
	storageRepo repository.StorageRepository,
) MockService {
	return &mockService{
		mockRepo:      mockRepo,
		openAIRepo:    openAIRepo,
		embeddingRepo: embeddingRepo,
		storageRepo:   storageRepo,
	}
}

func (s *mockService) List(query model.ListMocksQuery) ([]model.Mock, *model.Paginator, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	mocks, totalRecords, err := s.mockRepo.List(strings.TrimSpace(query.Sektor), strings.TrimSpace(query.Search), limit, offset)
	if err != nil {
		return nil, nil, err
	}

	paginator := model.NewPaginator(page, limit, totalRecords)
	return mocks, paginator, nil
}

func (s *mockService) Get(mockID string) (*model.Mock, error) {
	return s.mockRepo.GetByMockID(mockID)
}

func (s *mockService) Create(req model.CreateMockRequest) (*model.Mock, error) {
	vector, err := s.embeddingRepo.Embed(req.Keywords)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"mock_id":    req.MockID,
		"nama_mock":  req.NamaMock,
		"sektor":     req.Sektor,
		"keywords":   req.Keywords,
		"path_image": req.PathImage,
		"embedding":  vectorToString(vector),
	}

	return s.mockRepo.Create(payload)
}

func (s *mockService) Update(mockID string, req model.UpdateMockRequest) (*model.Mock, error) {
	payload := map[string]interface{}{}

	if req.MockID != "" {
		payload["mock_id"] = req.MockID
	}
	if req.NamaMock != "" {
		payload["nama_mock"] = req.NamaMock
	}
	if req.Sektor != "" {
		payload["sektor"] = req.Sektor
	}
	if req.PathImage != "" {
		payload["path_image"] = req.PathImage
	}
	if req.Keywords != "" {
		payload["keywords"] = req.Keywords
		vector, err := s.embeddingRepo.Embed(req.Keywords)
		if err != nil {
			return nil, err
		}
		payload["embedding"] = vectorToString(vector)
	}

	return s.mockRepo.Update(mockID, payload)
}

func (s *mockService) Delete(mockID string) error {
	return s.mockRepo.Delete(mockID)
}

// vectorToString converts []float64 to pgvector string format "[x,y,z]".
// PostgREST requires this format to cast into the vector column type.
func vectorToString(v []float64) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (s *mockService) UploadImage(file *multipart.FileHeader) (string, error) {
	return s.storageRepo.SaveImage(file)
}

func (s *mockService) GenerateKeywords(req model.GenerateKeywordsRequest) (model.GenerateKeywordsResponse, error) {
	keywords, err := s.openAIRepo.GenerateKeywordsFromImage(req.ImageURL)
	if err != nil {
		return model.GenerateKeywordsResponse{}, err
	}
	return model.GenerateKeywordsResponse{Keywords: keywords}, nil
}
