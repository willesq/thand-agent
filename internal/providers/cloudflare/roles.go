package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// LoadRoles loads Cloudflare roles from the API
func (p *cloudflareProvider) LoadRoles(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Loaded Cloudflare roles in %s", elapsed)
	}()

	accountRC := cloudflare.AccountIdentifier(p.accountID)
	//roles, err := p.client.ListAccountRoles(ctx, accountRC, cloudflare.ListAccountRolesParams{})
	roles, err := p.client.ListPermissionGroups(ctx, accountRC, cloudflare.ListPermissionGroupParams{})
	if err != nil {
		return fmt.Errorf("failed to list account roles: %w", err)
	}

	var rolesData []models.ProviderRole

	// Convert to slice and create fast lookup map
	for _, role := range roles {
		newRole := models.ProviderRole{
			Id:   role.ID,
			Name: role.Name,
			Role: role, // Store the full Cloudflare role object for later use
		}
		rolesData = append(rolesData, newRole)

		logrus.WithFields(logrus.Fields{
			"role":    role.Name,
			"role_id": role.ID,
		}).Debug("Loaded role")
	}

	p.SetRoles(rolesData)

	logrus.WithFields(logrus.Fields{
		"roles": len(rolesData),
	}).Debug("Loaded Cloudflare roles, building search index in background")

	return nil
}
