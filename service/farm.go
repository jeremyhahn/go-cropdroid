package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/pki/ca"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
)

type FarmServicer interface {
	Devices() ([]model.Device, error)
	GetFarmID() uint64
	GetChannels() *FarmChannels
	GetConfig() config.Farm
	GetConsistencyLevel() int
	GetPublicKey() string
	GetState() state.FarmStateMap
	GetStateID() uint64
	InitializeState(saveToStateStore bool) error
	IsRunning() bool
	Poll()
	PublishConfig(farmConfig config.Farm) error
	PublishState(farmState state.FarmStateMap) error
	PublishDeviceState(deviceState map[string]state.DeviceStateMap) error
	PublishDeviceDelta(deviceState map[string]state.DeviceStateDeltaMap) error
	RefreshHardwareVersions() error
	Run()
	RunCluster()
	RunWorkflow(workflow config.Workflow)
	SaveConfig(farmConfig config.Farm) error
	SetConfig(farmConfig config.Farm) error
	SetDeviceConfig(deviceConfig config.Device) error
	SetDeviceState(deviceType string, deviceState state.DeviceStateMap)
	SetConfigValue(session Session, farmID, deviceID uint64, key, value string) error
	SetMetricValue(deviceType string, key string, value float64) error
	SetSwitchValue(deviceType string, channelID int, value int) error
	Stop()
	WatchConfig() <-chan config.Farm
	WatchState() <-chan state.FarmStateMap
	WatchDeviceState() <-chan map[string]state.DeviceStateMap
	WatchDeviceDeltas() <-chan map[string]state.DeviceStateDeltaMap
	WatchFarmStateChange()
}

type DefaultFarmService struct {
	app                 *app.App
	idGenerator         util.IdGenerator
	farmID              uint64
	farmStateID         uint64
	mode                string
	consistencyLevel    int
	running             bool
	channels            *FarmChannels
	backoffTable        map[uint64]map[uint64]time.Time
	farmDAO             dao.FarmDAO
	deviceSettingDAO    dao.DeviceSettingDAO
	deviceMapper        mapper.DeviceMapper
	farmStateStore      state.FarmStateStorer
	deviceStateStore    state.DeviceStateStorer
	deviceDataStore     datastore.DeviceDataStore
	serviceRegistry     ServiceRegistry
	conditionService    ConditionServicer
	scheduleService     ScheduleService
	notificationService NotificationServicer
	farmStateQuitChan   chan int
	farmConfigQuitChan  chan int
	deviceStateQuitChan chan int
	pollTickerQuitChan  chan int
	FarmServicer
}

func CreateFarmService(
	app *app.App,
	farmDAO dao.FarmDAO,
	idGenerator util.IdGenerator,
	farmStateStore state.FarmStateStorer,
	deviceStateStore state.DeviceStateStorer,
	deviceDataStore datastore.DeviceDataStore,
	farmConfig config.Farm,
	consistencyLevel int,
	serviceRegistry ServiceRegistry,
	farmChannels *FarmChannels,
	deviceSettingDAO dao.DeviceSettingDAO,
	deviceMapper mapper.DeviceMapper) (FarmServicer, error) {

	farmID := farmConfig.Identifier()

	farmService := &DefaultFarmService{
		app:                 app,
		farmDAO:             farmDAO,
		idGenerator:         idGenerator,
		farmStateStore:      farmStateStore,
		deviceStateStore:    deviceStateStore,
		deviceDataStore:     deviceDataStore,
		farmID:              farmID,
		farmStateID:         idGenerator.NewFarmStateID(farmID),
		mode:                farmConfig.GetMode(),
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
		deviceSettingDAO:    deviceSettingDAO,
		deviceMapper:        deviceMapper,
		backoffTable:        make(map[uint64]map[uint64]time.Time, 0)}

	return farmService, nil
}

// Returns true if the farm service is running, false otherwise.
func (farm *DefaultFarmService) IsRunning() bool {
	return farm.running
}

