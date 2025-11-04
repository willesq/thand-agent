package thand

import (
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	emailProvider "github.com/thand-io/agent/internal/providers/email"
	slackProvider "github.com/thand-io/agent/internal/providers/slack"
	thandFunction "github.com/thand-io/agent/internal/workflows/functions/providers/thand"
)

type NotifierImpl interface {
	GetProviderName() string
	GetRecipients() []string
	GetCallFunction(toIdentity string) model.CallFunction
	GetPayload(toIdentity string) models.NotificationRequest
}

type defaultNotifierImpl struct {
	req thandFunction.NotifierRequest
}

func NewDefaultNotifierImpl(req thandFunction.NotifierRequest) NotifierImpl {
	return &defaultNotifierImpl{
		req: req,
	}
}

func (d *defaultNotifierImpl) GetRecipients() []string {
	return d.req.To
}

func (d *defaultNotifierImpl) GetCallFunction(toIdentity string) model.CallFunction {

	callMap := (&thandFunction.NotifierRequest{
		Provider: d.req.Provider,
		To:       []string{toIdentity},
	}).AsMap()

	return model.CallFunction{
		Call: thandFunction.ThandNotifyFunction,
		With: callMap,
	}
}

func (d *defaultNotifierImpl) GetProviderName() string {
	return d.req.Provider
}

func (d *defaultNotifierImpl) GetPayload(toIdentity string) models.NotificationRequest {

	switch d.GetProviderName() {
	case "slack":
		return d.GetSlackPayload(toIdentity)
	case "email":
		return d.GetEmailPayload(toIdentity)
	default:
		return models.NotificationRequest{}
	}

}

func (d *defaultNotifierImpl) GetEmailPayload(toIdentity string) models.NotificationRequest {

	notificationReq := d.req

	// Render HTML email using template
	html, err := RenderEmail("Workflow Notification", notificationReq.Message)
	if err != nil {
		logrus.WithError(err).Error("Failed to render email template")
		// Fallback to plain message if template fails
		// TODO: format markdown
		html = notificationReq.Message
	}

	emailReq := emailProvider.EmailNotificationRequest{
		To:      []string{toIdentity},
		Subject: "Workflow Notification",
		Body: emailProvider.EmailNotificationBody{
			Text: notificationReq.Message,
			HTML: html,
		},
	}

	var notificationPayload models.NotificationRequest
	err = common.ConvertInterfaceToInterface(emailReq, &notificationPayload)

	if err != nil {
		logrus.WithError(err).Error("Failed to convert email request")
		return models.NotificationRequest{}
	}

	return notificationPayload
}

func (d *defaultNotifierImpl) GetSlackPayload(toIdentity string) models.NotificationRequest {

	notificationReq := d.req

	slackReq := slackProvider.SlackNotificationRequest{
		To:   toIdentity,
		Text: notificationReq.Message,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", notificationReq.Message, false, false),
					nil,
					nil,
				),
			},
		},
	}

	var notificationPayload models.NotificationRequest
	err := common.ConvertInterfaceToInterface(slackReq, &notificationPayload)

	if err != nil {
		logrus.WithError(err).Error("Failed to convert slack request")
		return models.NotificationRequest{}
	}

	return notificationPayload
}
