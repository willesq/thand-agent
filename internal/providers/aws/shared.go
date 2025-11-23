package aws

import (
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/thand-io/agent/internal/models"
)

type awsData struct {
	permissions      []models.ProviderPermission
	permissionsMap   map[string]*models.ProviderPermission
	permissionsIndex bleve.Index

	roles      []models.ProviderRole
	rolesMap   map[string]*models.ProviderRole
	rolesIndex bleve.Index
}

var (
	sharedData     *awsData
	sharedDataOnce sync.Once
	sharedDataErr  error
)

func getSharedData() (*awsData, error) {
	sharedDataOnce.Do(func() {
		sharedData = &awsData{}
		var err error

		sharedData.permissions, sharedData.permissionsMap, err = loadPermissions()
		if err != nil {
			sharedDataErr = err
			return
		}

		sharedData.roles, sharedData.rolesMap, err = loadRoles()
		if err != nil {
			sharedDataErr = err
			return
		}

		sharedData.permissionsIndex, sharedData.rolesIndex, err = buildIndices(sharedData.permissions, sharedData.roles)
		if err != nil {
			sharedDataErr = err
			return
		}
	})
	return sharedData, sharedDataErr
}
