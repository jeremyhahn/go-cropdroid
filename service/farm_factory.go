package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type FarmFactory struct {
	app                       *app.App
	stateStore                state.FarmStorer
	configStore               store.FarmConfigStorer
	deviceConfigStore         store.DeviceConfigStorer
	deviceStateStore          state.DeviceStorer
	deviceDataStore           datastore.DeviceDatastore
	consistencyLevel          int
	deviceMapper          mapper.DeviceMapper
	changefeeders             map[string]datastore.Changefeeder
	deviceIndexMap        map[uint64]config.DeviceConfig
	channelIndexMap           map[int]config.ChannelConfig
	datastoreRegistry         datastore.DatastoreRegistry
	serviceRegistry           ServiceRegistry
	farmProvisionerChan       chan config.FarmConfig
	farmTickerProvisionerChan chan uint64
}

func NewFarmFactory(app *app.App, datastoreRegistry datastore.DatastoreRegistry,
	serviceRegistry ServiceRegistry, farmStateStore state.FarmStorer,
	farmConfigStore store.FarmConfigStorer, deviceStateStore state.DeviceStorer,
	deviceConfigStore store.DeviceConfigStorer, deviceDataStore datastore.DeviceDatastore,
	deviceMapper mapper.DeviceMapper, changefeeders map[string]datastore.Changefeeder,
	farmProvisionerChan chan config.FarmConfig, farmTickerProvisionerChan chan uint64) *FarmFactory {

	return &FarmFactory{
		app:                       app,
		stateStore:                farmStateStore,
		configStore:               farmConfigStore,
		deviceStateStore:          deviceStateStore,
		deviceConfigStore:         deviceConfigStore,
		deviceDataStore:           deviceDataStore,
		deviceMapper:          deviceMapper,
		changefeeders:             changefeeders,
		deviceIndexMap:        make(map[uint64]config.DeviceConfig, 0),
		channelIndexMap:           make(map[int]config.ChannelConfig, 0),
		datastoreRegistry:         datastoreRegistry,
		serviceRegistry:           serviceRegistry,
		farmProvisionerChan:       farmProvisionerChan,
		farmTickerProvisionerChan: farmTickerProvisionerChan}
}

func (fb *FarmFactory) RunProvisionerConsumer() {
	for {
		select {
		case farmConfig := <-fb.farmProvisionerChan:
			fb.app.Logger.Debugf("Processing provisioner request...")
			//farmConfigChangeChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
			//farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
			//farmService, err := fb.BuildService(farmConfig, farmConfigChangeChan, farmStateChangeChan)
			farmService, err := fb.BuildService(farmConfig)
			if err != nil {
				fb.app.Logger.Errorf("Error: %s", err)
			}
			fb.serviceRegistry.AddFarmService(farmService)
			fb.farmTickerProvisionerChan <- farmConfig.GetID()
		default:
			fb.app.Logger.Error("Error: Unable to process provisioner request")
		}
	}
}

func (fb *FarmFactory) BuildService(farmConfig config.FarmConfig) (FarmService, error) {

	farmID := farmConfig.GetID()
	farmName := farmConfig.GetName()
	deviceConfigs := farmConfig.GetDevices()

	farmDAO := fb.datastoreRegistry.GetFarmDAO()
	deviceDAO := fb.datastoreRegistry.GetDeviceDAO()
	deviceConfigDAO := fb.datastoreRegistry.GetDeviceConfigDAO()

	farmChannels := FarmChannels{
		FarmConfigChan:        make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE),
		FarmConfigChangeChan:  make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChan:         make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChangeChan:   make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmErrorChan:         make(chan common.FarmError, common.BUFFERED_CHANNEL_SIZE),
		FarmNotifyChan:        make(chan common.FarmNotification, common.BUFFERED_CHANNEL_SIZE),
		MetricChangedChan:     make(chan common.MetricValueChanged, common.BUFFERED_CHANNEL_SIZE),
		SwitchChangedChan:     make(chan common.SwitchValueChanged, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateChangeChan: make(chan common.DeviceStateChange, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateDeltaChan:  make(chan map[string]state.DeviceStateDeltaMap, common.BUFFERED_CHANNEL_SIZE)}

	// Build device services
	deviceFactory := NewDeviceFactory(fb.app, farmID, farmName, deviceDAO,
		fb.deviceStateStore, fb.deviceConfigStore, fb.consistencyLevel,
		fb.deviceMapper, fb.serviceRegistry, &farmChannels)

	deviceServices, err := deviceFactory.BuildServices(deviceConfigs,
		fb.deviceDataStore, farmConfig.GetMode())
	if err != nil {
		fb.app.Logger.Fatal(err)
	}

	// Build farm service
	farmService, err := CreateFarmService(fb.app, farmDAO, fb.stateStore,
		fb.configStore, fb.deviceConfigStore, fb.deviceDataStore, farmConfig, fb.consistencyLevel,
		fb.serviceRegistry, &farmChannels, deviceConfigDAO)
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

	fb.serviceRegistry.SetDeviceFactory(deviceFactory)
	fb.serviceRegistry.SetDeviceServices(farmID, deviceServices)
	fb.serviceRegistry.AddFarmService(farmService)

	fb.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
	fb.app.Logger.Debugf("Mode: %s", farmConfig.GetMode())
	fb.app.Logger.Debugf("Timezone: %s", farmConfig.GetTimezone())
	fb.app.Logger.Debugf("Polling interval: %d", farmConfig.GetInterval())

	return farmService, nil
}

func (fb *FarmFactory) GetFarmProvisionerChan() chan config.FarmConfig {
	return fb.farmProvisionerChan
}

func (fb *FarmFactory) GetDeviceIndexMap() map[uint64]config.DeviceConfig {
	return fb.deviceIndexMap
}

func (fb *FarmFactory) GetChannelIndexMap() map[int]config.ChannelConfig {
	return fb.channelIndexMap
}
