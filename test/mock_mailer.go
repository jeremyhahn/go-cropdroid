// +build broken

package test

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/stretchr/testify/mock"
)

type MockMailer struct {
	scope common.Scope
	common.Mailer
	mock.Mock
}

func NewMockMailer(scope common.Scope) common.Mailer {
	return &MockMailer{scope: scope}
}

func (mailer *MockMailer) Send(farmName, subject, message string) error {
	mailer.scope.GetLogger().Debugf("MockMailer: subject=%s, message=%s", subject, message)
	return nil
}
