package httpapi

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/clock"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/handler"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/security"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/auth"
	usecasebook "github.com/henriquesevero/rinohabits-api/internal/usecase/book"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/stats"
)

type Dependencies struct {
	Pool         *pgxpool.Pool
	TokenManager security.JWTTokenManager
	CORSOrigin   string
	UploadsDir   string
	APIBaseURL   string
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	uploadsDir := deps.UploadsDir
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	apiBaseURL := deps.APIBaseURL
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8090"
	}

	users := postgres.NewUserRepository(deps.Pool)
	habits := postgres.NewHabitRepository(deps.Pool)
	dailyLogs := postgres.NewDailyLogRepository(deps.Pool)
	hasher := security.NewBcryptHasher()
	systemClock := clock.NewSystemClock()

	authHandler := handler.NewAuthHandler(
		auth.NewRegisterUseCase(users, hasher),
		auth.NewLoginUseCase(users, hasher, deps.TokenManager),
		auth.NewGetCurrentUserUseCase(users),
	)

	habitHandler := handler.NewHabitHandler(
		usecasehabit.NewCreateHabitUseCase(habits),
		usecasehabit.NewListTodayHabitsUseCase(habits, dailyLogs, systemClock, users),
		usecasehabit.NewToggleHabitLogUseCase(habits, dailyLogs, systemClock, users),
		usecasehabit.NewCalculateStreakUseCase(habits, dailyLogs, systemClock, users),
		usecasehabit.NewDeleteHabitUseCase(habits),
	)

	books := postgres.NewBookRepository(deps.Pool)
	readingLogs := postgres.NewReadingLogRepository(deps.Pool)

	bookHandler := handler.NewBookHandler(
		usecasebook.NewCreateBookUseCase(books),
		usecasebook.NewListBooksUseCase(books),
		usecasebook.NewUpdateBookUseCase(books, systemClock),
		usecasebook.NewRegisterReadingUseCase(books, readingLogs, users, systemClock),
		usecasebook.NewDeleteBookUseCase(books),
		stats.NewGetReadingStatsUseCase(users, readingLogs, systemClock),
		books,
		uploadsDir,
		apiBaseURL,
	)

	statsHandler := handler.NewStatsHandler(
		stats.NewGetPeriodOverviewUseCase(users, habits, dailyLogs, systemClock),
		stats.NewGetTrendUseCase(users, habits, dailyLogs, systemClock),
		stats.NewGetCalendarUseCase(users, habits, dailyLogs, systemClock),
		stats.NewGetDailyBreakdownUseCase(users, habits, dailyLogs, systemClock),
	)

	protected := middleware.Authenticate(deps.TokenManager)

	os.MkdirAll(uploadsDir, 0755)

	mux.HandleFunc("GET /health", healthHandler(deps.Pool))
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("GET /me", protected(http.HandlerFunc(authHandler.Me)))
	mux.Handle("POST /habits", protected(http.HandlerFunc(habitHandler.Create)))
	mux.Handle("GET /habits/today", protected(http.HandlerFunc(habitHandler.Today)))
	mux.Handle("POST /habits/{id}/toggle", protected(http.HandlerFunc(habitHandler.ToggleLog)))
	mux.Handle("DELETE /habits/{id}", protected(http.HandlerFunc(habitHandler.Delete)))
	mux.Handle("GET /stats/overview", protected(http.HandlerFunc(statsHandler.Overview)))
	mux.Handle("GET /stats/trend", protected(http.HandlerFunc(statsHandler.Trend)))
	mux.Handle("GET /stats/calendar", protected(http.HandlerFunc(statsHandler.Calendar)))
	mux.Handle("GET /stats/daily", protected(http.HandlerFunc(statsHandler.DailyBreakdown)))
	mux.Handle("POST /books", protected(http.HandlerFunc(bookHandler.Create)))
	mux.Handle("GET /books", protected(http.HandlerFunc(bookHandler.List)))
	mux.Handle("PATCH /books/{id}", protected(http.HandlerFunc(bookHandler.Update)))
	mux.Handle("POST /books/{id}/reading", protected(http.HandlerFunc(bookHandler.RegisterReading)))
	mux.Handle("POST /books/{id}/cover", protected(http.HandlerFunc(bookHandler.UploadCover)))
	mux.Handle("DELETE /books/{id}", protected(http.HandlerFunc(bookHandler.Delete)))
	mux.Handle("GET /books/reading-stats", protected(http.HandlerFunc(bookHandler.ReadingStats)))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	return middleware.CORS(deps.CORSOrigin)(mux)
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
