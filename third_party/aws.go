package third_party

import (
	"encoding/json"
	"sync"
)

var (
	parsedEc2Docs map[string]string
	ec2DocsOnce   sync.Once
	ec2DocsErr    error
)

// GetParsedEc2Docs returns the pre-parsed EC2 docs map, parsing it once on first call
func GetParsedEc2Docs() (map[string]string, error) {
	ec2DocsOnce.Do(func() {
		ec2DocsErr = json.Unmarshal(ec2docs, &parsedEc2Docs)
	})
	return parsedEc2Docs, ec2DocsErr
}

type AwsManagedPolicies struct {
	Policies []AwsManagedPolicy `json:"policies"`
}

type AwsManagedPolicy struct {
	Name string `json:"name"`
}

var (
	parsedEc2Roles AwsManagedPolicies
	ec2RolesOnce   sync.Once
	ec2RolesErr    error
)

// GetParsedEc2Roles returns the pre-parsed EC2 roles struct, parsing it once on first call
func GetParsedEc2Roles() (AwsManagedPolicies, error) {
	ec2RolesOnce.Do(func() {
		ec2RolesErr = json.Unmarshal(ec2roles, &parsedEc2Roles)
	})
	return parsedEc2Roles, ec2RolesErr
}
