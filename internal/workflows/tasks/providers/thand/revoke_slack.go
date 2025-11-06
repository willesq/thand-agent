package thand

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/thand-io/agent/internal/models"
	thandFunction "github.com/thand-io/agent/internal/workflows/functions/providers/thand"
)

// createRevokeSlackBlocks creates the Slack Block Kit blocks for the revocation notification
func (r *revokeNotifier) createRevokeSlackBlocks() []slack.Block {

	notifyReq := r.req

	blocks := []slack.Block{}

	// Add the revocation message
	r.addRevokeMessageSection(&blocks, notifyReq)

	// Add divider
	blocks = append(blocks, slack.NewDividerBlock())

	// Add revocation details section
	r.addRevokeDetailsSection(&blocks, r.elevationReq)

	// Add divider before closing message
	blocks = append(blocks, slack.NewDividerBlock())

	// Add closing message
	r.addRevokeClosingMessageSection(&blocks)

	return blocks
}

// addRevokeMessageSection adds the revocation message block
func (r *revokeNotifier) addRevokeMessageSection(blocks *[]slack.Block, notifyReq *thandFunction.NotifierRequest) {
	message := "*Your access has been revoked*"

	if len(notifyReq.Message) > 0 {
		message = notifyReq.Message
	}

	*blocks = append(*blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			message,
			false,
			false,
		),
		nil,
		nil,
	))
}

// addRevokeDetailsSection builds and adds the revocation details section
func (r *revokeNotifier) addRevokeDetailsSection(blocks *[]slack.Block, elevateRequest *models.ElevateRequestInternal) {
	var revokeDetailsText strings.Builder
	revokeDetailsText.WriteString("*Revoked Access Details:*\n")

	if elevateRequest.Role != nil {
		revokeDetailsText.WriteString(fmt.Sprintf("• *Role:* %s\n", elevateRequest.Role.Name))
		if len(elevateRequest.Role.Description) > 0 {
			revokeDetailsText.WriteString(fmt.Sprintf("• *Description:* %s\n", elevateRequest.Role.Description))
		}
	}

	if len(elevateRequest.Providers) > 0 {
		revokeDetailsText.WriteString(fmt.Sprintf("• *Providers:* %s\n", strings.Join(elevateRequest.Providers, ", ")))
	}

	if len(elevateRequest.Duration) > 0 {
		revokeDetailsText.WriteString(fmt.Sprintf("• *Duration:* %s\n", elevateRequest.Duration))
	}

	*blocks = append(*blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			revokeDetailsText.String(),
			false,
			false,
		),
		nil,
		nil,
	))
}

// addRevokeClosingMessageSection adds a closing message
func (r *revokeNotifier) addRevokeClosingMessageSection(blocks *[]slack.Block) {
	*blocks = append(*blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			"Your access has been successfully revoked. If you need access again, please submit a new request.",
			false,
			false,
		),
		nil,
		nil,
	))
}
