package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	// ThandSyncWorkflowName is the name of the system synchronization workflow
	ThandSyncWorkflowName = "thand-sync"

	// Signal names
	SignalSystemUpdate = "SystemUpdate"
	SignalRemoteUpdate = "RemoteUpdate"

	// Query names
	QueryGetSystemState = "GetSystemState"
)

type ThandConfig struct {
	Endpoint string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint" default:"https://app.thand.io/"`
	Base     string `json:"base" yaml:"base" mapstructure:"base" default:"/"` // Base path for login endpoint e.g. /
	ApiKey   string `json:"api_key" yaml:"api_key" mapstructure:"api_key"`    // The API key for authenticating with Thand.io
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

// SystemSyncRequest is the input for the workflow
type SystemSyncRequest struct {
	AgentID string
}

// SystemSyncState represents the current state of the sync workflow
type SystemSyncState struct {
	LastSyncTime time.Time
	Status       string
}

// SystemChunk represents a batch of updates to be synced
type SystemChunk struct {
	// Configuration Definitions (Versioned)
	Roles     map[string]Role     `json:"roles,omitempty"`
	Workflows map[string]Workflow `json:"workflows,omitempty"`
	Providers map[string]Provider `json:"providers,omitempty"`

	// Provider Data (Aggregated from all providers)
	ProviderData map[string]ProviderData `json:"provider_data,omitempty"`
}

// ProviderData represents the data collected from a specific provider
type ProviderData struct {
	Identities    []Identity           `json:"identities,omitempty"`
	Users         []User               `json:"users,omitempty"`
	Groups        []Group              `json:"groups,omitempty"`
	Permissions   []ProviderPermission `json:"permissions,omitempty"`
	Resources     []ProviderResource   `json:"resources,omitempty"`
	ProviderRoles []ProviderRole       `json:"provider_roles,omitempty"`
}
