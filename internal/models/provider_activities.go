package models

import (
	"context"

	"github.com/sirupsen/logrus"
)

type providerActivities struct {
	provider ProviderImpl
}

func (a *providerActivities) SynchronizeIdentities(
	ctx context.Context,
	req SynchronizeUsersRequest,
) (*SynchronizeUsersResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeIdentities activity")

	return a.provider.SynchronizeIdentities(ctx, req)
}

func (a *providerActivities) SynchronizeResources(
	ctx context.Context,
	req SynchronizeResourcesRequest,
) (*SynchronizeResourcesResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeResources activity")

	return a.provider.SynchronizeResources(ctx, req)
}

func (a *providerActivities) SynchronizeUsers(
	ctx context.Context,
	req SynchronizeUsersRequest,
) (*SynchronizeUsersResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeUsers activity")

	return a.provider.SynchronizeUsers(ctx, req)
}

func (a *providerActivities) SynchronizeGroups(
	ctx context.Context,
	req SynchronizeGroupsRequest,
) (*SynchronizeGroupsResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeGroups activity")

	return a.provider.SynchronizeGroups(ctx, req)
}

func (a *providerActivities) SynchronizePermissions(
	ctx context.Context,
	req SynchronizePermissionsRequest,
) (*SynchronizePermissionsResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizePermissions activity")

	return a.provider.SynchronizePermissions(ctx, req)
}

func (a *providerActivities) SynchronizeRoles(
	ctx context.Context,
	req SynchronizeRolesRequest,
) (*SynchronizeRolesResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pagination": req.Pagination,
	}).Infoln("Starting SynchronizeRoles activity")

	return a.provider.SynchronizeRoles(ctx, req)
}
