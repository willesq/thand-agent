package workflows_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

const (
	// TemporalDefaultNamespace is the default namespace used for Temporal integration tests
	TemporalDefaultNamespace = "default"
	// TemporalDefaultPort is the default gRPC port for Temporal server
	TemporalDefaultPort = "7233"
)

// TestInfrastructure holds all test containers and clients
type TestInfrastructure struct {
	t   *testing.T
	ctx context.Context

	// LocalStack (AWS)
	localstackContainer testcontainers.Container
	LocalStackEndpoint  string

	// MailHog (SMTP testing)
	mailhogContainer testcontainers.Container
	MailHogSMTP      string // SMTP endpoint for sending (host:port)
	MailHogAPI       string // HTTP API endpoint for reading emails

	// PostgreSQL (for Temporal)
	postgresContainer testcontainers.Container

	// Temporal
	temporalContainer testcontainers.Container
	TemporalEndpoint  string
	TemporalClient    client.Client
}

// SetupTestInfrastructure creates and starts Temporal and LocalStack containers
func SetupTestInfrastructure(t *testing.T, ctx context.Context) *TestInfrastructure {
	t.Helper()

	infra := &TestInfrastructure{
		t:   t,
		ctx: ctx,
	}

	// Start LocalStack (AWS mock)
	infra.startLocalStack(ctx)

	// Start MailHog (SMTP testing)
	infra.startMailHog(ctx)

	// Start Temporal
	infra.startTemporal(ctx)

	return infra
}

// startLocalStack starts the LocalStack container
func (infra *TestInfrastructure) startLocalStack(ctx context.Context) {
	infra.t.Log("Starting LocalStack container...")

	container, err := localstack.Run(ctx,
		"localstack/localstack:3.0",
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "iam,sts,ses",
			"DEBUG":    "1",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/health").
				WithPort("4566/tcp").
				WithStartupTimeout(60*time.Second).
				WithPollInterval(2*time.Second),
		),
	)
	require.NoError(infra.t, err, "Failed to start LocalStack container")

	infra.localstackContainer = container

	host, err := container.Host(ctx)
	require.NoError(infra.t, err, "Failed to get LocalStack host")

	mappedPort, err := container.MappedPort(ctx, "4566/tcp")
	require.NoError(infra.t, err, "Failed to get LocalStack port")

	infra.LocalStackEndpoint = fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	infra.t.Logf("LocalStack started at %s", infra.LocalStackEndpoint)
}

// startMailHog starts the MailHog container for SMTP testing
// MailHog captures all outgoing emails and provides an API to read them
func (infra *TestInfrastructure) startMailHog(ctx context.Context) {
	infra.t.Log("Starting MailHog container...")

	req := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog:v1.0.1",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"}, // SMTP on 1025, HTTP API on 8025
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("1025/tcp").WithStartupTimeout(30*time.Second),
			wait.ForListeningPort("8025/tcp").WithStartupTimeout(30*time.Second),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(infra.t, err, "Failed to start MailHog container")

	infra.mailhogContainer = container

	host, err := container.Host(ctx)
	require.NoError(infra.t, err, "Failed to get MailHog host")

	smtpPort, err := container.MappedPort(ctx, "1025/tcp")
	require.NoError(infra.t, err, "Failed to get MailHog SMTP port")

	apiPort, err := container.MappedPort(ctx, "8025/tcp")
	require.NoError(infra.t, err, "Failed to get MailHog API port")

	infra.MailHogSMTP = net.JoinHostPort(host, smtpPort.Port())
	infra.MailHogAPI = fmt.Sprintf("http://%s:%s", host, apiPort.Port())

	infra.t.Logf("MailHog started - SMTP: %s, API: %s", infra.MailHogSMTP, infra.MailHogAPI)
}

// MailHogMessage represents an email captured by MailHog
type MailHogMessage struct {
	ID   string `json:"ID"`
	From struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"From"`
	To []struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"To"`
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"Body"`
	} `json:"Content"`
	Created time.Time `json:"Created"`
	Raw     struct {
		From string   `json:"From"`
		To   []string `json:"To"`
		Data string   `json:"Data"`
	} `json:"Raw"`
}

// MailHogMessages represents the response from MailHog API
type MailHogMessages struct {
	Total int              `json:"total"`
	Count int              `json:"count"`
	Start int              `json:"start"`
	Items []MailHogMessage `json:"items"`
}

// GetEmails retrieves all emails from MailHog
func (infra *TestInfrastructure) GetEmails() ([]MailHogMessage, error) {
	url := infra.MailHogAPI + "/api/v2/messages"
	resp, err := common.InvokeHttpRequest(&model.HTTPArguments{
		Method:   http.MethodGet,
		Endpoint: model.NewEndpoint(url),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get emails from MailHog: %w", err)
	}

	var messages MailHogMessages
	if err := json.Unmarshal(resp.Body(), &messages); err != nil {
		return nil, fmt.Errorf("failed to parse MailHog response: %w", err)
	}

	return messages.Items, nil
}

// GetEmailsForRecipient retrieves emails sent to a specific address
func (infra *TestInfrastructure) GetEmailsForRecipient(email string) ([]MailHogMessage, error) {
	allEmails, err := infra.GetEmails()
	if err != nil {
		return nil, err
	}

	var filtered []MailHogMessage
	for _, msg := range allEmails {
		for _, to := range msg.To {
			if fmt.Sprintf("%s@%s", to.Mailbox, to.Domain) == email {
				filtered = append(filtered, msg)
				break
			}
		}
	}
	return filtered, nil
}

// WaitForEmail waits for an email to arrive for a specific recipient
func (infra *TestInfrastructure) WaitForEmail(recipient string, timeout time.Duration) (*MailHogMessage, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		emails, err := infra.GetEmailsForRecipient(recipient)
		if err != nil {
			return nil, err
		}
		if len(emails) > 0 {
			return &emails[0], nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for email to %s", recipient)
}

// ExtractLinksFromEmail extracts all URLs from an email body
func (infra *TestInfrastructure) ExtractLinksFromEmail(msg *MailHogMessage) []string {
	// Match URLs in the email body
	urlRegex := regexp.MustCompile(`https?://[^\s<>"]+`)
	return urlRegex.FindAllString(msg.Content.Body, -1)
}

