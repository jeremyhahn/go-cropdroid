package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
)

type DefaultFarmService struct {
	app                 *app.App
	dao                 dao.FarmDAO
	stateStore          state.FarmStorer
	configStore         store.FarmConfigStorer
	deviceDataStore     datastore.DeviceDataStore
	farmID              uint64
	configClusterID     uint64
	mode                string
	consistencyLevel    int
	serviceRegistry     ServiceRegistry
	conditionService    ConditionService
	scheduleService     ScheduleService
	notificationService NotificationService
	channels            *FarmChannels
	running             bool
	backoffTable        map[uint64]map[uint64]time.Time
	deviceConfigDAO     dao.DeviceConfigDAO
	farmStateQuitChan   chan int
	farmConfigQuitChan  chan int
	deviceStateQuitChan chan int
	pollTickerQuitChan  chan int
	FarmService
	//observer.FarmConfigObserver
}

func CreateFarmService(app *app.App, farmDAO dao.FarmDAO, stateStore state.FarmStorer,
	configStore store.FarmConfigStorer, deviceConfigStore store.DeviceConfigStorer,
	deviceDataStore datastore.DeviceDataStore, farmConfig config.FarmConfig, consistencyLevel int,
	serviceRegistry ServiceRegistry, farmChannels *FarmChannels,
	deviceConfigDAO dao.DeviceConfigDAO) (FarmService, error) {

	farmID := farmConfig.GetID()
	mode := farmConfig.GetMode()

	backoffTable := make(map[uint64]map[uint64]time.Time, 0)

	deviceServices, err := serviceRegistry.GetDeviceServices(farmID)
	if err != nil {
		return nil, err
	}

	// Set farmConfig devices with latest configs (populated SystemInfo)
	// TODO: make this more efficient
	for _, ds := range deviceServices {
		deviceConfig, err := ds.GetConfig()
		if err != nil {
			return nil, err
		}
		farmConfig.SetDevice(deviceConfig)
	}

	farmService := &DefaultFarmService{
		app:                 app,
		dao:                 farmDAO,
		stateStore:          stateStore,
		configStore:         configStore,
		deviceDataStore:     deviceDataStore,
		farmID:              farmID,
		mode:                mode,
		consistencyLevel:    consistencyLevel,
		serviceRegistry:     serviceRegistry,
		conditionService:    serviceRegistry.GetConditionService(),
		scheduleService:     serviceRegistry.GetScheduleService(),
		notificationService: serviceRegistry.GetNotificationService(),
		channels:            farmChannels,
		running:             false,
		farmStateQuitChan:   make(chan int),
		farmConfigQuitChan:  make(chan int),
		deviceStateQuitChan: make(chan int),
		pollTickerQuitChan:  make(chan int),
		deviceConfigDAO:     deviceConfigDAO,
		backoffTable:        backoffTable}

	var configClusterID uint64
	if farmService.isRaftConfigStore(farmConfig) {
		configClusterID = util.NewClusterHash(farmConfig.GetOrganizationID(), farmID)
	} else {
		configClusterID = farmConfig.GetID()

		err = farmService.configStore.Put(configClusterID, farmConfig)
		//if err == state.ErrFarmNotFound {
		err = farmService.InitializeState(true)
		if err != nil {
			return nil, err
		}
		// } else if err != nil {
		// 	app.Logger.Errorf("Fatal farm service error (farmID=%d): %s", farmID, err)
		// 	return nil, err
		// }
	}

	farmService.configClusterID = configClusterID

	return farmService, nil
}

func (farm *DefaultFarmService) IsRunning() bool {
	return farm.running
}

func (farm *DefaultFarmService) isRaftConfigStore(farmConfig config.FarmConfig) bool {
	return farmConfig.GetConfigStore() == config.RAFT_MEMORY_STORE ||
		farmConfig.GetConfigStore() == config.RAFT_DISK_STORE
}

func (farm *DefaultFarmService) InitializeState(saveToStateStore bool) error {
	deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.farmID)
	if err != nil {
		return err
	}
	_state := state.NewFarmStateMap(farm.farmID)
	deviceStateMaps := make([]state.DeviceStateMap, 0)
	for _, d := range deviceServices {
		conf, _ := d.GetConfig()
		deviceID := conf.GetID()
		deviceStateMap := state.CreateEmptyDeviceStateMap(
			deviceID, len(conf.GetMetrics()), len(conf.GetChannels()))
		deviceStateMaps = append(deviceStateMaps, deviceStateMap)
		_state.SetDevice(conf.GetType(), deviceStateMap)
		farm.backoffTable[farm.farmID] = make(map[uint64]time.Time, 0)
	}
	if saveToStateStore {
		farm.stateStore.Put(farm.farmID, _state)
	}
	return nil
}

