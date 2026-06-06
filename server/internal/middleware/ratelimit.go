package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter stores a rate limiter per client IP address.
type ipLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

// newIPLimiter creates a new ipLimiter with the given rate and burst.
func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// getLimiter returns the rate limiter for the given IP, creating one if needed.
func (l *ipLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[ip]
	l.mu.RUnlock()

	if exists {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock.
	if limiter, exists = l.limiters[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(l.rps, l.burst)
	l.limiters[ip] = limiter
	return limiter
}

// RateLimitMiddleware returns a Gin middleware that limits requests per client
// IP address using a token-bucket algorithm.
func RateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	il := newIPLimiter(rps, burst)

	return func(c *gin.Context) {
		limiter := il.getLimiter(c.ClientIP())

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			return
		}

		c.Next()
	}
}
