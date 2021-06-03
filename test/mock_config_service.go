package test

import (
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockConfigService struct {
	service.ConfigService
	mock.Mock
}

func NewMockConfigService() service.ConfigService {
	return &MockConfigService{}
}