func (farm *DefaultFarmService) GetFarmID() uint64 {
	return farm.farmID
}

func (farm *DefaultFarmService) GetConfigClusterID() uint64 {
	return farm.configClusterID
}

func (farm *DefaultFarmService) GetConsistencyLevel() int {
	return farm.consistencyLevel
}

func (farm *DefaultFarmService) GetChannels() *FarmChannels {
	return farm.channels
}

func (farm *DefaultFarmService) GetPublicKey() string {
	// TODO: Replace w/ key defined in FarmConfig
	return string(farm.app.KeyPair.GetPublicBytes())
}

func (farm *DefaultFarmService) GetState() state.FarmStateMap {
	farmState, err := farm.stateStore.Get(farm.farmID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %+v", err)
		return state.NewFarmStateMap(farm.farmID)
	}
	return farmState
}

func (farm *DefaultFarmService) GetConfig() config.FarmConfig {
	farm.app.Logger.Debugf("Getting farm configuration. farm.id=%d, farm.configClusterID=%d",
		farm.GetFarmID(), farm.configClusterID)

	farmConfigID := farm.farmID
	if farm.app.Config.Mode == common.MODE_CLUSTER {
		farmConfigID = farm.configClusterID
	}

	conf, err := farm.configStore.Get(farmConfigID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil
	}
	return conf
}

func (farm *DefaultFarmService) SetConfig(farmConfig config.FarmConfig) error {
	if err := farm.configStore.Put(farm.configClusterID, farmConfig); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	farm.PublishConfig(farmConfig)
	return nil
}

// Stores the device state in the farm state
func (farm *DefaultFarmService) SetDeviceState(deviceType string, deviceState state.DeviceStateMap) {
	farm.app.Logger.Debugf("deviceType: %s, deviceState: %+v", deviceType, deviceState)
	state, err := farm.stateStore.Get(farm.farmID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return
	}
	if state == nil {
		// When running in cluster mode, the device state
		// may not have propagated to all nodes yet
		farm.app.Logger.Error("Farm state not found in state store! farm.farmID=%d", farm.farmID)
		return
	}
	state.SetDevice(deviceType, deviceState.Clone())
	farm.stateStore.Put(farm.farmID, state)
}

// Stores the specified device config in the farm and device config stores and publishes
// the whole farm configuration to connected websocket clients.
func (farm *DefaultFarmService) SetDeviceConfig(deviceConfig config.DeviceConfig) error {
	farm.app.Logger.Debugf("config: %+v", deviceConfig)
	deviceService, err := farm.serviceRegistry.GetDeviceService(farm.farmID, deviceConfig.GetType())
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	err = deviceService.SetConfig(deviceConfig)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	farmConfig, err := farm.configStore.Get(deviceConfig.GetFarmID(), farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	farmConfig.SetDevice(deviceConfig)
	farm.configStore.Put(farm.configClusterID, farmConfig)
	farm.PublishConfig(farmConfig)
	return nil
}

// Used by android client to set a specific configuration item in the settings menu
func (farm *DefaultFarmService) SetConfigValue(session Session, farmID, deviceID uint64, key, value string) error {

	farm.app.Logger.Debugf("Setting config farmID=%d, deviceID=%d, key=%s, value=%s",
		farmID, deviceID, key, value)

	configItem, err := farm.deviceConfigDAO.Get(deviceID, key)
	if err != nil {
		return err
	}
	configItem.SetValue(value)
	farm.deviceConfigDAO.Save(configItem)
	farm.app.Logger.Debugf("Saved configuration item: %+v", configItem)

	farmConfig, err := farm.dao.Get(farm.farmID, common.CONSISTENCY_CACHED)

	// This doesnt update raft state in cluster mode
	farm.channels.FarmConfigChangeChan <- farmConfig

	// This causes an infinite loop
	//farm.SetConfig(farmConfig)

	return nil
}

func (farm *DefaultFarmService) SetMetricValue(deviceType string, key string, value float64) error {

	farmState, err := farm.stateStore.Get(farm.farmID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil
	}

	if err := farmState.SetMetricValue(deviceType, key, value); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}

	farm.stateStore.Put(farm.farmID, farmState)

	metricDelta := map[string]float64{key: value}
	channelDelta := make(map[int]int, 0)
	delta := state.CreateDeviceStateDeltaMap(metricDelta, channelDelta)
	if err := farm.PublishDeviceDelta(map[string]state.DeviceStateDeltaMap{deviceType: delta}); err != nil {
		return err
	}

	return nil
}

