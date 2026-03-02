package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedRedirect(t *testing.T) {
	handler := &OIDCHandler{
		allowedDomains: []string{"http://localhost", "https://app.example.com"},
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "allowed http localhost",
			url:      "http://localhost/auth/callback",
			expected: true,
		},
		{
			name:     "allowed http localhost with query",
			url:      "http://localhost/dashboard?foo=bar",
			expected: true,
		},
		{
			name:     "allowed https domain",
			url:      "https://app.example.com/auth/callback",
			expected: true,
		},
		{
			name:     "disallowed domain",
			url:      "http://evil.com/phishing",
			expected: false,
		},
		{
			name:     "disallowed subdomain",
			url:      "https://malicious.example.com/callback",
			expected: false,
		},
		{
			name:     "empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "relative path",
			url:      "/auth/callback",
			expected: false,
		},
		{
			name:     "missing scheme",
			url:      "localhost/auth/callback",
			expected: false,
		},
		{
			name:     "javascript protocol",
			url:      "javascript:alert(1)",
			expected: false,
		},
		{
			name:     "data protocol",
			url:      "data:text/html,<script>alert(1)</script>",
			expected: false,
		},
		{
			name:     "wrong port",
			url:      "http://localhost:8080/callback",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isAllowedRedirect(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAllowedRedirectWithDomainWithoutScheme(t *testing.T) {
	handler := &OIDCHandler{
		allowedDomains: []string{"example.com"},
	}

	// When domain is configured without scheme, it should default to https
	assert.True(t, handler.isAllowedRedirect("https://example.com/callback"))
	assert.False(t, handler.isAllowedRedirect("http://example.com/callback"))
}
