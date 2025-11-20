package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

var ErrNotImplemented = errors.New("not implemented")

/*
name: aws-prod
description: Production AWS environment
provider: aws
config:

	region: us-east-1
	account_id: "123456789012"

enabled: true
*/
type Provider struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Provider    string       `json:"provider"`         // e.g. aws, gcp, azure
	Config      *BasicConfig `json:"config,omitempty"` // Provider-specific configuration
	Role        *Role        `json:"role,omitempty"`   // The base role for this provider
	Enabled     bool         `json:"enabled"`          // Whether this provider is enabled

	client ProviderImpl `json:"-" yaml:"-"`
}

func (p *Provider) GetClient() ProviderImpl {
	return p.client
}

func (p *Provider) HasPermission(user *User) bool {

	// If no user and no role then allow access
	// This is to allow access to public providers
	// e.g. for authentication
	// If a role is defined then we need a user to check against the role
	if user == nil && p.Role == nil {
		logrus.Debugf("Provider %s has no role defined and no user, allowing access", p.Name)
		return true
	} else if user == nil && p.Role != nil {
		// If we have a role defined but no user then deny access
		logrus.Debugf("Provider %s has a role defined but no user, denying access", p.Name)
		return false
	} else if user != nil && p.Role == nil {
		// If we have a user but no role then allow access
		logrus.Debugf("Provider %s has no role defined but has a user, allowing access", p.Name)
		return true
	}

	// Otherwise, if we have a role defined then check the user has that role
	return p.Role.HasPermission(user)
}

func (p *Provider) SetClient(client ProviderImpl) {
	p.client = client
}

func (p *Provider) GetConfig() *BasicConfig {
	return p.Config
}

// ProvidersResponse represents the response for a providers query
type ProvidersResponse struct {
	Version   string                      `json:"version"`
	Providers map[string]ProviderResponse `json:"providers"`
}

type ProviderResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Provider    string `json:"provider"` // e.g. aws, gcp, azure
	Enabled     bool   `json:"enabled"`
}

type ProviderCapability string

const (
	ProviderCapabilityRBAC       ProviderCapability = "rbac"
	ProviderCapabilityAuthorizer ProviderCapability = "authorizor"
	ProviderCapabilityNotifier   ProviderCapability = "notifier"
	ProviderCapabilityIdentities ProviderCapability = "identities" // Provider can return users, groups, etc.
)

func GetCapabilityFromString(cap string) (ProviderCapability, error) {
	switch strings.ToLower(cap) {
	case string(ProviderCapabilityRBAC):
		return ProviderCapabilityRBAC, nil
	case string(ProviderCapabilityAuthorizer):
		return ProviderCapabilityAuthorizer, nil
	case string(ProviderCapabilityNotifier):
		return ProviderCapabilityNotifier, nil
	default:
		return "", fmt.Errorf("unknown capability: %s", cap)
	}
}

/*
A user is assigned a role (e.g., "Manager").
This role has associated permissions (e.g., "approve reports," "view employee data").
These permissions, along with access to specific resources (e.g., "company financial reports"), constitute the user's entitlements.
*/

// Interface for provider implementations
type ProviderImpl interface {
	Initialize(provider Provider) error

	// Form base provider
	GetConfig() *BasicConfig
	GetName() string
	GetDescription() string
	GetProvider() string

	GetCapabilities() []ProviderCapability
	HasCapability(capability ProviderCapability) bool
	HasAnyCapability(capabilities ...ProviderCapability) bool

	ProviderNotifier
	ProviderAuthorizor
	ProviderRoleBasedAccessControl
	ProviderIdentities
}

type AuthorizeSessionResponse struct {
	Url string `json:"url"`
}

type RoleRequest struct {
	User     *User          `json:"user"`
	Role     *Role          `json:"role"`
	Duration *time.Duration `json:"duration,omitempty"` // Optional duration for temporary access
}

// IsValid checks if any of the fields are nil
// if they are then it returns false
func (r *RoleRequest) IsValid() bool {
	return r.User != nil && r.Role != nil
}

func (r *RoleRequest) GetUser() *User {
	return r.User
}

func (r *RoleRequest) GetRole() *Role {
	return r.Role
}

func (r *RoleRequest) GetDuration() *time.Duration {
	return r.Duration
}

type BaseProvider struct {
	provider     Provider
	capabilities []ProviderCapability
}

func NewBaseProvider(provider Provider, capabilities ...ProviderCapability) *BaseProvider {
	return &BaseProvider{
		provider:     provider,
		capabilities: capabilities,
	}
}

func (p *BaseProvider) GetConfig() *BasicConfig {
	return p.provider.Config
}

func (p *BaseProvider) SetConfig(config *BasicConfig) {
	p.provider.Config = config
}

func (p *BaseProvider) GetName() string {
	return p.provider.Name
}

func (p *BaseProvider) GetDescription() string {
	return p.provider.Description
}

func (p *BaseProvider) GetProvider() string {
	return p.provider.Provider
}

func (p *BaseProvider) GetCapabilities() []ProviderCapability {
	return p.capabilities
}

func (p *BaseProvider) HasCapability(capability ProviderCapability) bool {
	return slices.Contains(p.capabilities, capability)
}

func (p *BaseProvider) HasAnyCapability(capabilities ...ProviderCapability) bool {
	return slices.ContainsFunc(capabilities, p.HasCapability)
}

func (p *BaseProvider) Initialize(provider Provider) error {
	// Initialize the provider
	return nil
}

// ProviderDefinitions represents a collection of provider configurations loaded from a file or other source.
type ProviderDefinitions struct {
	Version   *version.Version    `yaml:"version" json:"version"`
	Providers map[string]Provider `yaml:"providers" json:"providers"`
}

// UnmarshalJSON converts Version to string from any type
func (h *ProviderDefinitions) UnmarshalJSON(data []byte) error {
	type Alias ProviderDefinitions
	aux := &struct {
		Version any `json:"version"`
		*Alias
	}{
		Alias: (*Alias)(h),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parsedVersion, err := version.NewVersion(convertVersionToString(aux.Version))

	if err != nil {
		return err
	}

	h.Version = parsedVersion

	return nil
}

// UnmarshalYAML converts Version to string from any type
func (h *ProviderDefinitions) UnmarshalYAML(unmarshal func(any) error) error {
	type Alias ProviderDefinitions
	aux := &struct {
		Version any `yaml:"version"`
		*Alias
	}{
		Alias: (*Alias)(h),
	}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	parsedVersion, err := version.NewVersion(convertVersionToString(aux.Version))

	if err != nil {
		return err
	}

	h.Version = parsedVersion

	return nil
}
