package thand

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	emailProvider "github.com/thand-io/agent/internal/providers/email"
	slackProvider "github.com/thand-io/agent/internal/providers/slack"
	thandFunction "github.com/thand-io/agent/internal/workflows/functions/providers/thand"
	taskModel "github.com/thand-io/agent/internal/workflows/tasks/model"
	"go.temporal.io/sdk/workflow"
)

const ThandNotifyTask = "notify"
const ThandApprovalEventType = "com.thand.approval"

type NotifyRequest struct {
	Approvals int                           `json:"approvals" default:"1"`
	Notifier  thandFunction.NotifierRequest `json:"notifier"`

	// Internal use only: entrypoint for resuming workflow
	Entrypoint string `json:"entrypoint"`
}

func (n *NotifyRequest) IsValid() bool {

	if n.Approvals == 0 {
		return false
	}

	if !n.Notifier.IsValid() {
		return false
	}

	return true
}

func (n *NotifyRequest) AsMap() map[string]any {
	return map[string]any{
		"approvals":  n.Approvals,
		"notifier":   n.Notifier.AsMap(),
		"entrypoint": n.Entrypoint,
	}
}

func (t *thandTask) executeNotifyTask(
	workflowTask *models.WorkflowTask,
	taskName string,
	call *taskModel.ThandTask,
) (any, error) {

	req := workflowTask.GetContextAsMap()

	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	var notifyReq NotifyRequest
	common.ConvertInterfaceToInterface(call.With, &notifyReq)

	if !notifyReq.IsValid() {
		return nil, errors.New("invalid notification request")
	}

	notificationReq := notifyReq.Notifier

	notifierProviders := t.config.GetProvidersByCapability(
		models.ProviderCapabilityNotifier)

	if !hasMatchingProvider(notificationReq, notifierProviders) {
		return nil, fmt.Errorf("no matching provider found for name: %s", notificationReq.Provider)
	}

	elevationReq, err := workflowTask.GetContextAsElevationRequest()

	if err != nil {
		return nil, fmt.Errorf("failed to get elevation request from input: %w", err)
	}

	if !elevationReq.IsValid() {
		return nil, errors.New("elevation request is not valid")
	}

	foundProvider := notificationReq.Provider

	if len(foundProvider) == 0 {
		return nil, errors.New("provider must be specified in the with block")
	}

	// Get server config to fetch provider
	providerConfig, err := t.config.Providers.GetProviderByName(foundProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"provider": providerConfig.Name,
	}).Info("Executing notification")

	var notificationPayload models.NotificationRequest

	switch providerConfig.Provider {
	case "slack":
		blocks := t.createSlackBlocks(workflowTask, elevationReq, &notifyReq)

		slackReq := slackProvider.SlackNotificationRequest{
			To: notificationReq.To,
			Text: fmt.Sprintf("Access request for role %s", func() string {
				if elevationReq.Role != nil {
					return elevationReq.Role.Name
				}
				return "unknown"
			}()),
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
		}
		err = common.ConvertInterfaceToInterface(slackReq, &notificationPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to convert slack request: %w", err)
		}
	case "email":
		emailReq := emailProvider.EmailNotificationRequest{
			To:      notificationReq.To,
			Subject: "Workflow Notification",
			Body: emailProvider.EmailNotificationBody{
				Text: notificationReq.Message,
				HTML: notificationReq.Message,
			},
		}
		err = common.ConvertInterfaceToInterface(emailReq, &notificationPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to convert email request: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Provider)
	}

	logrus.WithFields(logrus.Fields{
		"input": call.With,
	}).Info("Sending notification")

	if workflowTask.HasTemporalContext() {

		temporalContext := workflowTask.GetTemporalContext()

		serviceClient := t.config.GetServices()

		ao := workflow.ActivityOptions{
			TaskQueue:           serviceClient.GetTemporal().GetTaskQueue(),
			StartToCloseTimeout: time.Minute * 5,
		}
		aoctx := workflow.WithActivityOptions(temporalContext, ao)

		// Use Temporal activity to send notification
		err = workflow.ExecuteActivity(
			aoctx,
			thandFunction.ThandNotifyFunction,

			// args
			workflowTask,
			taskName,
			model.CallFunction{
				Call: thandFunction.ThandNotifyFunction,
				With: notificationReq.AsMap(),
			},
			notificationPayload,
		).Get(temporalContext, nil)

		if err != nil {
			return nil, fmt.Errorf("failed to send notification: %w", err)
		}

	} else {

		err = providerConfig.GetClient().SendNotification(
			workflowTask.GetContext(),
			notificationPayload,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return nil, nil
}

func hasMatchingProvider(notificationReq thandFunction.NotifierRequest, notifierProviders map[string]models.Provider) bool {

	// filter out providers to see if the name matches
	for _, provider := range notifierProviders {
		if strings.Compare(provider.Name, notificationReq.Provider) == 0 {
			return true
		} else if strings.Compare(provider.Provider, notificationReq.Provider) == 0 {
			return true
		}
	}

	return false
}

// createSlackBlocks creates the Slack Block Kit blocks for the notification
func (t *thandTask) createSlackBlocks(
	workflowTask *models.WorkflowTask,
	elevateRequest *models.ElevateRequestInternal,
	notifyReq *NotifyRequest,
) []slack.Block {
	blocks := []slack.Block{}

	// Add the user message section
	t.addUserMessageSection(&blocks, notifyReq)

	// Add divider
	blocks = append(blocks, slack.NewDividerBlock())

	// Add request details section
	t.addRequestDetailsSection(&blocks, elevateRequest)

	// Add identities section
	t.addIdentitiesSection(&blocks, elevateRequest)

	// Add inherited roles section
	t.addInheritedRolesSection(&blocks, elevateRequest)

	// Add permissions section
	t.addPermissionsSection(&blocks, elevateRequest)

	// Add resources section
	t.addResourcesSection(&blocks, elevateRequest)

	// Add user information section
	t.addUserInfoSection(&blocks, elevateRequest)

	// Add divider before action section
	blocks = append(blocks, slack.NewDividerBlock())

	// Add action section
	t.addActionSection(&blocks, workflowTask, notifyReq)

	return blocks
}

// addUserMessageSection adds the user message block if message is provided
func (t *thandTask) addUserMessageSection(blocks *[]slack.Block, notifyReq *NotifyRequest) {
	if len(notifyReq.Notifier.Message) > 0 {
		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				notifyReq.Notifier.Message,
				false,
				false,
			),
			nil,
			nil,
		))
	}
}

