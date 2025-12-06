package github

import (
	"github.com/thand-io/agent/internal/models"
)

var GitHubPermissions = []models.ProviderPermission{{
	Name:        "read",
	Description: "Read access",
}}
