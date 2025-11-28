package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/thand-io/agent/internal/common"
)

// createNewSession guides the user through creating a new session
func createNewSession() error {
	fmt.Println(headerStyle.Render("Create New Session"))
	fmt.Println()

	// Get available providers from config
	providers := getAvailableProviders()
	if len(providers) == 0 {

		// If there are no providers, then just kick off a general login
		return authKickStart()

	}

	var selectedProvider string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a provider:").
				Description("Choose the provider you want to create a session for").
				Options(providers...).
				Value(&selectedProvider),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	if len(selectedProvider) == 0 {
		fmt.Println(warningStyle.Render("No provider selected"))
		return nil
	}

	// Check if session already exists
	existingSession, _ := sessionManager.GetSession(cfg.GetLoginServerHostname(), selectedProvider)
	if existingSession != nil {
		now := time.Now().UTC()
		if existingSession.Expiry.After(now) {
			var overwrite bool
			overwriteForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Session already exists").
						Description(fmt.Sprintf("An active session for %s already exists. Do you want to replace it?", selectedProvider)).
						Value(&overwrite),
				),
			)

			if err := overwriteForm.Run(); err != nil {
				return err
			}

			if !overwrite {
				fmt.Println(infoStyle.Render("ℹ️  Session creation cancelled"))
				return nil
			}
		}
	}

	fmt.Println(infoStyle.Render("Starting authentication flow..."))
	fmt.Println(infoStyle.Render("Please complete the authentication in your browser"))
	fmt.Println()

	err = authProviderKickStart(
		selectedProvider,
	)

	if err != nil {
		return fmt.Errorf("failed to start authorization: %w", err)
	}

	// Wait for the session to be created (polling)
	fmt.Println(infoStyle.Render("Waiting for session to be created... (Press Ctrl+C to cancel)"))

	ctx, cleanup := common.WithInterrupt(context.Background())
	defer cleanup()

	newSession := sessionManager.AwaitProviderRefresh(
		ctx,
		cfg.GetLoginServerHostname(),
		selectedProvider,
	)

	if newSession == nil {
		if ctx.Err() != nil {
			fmt.Println()
			fmt.Println(warningStyle.Render("Authentication cancelled by user"))
			return nil
		}
		return fmt.Errorf("authentication timed out or failed")
	}

	fmt.Println(successStyle.Render("Session created successfully!"))
	fmt.Printf("Provider: %s\n", selectedProvider)
	fmt.Printf("Expires: %s\n", *newSession)

	return nil
}
