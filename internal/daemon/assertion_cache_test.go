package daemon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAssertionCache(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()

	assert.NotNil(t, cache)
	assert.Equal(t, 5*time.Minute, cache.ttl)
	assert.NotNil(t, cache.cleanupTicker)
	assert.NotNil(t, cache.stopCleanup)
}

func TestNewAssertionCache_DefaultValues(t *testing.T) {
	cache := NewAssertionCache(0, 0)
	defer cache.Stop()

	assert.Equal(t, 5*time.Minute, cache.ttl)
}

func TestAssertionCache_CheckAndAdd_FirstTime(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()

	// First time adding assertion ID should succeed
	result := cache.CheckAndAdd("assertion-123")
	assert.True(t, result, "First CheckAndAdd should return true")
}

func TestAssertionCache_CheckAndAdd_Duplicate(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()

	// Add assertion ID first time
	result1 := cache.CheckAndAdd("assertion-123")
	assert.True(t, result1, "First CheckAndAdd should return true")

	// Attempt to add same assertion ID (replay attack)
	result2 := cache.CheckAndAdd("assertion-123")
	assert.False(t, result2, "Duplicate CheckAndAdd should return false (replay detected)")
}

func TestAssertionCache_CheckAndAdd_EmptyID(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()

	// Empty assertion ID should fail
	result := cache.CheckAndAdd("")
	assert.False(t, result, "Empty assertion ID should return false")
}

func TestAssertionCache_CheckAndAdd_DifferentIDs(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()

	// Different assertion IDs should all succeed
	assert.True(t, cache.CheckAndAdd("assertion-1"))
	assert.True(t, cache.CheckAndAdd("assertion-2"))
	assert.True(t, cache.CheckAndAdd("assertion-3"))

	// But duplicates should fail
	assert.False(t, cache.CheckAndAdd("assertion-1"))
	assert.False(t, cache.CheckAndAdd("assertion-2"))
}

func TestAssertionCache_Cleanup_ExpiredEntries(t *testing.T) {
	// Use short TTL and cleanup interval for testing
	cache := NewAssertionCache(100*time.Millisecond, 50*time.Millisecond)
	defer cache.Stop()

	// Add assertion
	cache.CheckAndAdd("assertion-123")
	assert.Equal(t, 1, cache.Size(), "Cache should have 1 entry")

	// Wait for expiry and cleanup
	time.Sleep(200 * time.Millisecond)

	// Entry should be cleaned up
	assert.Equal(t, 0, cache.Size(), "Cache should be empty after cleanup")
}

func TestAssertionCache_Cleanup_KeepsValidEntries(t *testing.T) {
	// Use longer TTL with short cleanup interval
	cache := NewAssertionCache(5*time.Second, 50*time.Millisecond)
	defer cache.Stop()

	// Add assertion
	cache.CheckAndAdd("assertion-123")
	assert.Equal(t, 1, cache.Size())

	// Wait for cleanup cycle but not expiry
	time.Sleep(200 * time.Millisecond)

	// Entry should still be present (not expired)
	assert.Equal(t, 1, cache.Size(), "Valid entry should not be cleaned up")
}

func TestAssertionCache_Size(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()

	// Initially empty
	assert.Equal(t, 0, cache.Size())

	// Add assertions
	cache.CheckAndAdd("assertion-1")
	assert.Equal(t, 1, cache.Size())

	cache.CheckAndAdd("assertion-2")
	assert.Equal(t, 2, cache.Size())

	cache.CheckAndAdd("assertion-3")
	assert.Equal(t, 3, cache.Size())

	// Duplicate doesn't increase size
	cache.CheckAndAdd("assertion-1")
	assert.Equal(t, 3, cache.Size())
}

func TestAssertionCache_Stop(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)

	// Stop should complete without hanging
	cache.Stop()

	// Calling Stop multiple times should not panic
	cache.Stop()
}

func TestAssertionCache_ConcurrentAccess(t *testing.T) {
	cache := NewAssertionCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()

	// Test concurrent access from multiple goroutines
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			assertionID := time.Now().Format("20060102150405.000000")
			cache.CheckAndAdd(assertionID)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// All unique assertions should be cached
	assert.GreaterOrEqual(t, cache.Size(), 1, "Cache should have at least one entry")
}

func TestAssertionCache_ReplayAfterExpiry(t *testing.T) {
	// Use very short TTL for testing
	cache := NewAssertionCache(50*time.Millisecond, 25*time.Millisecond)
	defer cache.Stop()

	// Add assertion
	result1 := cache.CheckAndAdd("assertion-123")
	assert.True(t, result1, "First add should succeed")

	// Immediate duplicate should fail
	result2 := cache.CheckAndAdd("assertion-123")
	assert.False(t, result2, "Immediate duplicate should fail")

	// Wait for expiry and cleanup
	time.Sleep(150 * time.Millisecond)

	// After expiry, same assertion ID can be added again
	// Note: In production, SAML assertion IDs should be globally unique
	// and should never be reused, but we test the cache behavior here
	result3 := cache.CheckAndAdd("assertion-123")
	assert.True(t, result3, "After expiry, assertion ID can be added again")
}
