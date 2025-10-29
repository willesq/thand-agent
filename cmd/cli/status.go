package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
	"github.com/thand-io/agent/internal/models"
)

type statusMsg struct {
	execution *models.WorkflowExecutionInfo
	err       error
}

type errorMsg struct {
	err error
}

type tuiModel struct {
	workflowID  string
	execution   *models.WorkflowExecutionInfo
	spinner     spinner.Model
	loading     bool
	err         error
	lastUpdate  time.Time
	quitting    bool
	liveUpdates bool
	serverUrl   string
	authToken   string
}

func newTuiModel(workflowID, serverUrl, authToken string) tuiModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#3b82f6"))

	return tuiModel{
		workflowID:  workflowID,
		spinner:     s,
		loading:     true,
		liveUpdates: true,
		serverUrl:   serverUrl,
		authToken:   authToken,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchStatus)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case statusMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			// Check if this is a server configuration error
			if strings.Contains(msg.err.Error(), "Temporal service is not configured") ||
				strings.Contains(msg.err.Error(), "Workflow listing is only available in server mode") {
				m.liveUpdates = false
				return m, tea.Quit
			}
		} else {
			m.execution = msg.execution
			m.lastUpdate = time.Now()

			// Continue polling if workflow is still running
			if m.isWorkflowRunning() {
				return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
					return m.fetchStatus()
				})
			}
		}

	case errorMsg:
		m.loading = false
		m.err = msg.err
		m.liveUpdates = false
		return m, tea.Quit

	case tea.WindowSizeMsg:
		return m, nil
	}

	return m, nil
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	if m.loading {
		return fmt.Sprintf("\n %s Fetching workflow status...\n\n", m.spinner.View())
	}

	if m.execution == nil {
		return errorStyle.Render("No execution data available") + "\n"
	}

	var content strings.Builder

	// Title
	content.WriteString(workflowTitleStyle.Render("Workflow Execution Status"))
	content.WriteString("\n\n")

	// Current status section
	content.WriteString(m.renderStatusSection())
	content.WriteString("\n\n")

	// Task list
	content.WriteString(m.renderTaskList())
	content.WriteString("\n\n")

	// Last updated
	if !m.lastUpdate.IsZero() {
		content.WriteString(fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05")))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("Press q to quit")
	content.WriteString("\n")

	return content.String()
}

func (m tuiModel) renderStatusSection() string {
	var section strings.Builder

	// Current task status
	currentTask := m.execution.Task
	if currentTask == "" {
		currentTask = "Initializing"
	}

	var statusColor lipgloss.Color
	var statusText string
	switch strings.ToLower(m.execution.Status) {
	case "completed":
		statusColor = lipgloss.Color("#10b981")
		statusText = "COMPLETED"
	case "failed", "faulted":
		statusColor = lipgloss.Color("#ef4444")
		statusText = "FAILED"
	case "cancelled":
		statusColor = lipgloss.Color("#f59e0b")
		statusText = "CANCELLED"
	case "running":
		statusColor = lipgloss.Color("#3b82f6")
		statusText = "RUNNING"
	default:
		statusColor = lipgloss.Color("#6b7280")
		statusText = strings.ToUpper(m.execution.Status)
	}

	// Current task
	section.WriteString("Current Task: ")
	if m.isWorkflowRunning() {
		section.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), currentTask))
	} else {
		section.WriteString(currentTask)
	}
	section.WriteString("\n")

	// Overall status
	statusBadge := statusBadgeStyle.Copy().Background(statusColor).Render(statusText)
	section.WriteString(fmt.Sprintf("Status: %s", statusBadge))
	section.WriteString("\n")

	// Approval status
	if m.execution.Approved != nil {
		if *m.execution.Approved {
			approvalBadge := approvedStyle.Render("APPROVED")
			section.WriteString(fmt.Sprintf("Approval: %s", approvalBadge))
		} else {
			approvalBadge := rejectedStyle.Render("REJECTED")
			section.WriteString(fmt.Sprintf("Approval: %s", approvalBadge))
		}
	} else {
		approvalBadge := pendingApprovalStyle.Render("PENDING APPROVAL")
		section.WriteString(fmt.Sprintf("Approval: %s", approvalBadge))
	}
	section.WriteString("\n")

	// User info
	if m.execution.User != "" {
		section.WriteString(fmt.Sprintf("User: %s", m.execution.User))
		section.WriteString("\n")
	}

	if m.execution.Role != "" {
		section.WriteString(fmt.Sprintf("Role: %s", m.execution.Role))
		section.WriteString("\n")
	}

	return section.String()
}

func (m tuiModel) renderTaskList() string {
	var section strings.Builder

	section.WriteString("Workflow Tasks:\n")

	if len(m.execution.History) == 0 {
		section.WriteString(taskPendingStyle.Render("• No task history available"))
		section.WriteString("\n")
		return section.String()
	}

	currentTask := m.execution.Task

	for i, task := range m.execution.History {
		var style lipgloss.Style
		var icon string

		if task == currentTask && m.isWorkflowRunning() {
			// Current running task
			style = taskCurrentStyle
			icon = m.spinner.View()
		} else if i < len(m.execution.History)-1 || !m.isWorkflowRunning() {
			// Completed task
			style = taskCompletedStyle
			icon = "✓"
		} else {
			// Pending task
			style = taskPendingStyle
			icon = "○"
		}

		line := fmt.Sprintf("%s %s", icon, task)
		section.WriteString(style.Render(line))
		section.WriteString("\n")
	}

	return section.String()
}

func (m tuiModel) isWorkflowRunning() bool {
	if m.execution == nil {
		return false
	}
	status := strings.ToLower(m.execution.Status)
	return status == "running" || status == "pending" || status == "waiting" || status == "suspended"
}

func (m tuiModel) fetchStatus() tea.Msg {
	client := resty.New()
	url := fmt.Sprintf("%s/execution/%s", strings.TrimSuffix(m.serverUrl, "/"), m.workflowID)

	resp, err := client.R().
		SetAuthToken(m.authToken).
		SetHeader("Accept", "application/json").
		Get(url)

	if err != nil {
		return statusMsg{err: fmt.Errorf("failed to fetch status: %w", err)}
	}

	if resp.StatusCode() == http.StatusBadRequest {
		// Check if this is a configuration error
		return statusMsg{err: fmt.Errorf("polling is not supported by the remote server: %s for: %s", string(resp.Body()), url)}
	}

	if resp.StatusCode() != http.StatusOK {
		return statusMsg{err: fmt.Errorf("API error: %d - %s for: %s", resp.StatusCode(), string(resp.Body()), url)}
	}

	var apiResponse struct {
		Execution *models.WorkflowExecutionInfo `json:"execution"`
	}

	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return statusMsg{err: fmt.Errorf("failed to parse response: %w", err)}
	}

	return statusMsg{execution: apiResponse.Execution}
}

// runWorkflowStatusTUI starts the TUI for live workflow status updates
func runWorkflowStatusTUI(workflowID, serverUrl, authToken string) error {
	model := newTuiModel(workflowID, serverUrl, authToken)
	program := tea.NewProgram(model)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if live updates are not supported
	if !model.liveUpdates {
		return fmt.Errorf("live status updates not supported in this configuration")
	}

	return nil
}