// addRequestDetailsSection builds and adds the request details section
func (t *thandTask) addRequestDetailsSection(blocks *[]slack.Block, elevateRequest *models.ElevateRequestInternal) {
	var requestDetailsText strings.Builder
	requestDetailsText.WriteString("*Access Request Details:*\n")

	if elevateRequest.Role != nil {
		requestDetailsText.WriteString(fmt.Sprintf("• *Role:* %s\n", elevateRequest.Role.Name))
		if len(elevateRequest.Role.Description) > 0 {
			requestDetailsText.WriteString(fmt.Sprintf("• *Description:* %s\n", elevateRequest.Role.Description))
		}
	}

	if len(elevateRequest.Providers) > 0 {
		requestDetailsText.WriteString(fmt.Sprintf("• *Providers:* %s\n", strings.Join(elevateRequest.Providers, ", ")))
	}

	if len(elevateRequest.Reason) > 0 {
		requestDetailsText.WriteString(fmt.Sprintf("• *Reason:* %s\n", elevateRequest.Reason))
	}

	if len(elevateRequest.Duration) > 0 {
		requestDetailsText.WriteString(fmt.Sprintf("• *Duration:* %s\n", elevateRequest.Duration))
	}

	*blocks = append(*blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			requestDetailsText.String(),
			false,
			false,
		),
		nil,
		nil,
	))
}

