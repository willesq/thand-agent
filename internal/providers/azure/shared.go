package azure

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/data"
	"github.com/thand-io/agent/internal/models"
)

type azureData struct {
	permissions []models.ProviderPermission
	roles       []models.ProviderRole

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

		sharedData.permissions, err = loadPermissions()
		if err != nil {
			sharedDataErr = err
			return
		}

		sharedData.roles, err = loadRoles()
		if err != nil {
			sharedDataErr = err
			return
		}

	})
	return sharedData, sharedDataErr
}

func loadPermissions() ([]models.ProviderPermission, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed Azure permissions in %s", elapsed)
	}()

	// Get pre-parsed Azure permissions from data package
	azureOperations, err := data.GetParsedAzurePermissions()
	if err != nil {
		return nil, fmt.Errorf("failed to get parsed Azure permissions: %w", err)
	}

	var permissions []models.ProviderPermission

	for _, operation := range azureOperations {
		permission := models.ProviderPermission{
			Name:        operation.Name,
			Description: operation.Description,
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func loadRoles() ([]models.ProviderRole, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed Azure roles in %s", elapsed)
	}()

	// Get pre-parsed Azure roles from data package
	azureRoles, err := data.GetParsedAzureRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get parsed Azure roles: %w", err)
	}

	var roles []models.ProviderRole

	for _, role := range azureRoles {
		r := models.ProviderRole{
			Name:        role.Name,
			Description: role.Description,
		}
		roles = append(roles, r)
	}

	return roles, nil
}
