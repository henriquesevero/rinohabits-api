package config

import "os"

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	CORSOrigin         string
	UploadsDir         string
	APIBaseURL         string
	SupabaseURL        string
	SupabaseServiceKey string
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://rinohabits:rinohabits@localhost:5432/rinohabits?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
		CORSOrigin:         getEnv("CORS_ORIGIN", "*"),
		UploadsDir:         getEnv("UPLOADS_DIR", "./uploads"),
		APIBaseURL:         getEnv("API_BASE_URL", "http://localhost:8090"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
