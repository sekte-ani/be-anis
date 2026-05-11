package model

type RecommendRequest struct {
	Input string `json:"input" binding:"required"`
}

type MockReference struct {
	MockID     string  `json:"mock_id"`
	NamaMock   string  `json:"nama_mock"`
	Sektor     string  `json:"sektor"`
	Keywords   string  `json:"keywords"`
	PathImage  string  `json:"path_image"`
	Similarity float64 `json:"similarity"`
}

type RecommendResponse struct {
	RekomendasiFitur  string          `json:"rekomendasi_fitur"`
	SektorTerdeteksi  string          `json:"sektor_terdeteksi,omitempty"`
	MockReferences    []MockReference `json:"mock_references"`
}
