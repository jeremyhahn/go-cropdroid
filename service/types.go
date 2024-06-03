package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/state"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
)

var (
	//ErrFarmNotFound             = errors.New("farm config not found")
	ErrChannelNotFound          = errors.New("channel not found")
	ErrMetricNotFound           = errors.New("metric not found")
	ErrScheduleNotFound         = errors.New("channel schedule not found")
	ErrConditionNotFound        = errors.New("channel condition not found")
	ErrNoDeviceState            = errors.New("no device state")
	ErrCreateService            = errors.New("failed to create service")
	ErrDeviceNotFound           = errors.New("device not found")
	ErrWorkflowNotFound         = errors.New("workflow not found")
	ErrWorkflowStepNotFound     = errors.New("workflow step not found")
	ErrPermissionDenied         = errors.New("permission denied")
	ErrDeleteAdminAccount       = errors.New("admin account can't be deleted")
	ErrChangeAdminRole          = errors.New("admin role can't be changed")
	ErrResetPasswordUnsupported = errors.New("reset password feature unsupported by auth store")
)

type AlgorithmHandler interface {
	Handle() (bool, error)
}

type ScheduleHandler interface {
	Handle() error
}

type ConditionHandler interface {
	Handle() (bool, error)
}

type FarmChannels struct {
	FarmConfigChan        chan config.Farm
	FarmConfigChangeChan  chan config.Farm
	FarmStateChangeChan   chan state.FarmStateMap
	FarmErrorChan         chan common.FarmError
	FarmNotifyChan        chan common.FarmNotification
	DeviceStateChangeChan chan common.DeviceStateChange
	DeviceStateDeltaChan  chan map[string]state.DeviceStateDeltaMap
}

type UserCredentials struct {
	OrgID    uint64 `json:"orgId"`
	OrgName  string `json:"orgName"`
	Email    string `json:"email"`
	Password string `json:"password"`
	AuthType int    `json:"authType"`
}

// Claim structs are condensed models concerned only
// with users, roles, permissions, and licensing between
// the client and server. They get exchanged with every
// request and are used to generate a "Session" for working
// with business logic services in the "service" package.
type FarmClaim struct {
	ID    uint64   `json:"id"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

type OrganizationClaim struct {
	ID      uint64                           `json:"id"`
	Name    string                           `json:"name"`
	Farms   []FarmClaim                      `json:"farms"`
	Roles   []string                         `json:"roles"`
	License config.OrganizationLicenseStruct `json:"license"`
}

type AuthServicer interface {
	Activate(registrationID uint64) (model.User, error)
	Login(userCredentials *UserCredentials) (model.User, []config.Organization, []config.Farm, error)
	Register(userCredentials *UserCredentials, baseURI string) (model.User, error)
	ResetPassword(userCredentials *UserCredentials) error
}
