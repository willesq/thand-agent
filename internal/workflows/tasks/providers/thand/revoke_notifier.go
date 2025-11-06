package thand

import (
	"fmt"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/config"
	"github.com/thand-io/agent/internal/models"
	emailProvider "github.com/thand-io/agent/internal/providers/email"
	slackProvider "github.com/thand-io/agent/internal/providers/slack"
	thandFunction "github.com/thand-io/agent/internal/workflows/functions/providers/thand"
)

// revokeNotifier handles notifications sent to users after their access has been revoked
type revokeNotifier struct {
	config       *config.Config
	workflowTask *models.WorkflowTask
	elevationReq *models.ElevateRequestInternal
	req          *thandFunction.NotifierRequest
	providerKey  string
	revocations  map[string]any
}

// NewRevokeNotifier creates a new notifier for sending revocation notifications
func NewRevokeNotifier(
	config *config.Config,
	workflowTask *models.WorkflowTask,
	elevationReq *models.ElevateRequestInternal,
	req *thandFunction.NotifierRequest,
	providerKey string,
	revocations map[string]any,
) NotifierImpl {
	return &revokeNotifier{
		config:       config,
		workflowTask: workflowTask,
		elevationReq: elevationReq,
		req:          req,
		providerKey:  providerKey,
		revocations:  revocations,
	}
}

func (r *revokeNotifier) GetRecipients() []string {
	return r.req.To
}

func (r *revokeNotifier) GetCallFunction(toIdentity string) model.CallFunction {

	callMap := (&thandFunction.NotifierRequest{
		Provider: r.req.Provider,
		To:       []string{toIdentity},
	}).AsMap()

	return model.CallFunction{
		Call: thandFunction.ThandNotifyFunction,
		With: callMap,
	}
}

func (r *revokeNotifier) GetProviderName() string {
	return r.req.Provider
}

func (r *revokeNotifier) GetPayload(toIdentity string) models.NotificationRequest {

	elevationReq := r.elevationReq
	var notificationPayload models.NotificationRequest

	switch r.GetProviderName() {
	case "slack":
		blocks := r.createRevokeSlackBlocks()

		slackReq := slackProvider.SlackNotificationRequest{
			To: toIdentity,
			Text: fmt.Sprintf("Your access for role %s has been revoked", func() string {
				if elevationReq.Role != nil {
					return elevationReq.Role.Name
				}
				return "unknown"
			}()),
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
		}
		err := common.ConvertInterfaceToInterface(slackReq, &notificationPayload)
		if err != nil {
			logrus.WithError(err).Error("Failed to convert slack request")
			return models.NotificationRequest{}
		}
	case "email":
		plainText, html := r.createRevokeEmailBody()
		emailReq := emailProvider.EmailNotificationRequest{
			To:      []string{toIdentity},
			Subject: "Access Revoked",
			Body: emailProvider.EmailNotificationBody{
				Text: plainText,
				HTML: html,
			},
		}
		err := common.ConvertInterfaceToInterface(emailReq, &notificationPayload)
		if err != nil {
			logrus.WithError(err).Error("Failed to convert email request")
			return models.NotificationRequest{}
		}
	default:
		logrus.WithField("provider", r.GetProviderName()).Error("Unsupported provider type")
		return models.NotificationRequest{}
	}

	return notificationPayload
}
