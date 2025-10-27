package data

import (
	_ "embed"
	"sync"

	"github.com/thand-io/agent/internal/data/iam-dataset/generated/azure"
)

//go:embed iam-dataset/azure/built-in-roles.fb
var azureRolesFb []byte

//go:embed iam-dataset/azure/provider-operations.fb
var azurePermissionsFb []byte

type AzureBuiltInRole struct {
	Name        string
	Description string
}

type AzureResourceProviderOperation struct {
	Name        string
	Description string
}

var (
	parsedAzureRoles []AzureBuiltInRole
	azureRolesOnce   sync.Once
	azureRolesErr    error
)

var (
	parsedAzurePermissions []AzureResourceProviderOperation
	azurePermissionsOnce   sync.Once
	azurePermissionsErr    error
)

// GetParsedAzureRoles returns the pre-parsed Azure built-in roles from FlatBuffer
func GetParsedAzureRoles() ([]AzureBuiltInRole, error) {
	azureRolesOnce.Do(func() {
		// Parse FlatBuffer
		builtInRolesList := azure.GetRootAsBuiltInRolesList(azureRolesFb, 0)

		// Extract roles
		for i := 0; i < builtInRolesList.RolesLength(); i++ {
			var role azure.BuiltInRole
			if builtInRolesList.Roles(&role, i) {
				parsedAzureRoles = append(parsedAzureRoles, AzureBuiltInRole{
					Name:        string(role.Name()),
					Description: string(role.Description()),
				})
			}
		}
	})
	return parsedAzureRoles, azureRolesErr
}

// GetParsedAzurePermissions returns the pre-parsed Azure permissions from FlatBuffer
func GetParsedAzurePermissions() ([]AzureResourceProviderOperation, error) {
	azurePermissionsOnce.Do(func() {
		// Parse FlatBuffer
		resourceProvidersList := azure.GetRootAsResourceProvidersList(azurePermissionsFb, 0)

		// Extract operations from all providers
		for i := 0; i < resourceProvidersList.ProvidersLength(); i++ {
			var provider azure.ResourceProvider
			if resourceProvidersList.Providers(&provider, i) {
				for j := 0; j < provider.OperationsLength(); j++ {
					var operation azure.ResourceProviderOperation
					if provider.Operations(&operation, j) {
						parsedAzurePermissions = append(parsedAzurePermissions, AzureResourceProviderOperation{
							Name:        string(operation.Name()),
							Description: string(operation.Description()),
						})
					}
				}
			}
		}
	})
	return parsedAzurePermissions, azurePermissionsErr
}
