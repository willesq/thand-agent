package okta

import (
	"context"
	"fmt"
	"net/http"
)

// assignCustomRoleToUser assigns a custom role to a user via resource set assignments
func (p *oktaProvider) assignCustomRoleToUser(ctx context.Context, assignmentRequest ResourceSetAssignmentRequest) error {
	// Prepare the resource set assignment request

	// Use the request executor to make a direct API call
	// POST /api/internal/resourceSets/assignments
	reqExecutor := p.client.CloneRequestExecutor()

	req, err := reqExecutor.NewRequest(http.MethodPost, "/api/internal/resourceSets/assignments", assignmentRequest)
	if err != nil {
		return fmt.Errorf("failed to create assignment request: %w", err)
	}

	_, err = reqExecutor.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to assign custom role to user via resource sets: %w", err)
	}

	return nil
}
