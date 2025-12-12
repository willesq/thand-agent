package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

func (p *awsProvider) CanSynchronizeIdentities() bool {
	return true
}

func (p *awsProvider) SynchronizeIdentities(ctx context.Context, req models.SynchronizeIdentitiesRequest) (*models.SynchronizeIdentitiesResponse, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Refreshed AWS IAM users in %s", elapsed)
	}()

	input := &iam.ListUsersInput{}
	if req.Pagination != nil {
		if len(req.Pagination.Token) != 0 {
			input.Marker = &req.Pagination.Token
		}
		if req.Pagination.PageSize > 0 {
			maxItems := int32(req.Pagination.PageSize)
			input.MaxItems = &maxItems
		}
	}

	output, err := p.service.ListUsers(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list IAM users: %w", err)
	}

	var identities []models.Identity
	for _, user := range output.Users {
		userId := *user.UserId
		userName := *user.UserName
		//arn := *user.Arn

		// IAM users don't have emails
		identity := models.Identity{
			ID:    userId,
			Label: userName,
			User: &models.User{
				ID:       userId,
				Username: userName,
				Name:     userName,
				Source:   "iam",
				//Metadata: map[string]any{
				//	"arn": arn,
				//},
			},
		}
		identities = append(identities, identity)
	}

	response := &models.SynchronizeIdentitiesResponse{
		Identities: identities,
	}

	if output.IsTruncated && output.Marker != nil {
		response.Pagination = &models.PaginationOptions{
			Token:    *output.Marker,
			PageSize: req.Pagination.PageSize,
		}
	}

	return response, nil
}
