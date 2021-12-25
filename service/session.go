package service

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	logging "github.com/op/go-logging"
)

type Session interface {
	GetLogger() *logging.Logger
	SetLogger(*logging.Logger)
	GetFarmMembership() []uint64
	GetOrganizationMembership() []uint64
	GetRequestedOrganizationID() uint64
	GetRequestedFarmID() uint64
	HasRole(string) bool
	IsMemberOfOrganization(orgID uint64) bool
	IsMemberOfFarm(farmID uint64) bool
	GetFarmService() FarmService
	SetFarmService(FarmService)
	GetUser() common.UserAccount
	SetUser(user common.UserAccount)
	Close()
}

type DefaultSession struct {
	logger          *logging.Logger
	requestedOrgID  uint64
	requestedFarmID uint64
	orgClaims       []organizationClaim
	farmClaims      []farmClaim
	farmService     FarmService
	user            common.UserAccount
	Session
}

func CreateSession(logger *logging.Logger, orgClaims []organizationClaim,
	farmClaims []farmClaim, farmService FarmService, requestedOrgID,
	requestedFarmID uint64, user common.UserAccount) Session {

	return &DefaultSession{
		logger:          logger,
		requestedOrgID:  requestedOrgID,
		requestedFarmID: requestedFarmID,
		orgClaims:       orgClaims,
		farmClaims:      farmClaims,
		farmService:     farmService,
		user:            user}
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

func (session *DefaultSession) GetRequestedOrganizationID() uint64 {
	return session.requestedOrgID
}

func (session *DefaultSession) GetRequestedFarmID() uint64 {
	return session.requestedFarmID
}

func (session *DefaultSession) HasRole(role string) bool {
	return session.user.HasRole(role)
}

func (session *DefaultSession) IsMemberOfOrganization(organizationID uint64) bool {
	for _, orgClaim := range session.orgClaims {
		if orgClaim.ID == organizationID {
			return true
		}
	}
	return false
}

func (session *DefaultSession) IsMemberOfFarm(farmID uint64) bool {
	for _, farmClaim := range session.farmClaims {
		if farmClaim.ID == farmID {
			return true
		}
	}
	return false
}

func (session *DefaultSession) GetOrganizationMembership() []uint64 {
	ids := make([]uint64, len(session.orgClaims))
	for i, orgClaim := range session.orgClaims {
		ids[i] = orgClaim.ID
	}
	return ids
}

func (session *DefaultSession) GetFarmMembership() []uint64 {
	ids := make([]uint64, len(session.farmClaims))
	for i, farmClaim := range session.farmClaims {
		ids[i] = farmClaim.ID
	}
	return ids
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
