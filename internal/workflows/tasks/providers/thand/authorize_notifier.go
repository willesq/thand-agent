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

// authorizerNotifier handles notifications sent to users after their access request has been approved
type authorizerNotifier struct {
	config        *config.Config
	workflowTask  *models.WorkflowTask
	elevationReq  *models.ElevateRequestInternal
	req           *thandFunction.NotifierRequest
	providerKey   string
	authRequests  map[string]*models.AuthorizeRoleRequest
	authResponses map[string]*models.AuthorizeRoleResponse
}

// NewAuthorizerNotifier creates a new notifier for sending approval confirmation notifications
func NewAuthorizerNotifier(
	config *config.Config,
	workflowTask *models.WorkflowTask,
	elevationReq *models.ElevateRequestInternal,
	req *thandFunction.NotifierRequest,
	providerKey string,
	requests map[string]*models.AuthorizeRoleRequest,
	authorizations map[string]*models.AuthorizeRoleResponse,
) NotifierImpl {
	return &authorizerNotifier{
		config:        config,
		workflowTask:  workflowTask,
		elevationReq:  elevationReq,
		req:           req,
		providerKey:   providerKey,
		authRequests:  requests,
		authResponses: authorizations,
	}
}

func (a *authorizerNotifier) GetRecipients() []string {
	return a.req.To
}

func (a *authorizerNotifier) GetCallFunction(toIdentity string) model.CallFunction {

	callMap := (&thandFunction.NotifierRequest{
		Provider: a.req.Provider,
		To:       []string{toIdentity},
	}).AsMap()

	return model.CallFunction{
		Call: thandFunction.ThandNotifyFunction,
		With: callMap,
	}
}

func (a *authorizerNotifier) GetProviderName() string {
	return a.req.Provider
}

func (a *authorizerNotifier) GetPayload(toIdentity string) models.NotificationRequest {

	elevationReq := a.elevationReq
	var notificationPayload models.NotificationRequest

	switch a.GetProviderName() {
	case "slack":
		blocks := a.createAuthorizeSlackBlocks()

		slackReq := slackProvider.SlackNotificationRequest{
			To: toIdentity,
			Text: fmt.Sprintf("Your access request for role %s has been approved", func() string {
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
		plainText, html := a.createAuthorizeEmailBody()
		emailReq := emailProvider.EmailNotificationRequest{
			To:      []string{toIdentity},
			Subject: "Access Request Approved",
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
		logrus.WithField("provider", a.GetProviderName()).Error("Unsupported provider type")
		return models.NotificationRequest{}
	}

	return notificationPayload
}
