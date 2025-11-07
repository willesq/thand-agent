package cloudflare

import (
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus"
)

// buildSearchIndex builds the Bleve search index for Cloudflare permissions and roles
func (p *cloudflareProvider) buildSearchIndex() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Built Cloudflare search indices in %s", elapsed)
	}()

	// Create in-memory Bleve indices
	permissionsMapping := bleve.NewIndexMapping()
	permissionsIndex, err := bleve.NewMemOnly(permissionsMapping)
	if err != nil {
		logrus.Errorf("Failed to create permissions search index: %v", err)
		return
	}

	rolesMapping := bleve.NewIndexMapping()
	rolesIndex, err := bleve.NewMemOnly(rolesMapping)
	if err != nil {
		logrus.Errorf("Failed to create roles search index: %v", err)
		return
	}

	// Index permissions
	for _, perm := range p.permissions {
		if err := permissionsIndex.Index(perm.Name, perm); err != nil {
			logrus.Errorf("Failed to index permission %s: %v", perm.Name, err)
			return
		}
	}

	// Index roles
	for _, role := range p.roles {
		if err := rolesIndex.Index(role.Name, role); err != nil {
			logrus.Errorf("Failed to index role %s: %v", role.Name, err)
			return
		}
	}

	// Safely update the indices
	p.indexMu.Lock()
	p.permissionsIndex = permissionsIndex
	p.rolesIndex = rolesIndex
	p.indexMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"permissions": len(p.permissions),
		"roles":       len(p.roles),
	}).Debug("Cloudflare search indices ready")
}
