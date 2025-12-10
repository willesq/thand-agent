package data

import (
	_ "embed"
	"sync"

	"github.com/thand-io/agent/internal/data/iam-dataset/generated/aws"
)

//go:embed iam-dataset/aws/docs.fb
var awsDocsFb []byte

//go:embed iam-dataset/aws/managed_policies.fb
var awsRolesFb []byte

var (
	parsedAwsDocs map[string]string
	awsDocsOnce   sync.Once
	awsDocsErr    error
)

// GetParsedAwsDocs returns the pre-parsed AWS docs map from FlatBuffer
func GetParsedAwsDocs() (map[string]string, error) {

	awsDocsOnce.Do(func() {

		parsedAwsDocs = make(map[string]string)

		// Parse FlatBuffer
		permissionsList := aws.GetRootAsPermissionsList(awsDocsFb, 0)

		// Extract permissions
		for i := 0; i < permissionsList.PermissionsLength(); i++ {
			var permission aws.Permission
			if permissionsList.Permissions(&permission, i) {
				name := string(permission.Name())
				description := string(permission.Description())
				parsedAwsDocs[name] = description
			}
		}
	})

	return parsedAwsDocs, awsDocsErr
}

type AwsManagedPolicies struct {
	Policies []AwsManagedPolicy
}

type AwsManagedPolicy struct {
	Name string
}

var (
	parsedAwsRoles AwsManagedPolicies
	awsRolesOnce   sync.Once
	awsRolesErr    error
)

// GetParsedAwsRoles returns the pre-parsed AWS roles struct from FlatBuffer
func GetParsedAwsRoles() (AwsManagedPolicies, error) {
	awsRolesOnce.Do(func() {
		var policies []AwsManagedPolicy

		// Parse FlatBuffer
		managedPoliciesList := aws.GetRootAsManagedPoliciesList(awsRolesFb, 0)

		// Extract policies
		for i := 0; i < managedPoliciesList.PoliciesLength(); i++ {
			var policy aws.ManagedPolicy
			if managedPoliciesList.Policies(&policy, i) {
				name := string(policy.Name())
				policies = append(policies, AwsManagedPolicy{Name: name})
			}
		}

		parsedAwsRoles = AwsManagedPolicies{Policies: policies}
	})
	return parsedAwsRoles, awsRolesErr
}
