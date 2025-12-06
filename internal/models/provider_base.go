package models

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type BaseProvider struct {
	identifier   string
	name         string
	description  string
	provider     string
	config       *BasicConfig
	role         *Role
	capabilities []ProviderCapability

	// Add other common fields if necessary
	identity *IdentitySupport
	rbac     *RBACSupport

	impl ProviderImpl
}

func (p *BaseProvider) SetProviderImpl(impl ProviderImpl) {
	p.impl = impl
}

type IdentitySupport struct {
	mu sync.RWMutex

	// Identity management
	identities      []Identity
	identitiesMap   map[string]*Identity
	identitiesIndex bleve.Index
}

type RBACSupport struct {
	mu sync.RWMutex

	// Permission management
	permissions      []ProviderPermission
	permissionsMap   map[string]*ProviderPermission
	permissionsIndex bleve.Index

	// Role management
	roles      []ProviderRole
	rolesMap   map[string]*ProviderRole
	rolesIndex bleve.Index

	// Resource management
	resources      []ProviderResource
	resourcesMap   map[string]*ProviderResource
	resourcesIndex bleve.Index
}

func NewBaseProvider(identifier string, provider Provider, capabilities ...ProviderCapability) *BaseProvider {
	base := BaseProvider{
		identifier:   identifier,
		name:         provider.Name,
		description:  provider.Description,
		provider:     provider.Provider,
		config:       provider.Config,
		role:         provider.Role,
		capabilities: capabilities,
	}

	if base.HasCapability(ProviderCapabilityIdentities) {
		// Initialize identities map or other structures if needed
		base.identity = &IdentitySupport{
			identities:    make([]Identity, 0),
			identitiesMap: make(map[string]*Identity),
		}
	}

	if base.HasCapability(ProviderCapabilityRBAC) {
		// Initialize RBAC structures if needed
		base.rbac = &RBACSupport{
			permissions:    make([]ProviderPermission, 0),
			permissionsMap: make(map[string]*ProviderPermission),

			roles:    make([]ProviderRole, 0),
			rolesMap: make(map[string]*ProviderRole),
		}
	}

	return &base
}

func (p *BaseProvider) GetConfig() *BasicConfig {
	return p.config
}

func (p *BaseProvider) SetConfig(config *BasicConfig) {
	p.config = config
}

func (p *BaseProvider) SetPermissions(permissions []ProviderPermission) {
	p.SetPermissionsWithKey(permissions, func(p *ProviderPermission) string {
		return p.Name
	})
}

// Create the permissions map
func (p *BaseProvider) SetPermissionsWithKey(
	permissions []ProviderPermission,
	keyFunc func(p *ProviderPermission) string,
) {
	if p.rbac == nil {
		return
	}
	p.rbac.mu.Lock()
	defer p.rbac.mu.Unlock()
	p.rbac.permissions = permissions

	// Create the permissions map
	p.rbac.permissionsMap = make(map[string]*ProviderPermission)
	for i := range permissions {
		perm := &permissions[i]
		keyName := keyFunc(perm)
		p.rbac.permissionsMap[strings.ToLower(keyName)] = perm
	}

	// Trigger reindex
	go func() {
		err := p.buildPermissionIndices()
		if err != nil {
			logrus.WithError(err).Error("Failed to build rbac search indices")
			return
		}
	}()
}

func (p *BaseProvider) SetRoles(roles []ProviderRole) {
	p.SetRolesWithKey(roles, func(r *ProviderRole) string {
		return r.Name
	})
}

func (p *BaseProvider) SetRolesWithKey(
	roles []ProviderRole,
	keyFunc func(r *ProviderRole) string) {
	if p.rbac == nil {
		return
	}
	p.rbac.mu.Lock()
	defer p.rbac.mu.Unlock()
	p.rbac.roles = roles

	// Create the roles map
	p.rbac.rolesMap = make(map[string]*ProviderRole)
	for i := range roles {
		role := &roles[i]
		keyName := keyFunc(role)
		p.rbac.rolesMap[strings.ToLower(keyName)] = role
	}

	// Trigger reindex
	go func() {
		err := p.buildRoleIndices()
		if err != nil {
			logrus.WithError(err).Error("Failed to build role search indices")
			return
		}
	}()
}

