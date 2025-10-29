package cli

import (
	"github.com/charmbracelet/lipgloss"
)

// Shared styles for the CLI package
// All terminal colors and styling definitions are centralized here
var (
	// Primary styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

	// Status styles
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6"))

	// Session-specific styles
	expiredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Strikethrough(true)

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	// Elevation styles
	workflowTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	approvedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10b981")).
			Background(lipgloss.Color("#10b98115")).
			Padding(0, 1)

	rejectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ef4444")).
			Background(lipgloss.Color("#ef444415")).
			Padding(0, 1)

	pendingApprovalStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#f59e0b")).
				Background(lipgloss.Color("#f59e0b15")).
				Padding(0, 1)

	taskCompletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10b981")).
				Padding(0, 1).
				Margin(0, 0, 1, 2)

	taskCurrentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3b82f6")).
				Bold(true).
				Padding(0, 1).
				Margin(0, 0, 1, 2)

	taskPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")).
				Padding(0, 1).
				Margin(0, 0, 1, 2)

	statusBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)
)
