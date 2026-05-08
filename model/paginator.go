package model

// Paginator holds pagination metadata returned in list responses.
type Paginator struct {
	CurrentPage  int   `json:"current_page"`
	Limit        int   `json:"limit"`
	BackPage     int   `json:"back_page"`
	NextPage     int   `json:"next_page"`
	TotalRecords int64 `json:"total_records"`
	TotalPages   int   `json:"total_pages"`
}

// NewPaginator builds a Paginator from the current page, limit, and total record count.
func NewPaginator(page, limit int, totalRecords int64) *Paginator {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	backPage := page - 1
	if backPage < 1 {
		backPage = 1
	}

	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}

	return &Paginator{
		CurrentPage:  page,
		Limit:        limit,
		BackPage:     backPage,
		NextPage:     nextPage,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
	}
}
