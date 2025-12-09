package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/temporal"
)

type ThandActivities struct {
	Config *Config
}

func (a *ThandActivities) GetLocalConfiguration(ctx context.Context) (*models.SystemChunk, error) {

	if a.Config == nil {
		return nil, fmt.Errorf("configuration is not initialized")
	}

	chunk := &models.SystemChunk{
		Roles:        a.Config.Roles.Definitions,
		Workflows:    a.Config.Workflows.Definitions,
		Providers:    a.Config.Providers.Definitions,
		ProviderData: make(map[string]models.ProviderData),
	}

	return chunk, nil
}

func (a *ThandActivities) SynchronizeThandStart(
	ctx context.Context,
	endpoint *model.Endpoint,
	providerID string,
) (*models.SynchronizeStartResponse, error) {

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
		return nil, temporal.NewApplicationError("failed to invoke http request", "HttpRequestError", err)
	}

	if resp.StatusCode() != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result models.SynchronizeStartResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *ThandActivities) SynchronizeThandChunk(
	ctx context.Context,
	endpoint *model.Endpoint,
	providerID string,
	workflowID string,
	chunk models.SynchronizeChunkRequest,
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
		return temporal.NewApplicationError("failed to invoke http request", "HttpRequestError", err)
	}

	if resp.StatusCode() != http.StatusAccepted {
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func (a *ThandActivities) SynchronizeThandCommit(
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
		return temporal.NewApplicationError("failed to invoke http request", "HttpRequestError", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}
