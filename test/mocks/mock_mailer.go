package mocks

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockMailer struct {
	session service.Session
	common.Mailer
	mock.Mock
}

func NewMockMailer(session service.Session) common.Mailer {
	return &MockMailer{session: session}
}

func (mailer *MockMailer) Send(subject, message string) error {
	mailer.session.GetLogger().Debugf("MockMailer: subject=%s, message=%s", subject, message)
	return nil
}
