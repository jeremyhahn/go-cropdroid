package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
)

// "Global" farm factory used to manage all farms on the platform
type FarmFactory interface {
	BuildService(farmConfig config.FarmConfig) (FarmService, error)
	BuildClusterService(farmConfig config.FarmConfig) (FarmService, error)
	GetFarms(session Session) ([]config.FarmConfig, error)
	GetFarmProvisionerChan() chan config.FarmConfig
	GetDeviceIndexMap() map[uint64]config.DeviceConfig
	GetChannelIndexMap() map[int]config.ChannelConfig
}

type DefaultFarmFactory struct {
	app               *app.App
	farmStateStore    state.FarmStorer
	farmConfigStore   store.FarmConfigStorer
	deviceConfigStore store.DeviceConfigStorer
	deviceStateStore  state.DeviceStorer
	deviceDataStore   datastore.DeviceDataStore
	consistencyLevel  int
	deviceMapper      mapper.DeviceMapper
	changefeeders     map[string]datastore.Changefeeder
	// deviceIndexMap            map[uint64]config.DeviceConfig
	// channelIndexMap           map[int]config.ChannelConfig
	datastoreRegistry         datastore.DatastoreRegistry
	serviceRegistry           ServiceRegistry
	farmProvisionerChan       chan config.FarmConfig
	farmTickerProvisionerChan chan uint64
	idGenerator               util.IdGenerator
	FarmFactory
}

func NewFarmFactory(app *app.App, datastoreRegistry datastore.DatastoreRegistry,
	serviceRegistry ServiceRegistry, farmStateStore state.FarmStorer,
	farmConfigStore store.FarmConfigStorer, deviceStateStore state.DeviceStorer,
	deviceConfigStore store.DeviceConfigStorer, deviceDataStore datastore.DeviceDataStore,
	deviceMapper mapper.DeviceMapper, changefeeders map[string]datastore.Changefeeder,
	farmProvisionerChan chan config.FarmConfig, farmTickerProvisionerChan chan uint64,
	idGenerator util.IdGenerator) FarmFactory {

	return &DefaultFarmFactory{
		app:               app,
		farmStateStore:    farmStateStore,
		farmConfigStore:   farmConfigStore,
		deviceStateStore:  deviceStateStore,
		deviceConfigStore: deviceConfigStore,
		deviceDataStore:   deviceDataStore,
		deviceMapper:      deviceMapper,
		changefeeders:     changefeeders,
		// deviceIndexMap:            make(map[uint64]config.DeviceConfig, 0),
		// channelIndexMap:           make(map[int]config.ChannelConfig, 0),
		datastoreRegistry:         datastoreRegistry,
		serviceRegistry:           serviceRegistry,
		farmProvisionerChan:       farmProvisionerChan,
		farmTickerProvisionerChan: farmTickerProvisionerChan,
		idGenerator:               idGenerator}
}

func (ff *DefaultFarmFactory) BuildService(farmConfig config.FarmConfig) (FarmService, error) {

	farmID := farmConfig.GetID()
	farmName := farmConfig.GetName()
	deviceConfigs := farmConfig.GetDevices()

	farmDAO := ff.datastoreRegistry.GetFarmDAO()
	deviceConfigDAO := ff.datastoreRegistry.GetDeviceConfigDAO()

	farmChannels := FarmChannels{
		FarmConfigChan:       make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE),
		FarmConfigChangeChan: make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChan:        make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChangeChan:  make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmErrorChan:        make(chan common.FarmError, common.BUFFERED_CHANNEL_SIZE),
		FarmNotifyChan:       make(chan common.FarmNotification, common.BUFFERED_CHANNEL_SIZE),
		//MetricChangedChan:     make(chan common.MetricValueChanged, common.BUFFERED_CHANNEL_SIZE),
		//SwitchChangedChan:     make(chan common.SwitchValueChanged, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateChangeChan: make(chan common.DeviceStateChange, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateDeltaChan:  make(chan map[string]state.DeviceStateDeltaMap, common.BUFFERED_CHANNEL_SIZE)}

	// Build device services
	deviceFactory := NewDeviceFactory(ff.app, farmID, farmName,
		ff.deviceStateStore, ff.deviceConfigStore, ff.consistencyLevel,
		ff.deviceMapper, ff.serviceRegistry, &farmChannels)

	deviceServices, err := deviceFactory.BuildServices(deviceConfigs,
		ff.deviceDataStore, farmConfig.GetMode())
	if err != nil {
		ff.app.Logger.Fatal(err)
	}

	// Build farm service
	farmService, err := CreateFarmService(ff.app, farmDAO, ff.idGenerator, ff.farmStateStore,
		ff.farmConfigStore, ff.deviceConfigStore, ff.deviceDataStore, farmConfig, ff.consistencyLevel,
		ff.serviceRegistry, &farmChannels, deviceConfigDAO)
	if err != nil {
		return nil, err
	}

	/*
		deviceIndexMap := make(map[uint64]config.DeviceConfig, 0)
		channelIndexMap := make(map[int]config.ChannelConfig, 0)

			// Build farm device and channel indexes
			for i, device := range deviceConfigs {
				//if device.GetType() == common.CONTROLLER_TYPE_SERVER {
				//	continue
				//}
				deviceIndexMap[device.GetID()] = &deviceConfigs[i]
				if builder.app.DatastoreCDC && builder.app.DatastoreType == "cockroach" {
					// Device metric/channel state
					if _, ok := changefeeders[device.GetType()]; !ok {
						tableName := fmt.Sprintf("state_%d", device.GetID())
						changefeeders[device.GetType()] = cockroach.NewCockroachChangefeed(builder.app, tableName)
					}
				}
				channels := device.GetChannels()
				for _, channel := range channels {
					channelIndexMap[channel.GetID()] = &channels[i]
				}
			}*/

	ff.serviceRegistry.SetDeviceFactory(deviceFactory)
	ff.serviceRegistry.SetDeviceServices(farmID, deviceServices)
	ff.serviceRegistry.AddFarmService(farmService)

	ff.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
	ff.app.Logger.Debugf("Mode: %s", farmConfig.GetMode())
	ff.app.Logger.Debugf("Timezone: %s", farmConfig.GetTimezone())
	ff.app.Logger.Debugf("Polling interval: %d", farmConfig.GetInterval())

	return farmService, nil
}

func (ff *DefaultFarmFactory) GetFarmProvisionerChan() chan config.FarmConfig {
	return ff.farmProvisionerChan
}

// func (ff *DefaultFarmFactory) GetDeviceIndexMap() map[uint64]config.DeviceConfig {
// 	return ff.deviceIndexMap
// }

// func (ff *DefaultFarmFactory) GetChannelIndexMap() map[int]config.ChannelConfig {
// 	return ff.channelIndexMap
// }

// Returns all of the farms the user has access to within the current session
func (ff *DefaultFarmFactory) GetFarms(session Session) ([]config.FarmConfig, error) {
	farmIds := session.GetFarmMembership()
	// If requestedFarmID and requestedOrganizationID are 0
	// then this is probably a new user who hasnt been given
	// permissions to any orgs or farms yet, so there is no
	// way to get a farmService to perform a lookup for the
	// consistency level
	var consistencyLevel = common.CONSISTENCY_LOCAL
	farmService := session.GetFarmService()
	if farmService != nil {
		consistencyLevel = farmService.GetConsistencyLevel()
	}
	return ff.farmConfigStore.GetByIds(farmIds, consistencyLevel), nil
}
