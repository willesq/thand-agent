package daemon

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// correlationIDKey is the context key used to store the correlation ID
const correlationIDKey = "correlation_id"

// CorrelationMiddleware adds a unique correlation ID to each request for distributed tracing.
// The correlation ID can be used to trace requests across services and correlate log entries.
//
// Priority:
// 1. Uses existing X-Correlation-ID header if present (for distributed tracing)
// 2. Generates a new UUID if no correlation ID exists
//
// The correlation ID is:
// - Stored in the gin context for access by handlers
// - Added to the response header for client-side tracing
func CorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if correlation ID already exists in request header
		correlationID := c.GetHeader("X-Correlation-ID")

		// Generate new correlation ID if not present
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Store in context for handler access
		c.Set(correlationIDKey, correlationID)

		// Add to response header for tracing
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

// GetCorrelationID retrieves the correlation ID from the request context.
// Returns an empty string if no correlation ID is found.
//
// This is useful for handlers that need to include the correlation ID
// in logs or pass it to downstream services.
func GetCorrelationID(c *gin.Context) string {
	if id, exists := c.Get(correlationIDKey); exists {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return ""
}

// LogWithCorrelation creates a logrus entry with the correlation ID automatically included.
// This ensures all log entries for a request can be correlated and traced.
//
// Usage:
//
//	log := LogWithCorrelation(c)
//	log.Info("Processing SAML callback")
//	log.WithFields(logrus.Fields{...}).Warn("Security event detected")
//
// The correlation ID will automatically be included in all log entries created from this logger.
func LogWithCorrelation(c *gin.Context) *logrus.Entry {
	correlationID := GetCorrelationID(c)
	return logrus.WithField("correlation_id", correlationID)
}
