package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
)

// "Global" farm factory used to manage all farms on the platform
type FarmFactory interface {
	BuildService(farmStateStore state.FarmStateStorer,
		farmDAO dao.FarmDAO,
		eventLogDAO dao.EventLogDAO,
		deviceDataStore datastore.DeviceDataStore,
		deviceStateStore state.DeviceStateStorer,
		farmConfig config.Farm,
		farmChannels *FarmChannels) (FarmServicer, error)
	GetFarms(session Session) ([]config.Farm, error)
	GetFarmProvisionerChan() chan config.Farm
	GetDeviceIndexMap() map[uint64]config.Device
	GetChannelIndexMap() map[int]config.Channel
}

type DefaultFarmFactory struct {
	app               *app.App
	farmDAO           dao.FarmDAO
	datastoreRegistry dao.Registry
	deviceSettingDAO  dao.DeviceSettingDAO
	deviceMapper      mapper.DeviceMapper
	// deviceIndexMap            map[uint64]config.Device
	// channelIndexMap           map[int]config.Channel
	//datastoreRegistry         datastore.DatastoreRegistry
	serviceRegistry           ServiceRegistry
	farmProvisionerChan       chan config.Farm
	farmTickerProvisionerChan chan uint64
	FarmFactory
}

func NewFarmFactory(app *app.App, farmDAO dao.FarmDAO, datastoreRegistry dao.Registry,
	deviceSettingDAO dao.DeviceSettingDAO, serviceRegistry ServiceRegistry,
	deviceMapper mapper.DeviceMapper, farmProvisionerChan chan config.Farm,
	farmTickerProvisionerChan chan uint64,
	idGenerator util.IdGenerator) FarmFactory {

	return &DefaultFarmFactory{
		app:               app,
		farmDAO:           farmDAO,
		datastoreRegistry: datastoreRegistry,
		deviceSettingDAO:  deviceSettingDAO,
		deviceMapper:      deviceMapper,
		// deviceIndexMap:            make(map[uint64]config.Device, 0),
		// channelIndexMap:           make(map[int]config.Channel, 0),
		//datastoreRegistry:         datastoreRegistry,
		serviceRegistry:           serviceRegistry,
		farmProvisionerChan:       farmProvisionerChan,
		farmTickerProvisionerChan: farmTickerProvisionerChan}
}

func (ff *DefaultFarmFactory) BuildService(
	farmStateStore state.FarmStateStorer,
	farmDAO dao.FarmDAO,
	eventLogDAO dao.EventLogDAO,
	deviceDataStore datastore.DeviceDataStore,
	deviceStateStore state.DeviceStateStorer,
	farmConfig config.Farm,
	farmChannels *FarmChannels) (FarmServicer, error) {

	consistencyLevel := farmConfig.GetConsistencyLevel()
	farmName := farmConfig.GetName()
	deviceConfigs := farmConfig.GetDevices()

	// Build device services
	deviceFactory := NewDeviceFactory(ff.app, farmConfig.Identifier(), farmName,
		ff.datastoreRegistry, eventLogDAO, farmConfig.GetConfigStore(), consistencyLevel,
		deviceStateStore, ff.deviceMapper, ff.serviceRegistry, farmChannels)

	deviceServices, err := deviceFactory.BuildServices(deviceConfigs,
		deviceDataStore, farmConfig.GetMode())
	if err != nil {
		ff.app.Logger.Fatal(err)
	}

	// Build farm service
	farmService, err := CreateFarmService(ff.app, farmDAO, ff.app.IdGenerator,
		farmStateStore, deviceStateStore, deviceDataStore, farmConfig, consistencyLevel,
		ff.serviceRegistry, farmChannels, ff.deviceSettingDAO, ff.deviceMapper)
	if err != nil {
		return nil, err
	}

	// Build event log service
	eventLogService := NewEventLogService(ff.app, eventLogDAO, farmConfig.Identifier())

	// deviceIndexMap := make(map[uint64]config.Device, 0)
	// channelIndexMap := make(map[int]config.Channel, 0)

	// // Build farm device and channel indexes
	// for i, device := range deviceConfigs {
	// 	//if device.GetType() == common.CONTROLLER_TYPE_SERVER {
	// 	//	continue
	// 	//}
	// 	deviceIndexMap[device.GetID()] = &deviceConfigs[i]
	// 	if ff.app.DatastoreCDC && ff.app.DatastoreType == "cockroach" {
	// 		// Device metric/channel state
	// 		if _, ok := changefeeders[device.GetType()]; !ok {
	// 			tableName := fmt.Sprintf("state_%d", device.GetID())
	// 			changefeeders[device.GetType()] = cockroach.NewCockroachChangefeed(ff.app, tableName)
	// 		}
	// 	}
	// 	channels := device.GetChannels()
	// 	for _, channel := range channels {
	// 		channelIndexMap[channel.GetID()] = &channels[i]
	// 	}
	// }

	ff.serviceRegistry.SetDeviceFactory(deviceFactory)
	ff.serviceRegistry.SetDeviceServices(farmConfig.Identifier(), deviceServices)
	ff.serviceRegistry.AddFarmService(farmService)
	ff.serviceRegistry.AddEventLogService(eventLogService)

	return farmService, nil
}

func (ff *DefaultFarmFactory) GetFarmProvisionerChan() chan config.Farm {
	return ff.farmProvisionerChan
}

// func (ff *DefaultFarmFactory) GetDeviceIndexMap() map[uint64]config.Device {
// 	return ff.deviceIndexMap
// }

// func (ff *DefaultFarmFactory) GetChannelIndexMap() map[int]config.Channel {
// 	return ff.channelIndexMap
// }

// Returns all of the farms the user has access to within the current session
func (ff *DefaultFarmFactory) GetFarms(session Session) ([]config.Farm, error) {
	farmIds := session.GetFarmMembership()
	// If requestedFarmID and requestedOrganizationID are 0
	// then this is probably a new user who hasnt been given
	// permissions to any orgs or farms yet, so there is no
	// way to get a farmService to perform a lookup for the
	// consistency level
	//
	// var consistencyLevel = common.CONSISTENCY_LOCAL
	// farmService := session.GetFarmService()
	// if farmService != nil {
	// 	consistencyLevel = farmService.GetConsistencyLevel()
	// }

	// Update: 01/02/2022
	// The farm config store type is currently stored in farm object so this is a
	// chicken and egg problem, since the store type has to be known before
	// the lookup can be made. Hard coding consistency level for now.
	farmStructs, err := ff.farmDAO.GetByIds(farmIds, common.CONSISTENCY_LOCAL)
	if err != nil {
		ff.app.Logger.Error(err)
		return nil, err
	}
	farmConfigs := make([]config.Farm, len(farmStructs))
	for i, farm := range farmStructs {
		farmConfigs[i] = farm
	}
	return farmConfigs, err
}
