package aws

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/data"
	"github.com/thand-io/agent/internal/models"
)

type awsData struct {
	permissions []models.ProviderPermission
	roles       []models.ProviderRole
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
		logrus.Debugf("Parsed AWS permissions in %s", elapsed)
	}()

	// Get pre-parsed AWS permissions from data package
	docs, err := data.GetParsedAwsDocs()
	if err != nil {
		return nil, fmt.Errorf("failed to get parsed AWS permissions: %w", err)
	}

	var permissions []models.ProviderPermission

	// Convert to slice and create fast lookup map
	for name, description := range docs {
		perm := models.ProviderPermission{
			ID:          strings.ToLower(name),
			Name:        name,
			Description: description,
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func loadRoles() ([]models.ProviderRole, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed AWS roles in %s", elapsed)
	}()

	// Get pre-parsed AWS roles from data package
	docs, err := data.GetParsedAwsRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get parsed AWS roles: %w", err)
	}

	var roles []models.ProviderRole

	// Convert to slice and create fast lookup map
	for _, policy := range docs.Policies {
		role := models.ProviderRole{
			Name: policy.Name,
		}
		roles = append(roles, role)
	}

	return roles, nil
}