// addInheritedRolesSection adds inherited roles section if available
func (t *thandTask) addInheritedRolesSection(blocks *[]slack.Block, elevateRequest *models.ElevateRequestInternal) {
	if elevateRequest.Role != nil && len(elevateRequest.Role.Inherits) > 0 {
		var inheritsText strings.Builder
		inheritsText.WriteString("*Inherited Roles:*\n")

		for _, inheritedRole := range elevateRequest.Role.Inherits {
			inheritsText.WriteString(fmt.Sprintf("• %s\n", inheritedRole))
		}

		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				inheritsText.String(),
				false,
				false,
			),
			nil,
			nil,
		))
	}
}

// addPermissionsSection adds permissions section if available
func (t *thandTask) addPermissionsSection(blocks *[]slack.Block, elevateRequest *models.ElevateRequestInternal) {
	if elevateRequest.Role != nil && (len(elevateRequest.Role.Permissions.Allow) > 0 || len(elevateRequest.Role.Permissions.Deny) > 0) {
		var permissionsText strings.Builder
		permissionsText.WriteString("*Permissions:*\n")

		if len(elevateRequest.Role.Permissions.Allow) > 0 {
			permissionsText.WriteString("*Allowed:*\n")
			for _, perm := range elevateRequest.Role.Permissions.Allow {
				permissionsText.WriteString(fmt.Sprintf("- %s\n", perm))
			}
		}

		if len(elevateRequest.Role.Permissions.Deny) > 0 {
			permissionsText.WriteString("*Denied:*\n")
			for _, perm := range elevateRequest.Role.Permissions.Deny {
				permissionsText.WriteString(fmt.Sprintf("- %s\n", perm))
			}
		}

		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				permissionsText.String(),
				false,
				false,
			),
			nil,
			nil,
		))
	}
}

// addResourcesSection adds resources section if available
func (t *thandTask) addResourcesSection(blocks *[]slack.Block, elevateRequest *models.ElevateRequestInternal) {
	if elevateRequest.Role != nil && (len(elevateRequest.Role.Resources.Allow) > 0 || len(elevateRequest.Role.Resources.Deny) > 0) {
		var resourcesText strings.Builder
		resourcesText.WriteString("*Resources:*\n")

		if len(elevateRequest.Role.Resources.Allow) > 0 {
			resourcesText.WriteString("*Allowed:*\n")
			for _, resource := range elevateRequest.Role.Resources.Allow {
				resourcesText.WriteString(fmt.Sprintf("- %s\n", resource))
			}
		}

		if len(elevateRequest.Role.Resources.Deny) > 0 {
			resourcesText.WriteString("*Denied:*\n")
			for _, resource := range elevateRequest.Role.Resources.Deny {
				resourcesText.WriteString(fmt.Sprintf("- %s\n", resource))
			}
		}

		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				resourcesText.String(),
				false,
				false,
			),
			nil,
			nil,
		))
	}
}

// addUserInfoSection adds user information section if available
func (t *thandTask) addUserInfoSection(blocks *[]slack.Block, elevateRequest *models.ElevateRequestInternal) {
	if elevateRequest.User != nil {
		var userText strings.Builder
		userText.WriteString("*Requested by:*\n")
		userText.WriteString(fmt.Sprintf("*User:* %s\n", elevateRequest.User.Name))
		if len(elevateRequest.User.Email) > 0 {
			userText.WriteString(fmt.Sprintf("*Email:* %s\n", elevateRequest.User.Email))
		}

		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				userText.String(),
				false,
				false,
			),
			nil,
			nil,
		))
	}
}

