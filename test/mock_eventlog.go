// +build broken

package test

import (
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/datastore/gorm"
	"github.com/jeremyhahn/cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockEventLogService struct {
	scope      common.Scope
	dao        gorm.EventLogDAO
	controller string
	service.EventLogService
	mock.Mock
}

func NewMockEventLogService(scope common.Scope, dao gorm.EventLogDAO, controller string) service.EventLogService {
	return &MockEventLogService{
		scope:      scope,
		dao:        dao,
		controller: controller}
}

func (eventLog *MockEventLogService) Create(event, message string) {
	eventLog.scope.GetLogger().Debugf("[Create] event=%s, message=%s", event, message)
}

func (eventLog *MockEventLogService) GetAll() []entity.EventLog {
	eventLog.scope.GetLogger().Debugf("[GetAll]")
	return nil
}
