package models

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
)

type SynchronizeCapability string

const (
	SynchronizeRoles       SynchronizeCapability = "SynchronizeRoles"
	SynchronizePermissions SynchronizeCapability = "SynchronizePermissions"
	SynchronizeResources   SynchronizeCapability = "SynchronizeResources"
	SynchronizeIdentities  SynchronizeCapability = "SynchronizeIdentities"
	SynchronizeUsers       SynchronizeCapability = "SynchronizeUsers"
	SynchronizeGroups      SynchronizeCapability = "SynchronizeGroups"
)

type SynchronizeRequest struct {
	ProviderIdentifier string                  `json:"provider"` // Provider name
	Requests           []SynchronizeCapability `json:"requests,omitempty"`
}

type SynchronizeResponse struct {
	Roles       []ProviderRole       `json:"roles,omitempty"`
	Permissions []ProviderPermission `json:"permissions,omitempty"`
	Resources   []ProviderResource   `json:"resources,omitempty"`
	Identities  []Identity           `json:"identities,omitempty"`
}

func CreateTemporalIdentifier(providerIdentifier, base string) string {
	return strings.ToLower(providerIdentifier + "-" + base)
}

func GetNameFromFunction(i any) string {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Func && v.Type().NumIn() > 0 {
		receiverType := v.Type().In(0)
		for j := 0; j < receiverType.NumMethod(); j++ {
			method := receiverType.Method(j)
			if method.Func.Pointer() == v.Pointer() {
				return method.Name
			}
		}
	}
	fullName := runtime.FuncForPC(v.Pointer()).Name()
	parts := strings.Split(fullName, ".")
	return strings.TrimSuffix(parts[len(parts)-1], "-fm")
}

func SynchronizeWorkflow(ctx workflow.Context, syncReq SynchronizeRequest) (*SynchronizeResponse, error) {

	if len(syncReq.ProviderIdentifier) == 0 {
		return nil, fmt.Errorf("provider identifier is required")
	}

	log := workflow.GetLogger(ctx)
	log.Info("Starting synchronization workflow for provider: ", syncReq.ProviderIdentifier)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute all the synchronizations needed for RBAC
	// in parallel using the workflow parallel pattern

	syncResponse := &SynchronizeResponse{}
	errChan := workflow.NewChannel(ctx)
	syncCount := 0

	shouldSync := func(cap SynchronizeCapability) bool {
		if len(syncReq.Requests) == 0 {
			return true
		}
		return slices.Contains(syncReq.Requests, cap)
	}

	// Synchronize Identities
	if shouldSync(SynchronizeIdentities) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			req := SynchronizeIdentitiesRequest{}
			for {
				var resp SynchronizeIdentitiesResponse
				err := workflow.ExecuteActivity(
					ctx,
					CreateTemporalIdentifier(
						syncReq.ProviderIdentifier,
						GetNameFromFunction((*ProviderActivities).SynchronizeIdentities),
					),
					req,
				).
					Get(ctx, &resp)
				if err != nil {
					errChan.Send(ctx, err)
					return
				}
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			errChan.Send(ctx, nil)
		})
	}

	// Synchronize Users
	if shouldSync(SynchronizeUsers) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			req := SynchronizeUsersRequest{}
			for {
				var resp SynchronizeUsersResponse
				err := workflow.ExecuteActivity(
					ctx,
					CreateTemporalIdentifier(
						syncReq.ProviderIdentifier,
						GetNameFromFunction((*ProviderActivities).SynchronizeUsers),
					),
					req,
				).Get(ctx, &resp)
				if err != nil {
					errChan.Send(ctx, err)
					return
				}
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			errChan.Send(ctx, nil)
		})
	}

	// Synchronize Groups
	if shouldSync(SynchronizeGroups) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			req := SynchronizeGroupsRequest{}
			for {
				var resp SynchronizeGroupsResponse
				err := workflow.ExecuteActivity(
					ctx,
					CreateTemporalIdentifier(
						syncReq.ProviderIdentifier,
						GetNameFromFunction((*ProviderActivities).SynchronizeGroups),
					),
					req,
				).Get(ctx, &resp)
				if err != nil {
					errChan.Send(ctx, err)
					return
				}
				syncResponse.Identities = append(syncResponse.Identities, resp.Identities...)
				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			errChan.Send(ctx, nil)
		})
	}

	// Synchronize Resources
	if shouldSync(SynchronizeResources) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			req := SynchronizeResourcesRequest{}
			for {
				var resp SynchronizeResourcesResponse
				err := workflow.ExecuteActivity(
					ctx,
					CreateTemporalIdentifier(
						syncReq.ProviderIdentifier,
						GetNameFromFunction((*ProviderActivities).SynchronizeResources),
					),
					req,
				).Get(ctx, &resp)
				if err != nil {
					errChan.Send(ctx, err)
					return
				}
				syncResponse.Resources = append(syncResponse.Resources, resp.Resources...)
				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			errChan.Send(ctx, nil)
		})
	}

	// Synchronize Roles
	if shouldSync(SynchronizeRoles) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			req := SynchronizeRolesRequest{}
			for {
				var resp SynchronizeRolesResponse
				err := workflow.ExecuteActivity(
					ctx,
					CreateTemporalIdentifier(
						syncReq.ProviderIdentifier,
						GetNameFromFunction((*ProviderActivities).SynchronizeRoles),
					),
					req,
				).Get(ctx, &resp)
				if err != nil {
					errChan.Send(ctx, err)
					return
				}
				syncResponse.Roles = append(syncResponse.Roles, resp.Roles...)
				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			errChan.Send(ctx, nil)
		})
	}

	// Synchronize Permissions
	if shouldSync(SynchronizePermissions) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			req := SynchronizePermissionsRequest{}
			for {
				var resp SynchronizePermissionsResponse
				err := workflow.ExecuteActivity(
					ctx,
					CreateTemporalIdentifier(
						syncReq.ProviderIdentifier,
						GetNameFromFunction((*ProviderActivities).SynchronizePermissions),
					),
					req,
				).Get(ctx, &resp)
				if err != nil {
					errChan.Send(ctx, err)
					return
				}
				syncResponse.Permissions = append(syncResponse.Permissions, resp.Permissions...)
				if resp.Pagination == nil || len(resp.Pagination.Token) == 0 {
					break
				}
				req.Pagination = resp.Pagination
			}
			errChan.Send(ctx, nil)
		})
	}

	var errs []error
	for range syncCount {
		var err error
		errChan.Receive(ctx, &err)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		// Log errors but return what we have
		log.Error("Synchronization workflow encountered errors: ", errs)
	}

	return syncResponse, nil
}
