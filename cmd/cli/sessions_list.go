package cli

import (
	"fmt"
	"time"
)

// listSessions displays all current sessions with their status
func listSessions() error {
	fmt.Println(headerStyle.Render("Current Sessions"))
	fmt.Println()

	// Reload sessions to get the latest state
	if err := sessionManager.Load(cfg.GetLoginServerHostname()); err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	loginServer, err := sessionManager.GetLoginServer(cfg.GetLoginServerHostname())

	if err != nil {
		return fmt.Errorf("failed to get sessions for logon server: %w", err)
	}

	sessions := loginServer.GetSessions()

	if len(sessions) == 0 {
		fmt.Println(infoStyle.Render("ℹ️  No active sessions found"))
		return nil
	}

	currentTime := time.Now().UTC()

	for provider, session := range sessions {
		providerDisplay := headerStyle.Render(fmt.Sprintf("Provider: %s", provider))

		var statusDisplay string
		var expiryDisplay string

		sessionExpiryTime := session.Expiry.UTC()
		if sessionExpiryTime.Before(currentTime) {
			statusDisplay = expiredStyle.Render("EXPIRED")
			expiryDisplay = expiredStyle.Render(fmt.Sprintf("Expired: %s", session.Expiry.Format("2006-01-02 15:04:05")))
		} else {
			statusDisplay = activeStyle.Render("ACTIVE")
			timeUntilExpiry := time.Until(session.Expiry)
			expiryDisplay = activeStyle.Render(fmt.Sprintf("Expires: %s (%s)",
				session.Expiry.Format("2006-01-02 15:04:05"),
				formatDuration(timeUntilExpiry)))
		}

		versionDisplay := infoStyle.Render(fmt.Sprintf("Version: %d", session.Version))

		fmt.Println(providerDisplay)
		fmt.Println("  " + statusDisplay)
		fmt.Println("  " + expiryDisplay)
		fmt.Println("  " + versionDisplay)
		fmt.Println()
	}

	return nil
}
