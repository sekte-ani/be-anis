package model

import "time"

type Mock struct {
	ID        int64     `json:"id"`
	MockID    string    `json:"mock_id"`
	NamaMock  string    `json:"nama_mock"`
	Sektor    string    `json:"sektor"`
	Keywords  string    `json:"keywords"`
	PathImage string    `json:"path_image"`
	Embedding []float64 `json:"embedding,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ListMocksQuery struct {
	Sektor string `form:"sektor"`
	Search string `form:"search"`
}

type CreateMockRequest struct {
	MockID    string `json:"mock_id" binding:"required"`
	NamaMock  string `json:"nama_mock" binding:"required"`
	Sektor    string `json:"sektor" binding:"required"`
	Keywords  string `json:"keywords" binding:"required"`
	PathImage string `json:"path_image" binding:"required"`
}

type UpdateMockRequest struct {
	MockID    string `json:"mock_id,omitempty"`
	NamaMock  string `json:"nama_mock,omitempty"`
	Sektor    string `json:"sektor,omitempty"`
	Keywords  string `json:"keywords,omitempty"`
	PathImage string `json:"path_image,omitempty"`
}

type GenerateKeywordsRequest struct {
	ImageURL string `json:"image_url" binding:"required,url"`
}

type GenerateKeywordsResponse struct {
	Keywords string `json:"keywords"`
}