func (farm *DefaultFarmService) SetSwitchValue(deviceType string, channelID int, value int) error {

	farmState, err := farm.stateStore.Get(farm.farmID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil
	}

	if err := farmState.SetChannelValue(deviceType, channelID, value); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}

	farm.stateStore.Put(farm.farmID, farmState)

	metricDelta := make(map[string]float64, 0)
	channelDelta := map[int]int{channelID: value}
	delta := state.CreateDeviceStateDeltaMap(metricDelta, channelDelta)
	if err := farm.PublishDeviceDelta(map[string]state.DeviceStateDeltaMap{deviceType: delta}); err != nil {
		return err
	}

	return nil
}

// Publishes the entire farm configuration (all devices)
func (farm *DefaultFarmService) PublishConfig(farmConfig config.FarmConfig) error {
	select {
	//case farm.farmConfigChan <- farm.config:
	case farm.channels.FarmConfigChan <- farmConfig:
		farm.app.Logger.Error("PublishConfig fired! %+v", farmConfig)
	default:
		errmsg := "Farm config channel buffer full, discarding update!"
		farm.app.Logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

func (farm *DefaultFarmService) WatchConfig() <-chan config.FarmConfig {
	return farm.channels.FarmConfigChan
}

func (farm *DefaultFarmService) WatchState() <-chan state.FarmStateMap {
	return farm.channels.FarmStateChan
}

// Publishes a DeviceStateMap that contains ONLY the metrics and/or channels that've changed
func (farm *DefaultFarmService) PublishDeviceDelta(deviceState map[string]state.DeviceStateDeltaMap) error {
	select {
	case farm.channels.DeviceStateDeltaChan <- deviceState:
		farm.app.Logger.Error("PublishDeviceDelta fired!")
	default:
		errmsg := "Device delta state channel buffer full, discarding update!"
		farm.app.Logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

func (farm *DefaultFarmService) WatchDeviceDeltas() <-chan map[string]state.DeviceStateDeltaMap {
	return farm.channels.DeviceStateDeltaChan
}

// WatchFarmStateChange is intended to be run within a goroutine that watches the farmStateChangeChan
// and creates a delta that gets published to connected clients anytime the farm state changes.
func (farm *DefaultFarmService) WatchFarmStateChange() {
	farm.app.Logger.Debugf("Watching for farm state changes (farm.id=%d)", farm.GetFarmID())
	for {
		select {
		case newFarmState := <-farm.channels.FarmStateChangeChan:

			lastState, err := farm.stateStore.Get(farm.farmID)
			if err != nil {
				farm.app.Logger.Errorf("Error: %s", err)
				continue
			}

			newDeviceStates := newFarmState.GetDevices()

			farm.app.Logger.Debugf("New farm state published. last=%+v, new=%+v",
				lastState, newDeviceStates)

			if lastState == nil {
				for deviceType, stateMap := range newDeviceStates {
					farm.SetDeviceState(deviceType, stateMap)
				}
				continue
			}

			deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.GetFarmID())

			if err != nil {
				farm.app.Logger.Errorf("Error: %s", err)
			}

			for _, device := range deviceServices {
				deviceType := device.GetDeviceType()
				newDeviceState := newDeviceStates[deviceType]
				_, err := farm.OnDeviceStateChange(deviceType, newDeviceState)
				if err == state.ErrDeviceNotFound {
					farm.SetDeviceState(deviceType, newDeviceState)
				}
				if err != nil {
					farm.app.Logger.Errorf("Error: %s", err)
				}
			}
		case <-farm.farmStateQuitChan:
			farm.app.Logger.Debugf("Closing farm state channel. farmID=%d", farm.farmID)
			return
		}

	}
}

func (farm *DefaultFarmService) WatchFarmConfigChange() {
	farm.app.Logger.Debugf("Watching for farm config changes (configClusterID=%d)", farm.GetConfigClusterID())
	for {
		select {
		case newConfig := <-farm.channels.FarmConfigChangeChan:
			farm.app.Logger.Debugf("New config change for farm %d (configClusterID=%d)",
				farm.GetFarmID(), farm.GetConfigClusterID())

			// Calling farm.SetConfig here results in an infinite loop
			// since SetConfig calls configStore.Put which in turn
			// sends a farmConfigChangeChan message with the newly
			// written data.

			// farm.configMutex.Lock()
			// farm.config = newConfig
			// farm.configMutex.Unlock()

			newMode := newConfig.GetMode()

			if farm.mode != newMode {

				farm.app.Logger.Error("!!!modes dont match!!! old mode=%s, new mode=%s",
					farm.mode, newMode)

				deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.farmID)
				if err != nil {
					farm.app.Logger.Errorf("Error: %s", err)
					continue
				}

				for _, service := range deviceServices {
					var d device.IOSwitcher
					deviceType := service.GetDeviceType()
					deviceConfig, err := newConfig.GetDevice(deviceType)
					if err != nil {
						farm.app.Logger.Error(err)
						continue
					}
					switch newMode {
					case common.CONFIG_MODE_VIRTUAL:
						farmStateMap := state.NewFarmStateMap(farm.farmID)
						d = device.NewVirtualIOSwitch(farm.app, farmStateMap, "", deviceType)
					case common.CONFIG_MODE_SERVER:
						d = device.NewSmartSwitch(farm.app, deviceConfig.GetURI(), deviceType)
					}
					service.SetMode(newMode, d)
				}
			}
			farm.PublishConfig(newConfig)

		case <-farm.farmConfigQuitChan:
			farm.app.Logger.Debugf("Closing farm config channel. farmID=%d", farm.farmID)
			return
		}
	}
}

// WatchFarmStateChange is intended to be run within a goroutine that watches the farmStateChangeChan
// and creates a delta that gets published to connected clients anytime the farm state changes.
// func (farm *DefaultFarmService) WatchDeviceConfigChange() {
// 	farm.app.Logger.Debugf("Watching for device config changes (farm.id=%d)", farm.GetFarmID())
// 	for {
// 		select {
// 		case newDeviceConfig := <-farm.channels.DeviceConfigChangeChan:

// 		}
// 	}
// }

// Watches the DeviceStateChangeChan for real-time incoming device updates.
func (farm *DefaultFarmService) WatchDeviceStateChange() {

	farm.app.Logger.Debugf("Farm %d watching for incoming device state changes", farm.farmID)

	for {
		select {
		case newDeviceState := <-farm.channels.DeviceStateChangeChan:

			deviceID := newDeviceState.DeviceID
			deviceType := newDeviceState.DeviceType
			stateMap := newDeviceState.StateMap

			_, err := farm.OnDeviceStateChange(deviceType, stateMap)
			if err != nil {
				farm.error("WatchDeviceStateChange", "WatchDeviceStateChange", err)
				continue
			}

			farm.SetDeviceState(deviceType, stateMap)

			if err := farm.deviceDataStore.Save(deviceID, stateMap); err != nil {
				farm.app.Logger.Errorf("Error storing device data: %s", err)
				farm.error("WatchDeviceStateChange", "WatchDeviceStateChange", err)
				continue
			}

			deviceService, err := farm.serviceRegistry.GetDeviceService(farm.farmID, deviceType)
			if err != nil {
				farm.app.Logger.Errorf("Error getting device service: %s", err)
				continue
			}

			deviceConfig, err := deviceService.GetConfig()
			if err != nil {
				farm.app.Logger.Errorf("Error getting device config: %s", err)
				continue
			}

			if newDeviceState.IsPollEvent {
				farm.Manage(deviceConfig, farm.GetState())
			}
		case <-farm.deviceStateQuitChan:
			farm.app.Logger.Debugf("Closing device state channel. farmID=%d", farm.farmID)
			return
		}
	}
}

func (farm *DefaultFarmService) OnDeviceStateChange(deviceType string,
	newDeviceState state.DeviceStateMap) (state.DeviceStateDeltaMap, error) {

	farm.app.Logger.Debugf("newDeviceState=%+v", newDeviceState)
	var delta state.DeviceStateDeltaMap

	lastState, err := farm.stateStore.Get(farm.farmID)
	if err != nil && err != state.ErrFarmNotFound {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil, err
	}

	newChannelMap := make(map[int]int, len(newDeviceState.GetChannels()))
	for i, channel := range newDeviceState.GetChannels() {
		newChannelMap[i] = channel
	}

	if lastState == nil {
		delta = state.CreateDeviceStateDeltaMap(newDeviceState.GetMetrics(), newChannelMap)
	} else {
		d, err := lastState.Diff(deviceType, newDeviceState.GetMetrics(), newChannelMap)
		if err != nil {
			farm.app.Logger.Errorf("Error: %s", err)
			return nil, err
		}
		delta = d
	}
	if delta == nil {
		return nil, nil
	}
	if err := farm.PublishDeviceDelta(map[string]state.DeviceStateDeltaMap{deviceType: delta}); err != nil {
		return nil, err
	}
	return delta, nil
}

func (farm *DefaultFarmService) Run() {
	if farm.running == true {
		farm.app.Logger.Errorf("Farm %d already running!", farm.GetFarmID())
		return
	}
	farm.running = true

	go farm.WatchFarmStateChange()
	go farm.WatchFarmConfigChange()
	go farm.WatchDeviceStateChange()

	if !farm.app.DebugFlag {
		// Wait for top of the minute
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				_, _, secs := time.Now().Clock()
				if secs == 0 {
					ticker.Stop()
					return
				}
				farm.app.Logger.Infof("Waiting for top of the minute... %d sec left", 60-secs)
			}
		}
	}

	farm.poll()
	farm.Poll()
}

