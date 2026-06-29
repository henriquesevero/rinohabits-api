package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/clock"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/handler"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/security"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/auth"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type Dependencies struct {
	Pool         *pgxpool.Pool
	TokenManager security.JWTTokenManager
	CORSOrigin   string
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

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
	)

	protected := middleware.Authenticate(deps.TokenManager)

	mux.HandleFunc("GET /health", healthHandler(deps.Pool))
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("GET /me", protected(http.HandlerFunc(authHandler.Me)))
	mux.Handle("POST /habits", protected(http.HandlerFunc(habitHandler.Create)))
	mux.Handle("GET /habits/today", protected(http.HandlerFunc(habitHandler.Today)))
	mux.Handle("POST /habits/{id}/toggle", protected(http.HandlerFunc(habitHandler.ToggleLog)))

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