func (p *BaseProvider) SetResources(resources []ProviderResource) {
	p.SetResourcesWithKey(resources, func(r *ProviderResource) string {
		return r.Id
	})
}

func (p *BaseProvider) SetResourcesWithKey(
	resources []ProviderResource,
	keyFunc func(r *ProviderResource) string,
) {
	if p.rbac == nil {
		return
	}
	p.rbac.mu.Lock()
	defer p.rbac.mu.Unlock()
	p.rbac.resources = resources

	// Create the resources map
	p.rbac.resourcesMap = make(map[string]*ProviderResource)
	for i := range resources {
		resource := &resources[i]
		keyName := keyFunc(resource)
		p.rbac.resourcesMap[strings.ToLower(keyName)] = resource
	}

	// Trigger reindex
	go func() {
		err := p.buildResourceIndices()
		if err != nil {
			logrus.WithError(err).Error("Failed to build resources search indices")
			return
		}
	}()
}

func (p *BaseProvider) SetIdentities(identities []Identity) {
	p.SetIdentitiesWithKey(identities, func(i *Identity) []string {
		var keys []string
		keys = append(keys, i.ID)
		keys = append(keys, i.Label)
		if i.User != nil && len(i.User.Email) != 0 {
			keys = append(keys, i.User.Email)
		}
		if i.Group != nil {
			if len(i.Group.Name) != 0 {
				keys = append(keys, i.Group.Name)
			}
			if len(i.Group.Email) != 0 {
				keys = append(keys, i.Group.Email)
			}
		}
		return keys
	})
}

func (p *BaseProvider) SetIdentitiesWithKey(
	identities []Identity,
	keyFunc func(i *Identity) []string,
) {

	if p.identity == nil {
		return
	}
	p.identity.mu.Lock()
	defer p.identity.mu.Unlock()
	p.identity.identities = identities

	// Build the identities map
	for i := range identities {

		identity := &identities[i]

		keys := keyFunc(identity)

		for _, key := range keys {
			p.identity.identitiesMap[strings.ToLower(key)] = identity
		}
	}

	// Trigger reindex
	go func() {
		err := p.buildIdentitiyIndices()
		if err != nil {
			logrus.WithError(err).Error("Failed to build identity search indices")
			return
		}
	}()
}

func (p *BaseProvider) GetIdentifier() string {
	return p.identifier
}

func (p *BaseProvider) GetName() string {
	return p.name
}

func (p *BaseProvider) GetDescription() string {
	return p.description
}

