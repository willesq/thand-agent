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
	parsedEc2Docs map[string]string
	ec2DocsOnce   sync.Once
	ec2DocsErr    error
)

// GetParsedAwsDocs returns the pre-parsed AWS docs map from FlatBuffer
func GetParsedAwsDocs() (map[string]string, error) {
	ec2DocsOnce.Do(func() {

		parsedEc2Docs = make(map[string]string)

		// Parse FlatBuffer
		permissionsList := aws.GetRootAsPermissionsList(awsDocsFb, 0)

		// Extract permissions
		for i := 0; i < permissionsList.PermissionsLength(); i++ {
			var permission aws.Permission
			if permissionsList.Permissions(&permission, i) {
				name := string(permission.Name())
				description := string(permission.Description())
				parsedEc2Docs[name] = description
			}
		}
	})
	return parsedEc2Docs, ec2DocsErr
}

type AwsManagedPolicies struct {
	Policies []AwsManagedPolicy
}

type AwsManagedPolicy struct {
	Name string
}

var (
	parsedEc2Roles AwsManagedPolicies
	ec2RolesOnce   sync.Once
	ec2RolesErr    error
)

// GetParsedAwsRoles returns the pre-parsed AWS roles struct from FlatBuffer
func GetParsedAwsRoles() (AwsManagedPolicies, error) {
	ec2RolesOnce.Do(func() {
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

		parsedEc2Roles = AwsManagedPolicies{Policies: policies}
	})
	return parsedEc2Roles, ec2RolesErr
}
