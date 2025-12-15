package daemon

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	// csrfCookieName is the session key used to store CSRF tokens
	csrfCookieName = "_thand_csrf"

	// csrfTokenLength is the length of the random bytes for CSRF tokens (32 bytes = 256 bits)
	csrfTokenLength = 32
)

// generateCSRFToken generates a cryptographically secure random CSRF token.
// The token is 32 bytes (256 bits) of randomness, base64-encoded for storage and transmission.
//
// Returns the base64-encoded token string, or an error if random generation fails.
func generateCSRFToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// setCSRFToken generates a new CSRF token and stores it in the session.
// This should be called when presenting a form or page that will submit back to the server,
// particularly for IdP-initiated SAML flows where we need CSRF protection.
//
// The token is stored in the session (server-side) rather than as a cookie or hidden field
// to prevent token leakage.
//
// Returns the generated token string, or an error if generation or storage fails.
func setCSRFToken(c *gin.Context) (string, error) {
	token, err := generateCSRFToken()
	if err != nil {
		return "", err
	}

	session := sessions.Default(c)
	session.Set(csrfCookieName, token)
	if err := session.Save(); err != nil {
		return "", err
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id": GetCorrelationID(c),
	}).Debug("CSRF token generated and stored in session")

	return token, nil
}

// validateAndClearCSRFToken validates a CSRF token against the stored session token.
// This implements single-use CSRF tokens - the token is cleared from the session
// regardless of validation result to prevent reuse.
//
// Parameters:
//   - c: The gin context containing the session
//   - token: The CSRF token to validate (from request)
//
// Returns true if the token is valid and matches the stored token, false otherwise.
func validateAndClearCSRFToken(c *gin.Context, token string) bool {
	session := sessions.Default(c)
	storedToken := session.Get(csrfCookieName)

	// Clear the token from session regardless of validation result (single-use)
	session.Delete(csrfCookieName)
	session.Save()

	// Validate token exists in session
	if storedToken == nil {
		logrus.WithFields(logrus.Fields{
			"correlation_id": GetCorrelationID(c),
		}).Warn("CSRF validation failed: no token found in session")
		return false
	}

	// Validate token type
	storedTokenStr, ok := storedToken.(string)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"correlation_id": GetCorrelationID(c),
		}).Warn("CSRF validation failed: token in session is not a string")
		return false
	}

	// Validate token value matches
	valid := storedTokenStr == token

	if !valid {
		logrus.WithFields(logrus.Fields{
			"correlation_id": GetCorrelationID(c),
		}).Warn("CSRF validation failed: token mismatch")
	} else {
		logrus.WithFields(logrus.Fields{
			"correlation_id": GetCorrelationID(c),
		}).Debug("CSRF token validated successfully")
	}

	return valid
}
