package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/jeremyhahn/go-cropdroid/config/dao"
)

// "Global" farm factory used to manage all farms on the platform
type FarmFactory interface {
	BuildService(farmStateStore state.FarmStorer,
		farmDAO dao.FarmDAO,
		deviceDataStore datastore.DeviceDataStore,
		deviceStateStore state.DeviceStorer,
		farmConfig *config.Farm) (FarmService, error)
	GetFarms(session Session) ([]*config.Farm, error)
	GetFarmProvisionerChan() chan config.Farm
	GetDeviceIndexMap() map[uint64]config.Device
	GetChannelIndexMap() map[int]config.Channel
}

type DefaultFarmFactory struct {
	app              *app.App
	farmDAO          dao.FarmDAO
	deviceDAO        dao.DeviceDAO
	deviceSettingDAO dao.DeviceSettingDAO
	deviceMapper     mapper.DeviceMapper
	changefeeders    map[string]datastore.Changefeeder
	// deviceIndexMap            map[uint64]config.Device
	// channelIndexMap           map[int]config.Channel
	//datastoreRegistry         datastore.DatastoreRegistry
	serviceRegistry           ServiceRegistry
	farmProvisionerChan       chan config.Farm
	farmTickerProvisionerChan chan uint64
	farmChannels              *FarmChannels
	FarmFactory
}

func NewFarmFactory(app *app.App, farmDAO dao.FarmDAO, deviceDAO dao.DeviceDAO,
	deviceSettingDAO dao.DeviceSettingDAO, serviceRegistry ServiceRegistry,
	deviceMapper mapper.DeviceMapper, changefeeders map[string]datastore.Changefeeder,
	farmProvisionerChan chan config.Farm, farmTickerProvisionerChan chan uint64,
	farmChannels *FarmChannels, idGenerator util.IdGenerator) FarmFactory {

	return &DefaultFarmFactory{
		app:              app,
		farmDAO:          farmDAO,
		deviceDAO:        deviceDAO,
		deviceSettingDAO: deviceSettingDAO,
		deviceMapper:     deviceMapper,
		changefeeders:    changefeeders,
		// deviceIndexMap:            make(map[uint64]config.Device, 0),
		// channelIndexMap:           make(map[int]config.Channel, 0),
		//datastoreRegistry:         datastoreRegistry,
		serviceRegistry:           serviceRegistry,
		farmProvisionerChan:       farmProvisionerChan,
		farmTickerProvisionerChan: farmTickerProvisionerChan,
		farmChannels:              farmChannels}
}

func (ff *DefaultFarmFactory) BuildService(farmStateStore state.FarmStorer,
	farmDAO dao.FarmDAO,
	deviceDataStore datastore.DeviceDataStore,
	deviceStateStore state.DeviceStorer,
	farmConfig *config.Farm) (FarmService, error) {

	consistencyLevel := farmConfig.GetConsistencyLevel()

	farmID := farmConfig.GetID()
	farmName := farmConfig.GetName()
	deviceConfigs := farmConfig.GetDevices()

	// Build device services
	deviceFactory := NewDeviceFactory(ff.app, farmID, farmName,
		ff.deviceDAO, farmConfig.GetConfigStore(), consistencyLevel,
		deviceStateStore, ff.deviceMapper, ff.serviceRegistry, ff.farmChannels)

	deviceServices, err := deviceFactory.BuildServices(deviceConfigs,
		deviceDataStore, farmConfig.GetMode())
	if err != nil {
		ff.app.Logger.Fatal(err)
	}

	// Build farm service
	farmService, err := CreateFarmService(ff.app, farmDAO, ff.app.IdGenerator,
		farmStateStore, deviceDataStore, farmConfig, consistencyLevel,
		ff.serviceRegistry, ff.farmChannels, ff.deviceSettingDAO)
	if err != nil {
		return nil, err
	}

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
	ff.serviceRegistry.SetDeviceServices(farmID, deviceServices)
	ff.serviceRegistry.AddFarmService(farmService)

	ff.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
	ff.app.Logger.Debugf("Mode: %s", farmConfig.GetMode())
	ff.app.Logger.Debugf("Timezone: %s", farmConfig.GetTimezone())
	ff.app.Logger.Debugf("Polling interval: %d", farmConfig.GetInterval())

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
func (ff *DefaultFarmFactory) GetFarms(session Session) ([]*config.Farm, error) {
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
	// The config store type is currently stored at the farm level so this is a
	// chicken and egg problem, since the store type has to be known before
	// the lookup can be made.
	return ff.farmDAO.GetByIds(farmIds, common.CONSISTENCY_LOCAL)
}
