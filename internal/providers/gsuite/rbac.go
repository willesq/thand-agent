package gsuite

import (
	"context"

	"github.com/thand-io/agent/internal/models"
)

func (p *gsuiteProvider) GetAuthorizedAccessUrl(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
	resp *models.AuthorizeRoleResponse,
) string {

	return p.GetConfig().GetStringWithDefault(
		"sso_start_url", "https://accounts.google.com/")
}
