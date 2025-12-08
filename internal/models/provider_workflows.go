package models

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"go.temporal.io/api/enums/v1"
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
	Upstream           *model.Endpoint         `json:"upstream,omitempty"`
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

func SynchronizeUpstreamWorkflow(ctx workflow.Context, upstream *model.Endpoint, providerIdentifier string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var startResp SynchronizeStartResponse
	err := workflow.ExecuteActivity(
		ctx,
		CreateTemporalIdentifier(
			providerIdentifier,
			GetNameFromFunction((*ProviderActivities).SynchronizeThandStart),
		),
		upstream,
		providerIdentifier,
	).Get(ctx, &startResp)

	if err != nil {
		return err
	}

	upstreamWorkflowID := startResp.WorkflowID

	chunkChan := workflow.GetSignalChannel(ctx, "chunk")
	commitChan := workflow.GetSignalChannel(ctx, "commit")

	for {
		var chunk SynchronizeChunkRequest
		var commit bool

		selector := workflow.NewSelector(ctx)

		selector.AddReceive(chunkChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &chunk)
		})

		selector.AddReceive(commitChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			commit = true
		})

		selector.Select(ctx)

		if commit {
			break
		}

		err := workflow.ExecuteActivity(
			ctx,
			CreateTemporalIdentifier(
				providerIdentifier,
				GetNameFromFunction((*ProviderActivities).SynchronizeThandChunk),
			),
			upstream,
			providerIdentifier,
			upstreamWorkflowID,
			chunk,
		).Get(ctx, nil)

		if err != nil {
			workflow.GetLogger(ctx).Error("Failed to sync chunk", "error", err)
		}
	}

	return workflow.ExecuteActivity(
		ctx,
		CreateTemporalIdentifier(
			providerIdentifier,
			GetNameFromFunction((*ProviderActivities).SynchronizeThandCommit),
		),
		upstream,
		providerIdentifier,
		upstreamWorkflowID,
	).Get(ctx, nil)
}

func handleUpstreamSync(
	ctx workflow.Context,
	upstreamChan workflow.Channel,
	upstreamDone workflow.Channel,
	childFuture workflow.ChildWorkflowFuture,
) {
	defer upstreamDone.Close()

	var execution workflow.Execution
	if err := childFuture.GetChildWorkflowExecution().Get(ctx, &execution); err != nil {
		workflow.GetLogger(ctx).Error("Failed to start upstream workflow: ", err)
		// Drain channel
		for {
			var ignored any
			if !upstreamChan.Receive(ctx, &ignored) {
				break
			}
		}
		return
	}

	// Buffer for aggregating chunks
	var buffer SynchronizeChunkRequest
	bufferCount := 0
	const batchSize = 100
	const batchTimeout = 500 * time.Millisecond

	// Timer control
	timerCtx, cancelTimer := workflow.WithCancel(ctx)
	timer := workflow.NewTimer(timerCtx, batchTimeout)

	flush := func() {
		if bufferCount > 0 {
			err := childFuture.SignalChildWorkflow(ctx, "chunk", buffer).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("Failed to signal chunk: ", err)
			}
			// Reset buffer
			buffer = SynchronizeChunkRequest{}
			bufferCount = 0
		}
		// Reset timer
		cancelTimer()
		timerCtx, cancelTimer = workflow.WithCancel(ctx)
		timer = workflow.NewTimer(timerCtx, batchTimeout)
	}

	channelOpen := true
	for channelOpen {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(upstreamChan, func(c workflow.ReceiveChannel, more bool) {
			if !more {
				channelOpen = false
				return
			}

			var chunk SynchronizeChunkRequest
			c.Receive(ctx, &chunk)

			// Merge chunk into buffer
			if len(chunk.Roles) > 0 {
				buffer.Roles = append(buffer.Roles, chunk.Roles...)
				bufferCount += len(chunk.Roles)
			}
			if len(chunk.Permissions) > 0 {
				buffer.Permissions = append(buffer.Permissions, chunk.Permissions...)
				bufferCount += len(chunk.Permissions)
			}
			if len(chunk.Resources) > 0 {
				buffer.Resources = append(buffer.Resources, chunk.Resources...)
				bufferCount += len(chunk.Resources)
			}
			if len(chunk.Identities) > 0 {
				buffer.Identities = append(buffer.Identities, chunk.Identities...)
				bufferCount += len(chunk.Identities)
			}
			if len(chunk.Users) > 0 {
				buffer.Users = append(buffer.Users, chunk.Users...)
				bufferCount += len(chunk.Users)
			}
			if len(chunk.Groups) > 0 {
				buffer.Groups = append(buffer.Groups, chunk.Groups...)
				bufferCount += len(chunk.Groups)
			}

			if bufferCount >= batchSize {
				flush()
			}
		})

		selector.AddFuture(timer, func(f workflow.Future) {
			flush()
		})

		selector.Select(ctx)
	}

	// Flush any remaining items after channel closes
	flush()

	dCtx, _ := workflow.NewDisconnectedContext(ctx)
	err := childFuture.SignalChildWorkflow(dCtx, "commit", nil).Get(dCtx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Error("Failed to signal commit: ", err)
	}
}

