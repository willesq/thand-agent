package cloudflare

import (
	"context"
	"fmt"

	"github.com/thand-io/agent/internal/models"
)

// Cloudflare only supports roles when assigning accounts for access
// I am leaving this here for the future in case cloudflare adds support
// for more granular permissions in the future - aside from roles

// LoadPermissions loads Cloudflare permission reference data
func (p *cloudflareProvider) LoadPermissions() error {
	return nil
}

// GetPermission retrieves a specific permission by name
func (p *cloudflareProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	return nil, fmt.Errorf("permission not found")
}

// ListPermissions lists all permissions, optionally filtered
func (p *cloudflareProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {
	return []models.ProviderPermission{}, nil
}
