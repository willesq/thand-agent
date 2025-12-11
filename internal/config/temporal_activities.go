package config

import (
	"fmt"
	"strings"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/temporal"
)

type thandActivities struct {
	config *Config
}

// PatchProviderUpstream patches the provider's upstream endpoint in the Thand server
// This sends updates for users, groups, roles, permissions, resources, etc.
// So that Thand can paginate through the data when the provider is synchronized
func (t *thandActivities) PatchProviderUpstream(
	activityMethod models.SynchronizeCapability,
	providerIdentifier string,
	resp any,
) error {

	c := t.config

	if !c.HasThandService() {
		return temporal.NewNonRetryableApplicationError(
			"Thand service is not configured",
			"ThandServiceNotConfigured",
			nil,
		)
	}

	baseUrl := c.DiscoverThandServerApiUrl()
	providerSyncUrl := fmt.Sprintf("%s/providers/%s/sync",
		strings.TrimSuffix(baseUrl, "/"),
		strings.ToLower(providerIdentifier),
	)

	upstream := &model.Endpoint{
		EndpointConfig: &model.EndpointConfiguration{
			URI: &model.LiteralUri{Value: providerSyncUrl},
			Authentication: &model.ReferenceableAuthenticationPolicy{
				AuthenticationPolicy: &model.AuthenticationPolicy{
					Bearer: &model.BearerAuthenticationPolicy{
						Token: c.Thand.ApiKey,
					},
				},
			},
		},
	}

	// Make patch request
	err := models.PatchProviderUpstream(
		activityMethod,
		upstream,
		resp,
	)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to send pagination patch to server")
	}

	return err

}
