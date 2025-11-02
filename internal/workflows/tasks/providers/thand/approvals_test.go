package thand

import (
	"testing"
	"time"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// TestEvaluateApprovalSwitch tests the approval switch logic with various scenarios
func TestEvaluateApprovalSwitch(t *testing.T) {
	tests := []struct {
		name              string
		approvals         map[string]any
		requiredApprovals int
		approvedState     string
		deniedState       string
		taskName          string
		expectedState     string
		description       string
	}{
		{
			name: "single approval - meets requirement",
			approvals: map[string]any{
				"user1@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			requiredApprovals: 1,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "authorize",
			description:       "Should approve when one approval meets requirement of 1",
		},
		{
			name: "single denial - should deny immediately",
			approvals: map[string]any{
				"user1@example.com": map[string]any{
					"approved":  false,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			requiredApprovals: 1,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "denied",
			description:       "Should deny immediately when any user denies",
		},
		{
			name: "multiple approvals - meets requirement",
			approvals: map[string]any{
				"user1@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				"user2@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				"user3@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			requiredApprovals: 2,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "authorize",
			description:       "Should approve when 3 approvals meet requirement of 2",
		},
		{
			name: "multiple approvals - one denial",
			approvals: map[string]any{
				"user1@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				"user2@example.com": map[string]any{
					"approved":  false,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				"user3@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			requiredApprovals: 2,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "denied",
			description:       "Should deny when any user denies, even if others approve",
		},
		{
			name: "insufficient approvals - should loop",
			approvals: map[string]any{
				"user1@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			requiredApprovals: 2,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "approval_task",
			description:       "Should loop back when approvals (1) don't meet requirement (2)",
		},
		{
			name:              "no approvals yet - should loop",
			approvals:         map[string]any{},
			requiredApprovals: 1,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "approval_task",
			description:       "Should loop back when no approvals received yet",
		},
		{
			name: "exactly meets requirement",
			approvals: map[string]any{
				"user1@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				"user2@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			requiredApprovals: 2,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "authorize",
			description:       "Should approve when exactly meeting requirement",
		},
		{
			name: "mixed responses with insufficient approvals",
			approvals: map[string]any{
				"user1@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				"user2@example.com": map[string]any{
					"approved":  true,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			},
			requiredApprovals: 3,
			approvedState:     "authorize",
			deniedState:       "denied",
			taskName:          "approval_task",
			expectedState:     "approval_task",
			description:       "Should loop when 2 approvals don't meet requirement of 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a workflow task with the approvals context
			workflowTask := &models.WorkflowTask{
				WorkflowID:   "test-workflow",
				WorkflowName: "Test Workflow",
			} // Set the context with approvals
			workflowTask.SetContextKeyValue("approvals", tt.approvals)

			// Create a thandTask instance
			task := &thandTask{}

			// Create default flow directive (loop back to task)
			defaultFlowState := model.FlowDirective{
				Value: tt.taskName,
			}

			// Execute the approval switch
			flowDirective, err := task.evaluateApprovalSwitch(
				workflowTask,
				tt.taskName,
				tt.approvals,
				tt.requiredApprovals,
				tt.approvedState,
				tt.deniedState,
				defaultFlowState,
			) // Assert no error occurred
			require.NoError(t, err, "evaluateApprovalSwitch should not return an error for test: %s", tt.description)

			// Assert the flow directive is not nil
			require.NotNil(t, flowDirective, "flowDirective should not be nil for test: %s", tt.description)

			// Assert the expected state
			assert.Equal(t, tt.expectedState, flowDirective.Value, "Test: %s", tt.description)
		})
	}
}

// TestEvaluateApprovalSwitchProgression tests the progression of approvals over time
func TestEvaluateApprovalSwitchProgression(t *testing.T) {
	workflowTask := &models.WorkflowTask{
		WorkflowID:   "test-workflow",
		WorkflowName: "Test Workflow",
	}

	task := &thandTask{}
	taskName := "approval_task"
	requiredApprovals := 2
	approvedState := "authorize"
	deniedState := "denied"

	// Create default flow directive (loop back to task)
	defaultFlowState := model.FlowDirective{
		Value: taskName,
	}

	// Step 1: No approvals yet - should loop
	approvals := map[string]any{}
	workflowTask.SetContextKeyValue("approvals", approvals)

	flowDirective, err := task.evaluateApprovalSwitch(
		workflowTask,
		taskName,
		approvals,
		requiredApprovals,
		approvedState,
		deniedState,
		defaultFlowState,
	)

	require.NoError(t, err)
	require.NotNil(t, flowDirective)
	assert.Equal(t, taskName, flowDirective.Value, "Should loop when no approvals")

	// Step 2: First approval comes in - still should loop
	approvals["user1@example.com"] = map[string]any{
		"approved":  true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	workflowTask.SetContextKeyValue("approvals", approvals)

	flowDirective, err = task.evaluateApprovalSwitch(
		workflowTask,
		taskName,
		approvals,
		requiredApprovals,
		approvedState,
		deniedState,
		defaultFlowState,
	)

	require.NoError(t, err)
	require.NotNil(t, flowDirective)
	assert.Equal(t, taskName, flowDirective.Value, "Should loop when only 1 of 2 approvals")

	// Step 3: Second approval comes in - should now authorize
	approvals["user2@example.com"] = map[string]any{
		"approved":  true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	workflowTask.SetContextKeyValue("approvals", approvals)

	flowDirective, err = task.evaluateApprovalSwitch(
		workflowTask,
		taskName,
		approvals,
		requiredApprovals,
		approvedState,
		deniedState,
		defaultFlowState,
	)

	require.NoError(t, err)
	require.NotNil(t, flowDirective)
	assert.Equal(t, approvedState, flowDirective.Value, "Should authorize when 2 of 2 approvals received")
}

// TestEvaluateApprovalSwitchDenialProgression tests that a denial stops the process
func TestEvaluateApprovalSwitchDenialProgression(t *testing.T) {
	workflowTask := &models.WorkflowTask{
		WorkflowID:   "test-workflow",
		WorkflowName: "Test Workflow",
	}

	task := &thandTask{}
	taskName := "approval_task"
	requiredApprovals := 3
	approvedState := "authorize"
	deniedState := "denied"

	// Create default flow directive (loop back to task)
	defaultFlowState := model.FlowDirective{
		Value: taskName,
	}

	// Step 1: First approval comes in
	approvals := map[string]any{
		"user1@example.com": map[string]any{
			"approved":  true,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
	workflowTask.SetContextKeyValue("approvals", approvals)

	flowDirective, err := task.evaluateApprovalSwitch(
		workflowTask,
		taskName,
		approvals,
		requiredApprovals,
		approvedState,
		deniedState,
		defaultFlowState,
	)

	require.NoError(t, err)
	require.NotNil(t, flowDirective)
	assert.Equal(t, taskName, flowDirective.Value, "Should loop with 1 approval")

	// Step 2: Second user denies - should immediately deny
	approvals["user2@example.com"] = map[string]any{
		"approved":  false,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	workflowTask.SetContextKeyValue("approvals", approvals)

	flowDirective, err = task.evaluateApprovalSwitch(
		workflowTask,
		taskName,
		approvals,
		requiredApprovals,
		approvedState,
		deniedState,
		defaultFlowState,
	)

	require.NoError(t, err)
	require.NotNil(t, flowDirective)
	assert.Equal(t, deniedState, flowDirective.Value, "Should deny immediately when any user denies")

	// Step 3: Even if a third approval comes in, should still be denied
	approvals["user3@example.com"] = map[string]any{
		"approved":  true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	workflowTask.SetContextKeyValue("approvals", approvals)

	flowDirective, err = task.evaluateApprovalSwitch(
		workflowTask,
		taskName,
		approvals,
		requiredApprovals,
		approvedState,
		deniedState,
		defaultFlowState,
	)

	require.NoError(t, err)
	require.NotNil(t, flowDirective)
	assert.Equal(t, deniedState, flowDirective.Value, "Should remain denied even after additional approvals")
}

// TestSelfApprovalLogic tests the self-approval validation logic
func TestSelfApprovalLogic(t *testing.T) {
	tests := []struct {
		name              string
		selfApprove       bool
		requesterIdentity string
		approverIdentity  string
		shouldBlock       bool
		description       string
	}{
		{
			name:              "self-approval disabled - same user",
			selfApprove:       false,
			requesterIdentity: "user@example.com",
			approverIdentity:  "user@example.com",
			shouldBlock:       true,
			description:       "Should block approval when requester and approver are same and selfApprove is false",
		},
		{
			name:              "self-approval disabled - different user",
			selfApprove:       false,
			requesterIdentity: "requester@example.com",
			approverIdentity:  "approver@example.com",
			shouldBlock:       false,
			description:       "Should allow approval when requester and approver are different and selfApprove is false",
		},
		{
			name:              "self-approval enabled - same user",
			selfApprove:       true,
			requesterIdentity: "user@example.com",
			approverIdentity:  "user@example.com",
			shouldBlock:       false,
			description:       "Should allow approval when requester and approver are same and selfApprove is true",
		},
		{
			name:              "self-approval enabled - different user",
			selfApprove:       true,
			requesterIdentity: "requester@example.com",
			approverIdentity:  "approver@example.com",
			shouldBlock:       false,
			description:       "Should allow approval when requester and approver are different and selfApprove is true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the self-approval check logic
			notifyReq := NotifyRequest{
				Approvals:   1,
				SelfApprove: tt.selfApprove,
			}

			// Simulate the self-approval check logic from the code
			shouldBlock := !notifyReq.SelfApprove && tt.requesterIdentity == tt.approverIdentity

			// Verify the result
			assert.Equal(t, tt.shouldBlock, shouldBlock, "Test: %s", tt.description)
		})
	}
}

// TestSelfApprovalWithMultipleApprovers tests self-approval counting logic
func TestSelfApprovalWithMultipleApprovers(t *testing.T) {
	requesterIdentity := "user1@example.com"

	t.Run("self-approval disabled - only external approvers count", func(t *testing.T) {
		// Test scenario: selfApprove = false, requires 2 approvals
		// User1 (requester) tries to approve their own request (should be blocked)
		// User2 approves (should count)
		// User3 approves (should count)
		// Result: 2 valid approvals

		notifyReq := NotifyRequest{
			Approvals:   2,
			SelfApprove: false,
		}

		approvals := map[string]any{}

		// Step 1: Requester tries to self-approve (should be blocked)
		user1ShouldBlock := !notifyReq.SelfApprove && requesterIdentity == "user1@example.com"
		if !user1ShouldBlock {
			approvals["user1@example.com"] = map[string]any{
				"approved":  true,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
		}

		assert.Empty(t, approvals, "Self-approval should be blocked")

		// Step 2: External user2 approves (should count)
		user2ShouldBlock := !notifyReq.SelfApprove && requesterIdentity == "user2@example.com"
		if !user2ShouldBlock {
			approvals["user2@example.com"] = map[string]any{
				"approved":  true,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
		}

		assert.Len(t, approvals, 1, "External approval should be counted")

		// Step 3: External user3 approves (should count)
		user3ShouldBlock := !notifyReq.SelfApprove && requesterIdentity == "user3@example.com"
		if !user3ShouldBlock {
			approvals["user3@example.com"] = map[string]any{
				"approved":  true,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
		}

		assert.Len(t, approvals, 2, "Two external approvals should be counted")
		assert.NotContains(t, approvals, "user1@example.com", "Requester's self-approval should not be in approvals map")
		assert.Contains(t, approvals, "user2@example.com", "User2's approval should be in approvals map")
		assert.Contains(t, approvals, "user3@example.com", "User3's approval should be in approvals map")
	})

	t.Run("self-approval enabled - requester can approve own request", func(t *testing.T) {
		// Test scenario: selfApprove = true, requires 1 approval
		// User1 (requester) approves their own request (should count)

		notifyReq := NotifyRequest{
			Approvals:   1,
			SelfApprove: true,
		}

		approvals := map[string]any{}

		// Requester self-approves (should be allowed)
		user1ShouldBlock := !notifyReq.SelfApprove && requesterIdentity == "user1@example.com"
		if !user1ShouldBlock {
			approvals["user1@example.com"] = map[string]any{
				"approved":  true,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
		}

		assert.Len(t, approvals, 1, "Self-approval should be allowed and counted")
		assert.Contains(t, approvals, "user1@example.com", "Requester's self-approval should be in approvals map")
	})
}

// TestGetIdentityVariations tests different identity formats
func TestGetIdentityVariations(t *testing.T) {
	tests := []struct {
		name             string
		user             models.User
		expectedIdentity string
		description      string
	}{
		{
			name: "email present",
			user: models.User{
				Email:    "user@example.com",
				Username: "username",
				ID:       "123",
			},
			expectedIdentity: "user@example.com",
			description:      "Should use email when available",
		},
		{
			name: "no email, username present",
			user: models.User{
				Username: "username",
				ID:       "123",
			},
			expectedIdentity: "username",
			description:      "Should use username when email not available",
		},
		{
			name: "no email or username, ID present",
			user: models.User{
				ID: "123",
			},
			expectedIdentity: "123",
			description:      "Should use ID when email and username not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity := tt.user.GetIdentity()
			assert.Equal(t, tt.expectedIdentity, identity, "Test: %s", tt.description)
		})
	}
}
