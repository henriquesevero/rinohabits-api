package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func RateLimit(requestsPerMinute int, burst int) func(http.Handler) http.Handler {
	limiters := newLimiterStore(requestsPerMinute, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiters.allow(clientIP(r)) {
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type limiterStore struct {
	mu                sync.Mutex
	limiters          map[string]*rate.Limiter
	requestsPerMinute int
	burst             int
}

func newLimiterStore(requestsPerMinute, burst int) *limiterStore {
	s := &limiterStore{
		limiters:          make(map[string]*rate.Limiter),
		requestsPerMinute: requestsPerMinute,
		burst:             burst,
	}
	go s.evictPeriodically()
	return s
}

func (s *limiterStore) allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.limiters[key]
	if !ok {
		l = rate.NewLimiter(rate.Every(time.Minute/time.Duration(s.requestsPerMinute)), s.burst)
		s.limiters[key] = l
	}
	return l.Allow()
}

// Bounds memory: per-IP limiters are cheap to recreate, so the whole map is
// dropped periodically instead of tracking last-seen time per entry.
func (s *limiterStore) evictPeriodically() {
	for range time.Tick(10 * time.Minute) {
		s.mu.Lock()
		s.limiters = make(map[string]*rate.Limiter)
		s.mu.Unlock()
	}
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[len(parts)-1])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
