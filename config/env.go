package config

import (
	"bufio"
	"os"
	"strings"
)

type Env struct {
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
	OpenAIAPIKey        string
	EmbeddingServiceURL string
	AppBaseURL          string
	UploadDir           string
	Port                string
}

func LoadEnv() (*Env, error) {
	_ = loadDotEnv(".env")

	env := &Env{
		SupabaseURL:         os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:     os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceKey:  os.Getenv("SUPABASE_SERVICE_KEY"),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		EmbeddingServiceURL: getEnvOrDefault("EMBEDDING_SERVICE_URL", "http://localhost:8000"),
		AppBaseURL:          getEnvOrDefault("APP_BASE_URL", "https://anismockup.anitech.id"),
		UploadDir:           getEnvOrDefault("UPLOAD_DIR", "./uploads/img"),
		Port:                getEnvOrDefault("PORT", "5589"),
	}

	return env, nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func getEnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