// Initializes the farm state and device states with zero values
func (farm *DefaultFarmService) InitializeState(saveToStateStore bool) error {
	farm.app.Logger.Debugf("Initalizing farm state: farmID=%d, saveToStateStore=%t",
		farm.farmID, saveToStateStore)
	deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.farmID)
	if err != nil {
		return err
	}
	farmState := state.NewFarmStateMap(farm.farmStateID)
	for _, deviceService := range deviceServices {
		deviceConfig, err := deviceService.Config()
		if err != nil {
			// This for loop "fixes" a race condition / table locking problem with sqlite memory
			for err != nil && err.Error() == "database table is locked: devices" {
				deviceConfig, err = deviceService.Config()
				if err != nil {
					farm.app.Logger.Warning(err)
				}
				time.Sleep(50 * time.Millisecond)
			}
			farm.app.Logger.Warning(err)
			for deviceConfig == nil {
				farm.app.Logger.Warningf("Waiting for device config to become available...")
				time.Sleep(1 * time.Minute)
				deviceConfig, err = deviceService.Config()
				if err != nil {
					farm.app.Logger.Error(err)
					return err
				}
			}
		}
		deviceID := deviceConfig.Identifier()
		deviceType := deviceConfig.GetType()
		metrics := deviceConfig.GetMetrics()
		channels := deviceConfig.GetChannels()
		metricLen := len(metrics)
		deviceStateMap := state.CreateEmptyDeviceStateMap(deviceID, metricLen, len(channels))

		// Create default metric value state
		metricMap := make(map[string]float64, metricLen)
		for _, metric := range deviceConfig.GetMetrics() {
			metricMap[metric.Key] = 0
		}
		deviceStateMap.SetMetrics(metricMap)

		farmState.SetDevice(deviceType, deviceStateMap)

		if err := deviceService.SetState(deviceStateMap); err != nil {
			farm.app.Logger.Error(err)
			return err
		}

		farm.backoffTable[farm.farmID] = make(map[uint64]time.Time, 0)
	}
	if saveToStateStore {
		farm.farmStateStore.Put(farm.farmStateID, farmState)
	}
	return nil
}

// Returns a list of complete device models with the configuration and current state.
// "Server" devices are filtered  so only the IoT devices are returned.
func (farm *DefaultFarmService) Devices() ([]model.Device, error) {
	var devices []model.Device
	deviceConfigs := farm.GetConfig().GetDevices()
	for _, deviceConfig := range deviceConfigs {
		if deviceConfig.GetType() == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		deviceState, err := farm.deviceStateStore.Get(deviceConfig.ID)
		if err != nil {
			return nil, err
		}
		device, err := farm.deviceMapper.MapStateToDevice(deviceState, deviceConfig)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, nil
}

// Refreshes the farm state with the latest device hardware and firmware
// versions reported by each of the devices.
func (farm *DefaultFarmService) RefreshHardwareVersions() error {
	deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.farmID)
	for _, ds := range deviceServices {
		err := ds.RefreshSystemInfo()
		if err != nil {
			return err
		}
	}
	return err
}

// Returns the unique farm ID
func (farm *DefaultFarmService) GetFarmID() uint64 {
	return farm.farmID
}

// Returns the unique farm state ID
func (farm *DefaultFarmService) GetStateID() uint64 {
	return farm.farmStateID
}

// Returns the farm read consistency mode. When running
// in cluster mode, consistent reads are performed against
// the cluster quorum. All other modes and backends use
// local reads.
func (farm *DefaultFarmService) GetConsistencyLevel() int {
	return farm.consistencyLevel
}

// Returns the farm golang channels used for concurrent communication
func (farm *DefaultFarmService) GetChannels() *FarmChannels {
	return farm.channels
}

// Returns the farm RSA public key
func (farm *DefaultFarmService) GetPublicKey() string {
	// Try to load a certificate issued to the farm
	cert, err := farm.app.CA.PEM(fmt.Sprintf("%d", farm.farmID))
	if err == ca.ErrCertNotFound {
		// Fall back to the web server certificate
		cert, err = farm.app.CA.PEM(farm.app.Domain)
		if err == ca.ErrCertNotFound {
			farm.app.Logger.Error(err)
			return ""
		}
		return ""
	}
	return string(cert)
}

// Returns the current farm state, with each of its devices.
func (farm *DefaultFarmService) GetState() state.FarmStateMap {
	farmState, err := farm.farmStateStore.Get(farm.farmStateID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %+v", err)
		return state.NewFarmStateMap(farm.farmStateID)
	}
	return farmState
}

