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
	VAPIDPrivateKey    string
	VAPIDPublicKey     string
	VAPIDEmail         string
	GoogleBooksAPIKey  string
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://rinohabits:rinohabits@localhost:5432/rinohabits?sslmode=disable"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		CORSOrigin:         getEnv("CORS_ORIGIN", "*"),
		UploadsDir:         getEnv("UPLOADS_DIR", "./uploads"),
		APIBaseURL:         getEnv("API_BASE_URL", "http://localhost:8090"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),
		VAPIDPrivateKey:    getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDPublicKey:     getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDEmail:         getEnv("VAPID_EMAIL", "contato@henriquesevero.com"),
		GoogleBooksAPIKey:  getEnv("GOOGLE_BOOKS_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