func (farm *DefaultFarmService) Stop() {
	farm.app.Logger.Debugf("Stopping farm %d", farm.farmID)
	farm.farmStateQuitChan <- 0
	farm.farmConfigQuitChan <- 0
	farm.deviceStateQuitChan <- 0
	farm.pollTickerQuitChan <- 0
	farm.stateStore.Close()
	deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.farmID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
	}
	for _, deviceService := range deviceServices {
		deviceService.Stop()
	}
	farm.running = false
	farm.serviceRegistry.RemoveFarmService(farm.farmID)
}

func (farm *DefaultFarmService) Poll() {

	farmConfig, err := farm.configStore.Get(farm.configClusterID, common.CONSISTENCY_CACHED)
	if err != nil || farmConfig == nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.configClusterID)
		return
	}

	if farmConfig.GetInterval() > 0 {
		ticker := time.NewTicker(time.Duration(farmConfig.GetInterval()) * time.Second)
		for {
			select {
			case <-ticker.C:
				farm.poll()
			case <-farm.pollTickerQuitChan:
				ticker.Stop()
				return
			}
		}
	}
}

func (farm *DefaultFarmService) poll() {
	farm.app.Logger.Debugf("Polling farm: %d", farm.farmID)
	deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.farmID)
	if err != nil {
		farm.app.Logger.Error(err)
		return
	}
	for _, device := range deviceServices {
		device.Poll()
	}
}

