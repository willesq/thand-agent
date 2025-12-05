package aws

import (
	"context"
	"fmt"

	"github.com/thand-io/agent/internal/models"
)

func (p *awsProvider) SynchronizeRoles(ctx context.Context, req models.SynchronizeRolesRequest) (*models.SynchronizeRolesResponse, error) {
	data, err := getSharedData()
	if err != nil {
		return nil, fmt.Errorf("failed to get shared AWS data: %w", err)
	}

	return &models.SynchronizeRolesResponse{
		Roles: data.roles,
	}, nil
}
