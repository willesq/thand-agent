package gcp

import (
	"context"
	"fmt"

	"github.com/thand-io/agent/internal/models"
)

func (p *gcpProvider) SynchronizePermissions(
	ctx context.Context,
	req models.SynchronizePermissionsRequest,
) (*models.SynchronizePermissionsResponse, error) {
	config := p.GetConfig()
	stage := config.GetStringWithDefault("stage", "GA")

	data, err := getSharedData(stage)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared GCP data: %w", err)
	}

	return &models.SynchronizePermissionsResponse{
		Permissions: data.permissions,
	}, nil
}
