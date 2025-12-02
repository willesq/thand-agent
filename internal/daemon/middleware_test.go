package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/thand-io/agent/internal/models"
)

func TestMatchOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		pattern        string
		expectedResult bool
	}{
		{
			name:           "exact match",
			origin:         "https://app.thand.io",
			pattern:        "https://app.thand.io",
			expectedResult: true,
		},
		{
			name:           "wildcard subdomain match",
			origin:         "https://foo.app.thand.io",
			pattern:        "https://*.app.thand.io",
			expectedResult: true,
		},
		{
			name:           "wildcard subdomain match with different subdomain",
			origin:         "https://bar.app.thand.io",
			pattern:        "https://*.app.thand.io",
			expectedResult: true,
		},
		{
			name:           "wildcard subdomain match with numbers",
			origin:         "https://test123.app.thand.io",
			pattern:        "https://*.app.thand.io",
			expectedResult: true,
		},
		{
			name:           "wildcard subdomain match with hyphens",
			origin:         "https://my-test-app.app.thand.io",
			pattern:        "https://*.app.thand.io",
			expectedResult: true,
		},
		{
			name:           "no match - different domain",
			origin:         "https://foo.evil.com",
			pattern:        "https://*.app.thand.io",
			expectedResult: false,
		},
		{
			name:           "no match - different suffix",
			origin:         "https://foo.app.thand.com",
			pattern:        "https://*.app.thand.io",
			expectedResult: false,
		},
		{
			name:           "no match - missing subdomain",
			origin:         "https://app.thand.io",
			pattern:        "https://*.app.thand.io",
			expectedResult: false,
		},
		{
			name:           "wildcard with port",
			origin:         "https://foo.app.thand.io:8443",
			pattern:        "https://*.app.thand.io:8443",
			expectedResult: true,
		},
		{
			name:           "http scheme wildcard",
			origin:         "http://foo.app.thand.io",
			pattern:        "http://*.app.thand.io",
			expectedResult: true,
		},
		{
			name:           "scheme mismatch",
			origin:         "http://foo.app.thand.io",
			pattern:        "https://*.app.thand.io",
			expectedResult: false,
		},
		{
			name:           "empty origin",
			origin:         "",
			pattern:        "https://*.app.thand.io",
			expectedResult: false,
		},
		{
			name:           "nested subdomain - allowed by current implementation",
			origin:         "https://sub.foo.app.thand.io",
			pattern:        "https://*.app.thand.io",
			expectedResult: true,
		},
		{
			name:           "allow all wildcard",
			origin:         "https://any.domain.com",
			pattern:        "*",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchOrigin(tt.origin, tt.pattern)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                string
		origin              string
		method              string
		allowedOrigins      []string
		allowCredentials    bool
		expectedAllowOrigin string
		expectedStatus      int
		expectCORSHeaders   bool
	}{
		{
			name:                "valid origin - wildcard match",
			origin:              "https://test.app.thand.io",
			method:              "GET",
			allowedOrigins:      []string{"https://*.app.thand.io"},
			allowCredentials:    true,
			expectedAllowOrigin: "https://test.app.thand.io",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   true,
		},
		{
			name:                "valid origin - exact match",
			origin:              "https://localhost:8080",
			method:              "GET",
			allowedOrigins:      []string{"https://localhost:8080"},
			allowCredentials:    true,
			expectedAllowOrigin: "https://localhost:8080",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   true,
		},
		{
			name:                "invalid origin",
			origin:              "https://evil.com",
			method:              "GET",
			allowedOrigins:      []string{"https://*.app.thand.io"},
			allowCredentials:    true,
			expectedAllowOrigin: "",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   false,
		},
		{
			name:                "preflight request - valid origin",
			origin:              "https://test.app.thand.io",
			method:              "OPTIONS",
			allowedOrigins:      []string{"https://*.app.thand.io"},
			allowCredentials:    true,
			expectedAllowOrigin: "https://test.app.thand.io",
			expectedStatus:      http.StatusNoContent,
			expectCORSHeaders:   true,
		},
		{
			name:                "preflight request - invalid origin",
			origin:              "https://evil.com",
			method:              "OPTIONS",
			allowedOrigins:      []string{"https://*.app.thand.io"},
			allowCredentials:    true,
			expectedAllowOrigin: "",
			expectedStatus:      http.StatusForbidden,
			expectCORSHeaders:   false,
		},
		{
			name:                "no origin header",
			origin:              "",
			method:              "GET",
			allowedOrigins:      []string{"https://*.app.thand.io"},
			allowCredentials:    true,
			expectedAllowOrigin: "",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   false,
		},
		{
			name:                "multiple patterns - first match",
			origin:              "https://foo.app.thand.io",
			method:              "GET",
			allowedOrigins:      []string{"https://*.app.thand.io", "https://*.app.thand.com"},
			allowCredentials:    true,
			expectedAllowOrigin: "https://foo.app.thand.io",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   true,
		},
		{
			name:                "multiple patterns - second match",
			origin:              "https://foo.app.thand.com",
			method:              "GET",
			allowedOrigins:      []string{"https://*.app.thand.io", "https://*.app.thand.com"},
			allowCredentials:    true,
			expectedAllowOrigin: "https://foo.app.thand.com",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corsConfig := models.CORSConfig{
				AllowedOrigins:   tt.allowedOrigins,
				AllowCredentials: tt.allowCredentials,
			}

			router := gin.New()
			router.Use(CORSMiddleware(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})
			router.OPTIONS("/test", func(c *gin.Context) {
				// This shouldn't be reached for preflight - middleware handles it
				c.String(http.StatusOK, "OPTIONS")
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.method == "OPTIONS" {
				req.Header.Set("Access-Control-Request-Method", "POST")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectCORSHeaders {
				assert.Equal(t, tt.expectedAllowOrigin, w.Header().Get("Access-Control-Allow-Origin"))
				if tt.allowCredentials {
					assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				}
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}
