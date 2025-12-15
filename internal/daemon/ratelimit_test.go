package daemon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	assert.NotNil(t, rl)
	assert.Equal(t, 5.0, rl.rate)
	assert.Equal(t, 10, rl.burst)
	assert.NotNil(t, rl.cleanupTicker)
	assert.NotNil(t, rl.stopCleanup)
}

func TestRateLimiter_AllowWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	// Should allow first 10 requests (up to burst capacity)
	for i := 0; i < 10; i++ {
		result := rl.Allow("192.168.1.1")
		assert.True(t, result, "Request %d should be allowed within burst", i+1)
	}
}

func TestRateLimiter_DenyExceedingBurst(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	// Exhaust burst capacity (10 requests)
	for i := 0; i < 10; i++ {
		rl.Allow("192.168.1.1")
	}

	// 11th request should be denied (burst exhausted, no time for refill)
	result := rl.Allow("192.168.1.1")
	assert.False(t, result, "Request exceeding burst should be denied")
}

func TestRateLimiter_RefillTokens(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	// Exhaust burst capacity
	for i := 0; i < 10; i++ {
		rl.Allow("192.168.1.1")
	}

	// 11th request should be denied
	assert.False(t, rl.Allow("192.168.1.1"))

	// Wait for token refill (200ms = 1 token at 5 tokens/second)
	time.Sleep(200 * time.Millisecond)

	// Should allow 1 more request after refill
	result := rl.Allow("192.168.1.1")
	assert.True(t, result, "Request should be allowed after token refill")

	// But the next one should be denied (only 1 token refilled)
	result = rl.Allow("192.168.1.1")
	assert.False(t, result, "Next request should be denied (only 1 token refilled)")
}

func TestRateLimiter_IndependentIPs(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	// Exhaust burst for IP1
	for i := 0; i < 10; i++ {
		rl.Allow("192.168.1.1")
	}

	// IP1 should be rate limited
	assert.False(t, rl.Allow("192.168.1.1"))

	// But IP2 should still be allowed (independent bucket)
	assert.True(t, rl.Allow("192.168.1.2"))
	assert.True(t, rl.Allow("192.168.1.3"))
}

func TestRateLimiter_TokenCapAtBurst(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	// Use 5 tokens
	for i := 0; i < 5; i++ {
		rl.Allow("192.168.1.1")
	}

	// Wait long enough to refill more than burst (1 second = 5 tokens)
	time.Sleep(1 * time.Second)

	// Should have 10 tokens now (5 remaining + 5 refilled), capped at burst
	// Use 10 tokens
	for i := 0; i < 10; i++ {
		result := rl.Allow("192.168.1.1")
		assert.True(t, result, "Request %d should be allowed", i+1)
	}

	// 11th should be denied (tokens capped at burst)
	result := rl.Allow("192.168.1.1")
	assert.False(t, result, "Token bucket should be capped at burst limit")
}

func TestRateLimiter_Size(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	// Initially no buckets
	assert.Equal(t, 0, rl.Size())

	// Create buckets for different IPs
	rl.Allow("192.168.1.1")
	assert.Equal(t, 1, rl.Size())

	rl.Allow("192.168.1.2")
	assert.Equal(t, 2, rl.Size())

	rl.Allow("192.168.1.3")
	assert.Equal(t, 3, rl.Size())

	// Using same IP doesn't create new bucket
	rl.Allow("192.168.1.1")
	assert.Equal(t, 3, rl.Size())
}

func TestRateLimiter_Cleanup(t *testing.T) {
	// Create rate limiter with short cleanup interval for testing
	rl := &RateLimiter{
		rate:        5.0,
		burst:       10,
		stopCleanup: make(chan struct{}),
	}
	rl.cleanupTicker = time.NewTicker(100 * time.Millisecond)
	go rl.cleanup()
	defer rl.Stop()

	// Create a bucket
	rl.Allow("192.168.1.1")
	assert.Equal(t, 1, rl.Size())

	// Manually set lastRefill to old time to trigger cleanup
	value, _ := rl.buckets.Load("192.168.1.1")
	b := value.(*bucket)
	b.mu.Lock()
	b.lastRefill = time.Now().Add(-15 * time.Minute)
	b.mu.Unlock()

	// Wait for cleanup cycle
	time.Sleep(200 * time.Millisecond)

	// Bucket should be cleaned up
	assert.Equal(t, 0, rl.Size(), "Stale bucket should be cleaned up")
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)

	// Stop should complete without hanging
	rl.Stop()

	// Calling Stop multiple times should not panic
	rl.Stop()
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100.0, 200) // High limits for concurrent test
	defer rl.Stop()

	// Test concurrent access from multiple goroutines
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			ip := "192.168.1.1"
			rl.Allow(ip)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should have 1 bucket (all same IP)
	assert.Equal(t, 1, rl.Size())
}

func TestRateLimiter_LowRateScenario(t *testing.T) {
	// Test with low rate (1 request per second)
	rl := NewRateLimiter(1.0, 2)
	defer rl.Stop()

	// Burst of 2 allowed
	assert.True(t, rl.Allow("192.168.1.1"))
	assert.True(t, rl.Allow("192.168.1.1"))

	// 3rd denied
	assert.False(t, rl.Allow("192.168.1.1"))

	// Wait 1 second for 1 token refill
	time.Sleep(1 * time.Second)

	// Should allow 1 more
	assert.True(t, rl.Allow("192.168.1.1"))

	// Next should be denied
	assert.False(t, rl.Allow("192.168.1.1"))
}

func TestRateLimiter_HighRateScenario(t *testing.T) {
	// Test with high rate (100 requests per second)
	rl := NewRateLimiter(100.0, 50)
	defer rl.Stop()

	// Burst of 50 allowed
	for i := 0; i < 50; i++ {
		assert.True(t, rl.Allow("192.168.1.1"))
	}

	// 51st denied
	assert.False(t, rl.Allow("192.168.1.1"))

	// Wait 10ms for ~1 token refill (100 tokens/sec = 10ms per token)
	time.Sleep(10 * time.Millisecond)

	// Should allow 1 more
	assert.True(t, rl.Allow("192.168.1.1"))
}

func TestRateLimiter_PartialTokenConsumption(t *testing.T) {
	rl := NewRateLimiter(5.0, 10)
	defer rl.Stop()

	// Use half the burst
	for i := 0; i < 5; i++ {
		rl.Allow("192.168.1.1")
	}

	// Should still allow more (5 tokens remaining)
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow("192.168.1.1"))
	}

	// Now burst exhausted
	assert.False(t, rl.Allow("192.168.1.1"))
}
