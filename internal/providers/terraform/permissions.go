package terraform

import (
	"context"
	"fmt"
	"strings"

	"github.com/thand-io/agent/internal/models"
)

func (p *terraformProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	for _, perm := range p.permissions {
		if strings.Compare(perm.Name, permission) == 0 {
			return &perm, nil
		}
	}
	return nil, fmt.Errorf("permission %s not found", permission)
}

func (p *terraformProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {
	return p.permissions, nil
}
