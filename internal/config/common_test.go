package config

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// TestReadData_WorkflowStructure tests readData with realistic workflow structure
func TestReadData_WorkflowStructure(t *testing.T) {
	yamlInput := `version: "1.0"
workflows:
  deploy-workflow:
    name: "Deploy Application"
    description: "Workflow that triggers on push events"
    enabled: true`

	var definition models.WorkflowDefinitions
	result, err := common.ReadDataToInterface([]byte(yamlInput), definition)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the result has the expected structure
	assert.True(t, result.Version.Equal(version.Must(version.NewVersion("1.0"))))
	assert.NotNil(t, result.Workflows)

	workflow, exists := result.Workflows["deploy-workflow"]
	assert.True(t, exists, "workflow should exist")
	assert.Equal(t, "Deploy Application", workflow.Name)
	assert.Equal(t, "Workflow that triggers on push events", workflow.Description)
	assert.True(t, workflow.Enabled)
}

// TestReadData_YAMLWithProblematicKeysDirectValidation tests the core YAML parsing fix
func TestReadData_YAMLWithProblematicKeysDirectValidation(t *testing.T) {
	// Test serverless workflow YAML that uses 'on' in schedule - this would be broken with sigs.k8s.io/yaml
	yamlInput := `version: "1.0"
workflows:
  scheduled-workflow:
    name: "Scheduled Workflow"
    description: "Workflow with schedule trigger using on event consumption"
    enabled: true
    workflow:
      document:
        dsl: "1.0.1"
        namespace: "test"
        name: "scheduled-workflow"
        version: "1.0.0"
      schedule:
        on:
          any:
            - with:
                type: com.example.cron.trigger
                data: 
                  schedule: "0 0 * * *"
            - with:
                type: com.example.event.trigger
                data:
                  enabled: true
        cron: "0 0 * * *"
      do:
        - log:
            call: http
            with:
              method: POST
              url: "https://api.example.com/log"
              body:
                message: "Scheduled workflow executed"`

	var definition models.WorkflowDefinitions
	result, err := common.ReadDataToInterface([]byte(yamlInput), definition)
	require.NoError(t, err)
	require.NotNil(t, result)

	// The key test: marshal back to JSON and verify 'on' is preserved as a key
	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	t.Logf("Generated JSON: %s", jsonStr)

	// Verify that 'on' appears as a key in the JSON, not as boolean conversion
	assert.Contains(t, jsonStr, `"on":`, "JSON should contain 'on' as a key in schedule section")

	// Verify that boolean conversions ('true'/'false' as keys) don't appear from 'on' conversion
	// With the old sigs.k8s.io/yaml library, 'on:' would become '"true":' in JSON
	assert.NotContains(t, jsonStr, `"true":{"any"`, "JSON should not contain 'true' as a key from 'on' conversion")
	assert.NotContains(t, jsonStr, `"false":{"any"`, "JSON should not contain 'false' as a key from 'on' conversion")
}

func TestReadData_JSONInput(t *testing.T) {
	jsonInput := `{
		"version": "1.0",
		"workflows": {
			"test": {
				"name": "Test",
				"description": "Test workflow",
				"enabled": true
			}
		}
	}`

	var definition models.WorkflowDefinitions
	result, err := common.ReadDataToInterface([]byte(jsonInput), definition)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Version.Equal(version.Must(version.NewVersion("1.0"))))
	assert.NotNil(t, result.Workflows)

	workflow, exists := result.Workflows["test"]
	assert.True(t, exists)
	assert.Equal(t, "Test", workflow.Name)
}

func TestReadData_InvalidYAML(t *testing.T) {
	invalidYAML := `on:
  test: true
    invalid: indentation`

	var definition models.WorkflowDefinitions
	result, err := common.ReadDataToInterface([]byte(invalidYAML), definition)
	assert.Error(t, err, "should return error for invalid YAML")
	assert.Nil(t, result, "result should be nil for invalid YAML")
}

func TestReadData_EmptyInput(t *testing.T) {
	var definition models.WorkflowDefinitions
	result, err := common.ReadDataToInterface([]byte(""), definition)
	assert.Error(t, err, "should return error for empty input")
	assert.Nil(t, result, "result should be nil for empty input")
}

// Benchmark to ensure the YAML parsing performance is reasonable
func BenchmarkReadData_YAML(b *testing.B) {
	yamlInput := `version: "1.0"
workflows:
  test:
    name: "Test"
    description: "Benchmark test"
    enabled: true`

	var definition models.WorkflowDefinitions

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := common.ReadDataToInterface([]byte(yamlInput), definition)
		if err != nil {
			b.Fatal(err)
		}
	}
}