// addIdentitiesSection adds identities section if available
func (t *thandTask) addIdentitiesSection(blocks *[]slack.Block, elevateRequest *models.ElevateRequestInternal) {
	if len(elevateRequest.Identities) > 0 {
		var identitiesText strings.Builder
		identitiesText.WriteString("*Target Identities:*\n")

		for _, identity := range elevateRequest.Identities {
			identitiesText.WriteString(fmt.Sprintf("• %s\n", identity))
		}

		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				identitiesText.String(),
				false,
				false,
			),
			nil,
			nil,
		))
	}
}

// addActionSection adds action buttons section based on approval requirements
func (t *thandTask) addActionSection(
	blocks *[]slack.Block,
	workflowTask *models.WorkflowTask,
	notifyReq *NotifyRequest,
) {
	if notifyReq.Approvals > 0 {
		// Get current approvals from workflow context
		workflowContext := workflowTask.GetContextAsMap()
		approvals, ok := workflowContext["approvals"].([]any)
		if !ok {
			approvals = []any{}
		}

		// Count existing approved approvals
		approvedCount := 0
		for _, approval := range approvals {
			if approvalMap, ok := approval.(map[string]any); ok {
				if approved, exists := approvalMap["approved"]; exists {
					if approvedBool, ok := approved.(bool); ok && approvedBool {
						approvedCount++
					}
				}
			}
		}

		remainingApprovals := notifyReq.Approvals - approvedCount

		// Create dynamic message based on approval requirements
		var actionMessage string
		if notifyReq.Approvals == 1 {
			actionMessage = "*Action Required:*\n*One approval is required.* Please review the request and choose an action."
		} else if remainingApprovals <= 0 {
			actionMessage = "*Action Required:*\n*Sufficient approvals have been received.* Please review the request and choose an action."
		} else if remainingApprovals == 1 {
			actionMessage = fmt.Sprintf("*Action Required:*\n*%d more approval is needed (%d of %d received).* Please review the request and choose an action.", remainingApprovals, approvedCount, notifyReq.Approvals)
		} else {
			actionMessage = fmt.Sprintf("*Action Required:*\n*%d more approvals are needed (%d of %d received).* Please review the request and choose an action.", remainingApprovals, approvedCount, notifyReq.Approvals)
		}

		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				actionMessage,
				false,
				false,
			),
			nil,
			nil,
		))

		if remainingApprovals > 0 {

			*blocks = append(*blocks, slack.NewActionBlock(
				"",
				slack.NewButtonBlockElement(
					"approve",
					"Approve",
					slack.NewTextBlockObject(
						slack.PlainTextType,
						"✅ Approve",
						false,
						false,
					),
				).WithURL(t.createCallbackUrl(workflowTask, notifyReq, true)).WithStyle(slack.StylePrimary),
				slack.NewButtonBlockElement(
					"revoke",
					"Revoke",
					slack.NewTextBlockObject(
						slack.PlainTextType,
						"❌ Revoke",
						false,
						false,
					),
				).WithURL(t.createCallbackUrl(workflowTask, notifyReq, false)).WithStyle(slack.StyleDanger),
			))

		}
	} else {
		*blocks = append(*blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				"No action is required. This is a notification only.",
				false,
				false,
			),
			nil,
			nil,
		))
	}
}

func (t *thandTask) createCallbackUrl(
	workflowTask *models.WorkflowTask,
	notifyReq *NotifyRequest,
	approve bool,
) string {

	// Create an Event.
	event := cloudevents.NewEvent()
	event.SetSource("thand/agent")
	event.SetType(ThandApprovalEventType)
	event.SetData(cloudevents.ApplicationJSON, map[string]any{
		"approved": approve,
		"user":     "",
	})

	// Setup workflow for the next state
	signaledWorkflow := workflowTask.Clone().(*models.WorkflowTask)
	signaledWorkflow.SetInput(&event)

	if len(notifyReq.Entrypoint) > 0 {
		signaledWorkflow.SetEntrypoint(notifyReq.Entrypoint)
	}

	callbackUrl := t.config.GetResumeCallbackUrl(signaledWorkflow)

	return callbackUrl
}
