package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// SyncWithThand syncs the config with thand.io hosted services
func (c *Config) SynchronizeWithThand() error {

	if len(c.Thand.Endpoint) == 0 {
		return fmt.Errorf("no thand.io endpoint configured")
	}

	// 1. Loop through Providers and their capabilities to see what we can sync
	providers := c.GetProvidersByCapability(
		models.ProviderCapabilityIdentities,
		models.ProviderCapabilityRBAC,
	)

	for providerIdentifier, provider := range providers {

		fmt.Printf("Synchronizing provider: %s (%s)\n", provider.Name, providerIdentifier)

		if provider.GetClient() == nil {
			log.Println("Provider client is nil, skipping:", provider.Name)
			continue
		}

		c.SynchronizeProviderWithThand(provider.GetClient())
	}

	return nil
}

func (c *Config) SynchronizeProviderWithThand(provider models.ProviderImpl) error {

	// 2. Start Sync
	fmt.Println("Starting sync...")
	startResp, err := c.startSync(provider)
	if err != nil {
		logrus.WithError(err).Errorln("Failed to start sync")
		return fmt.Errorf("failed to start sync: %w", err)
	}

	fmt.Printf("Sync started. WorkflowID: %s, RunID: %s\n", startResp.WorkflowID, startResp.RunID)

	// 3. Prepare Data Chunks
	ctx := context.Background()

	identities, err := provider.ListIdentities(ctx)
	if err != nil {
		log.Printf("Failed to list identities: %v", err)
	}
	permissions, err := provider.ListPermissions(ctx)
	if err != nil {
		log.Printf("Failed to list permissions: %v", err)
	}
	roles, err := provider.ListRoles(ctx)
	if err != nil {
		log.Printf("Failed to list roles: %v", err)
	}
	resources, err := provider.ListResources(ctx)
	if err != nil {
		log.Printf("Failed to list resources: %v", err)
	}

	// 3. Send Chunks
	fullData := models.SynchronizeChunkRequest{
		Identities:  identities,
		Permissions: permissions,
		Roles:       roles,
		Resources:   resources,
	}

	if err := c.sendUnifiedChunks(provider.GetIdentifier(), startResp.WorkflowID, fullData); err != nil {
		logrus.WithError(err).Errorln("Failed to send data chunks")
		return fmt.Errorf("failed to send data chunks: %w", err)
	}

	// 4. Commit Sync
	// This signals that all chunks have been sent. The workflow will then prune missing records.
	fmt.Println("Committing sync...")
	if err := c.commitSync(provider.GetIdentifier(), startResp.WorkflowID); err != nil {
		logrus.WithError(err).Errorln("Failed to commit sync")
		return err
	}
	fmt.Println("Sync committed successfully")
	return nil
}

func (c *Config) sendUnifiedChunks(providerID, workflowID string, fullData models.SynchronizeChunkRequest) error {
	chunkSize := 100

	maxLen := 0
	if l := len(fullData.Identities); l > maxLen {
		maxLen = l
	}
	if l := len(fullData.Permissions); l > maxLen {
		maxLen = l
	}
	if l := len(fullData.Roles); l > maxLen {
		maxLen = l
	}
	if l := len(fullData.Resources); l > maxLen {
		maxLen = l
	}

	for i := 0; i < maxLen; i += chunkSize {
		chunk := models.SynchronizeChunkRequest{}
		hasData := false

		if i < len(fullData.Identities) {
			end := min(i+chunkSize, len(fullData.Identities))
			chunk.Identities = fullData.Identities[i:end]
			hasData = true
		}

		if i < len(fullData.Permissions) {
			end := min(i+chunkSize, len(fullData.Permissions))
			chunk.Permissions = fullData.Permissions[i:end]
			hasData = true
		}

		if i < len(fullData.Roles) {
			end := min(i+chunkSize, len(fullData.Roles))
			chunk.Roles = fullData.Roles[i:end]
			hasData = true
		}

		if i < len(fullData.Resources) {
			end := min(i+chunkSize, len(fullData.Resources))
			chunk.Resources = fullData.Resources[i:end]
			hasData = true
		}

		if hasData {
			fmt.Printf("Sending chunk %d...\n", i/chunkSize+1)
			if err := c.sendChunk(providerID, workflowID, chunk); err != nil {
				return fmt.Errorf("failed to send chunk: %w", err)
			}
		}
	}
	return nil
}

func (c *Config) startSync(provider models.ProviderImpl) (*models.SynchronizeStartResponse, error) {

	if len(c.Thand.Endpoint) == 0 {
		return nil, fmt.Errorf("no thand.io endpoint configured")
	}

	if len(c.Thand.ApiKey) == 0 {
		return nil, fmt.Errorf("no API key configured for thand.io")
	}

	url := fmt.Sprintf("%s/providers/%s/sync/start", c.Thand.Endpoint, provider.GetIdentifier())

	args := &model.HTTPArguments{
		Method: http.MethodPost,
		Endpoint: &model.Endpoint{
			EndpointConfig: &model.EndpointConfiguration{
				URI: &model.LiteralUri{Value: url},
			},
		},
		Headers: map[string]string{
			"Authorization": "Bearer " + c.Thand.ApiKey,
		},
	}

	resp, err := common.InvokeHttpRequest(args)
	if err != nil {
		return nil, err
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

func (c *Config) sendChunk(providerIdentifier, workflowID string, chunk interface{}) error {

	if len(c.Thand.Endpoint) == 0 {
		return fmt.Errorf("no thand.io endpoint configured")
	}

	if len(c.Thand.ApiKey) == 0 {
		return fmt.Errorf("no API key configured for thand.io")
	}

	url := fmt.Sprintf("%s/providers/%s/sync/%s/chunk", c.Thand.Endpoint, providerIdentifier, workflowID)

	chunkBytes, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %v", err)
	}

	args := &model.HTTPArguments{
		Method: http.MethodPost,
		Endpoint: &model.Endpoint{
			EndpointConfig: &model.EndpointConfiguration{
				URI: &model.LiteralUri{Value: url},
			},
		},
		Headers: map[string]string{
			"Authorization": "Bearer " + c.Thand.ApiKey,
		},
		Body: json.RawMessage(chunkBytes),
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

func (c *Config) commitSync(providerIdentifier, workflowID string) error {

	if len(c.Thand.Endpoint) == 0 {
		return fmt.Errorf("no thand.io endpoint configured")
	}

	if len(c.Thand.ApiKey) == 0 {
		return fmt.Errorf("no API key configured for thand.io")
	}

	url := fmt.Sprintf("%s/providers/%s/sync/%s/commit", c.Thand.Endpoint, providerIdentifier, workflowID)

	args := &model.HTTPArguments{
		Method: http.MethodPost,
		Endpoint: &model.Endpoint{
			EndpointConfig: &model.EndpointConfiguration{
				URI: &model.LiteralUri{Value: url},
			},
		},
		Headers: map[string]string{
			"Authorization": "Bearer " + c.Thand.ApiKey,
		},
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
