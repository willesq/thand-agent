package models

import (
	"strings"
	"time"
)

// Only to be used server side. The Code is already encrypted when
// sent to the server from the client.
type AuthWrapper struct {
	Version  int    `json:"version,omitempty"`
	Callback string `json:"callback"`
	Client   string `json:"client"`
	Provider string `json:"provider"`
	Code     string `json:"code,omitempty"` // Optional code if coming from client/cli
}

func NewAuthWrapper(
	callback string,
	client string,
	provider string,
	code string,
) AuthWrapper {
	return AuthWrapper{
		Callback: callback,
		Client:   client,
		Provider: provider,
		Code:     code,
	}
}

// Only to be used agent/cleint side. The code is to provde
// what client request was made to create the session.
type CodeWrapper struct {
	Version     int       `json:"version,omitempty"`
	LoginServer string    `json:"code"`
	ExpiresAt   time.Time `json:"created_at"`
}

func NewCodeWrapper(loginServer string) CodeWrapper {
	return CodeWrapper{
		Version:     1,
		LoginServer: loginServer,
		ExpiresAt:   time.Now().UTC().Add(5 * time.Minute),
	}
}

// IsValid checks if the code wrapper is still valid
func (cw CodeWrapper) IsValid(loginEndpoint string) bool {

	if !strings.EqualFold(cw.LoginServer, loginEndpoint) {
		return false
	}

	return time.Now().UTC().Before(cw.ExpiresAt)
}