// Returns the farm configuration
func (farm *DefaultFarmService) GetConfig() config.Farm {
	farm.app.Logger.Debugf("Getting farm configuration. farm.id=%d", farm.GetFarmID())
	conf, err := farm.farmDAO.Get(farm.farmID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil
	}
	return conf
}

// Saves the farm configuration to the database and broadcast it to connected clients
func (farm *DefaultFarmService) SetConfig(farmConfig config.Farm) error {
	if err := farm.SaveConfig(farmConfig); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	farm.PublishConfig(farmConfig)
	return nil
}

// Saves the configuration to the database
func (farm *DefaultFarmService) SaveConfig(farmConfig config.Farm) error {
	if err := farm.farmDAO.Save(farmConfig.(*config.FarmStruct)); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	return nil
}

// Stores the device state in the farm state
func (farm *DefaultFarmService) SetDeviceState(deviceType string, deviceState state.DeviceStateMap) {
	farm.app.Logger.Debugf("deviceType: %s, deviceState: %+v", deviceType, deviceState)
	farmState, err := farm.farmStateStore.Get(farm.farmStateID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return
	}
	if farmState == nil {
		// When running in cluster mode, the device state
		// may not have propagated to all nodes yet
		farm.app.Logger.Error("Farm state not found in state store! farm.farmStateID=%d",
			farm.farmStateID)
		return
	}
	farmState.SetDevice(deviceType, deviceState.Clone())
	farm.farmStateStore.Put(farm.farmStateID, farmState)
}

// Stores the specified device config in the farm and device config stores and publishes
// the whole farm configuration to connected websocket clients.
func (farm *DefaultFarmService) SetDeviceConfig(deviceConfig config.Device) error {
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
	farmConfig, err := farm.farmDAO.Get(deviceConfig.GetFarmID(), farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}
	farmConfig.SetDevice(deviceConfig.(*config.DeviceStruct))
	farm.farmDAO.Save(farmConfig)
	farm.PublishConfig(farmConfig)
	return nil
}

// Used by android client to set a specific configuration item in the settings menu
func (farm *DefaultFarmService) SetConfigValue(session Session, farmID, deviceID uint64, key, value string) error {

	farm.app.Logger.Debugf("Setting config farmID=%d, deviceID=%d, key=%s, value=%s",
		farmID, deviceID, key, value)

	configItem, err := farm.deviceSettingDAO.Get(farm.farmID,
		deviceID, key, farm.consistencyLevel)
	if err != nil {
		return err
	}
	configItem.SetValue(value)
	farm.deviceSettingDAO.Save(farm.farmID, configItem)
	farm.app.Logger.Debugf("Saved configuration item: %+v", configItem)

	farmConfig, err := farm.farmDAO.Get(farm.farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}

	// This doesnt update raft state in cluster mode ...
	farm.channels.FarmConfigChangeChan <- farmConfig

	// ... while this causes an infinite loop
	//farm.SetConfig(farmConfig)

	return nil
}

// Sets a device's metric value
func (farm *DefaultFarmService) SetMetricValue(deviceType string, key string, value float64) error {

	farmState, err := farm.farmStateStore.Get(farm.farmStateID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil
	}

	if err := farmState.SetMetricValue(deviceType, key, value); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}

	farm.farmStateStore.Put(farm.farmStateID, farmState)

	metricDelta := map[string]float64{key: value}
	channelDelta := make(map[int]int, 0)
	delta := state.CreateDeviceStateDeltaMap(metricDelta, channelDelta)
	if err := farm.PublishDeviceDelta(map[string]state.DeviceStateDeltaMap{deviceType: delta}); err != nil {
		return err
	}

	return nil
}

// Switches a device channel on or off
func (farm *DefaultFarmService) SetSwitchValue(deviceType string, channelID int, value int) error {

	farmState, err := farm.farmStateStore.Get(farm.farmStateID)
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil
	}

	if err := farmState.SetChannelValue(deviceType, channelID, value); err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return err
	}

	farm.farmStateStore.Put(farm.farmStateID, farmState)

	metricDelta := make(map[string]float64, 0)
	channelDelta := map[int]int{channelID: value}
	delta := state.CreateDeviceStateDeltaMap(metricDelta, channelDelta)
	if err := farm.PublishDeviceDelta(map[string]state.DeviceStateDeltaMap{deviceType: delta}); err != nil {
		return err
	}

	return nil
}

