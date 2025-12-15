package daemon

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// AssertionCache implements in-memory cache for SAML assertion ID replay protection.
// It provides thread-safe tracking of used assertion IDs to prevent replay attacks
// where an attacker captures a valid SAML assertion and attempts to reuse it.
type AssertionCache struct {
	cache         sync.Map      // map[string]*assertionEntry
	ttl           time.Duration // Time-to-live for cached assertions
	cleanupTicker *time.Ticker  // Periodic cleanup ticker
	stopCleanup   chan struct{} // Channel to stop cleanup goroutine
}

// assertionEntry represents a cached assertion with its timing information
type assertionEntry struct {
	addedAt time.Time // When the assertion was first cached
	expiry  time.Time // When the assertion entry expires
}

// NewAssertionCache creates a new assertion cache with the specified TTL and cleanup interval.
// The TTL should match the typical validity window of SAML assertions (usually 5 minutes).
// The cleanup interval determines how often expired entries are removed from memory.
func NewAssertionCache(ttl time.Duration, cleanupInterval time.Duration) *AssertionCache {
	if ttl == 0 {
		ttl = 5 * time.Minute // Default TTL matches typical SAML assertion validity
	}
	if cleanupInterval == 0 {
		cleanupInterval = 1 * time.Minute // Default cleanup every minute
	}

	ac := &AssertionCache{
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	ac.cleanupTicker = time.NewTicker(cleanupInterval)
	go ac.cleanup()

	logrus.WithFields(logrus.Fields{
		"ttl":              ttl,
		"cleanup_interval": cleanupInterval,
	}).Info("Assertion cache initialized")

	return ac
}

// CheckAndAdd atomically checks if an assertion ID exists in the cache and adds it if not.
// This method is the core of replay protection - it ensures that each assertion ID can
// only be used once within the TTL window.
//
// Returns true if the assertion was added (not a replay), false if it already exists (replay detected).
func (ac *AssertionCache) CheckAndAdd(assertionID string) bool {
	if assertionID == "" {
		logrus.Warn("Empty assertion ID provided to cache")
		return false
	}

	now := time.Now()
	entry := &assertionEntry{
		addedAt: now,
		expiry:  now.Add(ac.ttl),
	}

	// LoadOrStore is atomic - it returns the existing value if present,
	// or stores the new value and returns it. The 'loaded' bool indicates
	// whether the value was loaded (true) or stored (false).
	_, loaded := ac.cache.LoadOrStore(assertionID, entry)

	if loaded {
		// Assertion ID already exists - this is a replay attack
		logrus.WithFields(logrus.Fields{
			"assertion_id": assertionID,
			"event":        "replay_detected",
		}).Warn("SAML assertion replay attack detected")
		return false
	}

	// Successfully cached new assertion ID
	logrus.WithFields(logrus.Fields{
		"assertion_id": assertionID,
		"expiry":       entry.expiry,
	}).Debug("SAML assertion ID cached successfully")

	return true
}

// cleanup removes expired assertion entries from the cache.
// This goroutine runs periodically based on the cleanup interval and prevents
// unbounded memory growth by removing entries that have exceeded their TTL.
func (ac *AssertionCache) cleanup() {
	for {
		select {
		case <-ac.cleanupTicker.C:
			now := time.Now()
			count := 0

			// Iterate through all cache entries
			ac.cache.Range(func(key, value interface{}) bool {
				entry := value.(*assertionEntry)
				if now.After(entry.expiry) {
					ac.cache.Delete(key)
					count++
				}
				return true // Continue iteration
			})

			if count > 0 {
				logrus.WithField("count", count).Debug("Cleaned up expired assertion cache entries")
			}

		case <-ac.stopCleanup:
			// Graceful shutdown requested
			ac.cleanupTicker.Stop()
			logrus.Info("Assertion cache cleanup goroutine stopped")
			return
		}
	}
}

// Stop gracefully stops the cleanup goroutine.
// This should be called when the server is shutting down to prevent goroutine leaks.
func (ac *AssertionCache) Stop() {
	close(ac.stopCleanup)
}

// Size returns the current number of cached assertions.
// This is useful for monitoring and observability to track cache utilization.
func (ac *AssertionCache) Size() int {
	count := 0
	ac.cache.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
