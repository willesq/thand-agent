package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"golang.org/x/oauth2"
)

// RefreshIdentities fetches and caches user and team identities from GitHub
func (p *githubProvider) RefreshIdentities(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Refreshed GitHub identities in %s", elapsed)
	}()

	config := p.GetConfig()
	orgName, found := config.GetString("organization")
	if !found || len(orgName) == 0 {
		logrus.Warn("GitHub provider configuration missing 'organization', cannot fetch identities")
		return nil
	}

	var identities []models.Identity
	identitiesMap := make(map[string]*models.Identity)

	users, err := p.refreshUsers(ctx, orgName)
	if err != nil {
		logrus.WithError(err).Errorln("Error refreshing GitHub users")
	} else if len(users) > 0 {
		identities = append(identities, users...)
	}

	teams, err := p.refreshTeams(ctx, orgName)
	if err != nil {
		logrus.WithError(err).Errorln("Error refreshing GitHub teams")
	} else if len(teams) > 0 {
		identities = append(identities, teams...)
	}

	// Build map
	for i := range identities {
		identity := &identities[i]
		identitiesMap[strings.ToLower(identity.ID)] = identity
		identitiesMap[strings.ToLower(identity.Label)] = identity

		if identity.User != nil {
			identitiesMap[strings.ToLower(identity.User.Username)] = identity
		}
		if identity.Group != nil {
			identitiesMap[strings.ToLower(identity.Group.Name)] = identity
		}
	}

	p.indexMu.Lock()
	p.identities = identities
	p.identitiesMap = identitiesMap
	p.indexMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"count": len(identities),
		"org":   orgName,
	}).Debug("Refreshed GitHub identities")

	return nil
}

func (p *githubProvider) refreshUsers(ctx context.Context, orgName string) ([]models.Identity, error) {
	config := p.GetConfig()
	token, foundToken := config.GetString("token")
	if !foundToken || len(strings.TrimSpace(token)) == 0 {
		return nil, fmt.Errorf("GitHub token is missing or invalid in configuration")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var identities []models.Identity
	opts := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		members, resp, err := client.Organizations.ListMembers(ctx, orgName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch organization members: %w", err)
		}

		for _, member := range members {
			if member == nil {
				continue
			}

			// Extract user details
			login := member.GetLogin()
			id := member.GetID()

			if len(login) == 0 || id == 0 {
				continue
			}

			// Fetch full user details to get email
			// Note: This might hit rate limits for large organizations
			// fullUser, _, err := client.Users.Get(ctx, login)
			// if err != nil {
			//	logrus.Warnf("Failed to fetch user details for %s: %v", login, err)
			// } else {
			//	member = fullUser
			//}

			userId := fmt.Sprintf("%d", id)

			user := &models.User{
				ID:       userId,
				Username: login,
				Email:    member.GetEmail(),
				Source:   "github",
			}

			identity := models.Identity{
				ID:    userId,
				Label: login,
				User:  user,
			}
			identities = append(identities, identity)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return identities, nil
}

func (p *githubProvider) refreshTeams(ctx context.Context, orgName string) ([]models.Identity, error) {
	config := p.GetConfig()
	token, _ := config.GetString("token")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var identities []models.Identity
	opts := &github.ListOptions{PerPage: 100}

	for {
		teams, resp, err := client.Teams.ListTeams(ctx, orgName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch organization teams: %w", err)
		}

		for _, team := range teams {
			if team == nil {
				continue
			}

			name := team.GetName()
			slug := team.GetSlug()
			id := team.GetID()

			if len(name) == 0 || id == 0 {
				continue
			}

			teamId := fmt.Sprintf("%d", id)

			group := &models.Group{
				ID:   teamId,
				Name: name,
			}

			// Use slug as label if available, else name
			label := name
			if len(slug) != 0 {
				label = slug
			}

			identity := models.Identity{
				ID:    teamId,
				Label: label,
				Group: group,
			}
			identities = append(identities, identity)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return identities, nil
}

// GetIdentity retrieves a single identity by ID or name
func (p *githubProvider) GetIdentity(ctx context.Context, identity string) (*models.Identity, error) {
	p.indexMu.RLock()
	defer p.indexMu.RUnlock()

	if p.identitiesMap == nil {
		return nil, fmt.Errorf("identities not initialized")
	}

	if id, ok := p.identitiesMap[strings.ToLower(identity)]; ok {
		return id, nil
	}

	return nil, fmt.Errorf("identity not found: %s", identity)
}

// ListIdentities returns all cached identities
func (p *githubProvider) ListIdentities(ctx context.Context, filters ...string) ([]models.Identity, error) {
	p.indexMu.RLock()
	defer p.indexMu.RUnlock()

	if p.identities == nil {
		return nil, fmt.Errorf("identities not initialized")
	}

	// Filter implementation could be added here if needed
	// For now return all
	return p.identities, nil
}
