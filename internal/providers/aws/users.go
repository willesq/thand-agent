package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/temporal"
)

func (p *awsProvider) CanSynchronizeUsers() bool {
	return true
}

func (p *awsProvider) SynchronizeUsers(ctx context.Context, req models.SynchronizeUsersRequest) (*models.SynchronizeUsersResponse, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Refreshed AWS user identities in %s", elapsed)
	}()

	// 1. Get Identity Store ID
	resp, err := p.ssoAdminService.ListInstances(ctx, &ssoadmin.ListInstancesInput{})
	if err != nil {

		if req.Pagination == nil {

			// This is an initial request. If we've failed to get any users
			// this is probably a permission error.

			return nil, temporal.NewNonRetryableApplicationError(
				"Failed to list identity center instances",
				"IdentityCenterRequest",
				err,
			)
		}

		return nil, fmt.Errorf("failed to list SSO instances: %w", err)
	}

	if len(resp.Instances) == 0 {
		logrus.Warn("No SSO instances found, skipping user synchronization")
		return &models.SynchronizeUsersResponse{}, nil
	}

	identityStoreId := resp.Instances[0].IdentityStoreId
	if identityStoreId == nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"identity store ID not found in SSO instance",
			"IdentityCenterRequest",
		)
	}
	}

	if req.Pagination == nil {
		req.Pagination = &models.PaginationOptions{
			PageSize: 100,
		}
	}

	input := &identitystore.ListUsersInput{
		IdentityStoreId: identityStoreId,
		MaxResults:      aws.Int32(int32(req.Pagination.PageSize)),
	}

	if len(req.Pagination.Token) != 0 {
		input.NextToken = aws.String(req.Pagination.Token)
	}

	usersResp, err := p.identityStoreClient.ListUsers(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var identities []models.Identity
	for _, user := range usersResp.Users {
		userId := *user.UserId
		userName := *user.UserName

		var email string
		// Emails is a list of Email objects
		if len(user.Emails) > 0 {
			email = *user.Emails[0].Value
		}

		var displayName string
		if user.DisplayName != nil {
			displayName = *user.DisplayName
		} else {
			displayName = userName
		}

		identity := models.Identity{
			ID:    userId,
			Label: displayName,
			User: &models.User{
				ID:       userId,
				Username: userName,
				Email:    email,
				Name:     displayName,
				Source:   "aws-identity-center",
			},
		}
		identities = append(identities, identity)
	}

	response := &models.SynchronizeUsersResponse{
		Identities: identities,
	}

	if usersResp.NextToken != nil {
		response.Pagination = &models.PaginationOptions{
			Token:    *usersResp.NextToken,
			PageSize: req.Pagination.PageSize,
		}
	}

	logrus.WithFields(logrus.Fields{
		"count": len(identities),
	}).Debug("Refreshed AWS user identities")

	return response, nil
}
