package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"go.temporal.io/sdk/temporal"
)

// RegisterActivities registers provider-specific activities with the Temporal worker
func (b *BaseProvider) RegisterActivities(temporalClient TemporalImpl) error {
	return ErrNotImplemented
}

type ProviderActivities struct {
	provider ProviderImpl
}

func NewProviderActivities(provider ProviderImpl) *ProviderActivities {
	return &ProviderActivities{
		provider: provider,
	}
}

func (a *ProviderActivities) AuthorizeRole(
	ctx context.Context,
	req *AuthorizeRoleRequest,
) (*AuthorizeRoleResponse, error) {

	logrus.Infoln("Starting AuthorizeRole activity")
	return handleNotImplementedError(a.provider.AuthorizeRole(ctx, req))

}

func (a *ProviderActivities) RevokeRole(
	ctx context.Context,
	req *RevokeRoleRequest,
) (*RevokeRoleResponse, error) {

	logrus.Infoln("Starting RevokeRole activity")
	return handleNotImplementedError(a.provider.RevokeRole(ctx, req))

}

func (a *ProviderActivities) SynchronizeIdentities(
	ctx context.Context,
	req SynchronizeUsersRequest,
) (*SynchronizeUsersResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeIdentities activity")

	return handleNotImplementedError(a.provider.SynchronizeIdentities(ctx, req))
}

func (a *ProviderActivities) SynchronizeResources(
	ctx context.Context,
	req SynchronizeResourcesRequest,
) (*SynchronizeResourcesResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeResources activity")

	return handleNotImplementedError(a.provider.SynchronizeResources(ctx, req))
}

func (a *ProviderActivities) SynchronizeUsers(
	ctx context.Context,
	req SynchronizeUsersRequest,
) (*SynchronizeUsersResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeUsers activity")

	return handleNotImplementedError(a.provider.SynchronizeUsers(ctx, req))
}

func (a *ProviderActivities) SynchronizeGroups(
	ctx context.Context,
	req SynchronizeGroupsRequest,
) (*SynchronizeGroupsResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeGroups activity")

	return handleNotImplementedError(a.provider.SynchronizeGroups(ctx, req))
}

func (a *ProviderActivities) SynchronizePermissions(
	ctx context.Context,
	req SynchronizePermissionsRequest,
) (*SynchronizePermissionsResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizePermissions activity")

	return handleNotImplementedError(a.provider.SynchronizePermissions(ctx, req))
}

func (a *ProviderActivities) SynchronizeRoles(
	ctx context.Context,
	req SynchronizeRolesRequest,
) (*SynchronizeRolesResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeRoles activity")

	return handleNotImplementedError(a.provider.SynchronizeRoles(ctx, req))
}

func (a *ProviderActivities) SynchronizeThandStart(
	ctx context.Context,
	endpoint *model.Endpoint,
	providerID string,
) (*SynchronizeStartResponse, error) {

	if endpoint == nil || endpoint.EndpointConfig == nil || endpoint.EndpointConfig.URI == nil {
		return nil, fmt.Errorf("invalid endpoint configuration")
	}

	literalURI, ok := endpoint.EndpointConfig.URI.(*model.LiteralUri)
	if !ok {
		return nil, fmt.Errorf("endpoint URI must be a LiteralUri")
	}

	url := fmt.Sprintf("%s/providers/%s/sync/start", literalURI.Value, providerID)

	// Create a copy of the endpoint to avoid modifying the original if it's shared (though it shouldn't be across activity boundaries)
	ep := *endpoint
	epConfig := *endpoint.EndpointConfig
	epConfig.URI = &model.LiteralUri{Value: url}
	ep.EndpointConfig = &epConfig

	args := &model.HTTPArguments{
		Method:   http.MethodPost,
		Endpoint: &ep,
	}

	resp, err := common.InvokeHttpRequest(args)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result SynchronizeStartResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *ProviderActivities) SynchronizeThandChunk(
	ctx context.Context,
	endpoint *model.Endpoint,
	providerID string,
	workflowID string,
	chunk SynchronizeChunkRequest,
) error {

	if endpoint == nil || endpoint.EndpointConfig == nil || endpoint.EndpointConfig.URI == nil {
		return fmt.Errorf("invalid endpoint configuration")
	}

	literalURI, ok := endpoint.EndpointConfig.URI.(*model.LiteralUri)
	if !ok {
		return fmt.Errorf("endpoint URI must be a LiteralUri")
	}

	url := fmt.Sprintf("%s/providers/%s/sync/%s/chunk", literalURI.Value, providerID, workflowID)

	body, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	ep := *endpoint
	epConfig := *endpoint.EndpointConfig
	epConfig.URI = &model.LiteralUri{Value: url}
	ep.EndpointConfig = &epConfig

	args := &model.HTTPArguments{
		Method:   http.MethodPost,
		Endpoint: &ep,
		Body:     body,
	}

	resp, err := common.InvokeHttpRequest(args)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusAccepted {
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func (a *ProviderActivities) SynchronizeThandCommit(
	ctx context.Context,
	endpoint *model.Endpoint,
	providerID string,
	workflowID string,
) error {

	if endpoint == nil || endpoint.EndpointConfig == nil || endpoint.EndpointConfig.URI == nil {
		return fmt.Errorf("invalid endpoint configuration")
	}

	literalURI, ok := endpoint.EndpointConfig.URI.(*model.LiteralUri)
	if !ok {
		return fmt.Errorf("endpoint URI must be a LiteralUri")
	}

	url := fmt.Sprintf("%s/providers/%s/sync/%s/commit", literalURI.Value, providerID, workflowID)

	ep := *endpoint
	epConfig := *endpoint.EndpointConfig
	epConfig.URI = &model.LiteralUri{Value: url}
	ep.EndpointConfig = &epConfig

	args := &model.HTTPArguments{
		Method:   http.MethodPost,
		Endpoint: &ep,
	}

	resp, err := common.InvokeHttpRequest(args)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func handleNotImplementedError[T any](res T, err error) (T, error) {
	if err != nil {
		if errors.Is(err, ErrNotImplemented) {
			return res, temporal.NewNonRetryableApplicationError(
				"activity not implemented for this provider",
				"NotImplementedError",
				err,
			)
		}
	}
	return res, err
}
