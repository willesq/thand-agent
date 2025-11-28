package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"google.golang.org/api/cloudresourcemanager/v1"
)

// RefreshIdentities fetches and caches user and group identities from GCP IAM
func (p *gcpProvider) RefreshIdentities(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Refreshed GCP identities in %s", elapsed)
	}()

	projectId := p.GetProjectId()

	// Get current IAM policy to extract all members - request version 3 for conditions support
	policy, err := p.crmClient.Projects.GetIamPolicy(projectId, &cloudresourcemanager.GetIamPolicyRequest{
		Options: &cloudresourcemanager.GetPolicyOptions{
			RequestedPolicyVersion: 3,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %w", err)
	}

	var identities []models.Identity
	identitiesMap := make(map[string]*models.Identity)
	memberSet := make(map[string]bool) // To deduplicate members across bindings

	var userCount, groupCount int

	// Extract unique members from all bindings
	for _, binding := range policy.Bindings {
		for _, member := range binding.Members {
			if memberSet[member] {
				continue // Skip duplicates
			}
			memberSet[member] = true

			identity, identityType := parseMemberToIdentity(member)
			if identity == nil {
				continue
			}

			switch identityType {
			case "user":
				userCount++
			case "group":
				groupCount++
			}

			identities = append(identities, *identity)
		}
	}

	// Build the identities map
	for i := range identities {
		// Map by ID (lowercase)
		identitiesMap[strings.ToLower(identities[i].ID)] = &identities[i]
		// Map by label (lowercase)
		identitiesMap[strings.ToLower(identities[i].Label)] = &identities[i]

		// For users, also map by email
		if identities[i].User != nil && identities[i].User.Email != "" {
			identitiesMap[strings.ToLower(identities[i].User.Email)] = &identities[i]
		}
		// For groups, also map by name/email
		if identities[i].Group != nil {
			if identities[i].Group.Name != "" {
				identitiesMap[strings.ToLower(identities[i].Group.Name)] = &identities[i]
			}
			if identities[i].Group.Email != "" {
				identitiesMap[strings.ToLower(identities[i].Group.Email)] = &identities[i]
			}
		}
	}

	p.indexMu.Lock()
	p.identities = identities
	p.identitiesMap = identitiesMap
	p.indexMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"users":      userCount,
		"groups":     groupCount,
		"total":      len(identities),
		"project_id": projectId,
	}).Debug("Refreshed GCP identities from IAM policy")

	return nil
}

// parseMemberToIdentity converts a GCP IAM member string to an Identity
// Member format: "user:email@example.com", "group:group@example.com"
// Only users and groups are supported for now
func parseMemberToIdentity(member string) (*models.Identity, string) {
	parts := strings.SplitN(member, ":", 2)
	if len(parts) != 2 {
		return nil, ""
	}

	memberType := parts[0]
	memberValue := parts[1]

	switch memberType {
	case "user":
		// Regular user account
		name := extractNameFromEmail(memberValue)
		return &models.Identity{
			ID:    memberValue,
			Label: name,
			User: &models.User{
				ID:       memberValue,
				Email:    memberValue,
				Username: strings.Split(memberValue, "@")[0],
				Name:     name,
				Source:   "gcp",
			},
		}, "user"

	case "group":
		// Google group
		name := extractNameFromEmail(memberValue)
		return &models.Identity{
			ID:    memberValue,
			Label: name,
			Group: &models.Group{
				ID:    memberValue,
				Name:  name,
				Email: memberValue,
			},
		}, "group"

	default:
		// Skip service accounts, domains, allUsers, allAuthenticatedUsers, etc.
		return nil, ""
	}
}

// extractNameFromEmail extracts a display name from an email address
func extractNameFromEmail(email string) string {
	// Try to extract name from email format (e.g., "john.doe@example.com" -> "John Doe")
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return email
	}

	localPart := parts[0]
	// Replace dots, underscores, and hyphens with spaces
	name := strings.ReplaceAll(localPart, ".", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// Capitalize each word
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// GetIdentity retrieves a specific identity (user or group) from GCP
func (p *gcpProvider) GetIdentity(ctx context.Context, identity string) (*models.Identity, error) {
	// Try to get from cache first
	p.indexMu.RLock()
	identitiesMap := p.identitiesMap
	p.indexMu.RUnlock()

	if identitiesMap != nil {
		if id, exists := identitiesMap[strings.ToLower(identity)]; exists {
			return id, nil
		}
	}

	// If not in cache, refresh and try again
	if err := p.RefreshIdentities(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh identities: %w", err)
	}

	p.indexMu.RLock()
	defer p.indexMu.RUnlock()

	if id, exists := p.identitiesMap[strings.ToLower(identity)]; exists {
		return id, nil
	}

	return nil, fmt.Errorf("identity not found: %s", identity)
}

// ListIdentities lists all identities (users and groups) from GCP IAM
func (p *gcpProvider) ListIdentities(ctx context.Context, filters ...string) ([]models.Identity, error) {
	// Ensure we have fresh data
	if err := p.RefreshIdentities(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh identities: %w", err)
	}

	p.indexMu.RLock()
	identities := p.identities
	p.indexMu.RUnlock()

	// If no filters, return all identities
	if len(filters) == 0 {
		return identities, nil
	}

	// Apply filters
	var filtered []models.Identity
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, identity := range identities {
		// Check if any filter matches the identity
		if strings.Contains(strings.ToLower(identity.Label), filterText) ||
			strings.Contains(strings.ToLower(identity.ID), filterText) ||
			(identity.User != nil && strings.Contains(strings.ToLower(identity.User.Email), filterText)) ||
			(identity.User != nil && strings.Contains(strings.ToLower(identity.User.Name), filterText)) ||
			(identity.Group != nil && strings.Contains(strings.ToLower(identity.Group.Name), filterText)) ||
			(identity.Group != nil && strings.Contains(strings.ToLower(identity.Group.Email), filterText)) {
			filtered = append(filtered, identity)
		}
	}

	return filtered, nil
}
