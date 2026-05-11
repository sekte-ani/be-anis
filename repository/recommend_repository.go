package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"

	"be-anis/model"
)

type RecommendRepository interface {
	SearchMock(vector []float64, limit int, sektor string) ([]model.MockReference, error)
}

type recommendRepository struct {
	databaseURL string
}

func NewRecommendRepository(databaseURL string) RecommendRepository {
	return &recommendRepository{databaseURL: databaseURL}
}

func (r *recommendRepository) SearchMock(vector []float64, limit int, sektor string) ([]model.MockReference, error) {
	if r.databaseURL == "" {
		return nil, errors.New("DATABASE_URL is not configured")
	}
	if len(vector) == 0 {
		return nil, errors.New("empty embedding vector")
	}
	startedAt := time.Now()
	log.Printf(
		"[recommend][db] search_started vector_dims=%d limit=%d sektor=%q",
		len(vector), limit, sektor,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		log.Printf("[recommend][db] search_failed duration_ms=%d err=%v", time.Since(startedAt).Milliseconds(), err)
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	defer conn.Close(ctx)

	if err := pgxvector.RegisterTypes(ctx, conn); err != nil {
		log.Printf("[recommend][db] search_failed duration_ms=%d err=%v", time.Since(startedAt).Milliseconds(), err)
		return nil, fmt.Errorf("failed to register pgvector type: %w", err)
	}

	qVector := pgvector.NewVector(toFloat32(vector))

	if sektor != "" {
		refs, err := querySearch(ctx, conn, qVector, limit, sektor)
		if err != nil {
			log.Printf(
				"[recommend][db] search_failed duration_ms=%d sektor=%q err=%v",
				time.Since(startedAt).Milliseconds(), sektor, err,
			)
			return nil, err
		}
		if len(refs) > 0 {
			log.Printf(
				"[recommend][db] search_succeeded duration_ms=%d mode=sektor+vector sektor=%q result_count=%d",
				time.Since(startedAt).Milliseconds(), sektor, len(refs),
			)
			return refs, nil
		}
		// Sektor terdeteksi tapi DB tidak punya mock untuk sektor itu —
		// fallback ke vector-only search agar response tidak kosong.
		log.Printf(
			"[recommend][db] no_result_for_sektor duration_ms=%d sektor=%q fallback=vector_only",
			time.Since(startedAt).Milliseconds(), sektor,
		)
	}

	refs, err := querySearch(ctx, conn, qVector, limit, "")
	if err != nil {
		log.Printf(
			"[recommend][db] search_failed duration_ms=%d mode=vector_only err=%v",
			time.Since(startedAt).Milliseconds(), err,
		)
		return nil, err
	}
	log.Printf(
		"[recommend][db] search_succeeded duration_ms=%d mode=vector_only result_count=%d",
		time.Since(startedAt).Milliseconds(), len(refs),
	)
	return refs, nil
}

func querySearch(ctx context.Context, conn *pgx.Conn, qVector pgvector.Vector, limit int, sektor string) ([]model.MockReference, error) {
	const baseSelect = `SELECT mock_id, nama_mock, sektor, keywords, path_image,
	        1 - (embedding <=> $1::vector) AS similarity
	 FROM anis_mock
	 WHERE embedding IS NOT NULL`

	var (
		rows pgx.Rows
		err  error
	)
	if sektor == "" {
		rows, err = conn.Query(ctx,
			baseSelect+` ORDER BY embedding <=> $1::vector LIMIT $2`,
			qVector, limit,
		)
	} else {
		rows, err = conn.Query(ctx,
			baseSelect+` AND sektor = $3 ORDER BY embedding <=> $1::vector LIMIT $2`,
			qVector, limit, sektor,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("search_mock query failed: %w", err)
	}
	defer rows.Close()

	refs := make([]model.MockReference, 0, limit)
	for rows.Next() {
		var ref model.MockReference
		if err := rows.Scan(&ref.MockID, &ref.NamaMock, &ref.Sektor, &ref.Keywords, &ref.PathImage, &ref.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan search_mock row: %w", err)
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search_mock rows error: %w", err)
	}
	return refs, nil
}

func toFloat32(input []float64) []float32 {
	out := make([]float32, len(input))
	for i, v := range input {
		out[i] = float32(v)
	}
	return out
}
