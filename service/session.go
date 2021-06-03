package service

import (
	"github.com/jeremyhahn/cropdroid/common"
	logging "github.com/op/go-logging"
)

type Session interface {
	GetLogger() *logging.Logger
	SetLogger(*logging.Logger)
	GetFarmService() FarmService
	SetFarmService(FarmService)
	GetUser() common.UserAccount
	SetUser(user common.UserAccount)
	Close()
}

type DefaultSession struct {
	logger      *logging.Logger
	farmService FarmService
	user        common.UserAccount
	Session
}

func CreateSession(logger *logging.Logger, farmService FarmService, user common.UserAccount) Session {
	return &DefaultSession{
		logger:      logger,
		farmService: farmService,
		user:        user}
}

func CreateSystemSession(logger *logging.Logger, farmService FarmService) Session {
	return &DefaultSession{
		logger:      logger,
		farmService: farmService}
}

func (session *DefaultSession) GetLogger() *logging.Logger {
	return session.logger
}

func (session *DefaultSession) SetLogger(logger *logging.Logger) {
	session.logger = logger
}

func (session *DefaultSession) GetFarmService() FarmService {
	return session.farmService
}

func (session *DefaultSession) SetFarmService(farmService FarmService) {
	session.farmService = farmService
}

func (session *DefaultSession) GetUser() common.UserAccount {
	return session.user
}

func (session *DefaultSession) SetUser(user common.UserAccount) {
	session.user = user
}

func (session *DefaultSession) Close() {
	if session.logger != nil {
		if session.user != nil {
			session.GetLogger().Debugf("[common.Context] Closing session for %s (uid=%d)", session.user.GetEmail(), session.user.GetID())
		}
	}
}
