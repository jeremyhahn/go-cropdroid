package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockNotificationService struct {
	session service.Session
	mailer  common.Mailer
	service.NotificationService
	mock.Mock
}

func NewMockNotificationService(session service.Session, mailer common.Mailer) *MockNotificationService {
	return &MockNotificationService{session: session, mailer: mailer}
}

func (ns *MockNotificationService) Enqueue(notification model.Notification) error {
	ns.Called(notification)
	ns.session.GetLogger().Debugf("MockNotificationService: notification=%s", notification)
	return nil
}
