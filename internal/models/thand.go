package models

import "github.com/google/uuid"

type ThandConfig struct {
	Endpoint string `mapstructure:"endpoint"` // The Thand.io service endpoint
	ApiKey   string `mapstructure:"api_key"`  // The API key for authenticating with Thand.io
}

type SynchronizeStartRequest struct {
	ProviderID     uuid.UUID `json:"provider_id" binding:"required"`
	OrganisationID uuid.UUID `json:"organisation_id"`
}

type SynchronizeStartResponse struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

type SynchronizeChunkRequest struct {
	Identities  []Identity           `json:"identities"`
	Users       []User               `json:"users"`
	Groups      []Group              `json:"groups"`
	Roles       []ProviderRole       `json:"roles"`
	Permissions []ProviderPermission `json:"permissions"`
	Resources   []ProviderResource   `json:"resources"`
}

type SynchronizeCommitRequest struct {
	// Empty for now, but could contain summary stats
}
