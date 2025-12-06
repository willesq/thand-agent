package models

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

type SynchronizeRequest struct {
	//
}

type SynchronizeResponse struct {
	Roles       []ProviderRole       `json:"roles,omitempty"`
	Permissions []ProviderPermission `json:"permissions,omitempty"`
	Resources   []ProviderResource   `json:"resources,omitempty"`
	Identities  []Identity           `json:"identities,omitempty"`
}

func Synchronize(ctx workflow.Context, req SynchronizeRequest) (*SynchronizeResponse, error) {
	ao := workflow.ActivityOptions{

		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute all the synchronizations needed for RBAC
	// in parallel using the workflow parallel pattern

	var a *ProviderActivities

	syncResponse := &SynchronizeResponse{}
	errChan := workflow.NewChannel(ctx)
	syncCount := 4

	// Synchronize Identities
	workflow.Go(ctx, func(ctx workflow.Context) {
		req := SynchronizeIdentitiesRequest{}
		for {
			var resp SynchronizeIdentitiesResponse
			err := workflow.ExecuteActivity(ctx, a.SynchronizeIdentities, req).Get(ctx, &resp)
			if err != nil {
				errChan.Send(ctx, err)
				return
			}
			syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
			if resp.Pagination == nil || resp.Pagination.Token == "" {
				break
			}
			req.Pagination = resp.Pagination
		}
		errChan.Send(ctx, nil)
	})

	// Synchronize Users
	workflow.Go(ctx, func(ctx workflow.Context) {
		req := SynchronizeUsersRequest{}
		for {
			var resp SynchronizeUsersResponse
			err := workflow.ExecuteActivity(ctx, a.SynchronizeUsers, req).Get(ctx, &resp)
			if err != nil {
				errChan.Send(ctx, err)
				return
			}
			syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
			if resp.Pagination == nil || resp.Pagination.Token == "" {
				break
			}
			req.Pagination = resp.Pagination
		}
		errChan.Send(ctx, nil)
	})

	// Synchronize Groups
	workflow.Go(ctx, func(ctx workflow.Context) {
		req := SynchronizeGroupsRequest{}
		for {
			var resp SynchronizeGroupsResponse
			err := workflow.ExecuteActivity(ctx, a.SynchronizeGroups, req).Get(ctx, &resp)
			if err != nil {
				errChan.Send(ctx, err)
				return
			}
			syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
			if resp.Pagination == nil || resp.Pagination.Token == "" {
				break
			}
			req.Pagination = resp.Pagination
		}
		errChan.Send(ctx, nil)
	})

	// Synchronize Resources
	workflow.Go(ctx, func(ctx workflow.Context) {
		req := SynchronizeResourcesRequest{}
		for {
			var resp SynchronizeResourcesResponse
			err := workflow.ExecuteActivity(ctx, a.SynchronizeResources, req).Get(ctx, &resp)
			if err != nil {
				errChan.Send(ctx, err)
				return
			}
			syncResponse.Resources = append(syncResponse.Resources, resp.Resources...)
			if resp.Pagination == nil || resp.Pagination.Token == "" {
				break
			}
			req.Pagination = resp.Pagination
		}
		errChan.Send(ctx, nil)
	})

	// Synchronize Roles
	workflow.Go(ctx, func(ctx workflow.Context) {
		req := SynchronizeRolesRequest{}
		for {
			var resp SynchronizeRolesResponse
			err := workflow.ExecuteActivity(ctx, a.SynchronizeRoles, req).Get(ctx, &resp)
			if err != nil {
				errChan.Send(ctx, err)
				return
			}
			syncResponse.Roles = append(syncResponse.Roles, resp.Roles...)
			if resp.Pagination == nil || resp.Pagination.Token == "" {
				break
			}
			req.Pagination = resp.Pagination
		}
		errChan.Send(ctx, nil)
	})

	// Synchronize Permissions
	workflow.Go(ctx, func(ctx workflow.Context) {
		req := SynchronizePermissionsRequest{}
		for {
			var resp SynchronizePermissionsResponse
			err := workflow.ExecuteActivity(ctx, a.SynchronizePermissions, req).Get(ctx, &resp)
			if err != nil {
				errChan.Send(ctx, err)
				return
			}
			syncResponse.Permissions = append(syncResponse.Permissions, resp.Permissions...)
			if resp.Pagination == nil || resp.Pagination.Token == "" {
				break
			}
			req.Pagination = resp.Pagination
		}
		errChan.Send(ctx, nil)
	})

	var errs []error
	for range syncCount {
		var err error
		errChan.Receive(ctx, &err)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("synchronization failed: %v", errs)
	}

	return syncResponse, nil
}
