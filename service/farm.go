package service

import (
	"errors"
	"fmt"
	"hash/fnv"
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
)

type DefaultFarmService struct {
	app                 *app.App
	dao                 dao.FarmDAO
	stateStore          state.FarmStorer
	configStore         store.FarmConfigStorer
	deviceStore         datastore.DeviceDatastore
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
	backoffTable        map[uint64]map[int]time.Time
	FarmService
	deviceConfigDAO dao.DeviceConfigDAO
	//observer.FarmConfigObserver
}

func CreateFarmService(app *app.App, farmDAO dao.FarmDAO, stateStore state.FarmStorer,
	configStore store.FarmConfigStorer, deviceConfigStore store.DeviceConfigStorer,
	deviceStore datastore.DeviceDatastore, farmConfig config.FarmConfig, consistencyLevel int,
	serviceRegistry ServiceRegistry, farmChannels *FarmChannels,
	deviceConfigDAO dao.DeviceConfigDAO) (FarmService, error) {

	farmID := farmConfig.GetID()
	mode := farmConfig.GetMode()

	backoffTable := make(map[uint64]map[int]time.Time, 0)

	// Only used when clustering enabled
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(fmt.Sprintf("%d-%d", farmConfig.GetOrganizationID(), farmID)))
	configClusterID := clusterHash.Sum64()

	_state, err := stateStore.Get(farmID)
	if err == state.ErrFarmNotFound {
		_state = state.NewFarmStateMap(farmID)
		devices, err := serviceRegistry.GetDeviceServices(farmID)
		if err != nil {
			return nil, err
		}
		deviceStateMaps := make([]state.DeviceStateMap, 0)
		for _, d := range devices {
			conf, _ := d.GetConfig()
			deviceID := conf.GetID()
			deviceStateMap := state.CreateDeviceStateMapEmpty(deviceID)
			deviceStateMaps = append(deviceStateMaps, deviceStateMap)
			_state.SetDevice(conf.GetType(), deviceStateMap)
			backoffTable[deviceID] = make(map[int]time.Time, 0)
		}
		stateStore.Put(farmID, _state)
	} else if err != nil {
		app.Logger.Errorf("Fatal farm service error (farmID=%d): %s", farmID, err)
		return nil, err
	}

	return &DefaultFarmService{
		app:                 app,
		dao:                 farmDAO,
		stateStore:          stateStore,
		configStore:         configStore,
		deviceStore:         deviceStore,
		farmID:              farmID,
		mode:                mode,
		consistencyLevel:    consistencyLevel,
		configClusterID:     configClusterID,
		serviceRegistry:     serviceRegistry,
		conditionService:    serviceRegistry.GetConditionService(),
		scheduleService:     serviceRegistry.GetScheduleService(),
		notificationService: serviceRegistry.GetNotificationService(),
		channels:            farmChannels,
		running:             false,
		deviceConfigDAO: deviceConfigDAO,
		backoffTable:        backoffTable}, nil
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
	conf, err := farm.configStore.Get(farm.farmID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil
	}
	return conf
}

func (farm *DefaultFarmService) SetConfig(farmConfig config.FarmConfig) error {
	if err := farm.configStore.Put(farm.farmID, farmConfig); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	farm.PublishConfig(farmConfig)
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
					var d device.SmartSwitcher
					deviceType := service.GetDeviceType()
					deviceConfig, err := newConfig.GetDevice(deviceType)
					if err != nil {
						farm.app.Logger.Error(err)
						continue
					}
					switch newMode {
					case common.CONFIG_MODE_VIRTUAL:
						farmStateMap := state.NewFarmStateMap(farm.farmID)
						d = device.NewVirtualSmartSwitch(farm.app, farmStateMap, "", deviceType)
					case common.CONFIG_MODE_SERVER:
						d = device.NewSmartSwitch(farm.app, deviceConfig.GetURI(), deviceType)
					}
					service.SetMode(newMode, d)
				}
			}

			farm.PublishConfig(newConfig)
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

			farm.app.Logger.Errorf("Newest state: %+v", newDeviceState)

			deviceID := newDeviceState.DeviceID
			deviceType := newDeviceState.DeviceType
			stateMap := newDeviceState.StateMap

			_, err := farm.OnDeviceStateChange(deviceType, stateMap)
			if err != nil {
				farm.error("WatchDeviceStateChange", "WatchDeviceStateChange", err)
				continue
			}

			farm.SetDeviceState(deviceType, stateMap)

			if err := farm.deviceStore.Save(deviceID, stateMap); err != nil {
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

			farm.Manage(deviceConfig, farm.GetState())
		}
	}
}

