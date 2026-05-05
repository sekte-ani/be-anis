package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Vector []float64

func (v *Vector) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*v = nil
		return nil
	}

	if data[0] == '[' {
		var arr []float64
		if err := json.Unmarshal(data, &arr); err != nil {
			return fmt.Errorf("vector: invalid array: %w", err)
		}
		*v = arr
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("vector: expected array or pgvector string: %w", err)
	}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		*v = []float64{}
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]float64, len(parts))
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return fmt.Errorf("vector: parse element %d: %w", i, err)
		}
		out[i] = f
	}
	*v = out
	return nil
}

func (v Vector) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]float64(v))
}
