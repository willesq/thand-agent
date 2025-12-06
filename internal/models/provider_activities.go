package models

import (
	"context"

	"github.com/sirupsen/logrus"
)

type ProviderActivities struct {
	provider ProviderImpl
}

func (a *ProviderActivities) SynchronizeIdentities(
	ctx context.Context,
	req SynchronizeUsersRequest,
) (*SynchronizeUsersResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeIdentities activity")

	return a.provider.SynchronizeIdentities(ctx, req)
}

func (a *ProviderActivities) SynchronizeResources(
	ctx context.Context,
	req SynchronizeResourcesRequest,
) (*SynchronizeResourcesResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeResources activity")

	return a.provider.SynchronizeResources(ctx, req)
}

func (a *ProviderActivities) SynchronizeUsers(
	ctx context.Context,
	req SynchronizeUsersRequest,
) (*SynchronizeUsersResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeUsers activity")

	return a.provider.SynchronizeUsers(ctx, req)
}

func (a *ProviderActivities) SynchronizeGroups(
	ctx context.Context,
	req SynchronizeGroupsRequest,
) (*SynchronizeGroupsResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeGroups activity")

	return a.provider.SynchronizeGroups(ctx, req)
}

func (a *ProviderActivities) SynchronizePermissions(
	ctx context.Context,
	req SynchronizePermissionsRequest,
) (*SynchronizePermissionsResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizePermissions activity")

	return a.provider.SynchronizePermissions(ctx, req)
}

func (a *ProviderActivities) SynchronizeRoles(
	ctx context.Context,
	req SynchronizeRolesRequest,
) (*SynchronizeRolesResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeRoles activity")

	return a.provider.SynchronizeRoles(ctx, req)
}
