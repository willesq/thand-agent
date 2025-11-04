package thand

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// createAuthorizeEmailBody creates the email body for authorization confirmation
func (a *authorizerNotifier) createAuthorizeEmailBody() (string, string) {
	elevationReq := a.elevationReq
	notifyReq := a.req

	// Build plain text version
	var plainText strings.Builder
	plainText.WriteString("Good news! Your access request has been approved.\n\n")

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

	plainText.WriteString("\nYour access is now active. Please use it responsibly.")

	// Build data map for template
	data := map[string]any{
		"Providers": strings.Join(elevationReq.Providers, ", "),
		"Duration":  elevationReq.Duration,
	}

	if len(notifyReq.Message) > 0 {
		data["Message"] = notifyReq.Message
	}

	if elevationReq.Role != nil {
		data["Role"] = map[string]any{
			"Name":        elevationReq.Role.Name,
			"Description": elevationReq.Role.Description,
		}

		// Add permissions if available
		if len(elevationReq.Role.Permissions.Allow) > 0 {
			data["Permissions"] = elevationReq.Role.Permissions.Allow
		}
	}

	// Add provider access button info if we have a single provider
	if len(elevationReq.Providers) == 1 {
		providerName := elevationReq.Providers[0]
		data["ProviderName"] = providerName
		// You can customize URL generation based on your provider setup
		// For now, using a placeholder pattern
		data["ProviderURL"] = fmt.Sprintf("/provider/%s", providerName)
	}

	// Render HTML email using template
	html, err := RenderEmailWithTemplate("Access Request Approved", GetAuthorizeContentTemplate(), data)
	if err != nil {
		logrus.WithError(err).Error("Failed to render authorization email")
		return plainText.String(), ""
	}

	return plainText.String(), html
}
