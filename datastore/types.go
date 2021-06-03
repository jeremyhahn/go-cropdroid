package datastore

import (
	"encoding/json"

	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jeremyhahn/cropdroid/state"
)

type ChangefeedCallback func(Changefeed)

type Changefeeder interface {
	Subscribe(callback ChangefeedCallback)
}

type Changefeed interface {
	GetTable() string
	GetKey() int64
	GetValue() interface{}
	GetUpdated() string
	GetBytes() []byte
	GetRawMessage() map[string]*json.RawMessage
}

type ControllerStateDAO interface {
	//CreateTable(tableName string, controllerState state.ControllerStateMap) error
	Save(controllerID int, controllerState state.ControllerStateMap) error
	GetLast30Days(controllerID int, metric string) ([]float64, error)
}

type DatastoreRegistry interface {
	GetOrganizationDAO() dao.OrganizationDAO
	SetOrganizationDAO(dao dao.OrganizationDAO)
	GetFarmDAO() dao.FarmDAO
	SetFarmDAO(dao dao.FarmDAO)
	GetControllerDAO() dao.ControllerDAO
	SetControllerDAO(dao dao.ControllerDAO)
	GetControllerConfigDAO() dao.ControllerConfigDAO
	SetControllerConfigDAO(dao dao.ControllerConfigDAO)
	GetMetricDAO() dao.MetricDAO
	SetMetricDAO(dao dao.MetricDAO)
	GetChannelDAO() dao.ChannelDAO
	SetChannelDAO(dao dao.ChannelDAO)
	GetScheduleDAO() dao.ScheduleDAO
	SetScheduleDAO(dao dao.ScheduleDAO)
	GetConditionDAO() dao.ConditionDAO
	SetConditionDAO(dao dao.ConditionDAO)
	GetAlgorithmDAO() dao.AlgorithmDAO
	SetAlgorithmDAO(dao dao.AlgorithmDAO)
	GetUserDAO() dao.UserDAO
	SetUserDAO(dao.UserDAO)
	GetRoleDAO() dao.RoleDAO
	SetRoleDAO(dao.RoleDAO)
}