// ClearEmails deletes all emails from MailHog
func (infra *TestInfrastructure) ClearEmails() error {
	req, err := http.NewRequestWithContext(infra.ctx, http.MethodDelete, infra.MailHogAPI+"/api/v1/messages", nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code when clearing emails: %d", resp.StatusCode)
	}

	return nil
}

// startTemporal starts the Temporal container
// Uses temporalio/auto-setup which includes a development server setup
func (infra *TestInfrastructure) startTemporal(ctx context.Context) {
	infra.t.Log("Starting Temporal container...")

	// Use auto-setup with a PostgreSQL sidecar for persistence
	// First start PostgreSQL
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "temporal",
			"POSTGRES_PASSWORD": "temporal",
			"POSTGRES_DB":       "temporal",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	require.NoError(infra.t, err, "Failed to start PostgreSQL container")

	postgresHost, err := postgresContainer.Host(ctx)
	require.NoError(infra.t, err, "Failed to get PostgreSQL host")

	postgresPort, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(infra.t, err, "Failed to get PostgreSQL port")

	infra.t.Logf("PostgreSQL started at %s:%s", postgresHost, postgresPort.Port())

	// Get internal container IP for Temporal to connect to PostgreSQL
	postgresInspect, err := postgresContainer.Inspect(ctx)
	require.NoError(infra.t, err, "Failed to inspect PostgreSQL container")

	postgresIP := postgresInspect.NetworkSettings.Networks["bridge"].IPAddress
	infra.t.Logf("PostgreSQL internal IP: %s", postgresIP)

	// Now start Temporal auto-setup
	temporalReq := testcontainers.ContainerRequest{
		Image:        "temporalio/auto-setup:1.27.2",
		ExposedPorts: []string{TemporalDefaultPort + "/tcp"},
		Env: map[string]string{
			"DB":             "postgres12",
			"DB_PORT":        "5432",
			"POSTGRES_USER":  "temporal",
			"POSTGRES_PWD":   "temporal",
			"POSTGRES_SEEDS": postgresIP,
			// Enable worker versioning/deployments and force search attributes refresh
			"DYNAMIC_CONFIG_FILE_PATH": "/etc/temporal/dynamic_config.yaml",
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader: strings.NewReader(`
frontend.workerVersioningDataAPIs:
  - value: true
frontend.workerVersioningWorkflowAPIs:
  - value: true
frontend.enableDeployments:
  - value: true
system.forceSearchAttributesCacheRefreshOnRead:
  - value: true
`),
				ContainerFilePath: "/etc/temporal/dynamic_config.yaml",
				FileMode:          0644,
			},
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(TemporalDefaultPort+"/tcp").WithStartupTimeout(180*time.Second),
			wait.ForLog("Temporal server started").WithStartupTimeout(180*time.Second),
		).WithDeadline(240 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: temporalReq,
		Started:          true,
	})
	require.NoError(infra.t, err, "Failed to start Temporal container")

	infra.temporalContainer = container
	// Store postgres container for cleanup
	infra.postgresContainer = postgresContainer

	host, err := container.Host(ctx)
	require.NoError(infra.t, err, "Failed to get Temporal host")

	mappedPort, err := container.MappedPort(ctx, TemporalDefaultPort+"/tcp")
	require.NoError(infra.t, err, "Failed to get Temporal port")

	infra.TemporalEndpoint = net.JoinHostPort(host, mappedPort.Port())
	infra.t.Logf("Temporal started at %s", infra.TemporalEndpoint)

	// Wait for Temporal to be fully ready
	time.Sleep(5 * time.Second)

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  infra.TemporalEndpoint,
		Namespace: TemporalDefaultNamespace,
	})
	require.NoError(infra.t, err, "Failed to create Temporal client")

	infra.TemporalClient = c
	infra.t.Log("Temporal client connected")

	// Register custom search attributes for workflows
	infra.registerSearchAttributes(ctx)
}

