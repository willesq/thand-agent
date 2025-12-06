package gcp

import (
	"context"
	"fmt"

	"github.com/thand-io/agent/internal/models"
)

func (p *gcpProvider) SynchronizeRoles(ctx context.Context, req models.SynchronizeRolesRequest) (*models.SynchronizeRolesResponse, error) {
	config := p.GetConfig()
	stage := config.GetStringWithDefault("stage", "GA")

	data, err := getSharedData(stage)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared GCP data: %w", err)
	}

	return &models.SynchronizeRolesResponse{
		Roles: data.roles,
	}, nil
}
