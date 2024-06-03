package service

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/model"
	logging "github.com/op/go-logging"
)

type Session interface {
	GetLogger() *logging.Logger
	SetLogger(*logging.Logger)
	GetFarmMembership() []uint64
	GetOrganizationMembership() []uint64
	GetRequestedOrganizationID() uint64
	GetRequestedFarmID() uint64
	GetConsistencyLevel() int
	HasRole(string) bool
	IsMemberOfOrganization(orgID uint64) bool
	IsMemberOfFarm(farmID uint64) bool
	GetFarmService() FarmServicer
	SetFarmService(FarmServicer)
	GetUser() model.User
	SetUser(user model.User)
	Close()
}

type DefaultSession struct {
	logger           *logging.Logger
	requestedOrgID   uint64
	requestedFarmID  uint64
	consistencyLevel int
	orgClaims        []OrganizationClaim
	FarmClaims       []FarmClaim
	farmService      FarmServicer
	user             model.User
	Session
}

func CreateSession(
	logger *logging.Logger,
	orgClaims []OrganizationClaim,
	FarmClaims []FarmClaim,
	farmService FarmServicer,
	requestedOrgID, requestedFarmID uint64,
	consistencyLevel int,
	user model.User) Session {

	return &DefaultSession{
		logger:           logger,
		requestedOrgID:   requestedOrgID,
		requestedFarmID:  requestedFarmID,
		consistencyLevel: consistencyLevel,
		orgClaims:        orgClaims,
		FarmClaims:       FarmClaims,
		farmService:      farmService,
		user:             user}
}

func CreateSystemSession(
	logger *logging.Logger,
	farmService FarmServicer) Session {

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

func (session *DefaultSession) GetConsistencyLevel() int {
	return session.consistencyLevel
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
	for _, FarmClaim := range session.FarmClaims {
		if FarmClaim.ID == farmID {
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
	ids := make([]uint64, len(session.FarmClaims))
	for i, FarmClaim := range session.FarmClaims {
		ids[i] = FarmClaim.ID
	}
	return ids
}

func (session *DefaultSession) GetFarmService() FarmServicer {
	return session.farmService
}

func (session *DefaultSession) SetFarmService(farmService FarmServicer) {
	session.farmService = farmService
}

func (session *DefaultSession) GetUser() model.User {
	return session.user
}

func (session *DefaultSession) SetUser(user model.User) {
	session.user = user
}

func (session *DefaultSession) Close() {
	if session.logger != nil {
		if session.user != nil {
			session.GetLogger().Debugf("[common.Context] Closing session for %s (uid=%d)",
				session.user.GetEmail(), session.user.Identifier())
		}
	}
}

func (session *DefaultSession) String() string {
	return fmt.Sprintf("user=%s, farmID=%d", session.user, session.farmService.GetFarmID())
}

func (session *DefaultSession) Error(err error) {
	session.logger.Error("session: %+v, error: %s", session, err)
}