// Publishes the entire farm configuration (all devices)
func (farm *DefaultFarmService) PublishConfig(farmConfig config.Farm) error {
	select {
	case farm.channels.FarmConfigChan <- farmConfig:
		farm.app.Logger.Error("PublishConfig fired! %+v", farmConfig)
	default:
		errmsg := "farm config channel buffer full, discarding broadcast"
		farm.app.Logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

// Publishes a DeviceStateMap that contains ONLY the metrics and/or channels that've changed since the last update
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

// WatchFarmStateChange runs within a goroutine and listens to the farmStateChangeChan
// for incoming farm state change messages. When received, the new state is compared
// against the previous state to produce a delta that contains only the values that have
// changed since the last update, and broadcasts the delta to connected real-time clients.
func (farm *DefaultFarmService) WatchFarmStateChange() {
	farm.app.Logger.Debugf("Watching for farm state changes (farm.id=%d)", farm.GetFarmID())
	for {
		select {
		case newFarmState := <-farm.channels.FarmStateChangeChan:

			lastState, err := farm.farmStateStore.Get(farm.farmStateID)
			if err != nil {
				farm.app.Logger.Errorf("Error: %s", err)
				continue
			}

			newDeviceStates := newFarmState.GetDevices()

			farm.app.Logger.Debugf("Publishing new farm state. last=%+v, new=%+v",
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

			for _, deviceService := range deviceServices {
				deviceType := deviceService.DeviceType()
				// Use GetDevice here instead of accessing the map
				// directly to prevent concurrent read/write map error
				newDeviceState, err := newFarmState.GetDevice(deviceType)
				if err != nil {
					farm.app.Logger.Errorf("Error: %s", err)
					continue
				}
				_, err = farm.OnDeviceStateChange(deviceType, newDeviceState)
				if err == state.ErrDeviceNotFound {
					farm.SetDeviceState(deviceType, newDeviceState)
				}
				if err != nil {
					farm.app.Logger.Errorf("Error: %s", err)
				}
			}
		case <-farm.farmStateQuitChan:
			farm.app.Logger.Debugf("Closing farm state channel. farmStateID=%d", farm.farmStateID)
			return
		}

	}
}

// WatchFarmConfigChange runs within a goroutine and listening to FarmConfigChangeChan for
// incoming configuration updates
func (farm *DefaultFarmService) WatchFarmConfigChange() {
	farm.app.Logger.Debugf("Watching for farm config changes (farmID=%d)", farm.farmID)
	for {
		select {
		case newConfig := <-farm.channels.FarmConfigChangeChan:
			farm.app.Logger.Debugf("New config change for farm %d", farm.GetFarmID())

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
					deviceType := service.DeviceType()
					deviceConfig, err := newConfig.GetDevice(deviceType)
					if err != nil {
						farm.app.Logger.Error(err)
						continue
					}
					switch newMode {
					case common.CONFIG_MODE_VIRTUAL:
						farmStateMap := state.NewFarmStateMap(farm.farmStateID)
						d = device.NewVirtualIOSwitch(farm.app, farmStateMap, "", deviceType)
					case common.CONFIG_MODE_SERVER:
						d = device.NewSmartSwitch(farm.app, deviceConfig.GetURI(), deviceType)
					}
					service.SetMode(newMode, d)
				}
			}
			farm.app.Logger.Debugf("Publishing new farm config. farmID=%d", farm.farmID)
			farm.PublishConfig(newConfig)

		case <-farm.farmConfigQuitChan:
			farm.app.Logger.Debugf("Closing farm config channel. farmID=%d", farm.farmID)
			return
		}
	}
}

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

			if newDeviceState.IsPollEvent {
				deviceService, err := farm.serviceRegistry.GetDeviceService(farm.farmID, deviceType)
				if err != nil {
					farm.app.Logger.Errorf("Error getting device service: %s", err)
					continue
				}
				deviceConfig, err := deviceService.Config()
				if err != nil {
					farm.app.Logger.Errorf("Error getting device config: %s", err)
					continue
				}
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

	lastFarmState, err := farm.farmStateStore.Get(farm.farmStateID)
	if err != nil && err != state.ErrFarmNotFound {
		farm.app.Logger.Errorf("Error: %s", err)
		return nil, err
	}
	farm.app.Logger.Debugf("lastFarmState=%+v", lastFarmState)

	newChannelMap := make(map[int]int, len(newDeviceState.GetChannels()))
	for i, channel := range newDeviceState.GetChannels() {
		newChannelMap[i] = channel
	}

	if lastFarmState == nil {
		delta = state.CreateDeviceStateDeltaMap(newDeviceState.GetMetrics(), newChannelMap)
	} else {
		d, err := lastFarmState.Diff(deviceType, newDeviceState.GetMetrics(), newChannelMap)
		if err != nil {
			farm.app.Logger.Errorf("Error: %s", err)
			return nil, err
		}
		delta = d
	}
	if delta == nil {
		return nil, nil
	}

	farm.app.Logger.Debugf("Publishing device delta. deviceType=%s, delta: %+v", deviceType, delta)

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

	// if !farm.app.DebugFlag {
	// 	// Wait for top of the minute
	// 	ticker := time.NewTicker(time.Second)
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			_, _, secs := time.Now().Clock()
	// 			if secs == 0 {
	// 				ticker.Stop()
	// 				return
	// 			}
	// 			farm.app.Logger.Infof("Waiting for top of the minute... %d sec left", 60-secs)
	// 		}
	// 	}
	// }

	farm.poll()
	farm.Poll()
}

func (farm *DefaultFarmService) Stop() {
	farm.app.Logger.Debugf("Stopping farm %d", farm.farmID)
	farm.farmStateQuitChan <- 0
	farm.farmConfigQuitChan <- 0
	farm.deviceStateQuitChan <- 0
	farm.pollTickerQuitChan <- 0
	farm.farmStateStore.Close()
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

	farmConfig, err := farm.farmDAO.Get(farm.farmID, common.CONSISTENCY_LOCAL)
	if err != nil || farmConfig.ID == 0 {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
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

func (farm *DefaultFarmService) Manage(deviceConfig config.Device, farmState state.FarmStateMap) {

	//eventType := "Manage"

	farmConfig, err := farm.farmDAO.Get(farm.farmID, farm.consistencyLevel)
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
		farm.app.Logger.Warningf("No farm state, waiting until initialized. farm.id: %d", farmConfig.ID)
		return
	}

	for _, err := range farm.ManageMetrics(deviceConfig, farmState) {
		farm.app.Logger.Error(err.Error())
	}

	channels := deviceConfig.GetChannels()
	for _, err := range farm.ManageChannels(deviceConfig, farmState, channels) {
		farm.app.Logger.Error(err.Error())
	}
}

func (farm *DefaultFarmService) ManageMetrics(config config.Device, farmState state.FarmStateMap) []error {
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

func (farm *DefaultFarmService) ManageChannels(deviceConfig config.Device,
	farmState state.FarmStateMap, channels []*config.ChannelStruct) []error {

	farm.app.Logger.Debugf("Managing configured %s channels...", deviceConfig.GetType())

	var errors []error

	deviceService, err := farm.serviceRegistry.GetDeviceService(farm.farmID, deviceConfig.GetType())
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	for _, channel := range channels {

		if len(channel.GetName()) <= 0 {
			channel.SetName(fmt.Sprintf("channel%d", channel.GetBoardID()))
		}

		if !channel.IsEnabled() {
			continue
		}

		farm.app.Logger.Debugf("Managing channel %+v", channel)

		backoff := channel.GetBackoff()
		if backoff > 0 {

			farm.app.Logger.Debugf("farm.backoffTable=%+v", farm.backoffTable)

			if timer, ok := farm.backoffTable[farm.farmID][channel.ID]; ok {

				farm.app.Logger.Debugf("timer=%s, backoff=%d, channel=%+v",
					timer.String(), backoff, channel)

				if time.Since(timer).Minutes() < float64(backoff) {
					elapsed := time.Since(timer).Minutes()
					farm.app.Logger.Debugf("Waiting for %s backoff timer to expire. timer=%s, now=%s, elapsed=%.2f, backoff=%d",
						channel.GetName(), timer.String(), time.Now().String(), elapsed, backoff)
					return nil
				} else {
					delete(farm.backoffTable[farm.farmID], channel.Identifier())
				}
			}
		}

		if len(channel.GetConditions()) > 0 {

			// When running in cluster mode
			// if farm.backoffTable[farm.farmID] == nil {
			// 	farm.app.Logger.Error("farm.backoffTable[%d] is nil")
			// 	continue
			// }

			handler := NewChannelConditionHandler(farm.app.Logger, farm.idGenerator,
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

func (farm *DefaultFarmService) RunWorkflow(workflow config.Workflow) {
	farmConfig := farm.GetConfig()
	farm.app.Logger.Debugf("Managing %s workflow: %s", farmConfig.GetName(), workflow.GetName())
	go func() {
		for i, step := range workflow.GetSteps() {
			deviceService, err := farm.serviceRegistry.GetDeviceServiceByID(farm.farmID, step.GetDeviceID())
			if err != nil {
				farm.app.Logger.Error(err)
			}
			deviceConfig, err := deviceService.Config()
			if err != nil {
				farm.app.Logger.Error(err)
			}
			for _, channel := range deviceConfig.GetChannels() {
				if channel.ID == step.GetChannelID() {
					duration := step.GetDuration()
					boardID := channel.GetBoardID()
					_, err := deviceService.TimerSwitch(boardID,
						duration,
						fmt.Sprintf("%s workflow step #%d switching on %s for %d seconds",
							workflow.GetName(), i+1, channel.GetName(), duration))
					if err != nil {
						farm.app.Logger.Error(err)
						step.SetState(common.WORKFLOW_STATE_ERROR)
						workflow.SetStep(step)
						farmConfig.SetWorkflow(workflow.(*config.WorkflowStruct))
						farm.SetConfig(farmConfig)
						return
					}

					step.SetState(common.WORKFLOW_STATE_EXECUTING)
					workflow.SetStep(step)
					farmConfig.SetWorkflow(workflow.(*config.WorkflowStruct))
					farm.SetConfig(farmConfig)

					totalDuration := time.Duration(duration + step.GetWait())
					timeDuration := time.Duration(time.Second * totalDuration)
					time.Sleep(timeDuration)

					state, err := farm.GetState().GetChannelValue(deviceConfig.GetType(), boardID)
					if err != nil {
						farm.app.Logger.Error(err)
						step.SetState(common.WORKFLOW_STATE_ERROR)
						workflow.SetStep(step)
						farmConfig.SetWorkflow(workflow.(*config.WorkflowStruct))
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
					workflow.SetStep(step)
					farmConfig.SetWorkflow(workflow.(*config.WorkflowStruct))
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
			workflow.SetStep(step)
		}
		farmConfig.SetWorkflow(workflow.(*config.WorkflowStruct))
		farm.SetConfig(farmConfig)
	}()
}

// TODO: replace device service notify with this
func (farm *DefaultFarmService) notify(deviceType, eventType, message string) error {
	farmConfig, err := farm.farmDAO.Get(farm.farmID, farm.consistencyLevel)
	if err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
		return err
	}
	return farm.notificationService.Enqueue(&model.NotificationStruct{
		Device:    farmConfig.GetName(),
		Priority:  common.NOTIFICATION_PRIORITY_LOW,
		Title:     deviceType,
		Type:      eventType,
		Message:   message,
		Timestamp: time.Now()})
}

func (farm *DefaultFarmService) error(method, eventType string, err error) {
	farm.app.Logger.Errorf("Error: %s", method, err)
	farmConfig, _err := farm.farmDAO.Get(farm.farmID, farm.consistencyLevel)
	if _err != nil {
		farm.app.Logger.Errorf("Farm config not found: %d", farm.farmID)
		return
	}
	farm.notificationService.Enqueue(&model.NotificationStruct{
		Device:    farmConfig.GetName(),
		Priority:  common.NOTIFICATION_PRIORITY_HIGH,
		Title:     farmConfig.GetName(),
		Type:      eventType,
		Message:   err.Error(),
		Timestamp: time.Now()})
}
