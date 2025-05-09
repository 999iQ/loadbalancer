package middleware

import (
	"loadbalancer/internal/ratelimiter/bucket"
	"log"
	"net"
	"net/http"
	"strings"
)

// RateLimitMiddleware - возвращает новый http.Handler
func RateLimitMiddleware(bm *bucket.BucketManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := getIP(r)
		if err != nil {
			log.Printf("WARN: http.go - Internal Server Error: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !bm.Allow(ip) {
			log.Printf("WARN: http.go - IP: %s send too many requests\n", ip)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP - возвращает реальный IP клиента, если тот использует прокси
func getIP(r *http.Request) (string, error) {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0], nil
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}
