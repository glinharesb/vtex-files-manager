package auth

import (
	"net/http"
)

// Authenticator handles authentication for VTEX API requests using VTEX CLI token
type Authenticator struct {
	token string
}

// NewAuthenticator creates a new authenticator with VTEX CLI token
func NewAuthenticator(token string) *Authenticator {
	return &Authenticator{
		token: token,
	}
}

// AddAuthHeaders adds the authentication header to an HTTP request
// VTEX CLI token is the same as VtexIdclientAutCookie
func (a *Authenticator) AddAuthHeaders(req *http.Request) {
	req.Header.Set("VtexIdclientAutCookie", a.token)
}

// GetMethodName returns a human-readable name for the authentication method
func (a *Authenticator) GetMethodName() string {
	return "VTEX CLI Token"
}
