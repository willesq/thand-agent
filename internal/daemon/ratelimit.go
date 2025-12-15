package daemon

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RateLimiter implements IP-based rate limiting using the token bucket algorithm.
// This prevents denial-of-service attacks on SAML authentication endpoints by
// limiting the number of requests per IP address per second.
type RateLimiter struct {
	buckets       sync.Map      // map[string]*bucket (IP address -> bucket)
	rate          float64       // Tokens per second (e.g., 5.0)
	burst         int           // Maximum burst capacity (e.g., 10)
	cleanupTicker *time.Ticker  // Periodic cleanup ticker
	stopCleanup   chan struct{} // Channel to stop cleanup goroutine
}

// bucket represents a token bucket for a single IP address.
// The token bucket algorithm allows bursts of traffic up to the burst limit,
// then refills tokens at the configured rate.
type bucket struct {
	mu         sync.Mutex // Protects tokens and lastRefill
	tokens     float64    // Current number of available tokens
	lastRefill time.Time  // Last time tokens were refilled
}

// NewRateLimiter creates a new IP-based rate limiter with the specified rate and burst.
//
// Parameters:
//   - rate: Number of tokens to add per second (e.g., 5.0 allows 5 requests/second)
//   - burst: Maximum number of tokens in the bucket (allows traffic bursts)
//
// Example: NewRateLimiter(5.0, 10) allows bursts of up to 10 requests, then 5 requests/second sustained.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		rate:        rate,
		burst:       burst,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine to remove stale buckets (every 5 minutes)
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)
	go rl.cleanup()

	logrus.WithFields(logrus.Fields{
		"rate":  rate,
		"burst": burst,
	}).Info("Rate limiter initialized")

	return rl
}

// Middleware returns a gin middleware function for rate limiting.
// This middleware should be applied to routes that need protection against
// excessive requests (e.g., SAML callback endpoints).
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.Allow(ip) {
			logrus.WithFields(logrus.Fields{
				"ip":     ip,
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			}).Warn("Rate limit exceeded")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		c.Next()
	}
}

// Allow checks if a request from the given IP address should be allowed based on rate limiting.
// It implements the token bucket algorithm:
// 1. Refill tokens based on time elapsed since last refill
// 2. Cap tokens at burst limit
// 3. Check if at least 1 token is available
// 4. If yes, consume 1 token and allow request
//
// Returns true if the request should be allowed, false if rate limit is exceeded.
func (rl *RateLimiter) Allow(ip string) bool {
	now := time.Now()

	// Get or create bucket for this IP
	value, _ := rl.buckets.LoadOrStore(ip, &bucket{
		tokens:     float64(rl.burst), // Start with full bucket
		lastRefill: now,
	})

	b := value.(*bucket)
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time elapsed
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * rl.rate

	// Cap at burst size (don't accumulate tokens beyond burst)
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}

	b.lastRefill = now

	// Check if we have at least 1 token available
	if b.tokens >= 1.0 {
		b.tokens -= 1.0 // Consume 1 token
		return true
	}

	// Rate limit exceeded
	return false
}

// cleanup removes stale bucket entries to prevent unbounded memory growth.
// Buckets that haven't been accessed for 10 minutes are removed.
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			cutoff := time.Now().Add(-10 * time.Minute)
			count := 0

			rl.buckets.Range(func(key, value interface{}) bool {
				b := value.(*bucket)
				b.mu.Lock()
				shouldDelete := b.lastRefill.Before(cutoff)
				b.mu.Unlock()

				if shouldDelete {
					rl.buckets.Delete(key)
					count++
				}
				return true // Continue iteration
			})

			if count > 0 {
				logrus.WithField("count", count).Debug("Cleaned up stale rate limiter buckets")
			}

		case <-rl.stopCleanup:
			// Graceful shutdown requested
			rl.cleanupTicker.Stop()
			logrus.Info("Rate limiter cleanup goroutine stopped")
			return
		}
	}
}

// Stop gracefully stops the cleanup goroutine.
// This should be called when the server is shutting down to prevent goroutine leaks.
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// Size returns the current number of tracked IP addresses (buckets).
// This is useful for monitoring and observability to understand how many unique
// IPs are being rate-limited.
func (rl *RateLimiter) Size() int {
	count := 0
	rl.buckets.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
