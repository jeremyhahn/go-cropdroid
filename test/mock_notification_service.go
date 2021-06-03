// +build broken

package test

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockNotificationService struct {
	scope  common.Scope
	mailer common.Mailer
	service.NotificationService
	mock.Mock
}

func NewMockNotificationService(scope common.Scope, mailer common.Mailer) *MockNotificationService {
	return &MockNotificationService{scope: scope, mailer: mailer}
}

func (ns *MockNotificationService) Enqueue(notification common.Notification) error {
	ns.Called(notification)
	ns.scope.GetLogger().Debugf("MockNotificationService: notification=%s", notification)
	return nil
}
