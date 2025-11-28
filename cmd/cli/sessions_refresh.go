package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/thand-io/agent/internal/common"
)

// refreshSession allows the user to refresh/re-authenticate an existing session
func refreshSession() error {
	fmt.Println(headerStyle.Render("Refresh Session"))
	fmt.Println()

	// Reload sessions to get the latest state
	if err := sessionManager.Load(cfg.GetLoginServerHostname()); err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	loginServer, err := sessionManager.GetLoginServer(cfg.GetLoginServerHostname())
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	sessions := loginServer.GetSessions()

	if len(sessions) == 0 {
		fmt.Println(infoStyle.Render("ℹ️  No sessions to refresh"))
		return nil
	}

	// Build options from existing sessions
	var options []huh.Option[string]
	currentTime := time.Now().UTC()

	for provider, session := range sessions {
		var label string
		sessionExpiryTime := session.Expiry.UTC()
		if sessionExpiryTime.Before(currentTime) {
			label = fmt.Sprintf("%s (EXPIRED - %s)", provider, session.Expiry.Format("2006-01-02 15:04"))
		} else {
			timeUntilExpiry := time.Until(session.Expiry)
			label = fmt.Sprintf("%s (expires in %s)", provider, formatDuration(timeUntilExpiry))
		}
		options = append(options, huh.NewOption(label, provider))
	}

	var selectedProvider string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select session to refresh:").
				Options(options...).
				Value(&selectedProvider),
		),
	)

	err = form.Run()
	if err != nil {
		return err
	}

	if len(selectedProvider) == 0 {
		fmt.Println(warningStyle.Render("No session selected"))
		return nil
	}

	fmt.Println(infoStyle.Render("Starting re-authentication flow..."))
	fmt.Println(infoStyle.Render("Please complete the authentication in your browser"))
	fmt.Println()

	// So we should not be doing anything here other than
	// hitting the /authorize endpoint again

	err = authProviderKickStart(
		selectedProvider,
	)

	if err != nil {
		return fmt.Errorf("failed to start re-authorization: %w", err)
	}

	// Wait for the session to be refreshed
	fmt.Println(infoStyle.Render("Waiting for session to be refreshed... (Press Ctrl+C to cancel)"))

	ctx, cleanup := common.WithInterrupt(context.Background())
	defer cleanup()

	refreshedSession := sessionManager.AwaitProviderRefresh(
		ctx,
		cfg.GetLoginServerHostname(),
		selectedProvider,
	)

	if refreshedSession == nil {
		if ctx.Err() != nil {
			fmt.Println()
			fmt.Println(warningStyle.Render("Authentication cancelled by user"))
			return nil
		}
		return fmt.Errorf("authentication timed out or failed")
	}

	fmt.Println(successStyle.Render("Session refreshed successfully!"))
	fmt.Printf("Provider: %s\n", selectedProvider)
	fmt.Printf("New expiry: %s\n", *refreshedSession)

	return nil
}
