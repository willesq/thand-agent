package models

import (
	"context"
	"fmt"
)

type ProviderIdentities interface {
	GetIdentity(ctx context.Context, identity string) (*Identity, error)
	ListIdentities(ctx context.Context, filters ...string) ([]Identity, error)
	RefreshIdentities(ctx context.Context) error
}

/* Default implementations for identity resources */

func (p *BaseProvider) GetIdentity(ctx context.Context, identity string) (*Identity, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement GetIdentity", p.GetProvider())
}

func (p *BaseProvider) ListIdentities(ctx context.Context, filters ...string) ([]Identity, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement ListIdentities", p.GetProvider())
}

func (p *BaseProvider) RefreshIdentities(ctx context.Context) error {
	// Default implementation does nothing
	return fmt.Errorf("the provider '%s' does not implement RefreshIdentities", p.GetProvider())
}
