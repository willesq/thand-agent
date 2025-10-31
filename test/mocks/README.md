# Test Mocks

This directory contains mock implementations for testing.

## Structure

```
test/mocks/
└── providers/          # Mock provider registrations
    ├── aws.go         # AWS provider mock registration
    ├── azure.go       # Azure provider mock registration
    ├── gcp.go         # GCP provider mock registration
    └── kubernetes.go  # Kubernetes provider mock registration
```

## Provider Mocks

The provider mocks prevent tests from connecting to actual cloud services (AWS, GCP, Azure, Kubernetes) while still loading roles and permissions from embedded data.

### How It Works

1. **Mock implementations** are defined in each provider's directory:
   - `internal/providers/aws/mock.go`
   - `internal/providers/gcp/mock.go`
   - `internal/providers/azure/mock.go`
   - `internal/providers/kubernetes/mock.go`

2. **Mock registrations** in this directory (`test/mocks/providers/`) use `init()` functions to override the real providers:
   - They import the provider package and call `providers.Set(ProviderName, NewMockProvider())`
   - This happens automatically when the package is imported with `_` (blank import)

3. **Test helpers** import the mocks:
   - `internal/config/test_helpers.go` imports `_ "github.com/thand-io/agent/test/mocks/providers"`
   - This ensures mocks are registered before any tests run

### Provider Name Constants

Each provider defines a `ProviderName` constant:
- `aws.ProviderName = "aws"`
- `gcp.ProviderName = "gcp"`
- `azure.ProviderName = "azure"`
- `kubernetes.ProviderName = "kubernetes"`

These constants are used for registration to ensure consistency.

### Benefits

- **No cloud connections**: Tests run without AWS/GCP/Azure/Kubernetes API calls
- **Fast execution**: Tests complete in milliseconds instead of seconds
- **Offline testing**: No internet connection or cloud credentials required
- **Data loaded**: Roles and permissions still load from embedded IAM datasets
