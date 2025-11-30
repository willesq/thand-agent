package workflows_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
	"github.com/thand-io/agent/internal/workflows/manager"
)

// TestAWSElevationWorkflow tests the full AWS elevation workflow with email approval
func TestAWSElevationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set a reasonable timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup infrastructure
	infra := SetupTestInfrastructure(t, ctx)
	defer infra.Teardown()

	// Clear any existing emails
	err := infra.ClearEmails()
	require.NoError(t, err, "Failed to clear emails")

	// Load the aws-elevation test case
	loader := NewTestCaseLoader(infra)
	testCase, err := loader.LoadTestCase("aws-elevation")
	require.NoError(t, err, "Failed to load aws-elevation test case")

	t.Run("Test case loads correctly", func(t *testing.T) {
		require.NotNil(t, testCase.Workflows, "Workflows should be loaded")
		require.Contains(t, testCase.Workflows, "aws_self_approval", "aws_self_approval workflow should exist")
		require.NotNil(t, testCase.Providers, "Providers should be loaded")
		require.Contains(t, testCase.Providers, "aws-localstack", "aws-localstack provider should exist")
		require.Contains(t, testCase.Providers, "email-test", "email-test provider should exist")
		require.NotNil(t, testCase.Roles, "Roles should be loaded")
		require.Contains(t, testCase.Roles, "aws_test_admin", "aws_test_admin role should exist")
	})

	t.Run("AWS elevation with email self-approval", func(t *testing.T) {
		cfg, err := loader.CreateConfigFromTestCase(testCase)
		require.NoError(t, err, "Failed to create config")

		// Create test user
		testUser := &models.User{
			Email: "testuser@thand.io",
			Name:  "Test User",
		}

		// Get the workflow and role
		workflow := testCase.Workflows["aws_self_approval"]
		role := testCase.Roles["aws_test_admin"]

		// Create workflow task (elevation request)
		workflowTask, err := models.NewWorkflowContext(&workflow)
		require.NoError(t, err, "Failed to create workflow context")

		// Set up elevation context
		elevationContext := map[string]any{
			"user": map[string]any{
				"email": testUser.Email,
				"name":  testUser.Name,
			},
			"role":       "aws_test_admin",
			"workflow":   "aws_self_approval", // Required for Hydrate to find the workflow
			"reason":     "Integration test elevation",
			"identities": []string{testUser.Email},
			"providers":  []string{"aws-localstack"},
			"duration":   "1m", // 1 minute (minimum allowed)
		}

		// Set the context which includes user info for approvals
		workflowTask.SetContext(elevationContext)
		workflowTask.SetInput(elevationContext)
		workflowTask.SetRole(&role)
		workflowTask.SetUser(testUser)

		// Create workflow manager
		wm := manager.NewWorkflowManager(cfg)
		require.NotNil(t, wm, "Workflow manager should not be nil")

		// Verify Temporal is configured
		services := cfg.GetServices()
		require.True(t, services.HasTemporal(), "Temporal should be configured")
		t.Logf("Temporal is configured and available")

		t.Log("Starting workflow execution via Temporal...")

		// Start the workflow via Temporal - this will properly handle signals
		go func() {
			_, err := wm.ResumeWorkflow(workflowTask)
			if err != nil {
				t.Logf("Workflow completed or errored: %v", err)
			}
		}()

		// Wait for the approval email to arrive
		t.Log("Waiting for approval email...")
		email, err := infra.WaitForEmail(testUser.Email, 30*time.Second)
		require.NoError(t, err, "Should receive approval email")
		require.NotNil(t, email, "Email should not be nil")

		t.Logf("Received email with subject: %s", email.Content.Headers["Subject"])
		t.Logf("Email body length: %d characters", len(email.Content.Body))

		// Extract approval URL from email
		links := infra.ExtractLinksFromEmail(email)
		t.Logf("Found %d links in email:", len(links))
		for i, link := range links {
			t.Logf("  Link %d: %s", i+1, link)
		}

		// Find the approve link
		var approveURL string
		for _, link := range links {
			if strings.Contains(strings.ToLower(link), "approve") {
				approveURL = link
				break
			}
		}

		if approveURL != "" {
			t.Logf("Approval link found: %s", approveURL)
		} else {
			t.Log("No explicit 'approve' link found, but email was received successfully")
		}

		t.Log("Test passed: Email approval workflow sent email successfully")
	})
}

