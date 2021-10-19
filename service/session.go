package service

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
)

type Session interface {
	GetLogger() *logging.Logger
	SetLogger(*logging.Logger)
	GetFarms() ([]config.Farm, error)
	GetFarmService() FarmService
	SetFarmService(FarmService)
	GetUser() common.UserAccount
	SetUser(user common.UserAccount)
	Close()
}

type DefaultSession struct {
	logger      *logging.Logger
	orgID       uint64
	orgDAO      dao.OrganizationDAO
	farmDAO     dao.FarmDAO
	farmService FarmService
	user        common.UserAccount
	Session
}

func CreateSession(logger *logging.Logger, orgDAO dao.OrganizationDAO,
	farmDAO dao.FarmDAO, farmService FarmService,
	user common.UserAccount) Session {

	return &DefaultSession{
		logger:      logger,
		orgID:       0,
		orgDAO:      orgDAO,
		farmDAO:     farmDAO,
		farmService: farmService,
		user:        user}
}

func CreateSystemSession(logger *logging.Logger, farmService FarmService) Session {
	return &DefaultSession{
		logger:      logger,
		orgID:       0,
		farmService: farmService}
}

func (session *DefaultSession) GetLogger() *logging.Logger {
	return session.logger
}

func (session *DefaultSession) SetLogger(logger *logging.Logger) {
	session.logger = logger
}

func (session *DefaultSession) GetFarms() ([]config.Farm, error) {

	return session.farmDAO.GetByUserID(session.GetUser().GetID())
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

func (session *DefaultSession) String() string {
	return fmt.Sprintf("user=%s, farmID=%d", session.user, session.farmService.GetFarmID())
}
