package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	tokens     map[string]*bucket
	mu         sync.RWMutex
	rate       int           // tokens per interval
	interval   time.Duration // time interval
	capacity   int           // bucket capacity
	cleanupInt time.Duration // cleanup interval
}

// bucket represents a token bucket for a single client
type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per interval
// interval: time window for rate limiting
// capacity: maximum burst capacity
func NewRateLimiter(rate int, interval time.Duration, capacity int) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(map[string]*bucket),
		rate:       rate,
		interval:   interval,
		capacity:   capacity,
		cleanupInt: time.Minute * 5,
	}

	// Start cleanup goroutine to remove old buckets
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use API key if available, otherwise use IP
		identifier := c.GetHeader("X-API-Key")
		if identifier == "" {
			identifier = c.ClientIP()
		}

		if !rl.allow(identifier) {
			c.JSON(429, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if a request should be allowed
func (rl *RateLimiter) allow(identifier string) bool {
	rl.mu.Lock()
	b, exists := rl.tokens[identifier]
	if !exists {
		b = &bucket{
			tokens:     rl.capacity,
			lastRefill: time.Now(),
		}
		rl.tokens[identifier] = b
	}
	rl.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed / rl.interval * time.Duration(rl.rate))

	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > rl.capacity {
			b.tokens = rl.capacity
		}
		b.lastRefill = now
	}

	// Check if we can allow the request
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes old buckets periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInt)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for id, b := range rl.tokens {
			b.mu.Lock()
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(b.lastRefill) > time.Minute*10 {
				delete(rl.tokens, id)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
