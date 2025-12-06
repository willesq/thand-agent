package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"golang.org/x/oauth2"
)

func (p *githubProvider) CanSynchronizeGroups() bool {
	return true
}

// Sync fetches and caches user and team identities from GitHub
func (p *githubProvider) SynchronizeGroups(ctx context.Context, req models.SynchronizeGroupsRequest) (*models.SynchronizeGroupsResponse, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Refreshed GitHub identities in %s", elapsed)
	}()

	config := p.GetConfig()
	orgName, found := config.GetString("organization")

	if !found || len(orgName) == 0 {
		logrus.Warn("GitHub provider configuration missing 'organization', cannot fetch identities")
		return nil, fmt.Errorf("missing required GitHub configuration: organization")
	}

	token, _ := config.GetString("token")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	if req.Pagination == nil {
		req.Pagination = &models.PaginationOptions{
			Page:     1,
			PageSize: 100,
		}
	}

	var identities []models.Identity
	opts := &github.ListOptions{
		Page:    req.Pagination.Page,
		PerPage: req.Pagination.PageSize,
	}

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

	response := models.SynchronizeGroupsResponse{
		Identities: identities,
	}

	if resp.NextPage != 0 {
		response.Pagination = &models.PaginationOptions{
			Page:     resp.NextPage,
			PageSize: req.Pagination.PageSize,
		}
	}

	logrus.WithFields(logrus.Fields{
		"count": len(identities),
		"org":   orgName,
	}).Debug("Refreshed GitHub identities")

	return &response, nil
}
