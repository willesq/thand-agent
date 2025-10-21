package thand

import (
	"errors"
	"fmt"
	"maps"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/config"
	"github.com/thand-io/agent/internal/models"
)

const ThandRevokeTask = "revoke"

// ThandRevokeTask represents a custom task for Thand revocation
func (t *thandTask) executeRevokeTask(
	workflowTask *models.WorkflowTask,
) (any, error) {

	req := workflowTask.GetContextAsMap()

	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	return RevokeAuthorization(t.config, workflowTask, req)

}

func RevokeAuthorization(
	config *config.Config,
	workflowTask *models.WorkflowTask,
	req map[string]any,
) (any, error) {

	// Right - we need to take the role, policy and provider and make the request to
	// the provider to elevate.

	var elevateRequest models.ElevateRequestInternal
	if err := common.ConvertMapToInterface(req, &elevateRequest); err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	if !elevateRequest.IsValid() {
		return nil, errors.New("invalid elevate request")
	}

	user := elevateRequest.User
	role := elevateRequest.Role
	providers := elevateRequest.Providers
	duration, err := elevateRequest.AsDuration()

	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	// TODO use the duration to revoke the request

	logrus.WithFields(logrus.Fields{
		"user":     user,
		"role":     role,
		"provider": providers,
		"duration": duration,
	}).Info("Executing authorization logic")

	// First lets call the provider to execute the role request.
	primaryProvider := elevateRequest.Providers[0]

	providerCall, err := config.GetProviderByName(primaryProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	modelOutput := map[string]any{
		"revoked": true,
	}

	var authorizeResponse *models.AuthorizeRoleResponse

	// See if we can hydrate the authorization response
	if authorizationsMap, ok := req["authorizations"].(map[string]any); ok {
		if identityMap, ok := authorizationsMap[user.GetIdentity()].(map[string]any); ok {
			authorizeResponse = &models.AuthorizeRoleResponse{}
			if err := common.ConvertMapToInterface(identityMap, authorizeResponse); err != nil {
				return nil, fmt.Errorf("failed to convert authorize response: %w", err)
			}
		}
	}

	revokeOut, err := providerCall.GetClient().RevokeRole(
		workflowTask.GetContext(), &models.RevokeRoleRequest{
			RoleRequest: &models.RoleRequest{
				User: user,
				Role: role,
			},
			AuthorizeRoleResponse: authorizeResponse,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to revoke user: %w", err)
	}

	// If the revoke returned any output, merge it into modelOutput
	if revokeOut != nil {
		maps.Copy(modelOutput, map[string]any{
			"revocations": map[string]any{
				user.GetIdentity(): revokeOut,
			},
		})
	}

	return &modelOutput, nil
}
