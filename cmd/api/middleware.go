package main

import (
	"net/http"
	"qotd/cmd/api/types"
	"strings"
	"sync"
	"time"

	"github.com/rs/cors"
)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens, refillRate float64) *types.RateLimiter {
	return &types.RateLimiter{
		Tokens:     maxTokens,
		MaxTokens:  maxTokens,
		RefillRate: refillRate,
		LastRefill: time.Now(),
	}
}

// Global rate limiters for different IPs
var (
	globalRateLimiters = make(map[string]*types.RateLimiter)
	rateLimiterMutex   = sync.RWMutex{}
)

// getRateLimiter gets or creates a rate limiter for an IP address
func getRateLimiter(ip string) *types.RateLimiter {
	rateLimiterMutex.RLock()
	limiter, exists := globalRateLimiters[ip]
	rateLimiterMutex.RUnlock()

	if !exists {
		rateLimiterMutex.Lock()
		// Double-check after acquiring write lock
		if limiter, exists = globalRateLimiters[ip]; !exists {
			// 10 requests per minute (refill rate: 10/60 = 0.167 tokens per second)
			limiter = NewRateLimiter(10, 10.0/60.0)
			globalRateLimiters[ip] = limiter
		}
		rateLimiterMutex.Unlock()
	}

	return limiter
}

// cleanupOldRateLimiters removes old rate limiters to prevent memory leaks
func cleanupOldRateLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rateLimiterMutex.Lock()
		for ip, limiter := range globalRateLimiters {
			limiter.Mutex.Lock()
			// Remove limiters that haven't been used for 10 minutes
			if time.Since(limiter.LastRefill) > 10*time.Minute {
				delete(globalRateLimiters, ip)
			}
			limiter.Mutex.Unlock()
		}
		rateLimiterMutex.Unlock()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// rateLimitMiddleware applies rate limiting to requests
func (c *serverConfig) rateLimitMiddleware(next http.Handler) http.Handler {
	// Start cleanup goroutine once
	go cleanupOldRateLimiters()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health check
		// if strings.Contains(r.URL.Path, "healthcheck") {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		clientIP := getClientIP(r)
		limiter := getRateLimiter(clientIP)

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			c.logger.Error("[RL] ip: ", clientIP, " Rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *serverConfig) middleware(next http.Handler) http.Handler {
	handler := c.RecoverPanic(next)

	// Apply rate limiting
	handler = c.rateLimitMiddleware(handler)

	// Configure CORS based on environment
	corsOptions := cors.Options{
		AllowedOrigins: c.getAllowedOrigins(),
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Origin",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
		MaxAge:           300, // 5 minutes
		Debug:            c.env == "development",
	}

	corsHandler := cors.New(corsOptions)
	handler = corsHandler.Handler(handler)
	return handler
}

// getAllowedOrigins returns the list of allowed origins based on environment
func (c *serverConfig) getAllowedOrigins() []string {
	// Get allowed origins from environment variable
	allowedOrigins := getEnvAsString("CORS_ALLOWED_ORIGINS", "")

	if allowedOrigins != "" {
		// Split comma-separated origins and trim whitespace
		origins := strings.Split(allowedOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		return origins
	}

	// Default origins based on environment
	if c.env == "development" {
		return []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:8080",
		}
	}

	// Production: require explicit configuration
	return []string{}
}