func (farm *DefaultFarmService) Manage(deviceConfig config.DeviceConfig,
	farmState state.FarmStateMap) {

	//eventType := "Manage"

	farmConfig, err := farm.configStore.Get(farm.configClusterID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.configClusterID)
		return
	}

	deviceType := deviceConfig.GetType()

	if !deviceConfig.IsEnabled() {
		farm.app.Logger.Warningf("%s disabled...", deviceType)
		return
	}

	if farmConfig.GetMode() == common.CONFIG_MODE_MAINTENANCE {
		farm.app.Logger.Warning("Maintenance mode in progres...")
		return
	}

	if farmState == nil {
		farm.app.Logger.Warningf("No farm state, waiting until initialized. farm.id: %d", farmConfig.GetID())
		return
	}

	for _, err := range farm.ManageMetrics(deviceConfig, farmState) {
		farm.app.Logger.Error(err.Error())
	}

	channels := deviceConfig.GetChannels()
	channelConfigs := make([]config.ChannelConfig, len(channels))
	for i := range channels {
		channelConfigs[i] = &channels[i]
	}
	for _, err := range farm.ManageChannels(deviceConfig, farmState, channelConfigs) {
		farm.app.Logger.Error(err.Error())
	}
}

func (farm *DefaultFarmService) ManageMetrics(config config.DeviceConfig, farmState state.FarmStateMap) []error {
	var errors []error
	eventType := "ALARM"
	deviceType := config.GetType()

	farm.app.Logger.Debugf("Managing configured %s metrics...", deviceType)

	metricConfigs := config.GetMetrics()

	for _, metric := range metricConfigs {

		if !metric.IsEnabled() {
			continue
		}

		metricValue, err := farmState.GetMetricValue(config.GetType(), metric.GetKey())
		if err != nil {
			errors = append(errors, err)
			continue
		}

		farm.app.Logger.Debugf("notify=%t, metric=%s, value=%.2f, alarmLow=%.2f, alarmHigh=%.2f",
			metric.IsNotify(), metric.GetKey(), metricValue, metric.GetAlarmLow(), metric.GetAlarmHigh())

		if metric.IsNotify() && metricValue <= metric.GetAlarmLow() {
			message := fmt.Sprintf("%s LOW: %.2f", metric.GetName(), metricValue)
			farm.notify(deviceType, eventType, message)
		}

		if metric.IsNotify() && metricValue >= metric.GetAlarmHigh() {
			message := fmt.Sprintf("%s HIGH: %.2f", metric.GetName(), metricValue)
			farm.notify(deviceType, eventType, message)
		}
	}
	return errors
}

