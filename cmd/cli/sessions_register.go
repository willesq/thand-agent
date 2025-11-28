package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/thand-io/agent/internal/models"
)

// runSessionRegister handles the session register command
func runSessionRegister(cmd *cobra.Command) error {
	fmt.Println(titleStyle.Render("Session Registration"))
	fmt.Println()

	// Get login server from global config
	loginServerHostname := cfg.GetLoginServerHostname()
	if len(loginServerHostname) == 0 {
		return fmt.Errorf("no login server configured. Please set --login-server flag or configure in config.yaml")
	}

	// Get provider from flag or prompt
	provider, _ := cmd.Flags().GetString("provider")

	// If provider is empty, prompt for it
	if len(provider) == 0 {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Provider").
					Description("Enter the provider name (e.g., thand)").
					Value(&provider).
					Validate(func(s string) error {
						if len(s) == 0 {
							return fmt.Errorf("provider name is required")
						}
						return nil
					}),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("failed to get provider: %w", err)
		}
	}

	// Prompt for the session token
	var sessionToken string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Session Token").
				Description("Paste your encoded session token below").
				Value(&sessionToken).
				Validate(func(s string) error {
					if len(s) == 0 {
						return fmt.Errorf("session token is required")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("failed to get session token: %w", err)
	}

	// Trim whitespace from the session token
	sessionToken = strings.TrimSpace(sessionToken)

	fmt.Println()
	fmt.Println(infoStyle.Render("Decoding session token..."))

	// Decode the session token
	localSession, err := models.DecodedLocalSession(sessionToken)
	if err != nil {
		return fmt.Errorf("failed to decode session token: %w", err)
	}

	// Validate the session
	if localSession.IsExpired() {
		fmt.Println(warningStyle.Render("⚠️  Warning: This session has already expired"))

		var continueAnyway bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Continue anyway?").
					Description("The session token has expired. Do you still want to register it?").
					Value(&continueAnyway),
			),
		)

		if err := confirmForm.Run(); err != nil {
			return err
		}

		if !continueAnyway {
			fmt.Println(infoStyle.Render("ℹ️  Session registration cancelled"))
			return nil
		}
	}

	// Store the session in the session manager
	err = sessionManager.AddSession(loginServerHostname, provider, *localSession)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("✓ Session registered successfully!"))
	fmt.Println()
	fmt.Printf("  Login Server: %s\n", loginServerHostname)
	fmt.Printf("  Provider:     %s\n", provider)
	fmt.Printf("  Expires:      %s\n", localSession.Expiry.Format("2006-01-02 15:04:05"))

	if !localSession.IsExpired() {
		timeUntilExpiry := time.Until(localSession.Expiry)
		fmt.Printf("  Valid for:    %s\n", formatDuration(timeUntilExpiry))
	}

	return nil
}
