package cli

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
)

// removeSession allows the user to remove an existing session
func removeSession() error {
	fmt.Println(headerStyle.Render("Remove Session"))
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
		fmt.Println(infoStyle.Render("ℹ️  No sessions to remove"))
		return nil
	}

	// Build options from existing sessions
	var options []huh.Option[string]
	now := time.Now().UTC()

	for provider, session := range sessions {
		var label string
		if session.Expiry.Before(now) {
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
				Title("Select session to remove:").
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

	// Confirm removal
	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirm removal").
				Description(fmt.Sprintf("Are you sure you want to remove the session for %s?", selectedProvider)).
				Value(&confirm),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return err
	}

	if !confirm {
		fmt.Println(infoStyle.Render("ℹ️  Session removal cancelled"))
		return nil
	}

	// Remove the session
	if err := sessionManager.RemoveSession(cfg.GetLoginServerHostname(), selectedProvider); err != nil {
		return fmt.Errorf("failed to remove session: %w", err)
	}

	fmt.Println(successStyle.Render("Session removed successfully!"))

	return nil
}