type PaginatedRequest interface {
	SetPagination(*PaginationOptions)
}

type PaginatedResponse interface {
	GetPagination() *PaginationOptions
}

type workflowSyncer struct {
	providerIdentifier string
	upstreamChan       workflow.Channel
	errChan            workflow.Channel
}

type syncLoopParams[Req any, Resp PaginatedResponse, Item any] struct {
	ActivityMethod any
	InitialReq     Req

	GetItems   func(Resp) []Item
	Accumulate func([]Item)
	ToChunk    func([]Item) SynchronizeChunkRequest
}

func runSyncLoop[Req any, PReq interface {
	*Req
	PaginatedRequest
}, Resp PaginatedResponse, Item any](ctx workflow.Context, w *workflowSyncer, params syncLoopParams[Req, Resp, Item]) {
	req := params.InitialReq
	pendingSends := 0
	sendsDone := workflow.NewChannel(ctx)

	waitSends := func() {
		for pendingSends > 0 {
			sendsDone.Receive(ctx, nil)
			pendingSends--
		}
	}

	for {
		var resp Resp
		err := workflow.ExecuteActivity(
			ctx,
			CreateTemporalIdentifier(
				w.providerIdentifier,
				GetNameFromFunction(params.ActivityMethod),
			),
			req,
		).Get(ctx, &resp)

		if err != nil {
			waitSends()
			w.errChan.Send(ctx, err)
			return
		}

		items := params.GetItems(resp)
		params.Accumulate(items)

		if w.upstreamChan != nil {
			chunk := params.ToChunk(items)
			pendingSends++
			workflow.Go(ctx, func(ctx workflow.Context) {
				w.upstreamChan.Send(ctx, chunk)
				sendsDone.Send(ctx, true)
			})
		}

		pagination := resp.GetPagination()
		if pagination == nil || len(pagination.Token) == 0 {
			break
		}
		PReq(&req).SetPagination(pagination)
	}
	waitSends()
	w.errChan.Send(ctx, nil)
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

	// Channel for upstream chunks
	upstreamChan := workflow.NewChannel(ctx)
	upstreamDone := workflow.NewChannel(ctx)

	defer func() {
		upstreamChan.Close()
		if syncReq.Upstream != nil {
			dCtx, _ := workflow.NewDisconnectedContext(ctx)
			upstreamDone.Receive(dCtx, nil)
		}
	}()

	if syncReq.Upstream != nil {

		// When the SynchronizeWorkflow the upstream synchronization workflow
		// will continue to run
		cwo := workflow.ChildWorkflowOptions{
			ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
		}

		childCtx := workflow.WithChildOptions(ctx, cwo)
		childFuture := workflow.ExecuteChildWorkflow(
			childCtx,
			SynchronizeUpstreamWorkflow,
			syncReq.Upstream,
			syncReq.ProviderIdentifier,
		)

		workflow.Go(ctx, func(ctx workflow.Context) {
			handleUpstreamSync(ctx, upstreamChan, upstreamDone, childFuture)
		})
	}

	// Execute all the synchronizations needed for RBAC
	// in parallel using the workflow parallel pattern

	syncResponse := &SynchronizeResponse{}
	errChan := workflow.NewChannel(ctx)
	syncCount := 0

	syncer := &workflowSyncer{
		providerIdentifier: syncReq.ProviderIdentifier,
		upstreamChan: func() workflow.Channel {
			if syncReq.Upstream != nil {
				return upstreamChan
			}
			return nil
		}(),
		errChan: errChan,
	}

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
			runSyncLoop(ctx, syncer, syncLoopParams[SynchronizeIdentitiesRequest, SynchronizeIdentitiesResponse, Identity]{
				ActivityMethod: (*ProviderActivities).SynchronizeIdentities,
				InitialReq:     SynchronizeIdentitiesRequest{},
				GetItems:       func(r SynchronizeIdentitiesResponse) []Identity { return r.Identities },
				Accumulate:     func(items []Identity) { syncResponse.Identities = append(syncResponse.Identities, items...) },
				ToChunk:        func(items []Identity) SynchronizeChunkRequest { return SynchronizeChunkRequest{Identities: items} },
			})
		})
	}

	// Synchronize Users
	if shouldSync(SynchronizeUsers) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			runSyncLoop(ctx, syncer, syncLoopParams[SynchronizeUsersRequest, SynchronizeUsersResponse, Identity]{
				ActivityMethod: (*ProviderActivities).SynchronizeUsers,
				InitialReq:     SynchronizeUsersRequest{},
				GetItems:       func(r SynchronizeUsersResponse) []Identity { return r.Identities },
				Accumulate:     func(items []Identity) { syncResponse.Identities = append(syncResponse.Identities, items...) },
				ToChunk:        func(items []Identity) SynchronizeChunkRequest { return SynchronizeChunkRequest{Identities: items} },
			})
		})
	}

	// Synchronize Groups
	if shouldSync(SynchronizeGroups) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			runSyncLoop(ctx, syncer, syncLoopParams[SynchronizeGroupsRequest, SynchronizeGroupsResponse, Identity]{
				ActivityMethod: (*ProviderActivities).SynchronizeGroups,
				InitialReq:     SynchronizeGroupsRequest{},
				GetItems:       func(r SynchronizeGroupsResponse) []Identity { return r.Identities },
				Accumulate:     func(items []Identity) { syncResponse.Identities = append(syncResponse.Identities, items...) },
				ToChunk:        func(items []Identity) SynchronizeChunkRequest { return SynchronizeChunkRequest{Identities: items} },
			})
		})
	}

	// Synchronize Resources
	if shouldSync(SynchronizeResources) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			runSyncLoop(ctx, syncer, syncLoopParams[SynchronizeResourcesRequest, SynchronizeResourcesResponse, ProviderResource]{
				ActivityMethod: (*ProviderActivities).SynchronizeResources,
				InitialReq:     SynchronizeResourcesRequest{},
				GetItems:       func(r SynchronizeResourcesResponse) []ProviderResource { return r.Resources },
				Accumulate:     func(items []ProviderResource) { syncResponse.Resources = append(syncResponse.Resources, items...) },
				ToChunk: func(items []ProviderResource) SynchronizeChunkRequest {
					return SynchronizeChunkRequest{Resources: items}
				},
			})
		})
	}

	// Synchronize Roles
	if shouldSync(SynchronizeRoles) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			runSyncLoop(ctx, syncer, syncLoopParams[SynchronizeRolesRequest, SynchronizeRolesResponse, ProviderRole]{
				ActivityMethod: (*ProviderActivities).SynchronizeRoles,
				InitialReq:     SynchronizeRolesRequest{},
				GetItems:       func(r SynchronizeRolesResponse) []ProviderRole { return r.Roles },
				Accumulate:     func(items []ProviderRole) { syncResponse.Roles = append(syncResponse.Roles, items...) },
				ToChunk:        func(items []ProviderRole) SynchronizeChunkRequest { return SynchronizeChunkRequest{Roles: items} },
			})
		})
	}

	// Synchronize Permissions
	if shouldSync(SynchronizePermissions) {
		syncCount++
		workflow.Go(ctx, func(ctx workflow.Context) {
			runSyncLoop(ctx, syncer, syncLoopParams[SynchronizePermissionsRequest, SynchronizePermissionsResponse, ProviderPermission]{
				ActivityMethod: (*ProviderActivities).SynchronizePermissions,
				InitialReq:     SynchronizePermissionsRequest{},
				GetItems:       func(r SynchronizePermissionsResponse) []ProviderPermission { return r.Permissions },
				Accumulate: func(items []ProviderPermission) {
					syncResponse.Permissions = append(syncResponse.Permissions, items...)
				},
				ToChunk: func(items []ProviderPermission) SynchronizeChunkRequest {
					return SynchronizeChunkRequest{Permissions: items}
				},
			})
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