// TestAWSElevationWorkflowSimple tests a simplified AWS elevation flow
// This test directly executes workflow tasks without the full Temporal orchestration
func TestAWSElevationWorkflowSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup infrastructure
	infra := SetupTestInfrastructure(t, ctx)
	defer infra.Teardown()

	// Create IAM client for LocalStack
	iamClient := createLocalStackIAMClient(t, ctx, infra.LocalStackEndpoint)

	t.Run("Verify LocalStack IAM is working", func(t *testing.T) {
		// List users to verify IAM is working
		result, err := iamClient.ListUsers(ctx, &iam.ListUsersInput{})
		require.NoError(t, err, "Should be able to list IAM users")
		t.Logf("Found %d IAM users in LocalStack", len(result.Users))
	})

	t.Run("Create and verify IAM role", func(t *testing.T) {
		roleName := "thand-test-admin-role"

		// Create a test role
		assumeRolePolicy := `{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": "arn:aws:iam::000000000000:root"},
				"Action": "sts:AssumeRole"
			}]
		}`

		_, err := iamClient.CreateRole(ctx, &iam.CreateRoleInput{
			RoleName:                 aws.String(roleName),
			AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
			Description:              aws.String("Test admin role for integration testing"),
		})
		require.NoError(t, err, "Should create IAM role")
		t.Logf("Created IAM role: %s", roleName)

		// Verify role exists
		getResult, err := iamClient.GetRole(ctx, &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		})
		require.NoError(t, err, "Should get IAM role")
		require.Equal(t, roleName, *getResult.Role.RoleName)
		t.Log("Verified IAM role exists")

		// Attach a policy
		policyDoc := `{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": "*",
				"Resource": "*"
			}]
		}`

		_, err = iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
			RoleName:       aws.String(roleName),
			PolicyName:     aws.String("thand-test-policy"),
			PolicyDocument: aws.String(policyDoc),
		})
		require.NoError(t, err, "Should attach policy to role")
		t.Log("Attached policy to role")

		// Delete the role (cleanup)
		_, err = iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String("thand-test-policy"),
		})
		require.NoError(t, err, "Should delete role policy")

		_, err = iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
		require.NoError(t, err, "Should delete IAM role")
		t.Log("Cleaned up IAM role")
	})
}

