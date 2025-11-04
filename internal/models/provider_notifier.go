package models

import (
	"context"
	"fmt"
)

type NotificationRequest map[string]any

type ProviderNotifier interface {

	// Allow this provider to send notifications
	SendNotification(ctx context.Context, notification NotificationRequest) error
}

/* Default implementations for notifiers */

func (p *BaseProvider) SendNotification(ctx context.Context, notification NotificationRequest) error {
	// Default implementation does nothing
	return fmt.Errorf("the provider '%s' does not implement SendNotification", p.GetProvider())
}
