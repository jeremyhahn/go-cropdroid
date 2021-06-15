package mocks

import (
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockConfigService struct {
	service.ConfigService
	mock.Mock
}

func NewMockConfigService() service.ConfigService {
	return &MockConfigService{}
}