func (farm *DefaultFarmService) OnDeviceStateChange(deviceType string,
	newDeviceState state.DeviceStateMap) (state.DeviceStateDeltaMap, error) {

	farm.app.Logger.Debugf("newDeviceState=%+v", newDeviceState)

	lastState, err := farm.stateStore.Get(farm.farmID)
	if err != nil && err != state.ErrFarmNotFound {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil, err
	}

	var delta state.DeviceStateDeltaMap
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

// Stores the device state in the farm state
func (farm *DefaultFarmService) SetDeviceState(deviceType string, deviceState state.DeviceStateMap) {
	farm.app.Logger.Debugf("deviceType: %s, deviceState: %+v",
		deviceType, deviceState)

	state, err := farm.stateStore.Get(farm.farmID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return
	}
	state.SetDevice(deviceType, deviceState)
	farm.stateStore.Put(farm.farmID, state)
}

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

	farmConfig, err := farm.dao.Get(farm.farmID)

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

func (farm *DefaultFarmService) Poll() {

	farmConfig, err := farm.configStore.Get(farm.farmID, common.CONSISTENCY_CACHED)
	if err != nil || farmConfig == nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
		return
	}

	if farmConfig.GetInterval() > 0 {
		ticker := time.NewTicker(time.Duration(farmConfig.GetInterval()) * time.Second)
		quit := make(chan struct{})
		for {
			select {
			case <-ticker.C:
				farm.poll()
			case <-quit:
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
		//deviceType := device.GetDeviceType()
		go device.Poll(farm.channels.DeviceStateChangeChan)
		// if err := device.Poll(farm.channels.DeviceStateChangeChan); err != nil {
		// 	farm.app.Logger.Errorf("Error polling %s device", deviceType)
		// }
	}
}

func (farm *DefaultFarmService) Manage(deviceConfig config.DeviceConfig,
	farmState state.FarmStateMap) {

	//eventType := "Manage"

	farmConfig, err := farm.configStore.Get(farm.farmID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
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

	farm.app.Logger.Debugf("Processing configured %s metrics...", deviceType)

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

	farm.app.Logger.Debugf("Processing configured %s channels...",
		deviceConfig.GetType())

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

		farm.app.Logger.Debugf("Processing channel %+v", channel)

		backoff := channel.GetBackoff()
		if backoff > 0 {
			if timer, ok := farm.backoffTable[farm.farmID][channel.GetID()]; ok {
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
			handled, err := NewChannelConditionHandler(farm.app.Logger,
				deviceConfig, channel, farmState, farm, deviceService,
				farm.conditionService, farm.backoffTable[deviceConfig.GetID()]).Handle()
			if err != nil {
				farm.app.Logger.Debugf("Error processing %s conditions: %s", channel.GetName(), err)
				errors = append(errors, err)
				continue
			}
			if handled {
				farm.app.Logger.Debugf("Channel already handled by conditional, aborting schedule processing...")
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

// TODO: replace device service notify with this
func (farm *DefaultFarmService) notify(deviceType, eventType, message string) error {
	farmConfig, err := farm.configStore.Get(farm.farmID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
		return err
	}
	return farm.notificationService.Enqueue(&model.Notification{
		Device: farmConfig.GetName(),
		Priority:   common.NOTIFICATION_PRIORITY_LOW,
		Title:      deviceType,
		Type:       eventType,
		Message:    message,
		Timestamp:  time.Now()})
}

func (farm *DefaultFarmService) error(method, eventType string, err error) {
	farm.app.Logger.Errorf("Error: %s", method, err)
	farmConfig, _err := farm.configStore.Get(farm.farmID, farm.consistencyLevel)
	if _err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
		return
	}
	farm.notificationService.Enqueue(&model.Notification{
		Device: farmConfig.GetName(),
		Priority:   common.NOTIFICATION_PRIORITY_HIGH,
		Title:      farmConfig.GetName(),
		Type:       eventType,
		Message:    err.Error(),
		Timestamp:  time.Now()})
}