func (p *BaseProvider) GetProvider() string {
	return p.provider
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

func (p *BaseProvider) EnableCapability(capability ProviderCapability) {
	if !p.HasCapability(capability) {
		p.capabilities = append(p.capabilities, capability)
	}
}

func (p *BaseProvider) DisableCapability(capability ProviderCapability) {
	p.capabilities = slices.DeleteFunc(p.capabilities, func(c ProviderCapability) bool {
		return c == capability
	})
}

func (p *BaseProvider) Initialize(identifier string, provider Provider) error {
	// Initialize the provider
	return nil
}

func (p *BaseProvider) Synchronize(ctx context.Context, temporalService TemporalImpl) error {

	// Check if we have the relevant capabilities for synchronization
	if !p.HasAnyCapability(
		ProviderCapabilityIdentities,
		ProviderCapabilityRBAC,
	) {
		logrus.Infof("Provider %s does not have synchronization capabilities, skipping", p.GetName())
		return nil
	}

	if temporalService != nil {

		temporalClient := temporalService.GetClient()

		// Execute the provider workflow synchronize
		workflowOptions := client.StartWorkflowOptions{
			ID:        GetTemporalName(p.GetIdentifier(), TemporalSynchronizeWorkflowName),
			TaskQueue: temporalService.GetTaskQueue(),
			// Set a timeout for the workflow execution
			WorkflowExecutionTimeout: 30 * time.Minute,
		}

		// Only add versioning override if versioning is enabled
		if !temporalService.IsVersioningDisabled() {
			workflowOptions.VersioningOverride = &client.PinnedVersioningOverride{
				Version: worker.WorkerDeploymentVersion{
					DeploymentName: TemporalDeploymentName,
					BuildID:        common.GetBuildIdentifier(),
				},
			}
		}

		we, err := temporalClient.ExecuteWorkflow(
			ctx,
			workflowOptions,
			GetTemporalName(p.GetIdentifier(), TemporalSynchronizeWorkflowName),
			SynchronizeRequest{
				ProviderIdentifier: p.GetIdentifier(),
			},
		)

		if err != nil {
			return fmt.Errorf("failed to execute synchronize workflow: %w", err)
		}

		var resp SynchronizeResponse
		if err := we.Get(context.Background(), &resp); err != nil {
			return fmt.Errorf("failed to get synchronize workflow result: %w", err)
		}

		p.SetIdentities(resp.Identities)
		p.SetRoles(resp.Roles)
		p.SetPermissions(resp.Permissions)
		p.SetResources(resp.Resources)

		return nil
	}

	// Pure Go implementation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	syncResponse := &SynchronizeResponse{}

	// Helper to run sync
	runSync := func(name string, syncFunc func() error) {
		wg.Go(func() {
			if err := syncFunc(); err != nil {
				// Ignore not implemented errors
				if errors.Is(err, ErrNotImplemented) {
					return
				}
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s failed: %w", name, err))
				mu.Unlock()
			}
		})
	}

	if p.HasCapability(ProviderCapabilityIdentities) {
		// Synchronize Identities
		runSync("Identities", func() error {
			req := SynchronizeUsersRequest{}
			for {
				resp, err := p.SynchronizeIdentities(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				mu.Unlock()
				if resp.Pagination == nil || resp.Pagination.Token == "" {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Users
		runSync("Users", func() error {
			req := SynchronizeUsersRequest{}
			for {
				resp, err := p.SynchronizeUsers(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				mu.Unlock()
				if resp.Pagination == nil || resp.Pagination.Token == "" {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Groups
		runSync("Groups", func() error {
			req := SynchronizeGroupsRequest{}
			for {
				resp, err := p.SynchronizeGroups(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				mu.Unlock()
				if resp.Pagination == nil || resp.Pagination.Token == "" {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})
	}

	if p.HasCapability(ProviderCapabilityRBAC) {
		// Synchronize Resources
		runSync("Resources", func() error {
			req := SynchronizeResourcesRequest{}
			for {
				resp, err := p.SynchronizeResources(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Resources = append(syncResponse.Resources, resp.Resources...)
				mu.Unlock()
				if resp.Pagination == nil || resp.Pagination.Token == "" {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Roles
		runSync("Roles", func() error {
			req := SynchronizeRolesRequest{}
			for {
				resp, err := p.SynchronizeRoles(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Roles = append(syncResponse.Roles, resp.Roles...)
				mu.Unlock()
				if resp.Pagination == nil || resp.Pagination.Token == "" {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Permissions
		runSync("Permissions", func() error {
			req := SynchronizePermissionsRequest{}
			for {
				resp, err := p.SynchronizePermissions(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Permissions = append(syncResponse.Permissions, resp.Permissions...)
				mu.Unlock()
				if resp.Pagination == nil || resp.Pagination.Token == "" {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("synchronization failed: %v", errs)
	}

	p.SetIdentities(syncResponse.Identities)
	p.SetRoles(syncResponse.Roles)
	p.SetPermissions(syncResponse.Permissions)
	p.SetResources(syncResponse.Resources)

	return nil
}
