package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/supabase-community/postgrest-go"
	supabase "github.com/supabase-community/supabase-go"

	"be-anis/config"
	"be-anis/model"
)

const mockTable = "anis_mock"

type MockRepository interface {
	List(sektor, search string) ([]model.Mock, error)
	GetByMockID(mockID string) (*model.Mock, error)
	Create(payload map[string]interface{}) (*model.Mock, error)
	Update(mockID string, payload map[string]interface{}) (*model.Mock, error)
	Delete(mockID string) error
}

type mockRepository struct {
	clients *config.SupabaseClients
}

func NewMockRepository(clients *config.SupabaseClients) MockRepository {
	return &mockRepository{clients: clients}
}

func (r *mockRepository) client() *supabase.Client {
	if r.clients.Admin != nil {
		return r.clients.Admin
	}
	return r.clients.Public
}

func (r *mockRepository) List(sektor, search string) ([]model.Mock, error) {
	query := r.client().From(mockTable).Select("*", "exact", false)

	if sektor != "" {
		query = query.Eq("sektor", sektor)
	}

	if search != "" {
		pattern := fmt.Sprintf("%%%s%%", strings.ReplaceAll(search, ",", " "))
		filter := fmt.Sprintf("nama_mock.ilike.%s,keywords.ilike.%s", pattern, pattern)
		query = query.Or(filter, "")
	}

	query = query.Order("created_at", &postgrest.OrderOpts{Ascending: false})

	var mocks []model.Mock
	if _, err := query.ExecuteTo(&mocks); err != nil {
		return nil, err
	}
	return mocks, nil
}

func (r *mockRepository) GetByMockID(mockID string) (*model.Mock, error) {
	var mocks []model.Mock
	_, err := r.client().From(mockTable).
		Select("*", "exact", false).
		Eq("mock_id", mockID).
		Limit(1, "").
		ExecuteTo(&mocks)
	if err != nil {
		return nil, err
	}
	if len(mocks) == 0 {
		return nil, fmt.Errorf("mock with mock_id %q not found", mockID)
	}
	return &mocks[0], nil
}

func (r *mockRepository) Create(payload map[string]interface{}) (*model.Mock, error) {
	data, _, err := r.client().From(mockTable).
		Insert(payload, false, "", "representation", "exact").
		Execute()
	if err != nil {
		return nil, err
	}
	return decodeFirst(data)
}

func (r *mockRepository) Update(mockID string, payload map[string]interface{}) (*model.Mock, error) {
	data, _, err := r.client().From(mockTable).
		Update(payload, "representation", "exact").
		Eq("mock_id", mockID).
		Execute()
	if err != nil {
		return nil, err
	}
	return decodeFirst(data)
}

func (r *mockRepository) Delete(mockID string) error {
	_, _, err := r.client().From(mockTable).
		Delete("", "exact").
		Eq("mock_id", mockID).
		Execute()
	return err
}

func decodeFirst(data []byte) (*model.Mock, error) {
	var rows []model.Mock
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("failed to decode mock response: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no mock returned")
	}
	return &rows[0], nil
}
