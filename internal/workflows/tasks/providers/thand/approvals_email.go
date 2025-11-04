package thand

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// createApprovalEmailBody creates the email body for approval requests
func (a *approvalsNotifier) createApprovalEmailBody() (string, string) {
	elevationReq := a.elevationReq
	notifyReq := a.req

	// Build plain text version
	var plainText strings.Builder
	plainText.WriteString("A user has requested elevated access and requires your approval.\n\n")

	if elevationReq.User != nil {
		plainText.WriteString(fmt.Sprintf("Requested by: %s", elevationReq.User.Name))
		if len(elevationReq.User.Email) > 0 {
			plainText.WriteString(fmt.Sprintf(" (%s)", elevationReq.User.Email))
		}
		plainText.WriteString("\n\n")
	}

	if elevationReq.Role != nil {
		plainText.WriteString(fmt.Sprintf("Role: %s\n", elevationReq.Role.Name))
		if len(elevationReq.Role.Description) > 0 {
			plainText.WriteString(fmt.Sprintf("Description: %s\n", elevationReq.Role.Description))
		}
	}

	if len(elevationReq.Providers) > 0 {
		plainText.WriteString(fmt.Sprintf("Providers: %s\n", strings.Join(elevationReq.Providers, ", ")))
	}

	if len(elevationReq.Duration) > 0 {
		plainText.WriteString(fmt.Sprintf("Duration: %s\n", elevationReq.Duration))
	}

	if len(elevationReq.Reason) > 0 {
		plainText.WriteString(fmt.Sprintf("Reason: %s\n", elevationReq.Reason))
	}

	if len(elevationReq.Identities) > 0 {
		plainText.WriteString("\nTarget Identities:\n")
		for _, identity := range elevationReq.Identities {
			plainText.WriteString(fmt.Sprintf("• %s\n", identity))
		}
	}

	// Build data map for template
	data := map[string]any{
		"Providers":  strings.Join(elevationReq.Providers, ", "),
		"Duration":   elevationReq.Duration,
		"Reason":     elevationReq.Reason,
		"Identities": elevationReq.Identities,
	}

	if len(notifyReq.Notifier.Message) > 0 {
		data["Message"] = notifyReq.Notifier.Message
	}

	if elevationReq.User != nil {
		data["User"] = map[string]any{
			"Name":  elevationReq.User.Name,
			"Email": elevationReq.User.Email,
		}
	}

	if elevationReq.Role != nil {
		data["Role"] = map[string]any{
			"Name":        elevationReq.Role.Name,
			"Description": elevationReq.Role.Description,
		}
	}

	// Render HTML email using template
	html, err := RenderEmailWithTemplate("Access Request - Approval Required", GetApprovalContentTemplate(), data)
	if err != nil {
		logrus.WithError(err).Error("Failed to render approval email")
		return plainText.String(), ""
	}

	return plainText.String(), html
}
