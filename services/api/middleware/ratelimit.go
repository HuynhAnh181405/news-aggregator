package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Client represents a client for rate limiting purposes.
type Client struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client)
)

// RateLimitMiddleware provides rate limiting based on client IP.
// rps: requests per second allowed
// burst: maximum burst of requests allowed
func RateLimitMiddleware(rps float64, burst int) func(next http.Handler) http.Handler {
	cleanUpClients() // Start cleanup goroutine

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			mu.Lock()
			client, exists := clients[ip]
			if !exists {
				client = &Client{Limiter: rate.NewLimiter(rate.Limit(rps), burst)}
				clients[ip] = client
			}
			client.LastSeen = time.Now()
			mu.Unlock()

			if !client.Limiter.Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				log.Printf("Rate limit exceeded for IP: %s", ip)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client's IP address.
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded // In a production environment behind a load balancer
	}
	return r.RemoteAddr
}

// cleanUpClients periodically removes old client entries to prevent memory leaks.
func cleanUpClients() {
	go func() {
		for {
			time.Sleep(time.Minute) // Clean up every minute
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.LastSeen) > 5*time.Minute { // Remove clients inactive for 5 mins
					delete(clients, ip)
					log.Printf("Cleaned up inactive client: %s", ip)
				}
			}
			mu.Unlock()
		}
	}()
}
