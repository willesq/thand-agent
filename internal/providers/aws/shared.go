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

	indexReady chan struct{}
}

var (
	sharedData     *awsData
	sharedDataOnce sync.Once
	sharedDataErr  error
)

func getSharedData() (*awsData, error) {
	sharedDataOnce.Do(func() {
		sharedData = &awsData{
			indexReady: make(chan struct{}),
		}
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

		// Build indices in background
		go func() {
			pIdx, rIdx, err := buildIndices(sharedData.permissions, sharedData.roles)
			if err != nil {
				// Log error but don't fail the whole provider initialization
				// The provider will continue without search capabilities
				return
			}
			sharedData.permissionsIndex = pIdx
			sharedData.rolesIndex = rIdx
			close(sharedData.indexReady)
		}()
	})
	return sharedData, sharedDataErr
}
