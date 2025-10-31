//go:build test

package mocks

import (
	"github.com/thand-io/agent/internal/providers"
	"github.com/thand-io/agent/internal/providers/aws"
)

// SetupMockAwsProvider overrides the AWS provider with the mock version for testing.
// Call this function explicitly in your test setup.
func SetupMockAwsProvider() {
	providers.Set("aws", aws.NewMockAwsProvider())
}