// Teardown stops and removes all containers
func (infra *TestInfrastructure) Teardown() {
	infra.t.Log("Tearing down test infrastructure...")

	terminateCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if infra.TemporalClient != nil {
		infra.TemporalClient.Close()
	}

	if infra.temporalContainer != nil {
		if err := infra.temporalContainer.Terminate(terminateCtx); err != nil {
			infra.t.Logf("Warning: Failed to terminate Temporal container: %v", err)
		}
	}

	if infra.postgresContainer != nil {
		if err := infra.postgresContainer.Terminate(terminateCtx); err != nil {
			infra.t.Logf("Warning: Failed to terminate PostgreSQL container: %v", err)
		}
	}

	if infra.mailhogContainer != nil {
		if err := infra.mailhogContainer.Terminate(terminateCtx); err != nil {
			infra.t.Logf("Warning: Failed to terminate MailHog container: %v", err)
		}
	}

	if infra.localstackContainer != nil {
		if err := infra.localstackContainer.Terminate(terminateCtx); err != nil {
			infra.t.Logf("Warning: Failed to terminate LocalStack container: %v", err)
		}
	}

	infra.t.Log("Test infrastructure teardown complete")
}

// registerSearchAttributes registers custom search attributes needed for workflows
func (infra *TestInfrastructure) registerSearchAttributes(ctx context.Context) {
	// Wait a bit for Temporal's search attribute system to be ready
	time.Sleep(2 * time.Second)

	// Use the actual typed search attributes from models.temporal.go
	// Each SearchAttributeKey has GetName() and GetValueType() methods
	searchAttributes := []interface {
		GetName() string
		GetValueType() enums.IndexedValueType
	}{
		models.TypedSearchAttributeStatus,
		models.TypedSearchAttributeTask,
		models.TypedSearchAttributeUser,
		models.TypedSearchAttributeRole,
		models.TypedSearchAttributeWorkflow,
		models.TypedSearchAttributeProviders,
		models.TypedSearchAttributeReason,
		models.TypedSearchAttributeDuration,
		models.TypedSearchAttributeIdentities,
		models.TypedSearchAttributeApproved,
	}

	operatorClient := infra.TemporalClient.OperatorService()

	registered := 0
	for _, attr := range searchAttributes {
		// Try to add the search attribute - it may already exist
		_, err := operatorClient.AddSearchAttributes(ctx, &operatorservice.AddSearchAttributesRequest{
			Namespace: TemporalDefaultNamespace,
			SearchAttributes: map[string]enums.IndexedValueType{
				attr.GetName(): attr.GetValueType(),
			},
		})
		if err != nil {
			// Log but don't fail - some may already exist or hit limits
			infra.t.Logf("Note: Search attribute '%s' (%s) registration: %v",
				attr.GetName(), attr.GetValueType().String(), err)
		} else {
			registered++
			infra.t.Logf("Registered search attribute: %s (%s)",
				attr.GetName(), attr.GetValueType().String())
		}
	}

	infra.t.Logf("Custom search attributes registered: %d/%d", registered, len(searchAttributes))

	// Wait for the search attributes to propagate to the visibility store
	time.Sleep(3 * time.Second)
}

// TestTemporalAndLocalStackSetup verifies that all containers start correctly
func TestTemporalAndLocalStackSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set a reasonable timeout for container startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup infrastructure
	infra := SetupTestInfrastructure(t, ctx)
	defer infra.Teardown()

	// Verify LocalStack is accessible
	t.Run("LocalStack is accessible", func(t *testing.T) {
		require.NotEmpty(t, infra.LocalStackEndpoint, "LocalStack endpoint should be set")
		t.Logf("LocalStack endpoint: %s", infra.LocalStackEndpoint)
	})

	// Verify MailHog is accessible
	t.Run("MailHog is accessible", func(t *testing.T) {
		require.NotEmpty(t, infra.MailHogSMTP, "MailHog SMTP endpoint should be set")
		require.NotEmpty(t, infra.MailHogAPI, "MailHog API endpoint should be set")
		t.Logf("MailHog SMTP: %s, API: %s", infra.MailHogSMTP, infra.MailHogAPI)

		// Verify we can query the API
		emails, err := infra.GetEmails()
		require.NoError(t, err, "Should be able to query MailHog API")
		t.Logf("MailHog has %d emails", len(emails))
	})

	// Verify Temporal is accessible
	t.Run("Temporal is accessible", func(t *testing.T) {
		require.NotEmpty(t, infra.TemporalEndpoint, "Temporal endpoint should be set")
		require.NotNil(t, infra.TemporalClient, "Temporal client should be connected")
		t.Logf("Temporal endpoint: %s", infra.TemporalEndpoint)

		// Try to list workflows to verify connection
		_, err := infra.TemporalClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: TemporalDefaultNamespace,
			PageSize:  1,
		})
		require.NoError(t, err, "Should be able to list workflows")
	})
}
