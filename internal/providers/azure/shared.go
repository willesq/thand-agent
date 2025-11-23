package azure

import (
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

type azureData struct {
	permissions      []models.ProviderPermission
	permissionsMap   map[string]*models.ProviderPermission
	permissionsIndex bleve.Index

	roles      []models.ProviderRole
	rolesMap   map[string]*models.ProviderRole
	rolesIndex bleve.Index

	indexReady chan struct{}
}

var (
	sharedData     *azureData
	sharedDataOnce sync.Once
	sharedDataErr  error
)

func getSharedData() (*azureData, error) {
	sharedDataOnce.Do(func() {
		sharedData = &azureData{
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
			defer close(sharedData.indexReady)
			pIdx, rIdx, err := buildIndices(sharedData.permissions, sharedData.roles)
			if err != nil {
				logrus.WithError(err).Error("Failed to build Azure search indices")
				return
			}
			sharedData.permissionsIndex = pIdx
			sharedData.rolesIndex = rIdx
		}()
	})
	return sharedData, sharedDataErr
}
