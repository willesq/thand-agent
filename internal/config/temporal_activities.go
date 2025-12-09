package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/temporal"
)

type ThandActivities struct {
	Config *Config
}

type GetLocalConfigurationChunkResponse struct {
	Chunk      models.SystemChunk
	NextCursor *models.ConfigurationCursor
}

const MaxChunkSize = 1024 * 1024 // 1MB

func (a *ThandActivities) GetLocalConfigurationChunk(ctx context.Context, cursor *models.ConfigurationCursor) (*GetLocalConfigurationChunkResponse, error) {

	if a.Config == nil {
		return nil, fmt.Errorf("configuration is not initialized")
	}

	chunk := models.SystemChunk{
		Roles:        make(map[string]models.Role),
		Workflows:    make(map[string]models.Workflow),
		Providers:    make(map[string]models.Provider),
		ProviderData: make(map[string]models.ProviderData),
	}

	currentSize := 0

	// Parse cursor
	if cursor == nil {
		cursor = &models.ConfigurationCursor{Section: "roles", Offset: 0}
	} else if cursor.Section == "done" {
		return &GetLocalConfigurationChunkResponse{Chunk: chunk, NextCursor: nil}, nil
	}

	// Helper to check size and add
	// Returns true if we should stop (chunk full)
	shouldStop := func(key string, item interface{}) bool {
		bytes, _ := json.Marshal(item)
		itemSize := len(bytes) + len(key) + 10 // +10 for JSON overhead
		if currentSize+itemSize > MaxChunkSize {
			return true
		}
		currentSize += itemSize
		return false
	}

	// 1. Process Roles
	if cursor.Section == "roles" {
		// Sort keys for deterministic iteration
		keys := make([]string, 0, len(a.Config.Roles.Definitions))
		for k := range a.Config.Roles.Definitions {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i := cursor.Offset; i < len(keys); i++ {
			key := keys[i]
			item := a.Config.Roles.Definitions[key]
			if shouldStop(key, item) {
				return &GetLocalConfigurationChunkResponse{
					Chunk: chunk,
					NextCursor: &models.ConfigurationCursor{
						Section: "roles",
						Offset:  i,
					},
				}, nil
			}
			chunk.Roles[key] = item
		}
		// Finished roles, move to workflows
		cursor.Section = "workflows"
		cursor.Offset = 0
	}

	// 2. Process Workflows
	if cursor.Section == "workflows" {
		keys := make([]string, 0, len(a.Config.Workflows.Definitions))
		for k := range a.Config.Workflows.Definitions {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i := cursor.Offset; i < len(keys); i++ {
			key := keys[i]
			item := a.Config.Workflows.Definitions[key]
			if shouldStop(key, item) {
				return &GetLocalConfigurationChunkResponse{
					Chunk: chunk,
					NextCursor: &models.ConfigurationCursor{
						Section: "workflows",
						Offset:  i,
					},
				}, nil
			}
			chunk.Workflows[key] = item
		}
		// Finished workflows, move to providers
		cursor.Section = "providers"
		cursor.Offset = 0
	}

	// 3. Process Providers
	if cursor.Section == "providers" {
		keys := make([]string, 0, len(a.Config.Providers.Definitions))
		for k := range a.Config.Providers.Definitions {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i := cursor.Offset; i < len(keys); i++ {
			key := keys[i]
			item := a.Config.Providers.Definitions[key]
			if shouldStop(key, item) {
				return &GetLocalConfigurationChunkResponse{
					Chunk: chunk,
					NextCursor: &models.ConfigurationCursor{
						Section: "providers",
						Offset:  i,
					},
				}, nil
			}
			chunk.Providers[key] = item
		}
		// Finished providers
		cursor.Section = "done"
	}

	return &GetLocalConfigurationChunkResponse{
		Chunk:      chunk,
		NextCursor: nil, // Done
	}, nil
}

func (a *ThandActivities) SynchronizeThandStart(
	ctx context.Context,
	providerID string,
) (*models.SynchronizeStartResponse, error) {

	if a.Config == nil {
		return nil, fmt.Errorf("configuration is not initialized")
	}

	endpoint := a.Config.Thand.Endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("thand endpoint is not configured")
	}

	url := fmt.Sprintf("%s/sync/start", endpoint)

	ep := model.Endpoint{
		EndpointConfig: &model.EndpointConfiguration{
			URI: &model.LiteralUri{Value: url},
		},
	}

	args := &model.HTTPArguments{
		Method:   http.MethodPost,
		Endpoint: &ep,
		Headers:  make(map[string]string),
	}

	if a.Config.Thand.ApiKey != "" {
		args.Headers["Authorization"] = "Bearer " + a.Config.Thand.ApiKey
	}

	resp, err := common.InvokeHttpRequest(args)
	if err != nil {
		return nil, temporal.NewApplicationError("failed to invoke http request", "HttpRequestError", err)
	}

	if resp.StatusCode() != http.StatusAccepted && resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result models.SynchronizeStartResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	return &result, nil
}

func (a *ThandActivities) SynchronizeThandChunk(
	ctx context.Context,
	providerID string,
	workflowID string,
	chunk models.SystemChunk,
) error {

	if a.Config == nil {
		return fmt.Errorf("configuration is not initialized")
	}

	endpoint := a.Config.Thand.Endpoint
	if endpoint == "" {
		return fmt.Errorf("thand endpoint is not configured")
	}

	url := fmt.Sprintf("%s/sync/%s/chunk", endpoint, workflowID)

	body, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	ep := model.Endpoint{
		EndpointConfig: &model.EndpointConfiguration{
			URI: &model.LiteralUri{Value: url},
		},
	}

	args := &model.HTTPArguments{
		Method:   http.MethodPost,
		Endpoint: &ep,
		Body:     body,
		Headers:  make(map[string]string),
	}

	args.Headers["Content-Type"] = "application/json"
	if a.Config.Thand.ApiKey != "" {
		args.Headers["Authorization"] = "Bearer " + a.Config.Thand.ApiKey
	}

	resp, err := common.InvokeHttpRequest(args)
	if err != nil {
		return temporal.NewApplicationError("failed to invoke http request", "HttpRequestError", err)
	}

	if resp.StatusCode() != http.StatusAccepted && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func (a *ThandActivities) SynchronizeThandCommit(
	ctx context.Context,
	providerID string,
	workflowID string,
) error {

	if a.Config == nil {
		return fmt.Errorf("configuration is not initialized")
	}

	endpoint := a.Config.Thand.Endpoint
	if endpoint == "" {
		return fmt.Errorf("thand endpoint is not configured")
	}

	url := fmt.Sprintf("%s/sync/%s/commit", endpoint, workflowID)

	ep := model.Endpoint{
		EndpointConfig: &model.EndpointConfiguration{
			URI: &model.LiteralUri{Value: url},
		},
	}

	args := &model.HTTPArguments{
		Method:   http.MethodPost,
		Endpoint: &ep,
		Headers:  make(map[string]string),
	}

	if a.Config.Thand.ApiKey != "" {
		args.Headers["Authorization"] = "Bearer " + a.Config.Thand.ApiKey
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
