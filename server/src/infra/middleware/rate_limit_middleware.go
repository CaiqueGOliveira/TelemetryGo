package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	RPS   float64
	Burst int
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		return func(ctx *gin.Context) { ctx.Next() }
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*clientLimiter)
	)

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for key, client := range clients {
				if time.Since(client.lastSeen) > 10*time.Minute {
					delete(clients, key)
				}
			}
			mu.Unlock()
		}
	}()

	return func(ctx *gin.Context) {
		key := ctx.ClientIP()
		if userID := ctx.GetString("user_id"); userID != "" {
			key = "user:" + userID
		}

		mu.Lock()
		client, ok := clients[key]
		if !ok {
			client = &clientLimiter{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			clients[key] = client
		}
		client.lastSeen = time.Now()
		mu.Unlock()

		if !client.limiter.Allow() {
			ctx.Header("Retry-After", "1")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		ctx.Next()
	}
}
