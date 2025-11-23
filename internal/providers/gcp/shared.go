package gcp

import (
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/thand-io/agent/internal/models"
)

type gcpData struct {
	permissions      []models.ProviderPermission
	permissionsMap   map[string]*models.ProviderPermission
	permissionsIndex bleve.Index

	roles      []models.ProviderRole
	rolesMap   map[string]*models.ProviderRole
	rolesIndex bleve.Index

	indexReady chan struct{}
}

type gcpSingleton struct {
	data *gcpData
	err  error
	once sync.Once
}

var (
	sharedDataMap = make(map[string]*gcpSingleton)
	sharedDataMu  sync.Mutex
)

func getSharedData(stage string) (*gcpData, error) {
	sharedDataMu.Lock()
	singleton, ok := sharedDataMap[stage]
	if !ok {
		singleton = &gcpSingleton{}
		sharedDataMap[stage] = singleton
	}
	sharedDataMu.Unlock()

	singleton.once.Do(func() {
		data := &gcpData{
			indexReady: make(chan struct{}),
		}
		var err error

		data.permissions, data.permissionsMap, err = loadPermissions(stage)
		if err != nil {
			singleton.err = err
			return
		}

		data.roles, data.rolesMap, err = loadRoles(stage)
		if err != nil {
			singleton.err = err
			return
		}

		// Build indices in background
		go func() {
			pIdx, rIdx, err := buildIndices(data.permissions, data.roles)
			if err != nil {
				// Log error but don't fail the whole provider initialization
				return
			}
			data.permissionsIndex = pIdx
			data.rolesIndex = rIdx
			close(data.indexReady)
		}()

		singleton.data = data
	})

	return singleton.data, singleton.err
}
