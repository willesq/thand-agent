package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

func (p *kubernetesProvider) LoadRoles() error {
	// Kubernetes doesn't have predefined roles like AWS IAM managed policies
	// Roles come from configuration, not from the provider itself
	p.roles = []models.ProviderRole{}

	// Create empty index
	mapping := bleve.NewIndexMapping()
	rolesIndex, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return fmt.Errorf("failed to create roles search index: %w", err)
	}

	p.rolesIndex = rolesIndex

	logrus.Debug("Kubernetes provider: No built-in roles (roles come from configuration)")
	return nil
}

func (p *kubernetesProvider) GetRole(ctx context.Context, role string) (*models.ProviderRole, error) {
	// First try exact match
	for _, r := range p.roles {
		if r.Name == role {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("role %s not found", role)
}

func (p *kubernetesProvider) ListRoles(ctx context.Context, filters ...string) ([]models.ProviderRole, error) {
	if len(filters) == 0 {
		return p.roles, nil
	}

	query := strings.Join(filters, " ")
	searchQuery := bleve.NewMatchQuery(query)
	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = len(p.roles) // Return all matches

	searchResult, err := p.rolesIndex.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var results []models.ProviderRole
	for _, hit := range searchResult.Hits {
		// Find the role by ID
		for _, role := range p.roles {
			if role.Name == hit.ID {
				results = append(results, role)
				break
			}
		}
	}

	return results, nil
}
