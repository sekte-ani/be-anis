package config

import (
	"fmt"

	supabase "github.com/supabase-community/supabase-go"
)

type SupabaseClients struct {
	Public *supabase.Client
	Admin  *supabase.Client
}

func NewSupabaseClients(env *Env) (*SupabaseClients, error) {
	if env.SupabaseURL == "" || env.SupabaseAnonKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY are required")
	}

	publicClient, err := supabase.NewClient(env.SupabaseURL, env.SupabaseAnonKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init supabase public client: %w", err)
	}

	var adminClient *supabase.Client
	if env.SupabaseServiceKey != "" {
		adminClient, err = supabase.NewClient(env.SupabaseURL, env.SupabaseServiceKey, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to init supabase admin client: %w", err)
		}
	}

	return &SupabaseClients{
		Public: publicClient,
		Admin:  adminClient,
	}, nil
}
