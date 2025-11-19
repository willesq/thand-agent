package terraform

import (
	"context"
	"fmt"

	"github.com/thand-io/agent/internal/models"
)

func (p *terraformProvider) GetRole(ctx context.Context, role string) (*models.ProviderRole, error) {
	// TODO: Implement Terraform GetRole logic
	return nil, fmt.Errorf("terraform has no concept of roles")
}

func (p *terraformProvider) ListRoles(ctx context.Context, filters ...string) ([]models.ProviderRole, error) {
	// TODO: Implement Terraform ListRoles logic
	return nil, fmt.Errorf("terraform has no concept of roles")
}