func (farm *DefaultFarmService) ManageChannels(deviceConfig config.DeviceConfig,
	farmState state.FarmStateMap, channels []config.ChannelConfig) []error {

	farm.app.Logger.Debugf("Managing configured %s channels...", deviceConfig.GetType())

	var errors []error

	deviceService, err := farm.serviceRegistry.GetDeviceService(farm.farmID, deviceConfig.GetType())
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	for _, channel := range channels {

		if len(channel.GetName()) <= 0 {
			channel.SetName(fmt.Sprintf("channel%d", channel.GetChannelID()))
		}

		if !channel.IsEnabled() {
			continue
		}

		farm.app.Logger.Debugf("Managing channel %+v", channel)

		backoff := channel.GetBackoff()
		if backoff > 0 {

			farm.app.Logger.Debugf("farm.backoffTable=%+v", farm.backoffTable)

			if timer, ok := farm.backoffTable[farm.farmID][channel.GetID()]; ok {

				farm.app.Logger.Debugf("timer=%s, backoff=%d, channel=%+v",
					timer.String(), backoff, channel)

				if time.Since(timer).Minutes() < float64(backoff) {
					elapsed := time.Since(timer).Minutes()
					farm.app.Logger.Debugf("Waiting for %s backoff timer to expire. timer=%s, now=%s, elapsed=%.2f, backoff=%d",
						channel.GetName(), timer.String(), time.Now().String(), elapsed, backoff)
					return nil
				} else {
					delete(farm.backoffTable[farm.farmID], channel.GetID())
				}
			}
		}

		if len(channel.GetConditions()) > 0 {

			// When running in cluster mode
			// if farm.backoffTable[farm.farmID] == nil {
			// 	farm.app.Logger.Error("farm.backoffTable[%d] is nil")
			// 	continue
			// }

			handler := NewChannelConditionHandler(farm.app.Logger,
				deviceConfig, channel, farmState, farm, deviceService,
				farm.conditionService, farm.backoffTable[farm.farmID])
			handled, err := handler.Handle()
			if err != nil {
				farm.app.Logger.Debugf("Error processing %s conditions: %s", channel.GetName(), err)
				errors = append(errors, err)
				continue
			}
			if handled {
				farm.app.Logger.Debugf("Channel %s already handled by conditional, aborting schedule processing...", channel.GetName())
				continue
			}
		}

		if len(channel.GetSchedule()) > 0 {
			if err := NewChannelScheduleHandler(farm.app.Logger, deviceConfig, channel,
				farmState, deviceService, farm.scheduleService, farm).Handle(); err != nil {

				farm.app.Logger.Debugf("Error processing %s schedules: %s", channel.GetName(), err)
				errors = append(errors, err)
				continue
			}
		}

	}
	return errors
}

