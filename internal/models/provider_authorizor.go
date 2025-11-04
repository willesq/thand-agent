package models

import (
	"context"
	"fmt"
)

type ProviderAuthorizor interface {

	// Allow this provider to authorize a user
	AuthorizeSession(ctx context.Context, auth *AuthorizeUser) (*AuthorizeSessionResponse, error)
	CreateSession(ctx context.Context, auth *AuthorizeUser) (*Session, error)
	ValidateSession(ctx context.Context, session *Session) error
	RenewSession(ctx context.Context, session *Session) (*Session, error)
}

/* Default implementations for authorizers */

func (p *BaseProvider) AuthorizeSession(ctx context.Context, auth *AuthorizeUser) (*AuthorizeSessionResponse, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement AuthorizeSession", p.GetProvider())
}

func (p *BaseProvider) CreateSession(ctx context.Context, auth *AuthorizeUser) (*Session, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement CreateSession", p.GetProvider())
}

func (p *BaseProvider) ValidateSession(ctx context.Context, session *Session) error {
	// Default implementation does nothing
	return fmt.Errorf("the provider '%s' does not implement ValidateSession", p.GetProvider())
}

func (p *BaseProvider) RenewSession(ctx context.Context, session *Session) (*Session, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement RenewSession", p.GetProvider())
}
