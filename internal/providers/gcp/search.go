package gcp

import (
	"fmt"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

func buildIndices(permissions []models.ProviderPermission, roles []models.ProviderRole) (bleve.Index, bleve.Index, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Built GCP search indices in %s", elapsed)
	}()

	// Create in-memory Bleve indices
	permissionsMapping := bleve.NewIndexMapping()
	permissionsIndex, err := bleve.NewMemOnly(permissionsMapping)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create permissions search index: %v", err)
	}

	rolesMapping := bleve.NewIndexMapping()
	rolesIndex, err := bleve.NewMemOnly(rolesMapping)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create roles search index: %v", err)
	}

	// Index permissions
	for _, perm := range permissions {
		if err := permissionsIndex.Index(perm.Name, perm); err != nil {
			return nil, nil, fmt.Errorf("failed to index permission %s: %v", perm.Name, err)
		}
	}

	// Index roles
	for _, role := range roles {
		if err := rolesIndex.Index(role.Name, role); err != nil {
			return nil, nil, fmt.Errorf("failed to index role %s: %v", role.Name, err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"permissions": len(permissions),
		"roles":       len(roles),
	}).Debug("GCP search indices ready")

	return permissionsIndex, rolesIndex, nil
}
