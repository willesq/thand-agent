package workflows_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/wait"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// TestInfrastructure holds all test containers and clients
type TestInfrastructure struct {
	t   *testing.T
	ctx context.Context

	// LocalStack
	localstackContainer testcontainers.Container
	LocalStackEndpoint  string

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

	// Start LocalStack
	infra.startLocalStack(ctx)

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
		Image:        "temporalio/auto-setup:1.25.0",
		ExposedPorts: []string{"7233/tcp"},
		Env: map[string]string{
			"DB":                   "postgres12",
			"DB_PORT":              "5432",
			"POSTGRES_USER":        "temporal",
			"POSTGRES_PWD":         "temporal",
			"POSTGRES_SEEDS":       postgresIP,
			"DYNAMIC_CONFIG_VALUE": "system.forceSearchAttributesCacheRefreshOnRead=true",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("7233/tcp").WithStartupTimeout(180*time.Second),
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

	mappedPort, err := container.MappedPort(ctx, "7233/tcp")
	require.NoError(infra.t, err, "Failed to get Temporal port")

	infra.TemporalEndpoint = net.JoinHostPort(host, mappedPort.Port())
	infra.t.Logf("Temporal started at %s", infra.TemporalEndpoint)

	// Wait for Temporal to be fully ready
	time.Sleep(5 * time.Second)

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  infra.TemporalEndpoint,
		Namespace: "default",
	})
	require.NoError(infra.t, err, "Failed to create Temporal client")

	infra.TemporalClient = c
	infra.t.Log("Temporal client connected")
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

	if infra.localstackContainer != nil {
		if err := infra.localstackContainer.Terminate(terminateCtx); err != nil {
			infra.t.Logf("Warning: Failed to terminate LocalStack container: %v", err)
		}
	}

	infra.t.Log("Test infrastructure teardown complete")
}

// TestTemporalAndLocalStackSetup verifies that both containers start correctly
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

	// Verify Temporal is accessible
	t.Run("Temporal is accessible", func(t *testing.T) {
		require.NotEmpty(t, infra.TemporalEndpoint, "Temporal endpoint should be set")
		require.NotNil(t, infra.TemporalClient, "Temporal client should be connected")
		t.Logf("Temporal endpoint: %s", infra.TemporalEndpoint)

		// Try to list workflows to verify connection
		_, err := infra.TemporalClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: "default",
			PageSize:  1,
		})
		require.NoError(t, err, "Should be able to list workflows")
	})
}