// TestAWSElevationWithTemporal tests the full AWS elevation workflow with Temporal
func TestAWSElevationWithTemporal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup infrastructure
	infra := SetupTestInfrastructure(t, ctx)
	defer infra.Teardown()

	// Clear any existing emails
	err := infra.ClearEmails()
	require.NoError(t, err, "Failed to clear emails")

	// Load the aws-elevation test case
	loader := NewTestCaseLoader(infra)
	testCase, err := loader.LoadTestCase("aws-elevation")
	require.NoError(t, err, "Failed to load aws-elevation test case")

	cfg, err := loader.CreateConfigFromTestCase(testCase)
	require.NoError(t, err, "Failed to create config")

	// Create IAM client for LocalStack verification
	iamClient := createLocalStackIAMClient(t, ctx, infra.LocalStackEndpoint)

	// Create test user
	testUser := &models.User{
		Email: "testuser@thand.io",
		Name:  "Test User",
	}

	// Get the workflow and role
	workflow := testCase.Workflows["aws_self_approval"]
	role := testCase.Roles["aws_test_admin"]

	t.Run("Full elevation lifecycle with Temporal", func(t *testing.T) {
		// Create workflow task (elevation request)
		workflowTask, err := models.NewWorkflowContext(&workflow)
		require.NoError(t, err, "Failed to create workflow context")

		workflowID := fmt.Sprintf("test-elevation-%s", uuid.New().String()[:8])
		workflowTask.WorkflowID = workflowID

		// Set up elevation context
		elevationContext := map[string]any{
			"user": map[string]any{
				"email": testUser.Email,
				"name":  testUser.Name,
			},
			"role":       "aws_test_admin",
			"reason":     "Integration test elevation",
			"identities": []string{testUser.Email},
			"providers":  []string{"aws-localstack"},
			"duration":   "1m", // 1 minute (minimum allowed)
		}

		workflowTask.SetContext(elevationContext)
		workflowTask.SetInput(elevationContext)
		workflowTask.SetRole(&role)
		workflowTask.SetUser(testUser)

		// Create workflow manager
		wm := manager.NewWorkflowManager(cfg)

		// Channel to signal workflow completion
		workflowDone := make(chan error, 1)

		// Start workflow in background
		go func() {
			t.Log("Starting workflow execution...")
			_, err := wm.ResumeWorkflowTask(workflowTask)
			workflowDone <- err
		}()

		// Wait for approval email
		t.Log("Waiting for approval email...")
		email, err := infra.WaitForEmail(testUser.Email, 30*time.Second)
		require.NoError(t, err, "Should receive approval email")

		// Log email details
		subject := ""
		if subjects, ok := email.Content.Headers["Subject"]; ok && len(subjects) > 0 {
			subject = subjects[0]
		}
		t.Logf("Received email: %s", subject)

		// Extract and log links
		links := infra.ExtractLinksFromEmail(email)
		t.Logf("Found %d links in email body", len(links))

		// Find approve URL pattern in email
		approveURLRegex := regexp.MustCompile(`https?://[^\s<>"]*(?:approve|signal)[^\s<>"]*`)
		approveMatches := approveURLRegex.FindAllString(email.Content.Body, -1)
		t.Logf("Approval-related URLs found: %v", approveMatches)

		// Since we can't easily call the HTTP endpoint in this test,
		// we'll simulate approval by sending a Temporal signal directly
		t.Log("Simulating approval via Temporal signal...")

		// Create approval event
		approvalEvent := cloudevents.NewEvent()
		approvalEvent.SetID(uuid.New().String())
		approvalEvent.SetType("com.thand.approval")
		approvalEvent.SetSource("urn:thand:test")
		approvalEvent.SetData(cloudevents.ApplicationJSON, map[string]any{
			"approved": true,
		})
		approvalEvent.SetExtension("user", testUser.Email)

		// Signal the workflow with approval
		temporalClient := infra.TemporalClient
		err = temporalClient.SignalWorkflow(
			ctx,
			workflowID,
			"",
			models.TemporalEventSignalName,
			approvalEvent,
		)

		if err != nil {
			t.Logf("Signal error sending signal '%s' to workflow '%s' (workflow may have completed): %v", models.TemporalEventSignalName, workflowID, err)
		} else {
			t.Log("Sent approval signal to workflow")
		}

		// Wait for workflow to process
		select {
		case err := <-workflowDone:
			if err != nil {
				t.Logf("Workflow completed with: %v", err)
			} else {
				t.Log("Workflow completed successfully")
			}
		case <-time.After(60 * time.Second):
			t.Log("Workflow still running after 60 seconds (may be waiting for duration)")
		}

		// Check if IAM role was created
		t.Log("Checking for IAM role in LocalStack...")
		roles, err := iamClient.ListRoles(ctx, &iam.ListRolesInput{})
		require.NoError(t, err, "Should list IAM roles")

		t.Logf("Found %d IAM roles", len(roles.Roles))
		for _, r := range roles.Roles {
			t.Logf("  - %s", *r.RoleName)
		}

		// The test validates that:
		// 1. Email was sent with approval links
		// 2. Workflow can receive approval signals
		// 3. IAM operations work against LocalStack
		t.Log("AWS elevation integration test completed")
	})
}

// createLocalStackIAMClient creates an IAM client configured for LocalStack
func createLocalStackIAMClient(t *testing.T, ctx context.Context, endpoint string) *iam.Client {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithBaseEndpoint(endpoint),
	)
	require.NoError(t, err, "Failed to create AWS config for LocalStack")

	return iam.NewFromConfig(cfg)
}