// func (farm *DefaultFarmService) ManageWorkflows() error {
// 	farmConfig := farm.GetConfig()
// 	farm.app.Logger.Debugf("Managing %s workflows...", farmConfig.GetName())
// 	for _, workflow := range farmConfig.GetWorkflows() {
// 		if err := farm.RunWorkflow(&workflow); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (farm *DefaultFarmService) RunWorkflow(workflow config.WorkflowConfig) {
	farmConfig := farm.GetConfig()
	farm.app.Logger.Debugf("Managing %s workflow: %s", farmConfig.GetName(), workflow.GetName())
	go func() {
		for i, step := range workflow.GetSteps() {
			deviceService, err := farm.serviceRegistry.GetDeviceServiceByID(farm.farmID, step.GetDeviceID())
			if err != nil {
				farm.app.Logger.Error(err)
			}
			deviceConfig, err := deviceService.GetConfig()
			if err != nil {
				farm.app.Logger.Error(err)
			}
			for _, channel := range deviceConfig.GetChannels() {
				if channel.GetID() == step.GetChannelID() {
					duration := step.GetDuration()
					boardID := channel.GetChannelID()
					_, err := deviceService.TimerSwitch(boardID,
						duration,
						fmt.Sprintf("%s workflow step #%d switching on %s for %d seconds",
							workflow.GetName(), i+1, channel.GetName(), duration))
					if err != nil {
						farm.app.Logger.Error(err)
						step.SetState(common.WORKFLOW_STATE_ERROR)
						workflow.SetStep(&step)
						farmConfig.SetWorkflow(workflow)
						farm.SetConfig(farmConfig)
						return
					}

					step.SetState(common.WORKFLOW_STATE_EXECUTING)
					workflow.SetStep(&step)
					farmConfig.SetWorkflow(workflow)
					farm.SetConfig(farmConfig)

					totalDuration := time.Duration(duration + step.GetWait())
					timeDuration := time.Duration(time.Second * totalDuration)
					time.Sleep(timeDuration)

					state, err := farm.GetState().GetChannelValue(deviceConfig.GetType(), boardID)
					if err != nil {
						farm.app.Logger.Error(err)
						step.SetState(common.WORKFLOW_STATE_ERROR)
						workflow.SetStep(&step)
						farmConfig.SetWorkflow(workflow)
						farm.SetConfig(farmConfig)
						return
					}
					for state != common.SWITCH_OFF {
						farm.app.Logger.Errorf("Workflow %s waiting for channel %s timer to expire, expected OFF state...",
							workflow.GetName(), channel.GetName())

						time.Sleep(5 * time.Second)
						state, _ = farm.GetState().GetChannelValue(deviceConfig.GetType(), boardID)
					}

					step.SetState(common.WORKFLOW_STATE_COMPLETED)
					workflow.SetStep(&step)
					farmConfig.SetWorkflow(workflow)
					farm.SetConfig(farmConfig)
				}
			}
		}
		now := time.Now().In(farm.app.Location)
		nowHr, nowMin, nowSec := now.Clock()
		nowDateTime := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin, nowSec, 0, farm.app.Location)
		workflow.SetLastCompleted(&nowDateTime)
		for _, step := range workflow.GetSteps() {
			step.SetState(common.WORKFLOW_STATE_READY)
			workflow.SetStep(&step)
		}
		farmConfig.SetWorkflow(workflow)
		farm.SetConfig(farmConfig)
	}()
}

// TODO: replace device service notify with this
func (farm *DefaultFarmService) notify(deviceType, eventType, message string) error {
	farmConfig, err := farm.configStore.Get(farm.configClusterID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
		return err
	}
	return farm.notificationService.Enqueue(&model.Notification{
		Device:    farmConfig.GetName(),
		Priority:  common.NOTIFICATION_PRIORITY_LOW,
		Title:     deviceType,
		Type:      eventType,
		Message:   message,
		Timestamp: time.Now()})
}

func (farm *DefaultFarmService) error(method, eventType string, err error) {
	farm.app.Logger.Errorf("Error: %s", method, err)
	farmConfig, _err := farm.configStore.Get(farm.configClusterID, farm.consistencyLevel)
	if _err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
		return
	}
	farm.notificationService.Enqueue(&model.Notification{
		Device:    farmConfig.GetName(),
		Priority:  common.NOTIFICATION_PRIORITY_HIGH,
		Title:     farmConfig.GetName(),
		Type:      eventType,
		Message:   err.Error(),
		Timestamp: time.Now()})
}
