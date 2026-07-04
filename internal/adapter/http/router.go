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
	"github.com/henriquesevero/rinohabits-api/internal/adapter/storage"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/auth"
	usecasebook "github.com/henriquesevero/rinohabits-api/internal/usecase/book"
	usecasecourse "github.com/henriquesevero/rinohabits-api/internal/usecase/course"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/stats"
)

type Dependencies struct {
	Pool               *pgxpool.Pool
	TokenManager       security.JWTTokenManager
	CORSOrigin         string
	UploadsDir         string
	APIBaseURL         string
	SupabaseURL        string
	SupabaseServiceKey string
	VAPIDPrivateKey    string
	VAPIDPublicKey     string
	VAPIDEmail         string
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

	var fileStorage port.FileStorage
	if deps.SupabaseURL != "" && deps.SupabaseServiceKey != "" {
		fileStorage = storage.NewSupabaseStorage(deps.SupabaseURL, deps.SupabaseServiceKey)
	} else {
		os.MkdirAll(uploadsDir, 0755)
		fileStorage = storage.NewLocalStorage(uploadsDir, apiBaseURL)
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
		auth.NewChangeEmailUseCase(users, hasher),
		auth.NewChangePasswordUseCase(users, hasher),
		auth.NewDeleteAccountUseCase(users, hasher),
	)

	habitHandler := handler.NewHabitHandler(
		usecasehabit.NewCreateHabitUseCase(habits),
		usecasehabit.NewListTodayHabitsUseCase(habits, dailyLogs, systemClock, users),
		usecasehabit.NewToggleHabitLogUseCase(habits, dailyLogs, systemClock, users),
		usecasehabit.NewCalculateStreakUseCase(habits, dailyLogs, systemClock, users),
		usecasehabit.NewUpdateHabitUseCase(habits),
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
		fileStorage,
	)

	courses := postgres.NewCourseRepository(deps.Pool)
	courseLogs := postgres.NewCourseLogRepository(deps.Pool)

	courseHandler := handler.NewCourseHandler(
		usecasecourse.NewCreateCourseUseCase(courses),
		usecasecourse.NewListCoursesUseCase(courses),
		usecasecourse.NewUpdateCourseUseCase(courses, systemClock),
		usecasecourse.NewRegisterStudyUseCase(courses, courseLogs, users, systemClock),
		usecasecourse.NewDeleteCourseUseCase(courses),
		courses,
		fileStorage,
	)

	pushSubs := postgres.NewPushSubscriptionRepository(deps.Pool)
	notificationHandler := handler.NewNotificationHandler(pushSubs, deps.VAPIDPublicKey, deps.VAPIDPrivateKey, deps.VAPIDEmail)

	statsHandler := handler.NewStatsHandler(
		stats.NewGetPeriodOverviewUseCase(users, habits, dailyLogs, systemClock),
		stats.NewGetTrendUseCase(users, habits, dailyLogs, systemClock),
		stats.NewGetCalendarUseCase(users, habits, dailyLogs, systemClock),
		stats.NewGetDailyBreakdownUseCase(users, habits, dailyLogs, systemClock),
	)

	protected := middleware.Authenticate(deps.TokenManager)

	mux.HandleFunc("GET /health", healthHandler(deps.Pool))
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.Handle("GET /me", protected(http.HandlerFunc(authHandler.Me)))
	mux.Handle("PATCH /me/email", protected(http.HandlerFunc(authHandler.ChangeEmail)))
	mux.Handle("PATCH /me/password", protected(http.HandlerFunc(authHandler.ChangePassword)))
	mux.Handle("DELETE /me", protected(http.HandlerFunc(authHandler.DeleteAccount)))
	mux.Handle("POST /habits", protected(http.HandlerFunc(habitHandler.Create)))
	mux.Handle("GET /habits/today", protected(http.HandlerFunc(habitHandler.Today)))
	mux.Handle("POST /habits/{id}/toggle", protected(http.HandlerFunc(habitHandler.ToggleLog)))
	mux.Handle("PATCH /habits/{id}", protected(http.HandlerFunc(habitHandler.Update)))
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
	mux.Handle("POST /courses", protected(http.HandlerFunc(courseHandler.Create)))
	mux.Handle("GET /courses", protected(http.HandlerFunc(courseHandler.List)))
	mux.Handle("PATCH /courses/{id}", protected(http.HandlerFunc(courseHandler.Update)))
	mux.Handle("POST /courses/{id}/study", protected(http.HandlerFunc(courseHandler.RegisterStudy)))
	mux.Handle("POST /courses/{id}/cover", protected(http.HandlerFunc(courseHandler.UploadCover)))
	mux.Handle("DELETE /courses/{id}", protected(http.HandlerFunc(courseHandler.Delete)))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))
	mux.Handle("POST /notifications/subscribe", protected(http.HandlerFunc(notificationHandler.Subscribe)))
	mux.Handle("DELETE /notifications/subscribe", protected(http.HandlerFunc(notificationHandler.Unsubscribe)))
	mux.Handle("POST /notifications/test", protected(http.HandlerFunc(notificationHandler.TestNotify)))

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
