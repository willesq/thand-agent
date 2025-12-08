package models

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func (p *BaseProvider) Synchronize(
	ctx context.Context,
	temporalService TemporalImpl,
	syncRequest *SynchronizeRequest,
) error {
	return Synchronize(ctx, temporalService, p, syncRequest)
}

// Synchronize performs synchronization of identities, roles, permissions, and resources
// for the given provider. It can use Temporal workflows if a Temporal service
// is provided, otherwise it falls back to a pure Go implementation.
// The SynchronizeRequest can specify which capabilities to synchronize.
// and can be nil to use default behavior.
func Synchronize(
	ctx context.Context,
	temporalService TemporalImpl,
	provider ProviderImpl,
	syncRequest *SynchronizeRequest,
) error {

	// Check if we have the relevant capabilities for synchronization
	if !provider.HasAnyCapability(
		ProviderCapabilityIdentities,
		ProviderCapabilityRBAC,
	) {
		logrus.Infof("Provider %s does not have synchronization capabilities, skipping", provider.GetName())
		return nil
	}

	// Set default values
	if syncRequest == nil {
		syncRequest = &SynchronizeRequest{}
	}

	if len(syncRequest.ProviderIdentifier) == 0 {
		syncRequest.ProviderIdentifier = provider.GetIdentifier()
	}

	if len(syncRequest.Requests) == 0 {

		requests := getSynchronizationRequests(provider)

		if len(requests) == 0 {
			logrus.Infof("Provider %s does not have overridden synchronization methods, skipping", provider.GetName())
			return nil
		}
		syncRequest.Requests = requests

	}

	if temporalService != nil {

		temporalClient := temporalService.GetClient()

		// Execute the provider workflow synchronize
		workflowOptions := client.StartWorkflowOptions{
			ID:        GetTemporalName(provider.GetIdentifier(), TemporalSynchronizeWorkflowName),
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
			GetTemporalName(provider.GetIdentifier(), TemporalSynchronizeWorkflowName),
			syncRequest,
		)

		if err != nil {
			return fmt.Errorf("failed to execute synchronize workflow: %w", err)
		}

		var resp SynchronizeResponse
		if err := we.Get(context.Background(), &resp); err != nil {
			return fmt.Errorf("failed to get synchronize workflow result: %w", err)
		}

		if len(resp.Identities) > 0 {
			logrus.WithFields(logrus.Fields{
				"identities": len(resp.Identities),
			}).Info("Setting synchronized identities")
			provider.SetIdentities(resp.Identities)
		}
		if len(resp.Roles) > 0 {
			logrus.WithFields(logrus.Fields{
				"roles": len(resp.Roles),
			}).Info("Setting synchronized roles")
			provider.SetRoles(resp.Roles)
		}
		if len(resp.Permissions) > 0 {
			logrus.WithFields(logrus.Fields{
				"permissions": len(resp.Permissions),
			}).Info("Setting synchronized permissions")
			provider.SetPermissions(resp.Permissions)
		}
		if len(resp.Resources) > 0 {
			logrus.WithFields(logrus.Fields{
				"resources": len(resp.Resources),
			}).Info("Setting synchronized resources")
			provider.SetResources(resp.Resources)
		}

		return nil
	}

	// Pure Go implementation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	syncResponse := &SynchronizeResponse{}

	activities := NewProviderActivities(provider)
	var upstreamWorkflowID string

	if syncRequest.Upstream != nil {
		resp, err := activities.SynchronizeThandStart(ctx, syncRequest.Upstream, provider.GetIdentifier())
		if err != nil {
			return fmt.Errorf("failed to start upstream sync: %w", err)
		}
		upstreamWorkflowID = resp.WorkflowID
	}

	// Helper to run sync
	runSync := func(name SynchronizeCapability, syncFunc func() error) {

		if !slices.Contains(syncRequest.Requests, name) {
			logrus.Infof("Skipping synchronization for %s as it's not requested", name)
			return
		}

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

	if provider.HasCapability(ProviderCapabilityIdentities) {
		// Synchronize Identities
		runSync(SynchronizeIdentities, func() error {
			req := SynchronizeUsersRequest{}
			for {
				resp, err := provider.SynchronizeIdentities(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				mu.Unlock()

				if syncRequest.Upstream != nil {
					chunk := SynchronizeChunkRequest{Identities: resp.Identities}
					if err := activities.SynchronizeThandChunk(ctx, syncRequest.Upstream, provider.GetIdentifier(), upstreamWorkflowID, chunk); err != nil {
						return fmt.Errorf("failed to send chunk: %w", err)
					}
				}

				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Users
		runSync(SynchronizeUsers, func() error {
			req := SynchronizeUsersRequest{}
			for {
				resp, err := provider.SynchronizeUsers(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				mu.Unlock()

				if syncRequest.Upstream != nil {
					chunk := SynchronizeChunkRequest{Identities: resp.Identities}
					if err := activities.SynchronizeThandChunk(ctx, syncRequest.Upstream, provider.GetIdentifier(), upstreamWorkflowID, chunk); err != nil {
						return fmt.Errorf("failed to send chunk: %w", err)
					}
				}

				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Groups
		runSync(SynchronizeGroups, func() error {
			req := SynchronizeGroupsRequest{}
			for {
				resp, err := provider.SynchronizeGroups(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				mu.Unlock()

				if syncRequest.Upstream != nil {
					chunk := SynchronizeChunkRequest{Identities: resp.Identities}
					if err := activities.SynchronizeThandChunk(ctx, syncRequest.Upstream, provider.GetIdentifier(), upstreamWorkflowID, chunk); err != nil {
						return fmt.Errorf("failed to send chunk: %w", err)
					}
				}

				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})
	}

	if provider.HasCapability(ProviderCapabilityRBAC) {
		// Synchronize Resources
		runSync(SynchronizeResources, func() error {
			req := SynchronizeResourcesRequest{}
			for {
				resp, err := provider.SynchronizeResources(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Resources = append(syncResponse.Resources, resp.Resources...)
				mu.Unlock()

				if syncRequest.Upstream != nil {
					chunk := SynchronizeChunkRequest{Resources: resp.Resources}
					if err := activities.SynchronizeThandChunk(ctx, syncRequest.Upstream, provider.GetIdentifier(), upstreamWorkflowID, chunk); err != nil {
						return fmt.Errorf("failed to send chunk: %w", err)
					}
				}

				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Roles
		runSync(SynchronizeRoles, func() error {
			req := SynchronizeRolesRequest{}
			for {
				resp, err := provider.SynchronizeRoles(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Roles = append(syncResponse.Roles, resp.Roles...)
				mu.Unlock()

				if syncRequest.Upstream != nil {
					chunk := SynchronizeChunkRequest{Roles: resp.Roles}
					if err := activities.SynchronizeThandChunk(ctx, syncRequest.Upstream, provider.GetIdentifier(), upstreamWorkflowID, chunk); err != nil {
						return fmt.Errorf("failed to send chunk: %w", err)
					}
				}

				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})

		// Synchronize Permissions
		runSync(SynchronizePermissions, func() error {
			req := SynchronizePermissionsRequest{}
			for {
				resp, err := provider.SynchronizePermissions(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				syncResponse.Permissions = append(syncResponse.Permissions, resp.Permissions...)
				mu.Unlock()

				if syncRequest.Upstream != nil {
					chunk := SynchronizeChunkRequest{Permissions: resp.Permissions}
					if err := activities.SynchronizeThandChunk(ctx, syncRequest.Upstream, provider.GetIdentifier(), upstreamWorkflowID, chunk); err != nil {
						return fmt.Errorf("failed to send chunk: %w", err)
					}
				}

				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			return nil
		})
	}

	logrus.WithFields(logrus.Fields{
		"requests": len(syncRequest.Requests),
	}).Info("Waiting for synchronization tasks to complete")

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("synchronization failed: %v", errs)
	}

	if syncRequest.Upstream != nil {
		if err := activities.SynchronizeThandCommit(ctx, syncRequest.Upstream, provider.GetIdentifier(), upstreamWorkflowID); err != nil {
			return fmt.Errorf("failed to commit upstream sync: %w", err)
		}
	}

	if len(syncResponse.Identities) > 0 {
		logrus.WithFields(logrus.Fields{
			"identities": len(syncResponse.Identities),
		}).Info("Setting synchronized identities")
		provider.SetIdentities(syncResponse.Identities)
	}
	if len(syncResponse.Roles) > 0 {
		logrus.WithFields(logrus.Fields{
			"roles": len(syncResponse.Roles),
		}).Info("Setting synchronized roles")
		provider.SetRoles(syncResponse.Roles)
	}
	if len(syncResponse.Permissions) > 0 {
		logrus.WithFields(logrus.Fields{
			"permissions": len(syncResponse.Permissions),
		}).Info("Setting synchronized permissions")
		provider.SetPermissions(syncResponse.Permissions)
	}
	if len(syncResponse.Resources) > 0 {
		logrus.WithFields(logrus.Fields{
			"resources": len(syncResponse.Resources),
		}).Info("Setting synchronized resources")
		provider.SetResources(syncResponse.Resources)
	}

	return nil
}

func getSynchronizationRequests(provider ProviderImpl) []SynchronizeCapability {
	requests := make([]SynchronizeCapability, 0)

	// Determine which capabilities to synchronize
	// Check if the underlying provider has been overridden to
	// support identities, roles, permissions, resources

	if provider.CanSynchronizeIdentities() {
		requests = append(requests, SynchronizeIdentities)
	}

	if provider.CanSynchronizeUsers() {
		requests = append(requests, SynchronizeUsers)
	}

	if provider.CanSynchronizeGroups() {
		requests = append(requests, SynchronizeGroups)
	}

	if provider.CanSynchronizeResources() {
		requests = append(requests, SynchronizeResources)
	}

	if provider.CanSynchronizeRoles() {
		requests = append(requests, SynchronizeRoles)
	}

	if provider.CanSynchronizePermissions() {
		requests = append(requests, SynchronizePermissions)
	}

	return requests
}

func (p *BaseProvider) CanSynchronizeRoles() bool {
	return false
}

func (p *BaseProvider) CanSynchronizePermissions() bool {
	return false
}

func (p *BaseProvider) CanSynchronizeUsers() bool {
	return false
}

func (p *BaseProvider) CanSynchronizeGroups() bool {
	return false
}

func (p *BaseProvider) CanSynchronizeIdentities() bool {
	return false
}

func (p *BaseProvider) CanSynchronizeResources() bool {
	return false
}
