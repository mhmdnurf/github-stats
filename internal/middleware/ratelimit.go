package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	tokens       float64
	lastRefilled time.Time
}

type RateLimiter struct {
	mu            sync.Mutex
	visitors      map[string]*visitor
	ratePerSecond float64
	burst         float64
	idleTTL       time.Duration
	now           func() time.Time
}

func NewRateLimiter(ratePerSecond float64, burst int) *RateLimiter {
	if ratePerSecond <= 0 {
		ratePerSecond = 1
	}

	if burst <= 0 {
		burst = 1
	}

	return &RateLimiter{
		visitors:      make(map[string]*visitor),
		ratePerSecond: ratePerSecond,
		burst:         float64(burst),
		idleTTL:       10 * time.Minute,
		now:           time.Now,
	}
}

func (limiter *RateLimiter) allow(key string) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := limiter.now()

	limiter.evictStale(now)

	v, ok := limiter.visitors[key]
	if !ok {
		v = &visitor{
			tokens:       limiter.burst - 1,
			lastRefilled: now,
		}
		limiter.visitors[key] = v
		return true
	}

	elapsed := now.Sub(v.lastRefilled).Seconds()
	v.tokens += elapsed * limiter.ratePerSecond
	if v.tokens > limiter.burst {
		v.tokens = limiter.burst
	}
	v.lastRefilled = now

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

func (limiter *RateLimiter) evictStale(now time.Time) {
	for key, v := range limiter.visitors {
		if now.Sub(v.lastRefilled) > limiter.idleTTL {
			delete(limiter.visitors, key)
		}
	}
}

func (limiter *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		key := clientKey(request)

		if !limiter.allow(key) {
			writer.Header().Set("Retry-After", "1")
			writer.Header().Set("Cache-Control", "no-store")
			http.Error(
				writer,
				"rate limit exceeded",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func clientKey(request *http.Request) string {
	forwarded := strings.TrimSpace(
		request.Header.Get("X-Forwarded-For"),
	)
	if forwarded != "" {
		if commaIndex := strings.IndexByte(forwarded, ','); commaIndex != -1 {
			forwarded = forwarded[:commaIndex]
		}

		return strings.TrimSpace(forwarded)
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}

	return host
}
