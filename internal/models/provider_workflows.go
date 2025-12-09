package models

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/thand-io/agent/internal/common"
	"go.temporal.io/sdk/temporal"
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
	// Everything will be updated using local activities,
	// so we can just return an empty response for now
}

func CreateTemporalWorkflowIdentifier(workflowName string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", common.GetClientIdentifier(), workflowName))
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

func runSyncLoop[Req any, Resp any](
	ctx workflow.Context,
	providerID string,
	activityMethod any,
	req Req,
	setPagination func(*Req, *PaginationOptions),
	getPagination func(Resp) *PaginationOptions,
) error {

	ao := workflow.LocalActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    100 * time.Second,
			MaximumAttempts:    10,
		},
	}

	ctx = workflow.WithLocalActivityOptions(ctx, ao)

	for {
		var resp Resp
		err := workflow.ExecuteLocalActivity(
			ctx,
			CreateTemporalProviderWorkflowName(
				providerID,
				GetNameFromFunction(activityMethod),
			),
			req,
		).Get(ctx, &resp)

		if err != nil {
			return err
		}

		pagination := getPagination(resp)
		if pagination == nil || len(pagination.Token) == 0 {
			break
		}
		setPagination(&req, pagination)
	}
	return nil
}

func ProviderSynchronizeWorkflow(ctx workflow.Context, syncReq SynchronizeRequest) (*SynchronizeResponse, error) {

	if len(syncReq.ProviderIdentifier) == 0 {
		return nil, fmt.Errorf("provider identifier is required")
	}

	log := workflow.GetLogger(ctx)
	log.Info("Starting synchronization workflow for provider: ", syncReq.ProviderIdentifier)

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
			err := runSyncLoop(ctx, syncReq.ProviderIdentifier, (*ProviderActivities).SynchronizeIdentities, SynchronizeIdentitiesRequest{},
				func(r *SynchronizeIdentitiesRequest, p *PaginationOptions) { r.Pagination = p },
				func(r SynchronizeIdentitiesResponse) *PaginationOptions { return r.Pagination },
			)
			errChan.Send(ctx, err)
		})
	}

	// Synchronize Users
	if shouldSync(SynchronizeUsers) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			err := runSyncLoop(ctx, syncReq.ProviderIdentifier, (*ProviderActivities).SynchronizeUsers, SynchronizeUsersRequest{},
				func(r *SynchronizeUsersRequest, p *PaginationOptions) { r.Pagination = p },
				func(r SynchronizeUsersResponse) *PaginationOptions { return r.Pagination },
			)
			errChan.Send(ctx, err)
		})
	}

	// Synchronize Groups
	if shouldSync(SynchronizeGroups) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			err := runSyncLoop(ctx, syncReq.ProviderIdentifier, (*ProviderActivities).SynchronizeGroups, SynchronizeGroupsRequest{},
				func(r *SynchronizeGroupsRequest, p *PaginationOptions) { r.Pagination = p },
				func(r SynchronizeGroupsResponse) *PaginationOptions { return r.Pagination },
			)
			errChan.Send(ctx, err)
		})
	}

	// Synchronize Resources
	if shouldSync(SynchronizeResources) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			err := runSyncLoop(ctx, syncReq.ProviderIdentifier, (*ProviderActivities).SynchronizeResources, SynchronizeResourcesRequest{},
				func(r *SynchronizeResourcesRequest, p *PaginationOptions) { r.Pagination = p },
				func(r SynchronizeResourcesResponse) *PaginationOptions { return r.Pagination },
			)
			errChan.Send(ctx, err)
		})
	}

	// Synchronize Roles
	if shouldSync(SynchronizeRoles) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			err := runSyncLoop(ctx, syncReq.ProviderIdentifier, (*ProviderActivities).SynchronizeRoles, SynchronizeRolesRequest{},
				func(r *SynchronizeRolesRequest, p *PaginationOptions) { r.Pagination = p },
				func(r SynchronizeRolesResponse) *PaginationOptions { return r.Pagination },
			)
			errChan.Send(ctx, err)
		})
	}

	// Synchronize Permissions
	if shouldSync(SynchronizePermissions) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			err := runSyncLoop(ctx, syncReq.ProviderIdentifier, (*ProviderActivities).SynchronizePermissions, SynchronizePermissionsRequest{},
				func(r *SynchronizePermissionsRequest, p *PaginationOptions) { r.Pagination = p },
				func(r SynchronizePermissionsResponse) *PaginationOptions { return r.Pagination },
			)
			errChan.Send(ctx, err)
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
		return nil, fmt.Errorf("synchronization failed: %v", errs)
	}

	return &SynchronizeResponse{}, nil
}
