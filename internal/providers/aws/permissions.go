package aws

import (
	"context"
	"fmt"

	"github.com/thand-io/agent/internal/models"
)

func (p *awsProvider) SynchronizePermissions(
	ctx context.Context,
	req models.SynchronizePermissionsRequest,
) (*models.SynchronizePermissionsResponse, error) {
	data, err := getSharedData()
	if err != nil {
		return nil, fmt.Errorf("failed to get shared AWS data: %w", err)
	}

	return &models.SynchronizePermissionsResponse{
		Permissions: data.permissions,
	}, nil
}
