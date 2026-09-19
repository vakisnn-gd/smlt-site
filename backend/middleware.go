package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := false
		if origin != "" {
			for _, o := range strings.Split(allowedOrigin, ",") {
				if strings.TrimSpace(o) == origin {
					allowed = true
					break
				}
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// rateLimitMiddleware applies a per-IP limit inside the given bucket.
// Buckets are independent: auth attempts do not consume demonlist quota etc.
func rateLimitMiddleware(bucket string, maxAttempts int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := bucket + "|" + clientIP(r)
		if !checkRateLimit(key, maxAttempts, 15*time.Minute, 15*time.Minute) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "900")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Слишком много попыток. Подождите 15 минут.",
			})
			return
		}
		next(w, r)
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			if cookie, err := r.Cookie("smlt_session"); err == nil && cookie.Value != "" {
				auth = "Bearer " + cookie.Value
			}
		}
		if !strings.HasPrefix(auth, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Требуется авторизация"})
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if len(token) == 0 || !validateToken(token) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Неверный токен"})
			return
		}
		next(w, r)
	}
}

func bodyLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next(w, r)
	}
}

func methodCheck(allowedMethods string, next http.HandlerFunc) http.HandlerFunc {
	methods := make(map[string]bool)
	for _, m := range strings.Split(allowedMethods, ",") {
		methods[strings.TrimSpace(m)] = true
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if !methods[r.Method] {
			w.Header().Set("Allow", allowedMethods)
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Метод не поддерживается"})
			return
		}
		next(w, r)
	}
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self' https://api.demonlist.org; frame-src 'self' https://www.youtube.com https://www.youtube-nocookie.com; object-src 'none'; frame-ancestors 'none'; base-uri 'self'")
		next.ServeHTTP(w, r)
	})
}
