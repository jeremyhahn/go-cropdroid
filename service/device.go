package service

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/model"

	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type DeviceServicer interface {
	SetMetricValue(key string, value float64) error
	DeviceType() string
	Config() (config.Device, error)
	ID() uint64
	State() (state.DeviceStateMap, error)
	View() (viewmodel.DeviceView, error)
	History(metric string) ([]float64, error)
	Device() (model.Device, error)
	Manage(farmState state.FarmStateMap)
	Poll() error
	SetConfig(config config.Device) error
	SetMode(mode string, device device.IOSwitcher)
	SetState(deviceStateMap state.DeviceStateMap) error
	Stop()
	Switch(channelID, position int, logMessage string) (*common.Switch, error)
	TimerSwitch(channelID, duration int, logMessage string) (common.TimerEvent, error)
	ManageMetrics(config config.Device, farmState state.FarmStateMap) []error
	ManageChannels(deviceConfig config.Device, farmState state.FarmStateMap, channels []model.Channel) []error
	ChannelConfig(channelID int) (config.Channel, error)
	RefreshSystemInfo() error
}

type IOSwitchDeviceService struct {
	app             *app.App
	farmID          uint64
	deviceID        uint64
	farmName        string
	consistency     int
	stateStore      state.DeviceStateStorer
	deviceDAO       dao.DeviceDAO
	eventLogDAO     dao.EventLogDAO
	deviceStore     datastore.DeviceDataStore
	device          device.IOSwitcher
	deviceMutex     *sync.RWMutex
	mapper          mapper.DeviceMapper
	eventLogService EventLogServicer
	farmChannels    *FarmChannels
	DeviceServicer
}

func NewDeviceService(
	app *app.App,
	farmID, deviceID uint64,
	farmName string,
	stateStore state.DeviceStateStorer,
	deviceDAO dao.DeviceDAO,
	eventLogDAO dao.EventLogDAO,
	deviceDatastore datastore.DeviceDataStore,
	deviceMapper mapper.DeviceMapper,
	device device.IOSwitcher,
	farmChannels *FarmChannels,
	consistency int) (DeviceServicer, error) {

	// deviceConfig, err := deviceDAO.Get(farmID, deviceID, consistency)
	// if err != nil {
	// 	return nil, err
	// }

	// //if deviceConfig.IsEnabled() && deviceConfig.GetURI() != "" {
	// if deviceConfig.IsEnabled() {
	// 	deviceInfo, err := device.SystemInfo()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	deviceConfig.SetHardwareVersion(deviceInfo.GetHardwareVersion())
	// 	deviceConfig.SetFirmwareVersion(deviceInfo.GetFirmwareVersion())
	// 	deviceDAO.Save(deviceConfig)
	// }

	app.Logger.Errorf("Creating IOSwitchDeviceService for %s", device.GetType())

	return &IOSwitchDeviceService{
		app:             app,
		deviceID:        deviceID,
		farmID:          farmID,
		farmName:        farmName,
		stateStore:      stateStore,
		deviceDAO:       deviceDAO,
		eventLogDAO:     eventLogDAO,
		deviceStore:     deviceDatastore,
		mapper:          deviceMapper,
		device:          device,
		deviceMutex:     &sync.RWMutex{},
		eventLogService: NewEventLogService(app, eventLogDAO, farmID),
		farmChannels:    farmChannels,
		consistency:     consistency}, nil
}

// Retrives the latest hardware and firmware versions from the device
func (service *IOSwitchDeviceService) RefreshSystemInfo() error {
	deviceConfig, err := service.deviceDAO.Get(service.farmID, service.deviceID, service.consistency)
	if err != nil {
		return err
	}
	if deviceConfig.IsEnabled() {
		deviceInfo, err := service.device.SystemInfo()
		if err != nil {
			return err
		}
		deviceConfig.SetHardwareVersion(deviceInfo.GetHardwareVersion())
		deviceConfig.SetFirmwareVersion(deviceInfo.GetFirmwareVersion())
		service.deviceDAO.Save(deviceConfig)
	}
	return nil
}

// Closes the device state store
func (service *IOSwitchDeviceService) Stop() {
	service.app.Logger.Debugf("closing device state store. deviceID=%d, farmName=%s",
		service.deviceID, service.farmName)
	service.stateStore.Close()
}

// Returns the unique device ID
func (service *IOSwitchDeviceService) ID() uint64 {
	return service.deviceID
}

// Returns the device type
func (service *IOSwitchDeviceService) DeviceType() string {
	return service.device.GetType()
}

// Returns the device configuration
func (service *IOSwitchDeviceService) Config() (config.Device, error) {
	return service.deviceDAO.Get(service.farmID, service.deviceID, service.consistency)
}

// Sets the device configuration
func (service *IOSwitchDeviceService) SetConfig(deviceConfig config.Device) error {
	return service.deviceDAO.Save(deviceConfig.(*config.DeviceStruct))
}

// Returns the current device state
func (service *IOSwitchDeviceService) State() (state.DeviceStateMap, error) {
	return service.stateStore.Get(service.deviceID)
}

// Sets the current device state
func (service *IOSwitchDeviceService) SetState(deviceStateMap state.DeviceStateMap) error {
	return service.stateStore.Put(service.deviceID, deviceStateMap.(*state.DeviceState))
}

// Sets the device operational mode
func (service *IOSwitchDeviceService) SetMode(mode string, d device.IOSwitcher) {
	service.deviceMutex.Lock()
	defer service.deviceMutex.Unlock()
	service.device = d
}

// Returns a complete device viewmodel that contains the device configuration and
// current state, with all metrics and channels sorted by name.
func (service *IOSwitchDeviceService) View() (viewmodel.DeviceView, error) {
	device, err := service.Device()
	if err != nil {
		return nil, err
	}
	metrics := device.GetMetrics()
	channels := device.GetChannels()
	sort.SliceStable(metrics, func(i, j int) bool {
		return strings.ToLower(metrics[i].GetName()) < strings.ToLower(metrics[j].GetName())
	})
	sort.SliceStable(channels, func(i, j int) bool {
		return strings.ToLower(channels[i].GetName()) < strings.ToLower(channels[j].GetName())
	})
	return viewmodel.NewDeviceView(service.app, metrics, channels), err
}

// Returns a historical data set for the requested metric
func (service *IOSwitchDeviceService) History(metric string) ([]float64, error) {
	values, err := service.deviceStore.GetLast30Days(service.deviceID, metric)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// GetDevice combines DeviceState and Config to return a fully populated domain model
// Device instance including child Metric and Channel objects with their current state. This
// operation is more costly than working with the "indexed" maps and functions in FarmState.
func (service *IOSwitchDeviceService) Device() (model.Device, error) {
	deviceState, err := service.stateStore.Get(service.deviceID)
	if err != nil {
		return nil, err
	}
	deviceConfig, err := service.deviceDAO.Get(service.farmID,
		service.deviceID, service.consistency)
	if err != nil {
		return nil, err
	}
	device, err := service.mapper.MapStateToDevice(deviceState, deviceConfig)
	if err != nil {
		return nil, err
	}
	return device, nil
}

// Poll gets the current state from the device and sends it to the FarmService asynchronously
// on the DeviceStateChangeChan channel. The FarmService#WatchDeviceStateChange watches
// for the DeviceStateChange to update the farm state and publish the new state to connected
// websocket clients.
func (service *IOSwitchDeviceService) Poll() error {
	deviceID := service.deviceID
	deviceType := service.device.GetType()
	eventType := "Poll"
	deviceConfig, err := service.deviceDAO.Get(service.farmID,
		service.deviceID, service.consistency)
	if err != nil {
		service.error(eventType, eventType, err)
		return err
	}
	if !deviceConfig.IsEnabled() {
		service.app.Logger.Warningf("%s disabled...", deviceType)
		return nil
	}
	service.app.Logger.Debugf("Polling %s state...", deviceType)
	state, err := service.device.State()
	if err != nil {
		service.error(eventType, eventType, err)
		return err
	}
	state.SetID(deviceID)
	state.SetFarmID(service.farmID)
	service.stateStore.Put(deviceID, state)
	service.farmChannels.DeviceStateChangeChan <- common.DeviceStateChange{
		DeviceID:    deviceID,
		DeviceType:  deviceType,
		StateMap:    state,
		IsPollEvent: true}
	return nil
}

// Toggles a switch to the requested permission, updates the current device state
// and broadcasts the new state to connected websocket clients.
func (service *IOSwitchDeviceService) Switch(channelID, position int, logMessage string) (*common.Switch, error) {
	eventType := "Switch"
	deviceType := service.device.GetType()
	switchPosition := util.NewSwitchPosition(position)
	channelConfig, err := service.ChannelConfig(channelID)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	channelName := channelConfig.GetName()
	if logMessage == "" {
		logMessage = fmt.Sprintf("Switching %s %s", strings.ToLower(channelName),
			switchPosition.ToString())
	}
	service.notify(eventType, logMessage)
	service.app.Logger.Debug(fmt.Sprintf("Switching %s (channel=%d), %s", channelName, channelID,
		switchPosition.ToString()))
	_switch, err := service.device.Switch(channelConfig.GetBoardID(), position)
	if err != nil {
		return _switch, err
	}
	deviceStateMap, err := service.stateStore.Get(service.deviceID)
	if err != nil {
		return nil, err
	}
	channels := deviceStateMap.GetChannels()
	channels[channelID] = position
	deviceStateMap.SetChannels(channels)
	service.stateStore.Put(service.deviceID, deviceStateMap)
	service.farmChannels.DeviceStateChangeChan <- common.DeviceStateChange{
		DeviceID:   service.deviceID,
		DeviceType: deviceType,
		StateMap:   deviceStateMap}
	// service.farmChannels.SwitchChangedChan <- common.SwitchValueChanged{
	// 	DeviceType: "",
	// 	ChannelID:  1,
	// 	Value:      common,
	// }
	service.eventLogService.Create(service.deviceID, deviceType,
		eventType, logMessage)
	return _switch, nil
}

// Toggles a device switch to the ON position for the requested duration (seconds) and
// then sends a 2nd request to the device to toggle the switch to the OFF position after
// the duration has lapsed. Devices should be designed to turn their switches off after
// the specified duration when the TimerSwitch function is called. The 2nd call to turn the
// switch off is only a safety mechanism an should never be relied on to turn a Timerswtich off.
func (service *IOSwitchDeviceService) TimerSwitch(channelID, duration int, logMessage string) (common.TimerEvent, error) {
	eventType := "TimerSwitch"
	deviceID := service.deviceID
	deviceType := service.device.GetType()
	channelConfig, err := service.ChannelConfig(channelID)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	event, err := service.device.TimerSwitch(channelID, duration)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	deviceStateMap, err := service.stateStore.Get(deviceID)
	if err != nil {
		return nil, err
	}
	channels := deviceStateMap.GetChannels()
	channels[channelID] = common.SWITCH_ON
	deviceStateMap.SetChannels(channels)
	service.stateStore.Put(deviceID, deviceStateMap)
	// service.farmChannels.DeviceStateChangeChan <- model.DeviceStateChange{
	// 	DeviceID:   deviceID,
	// 	DeviceType: deviceType,
	// 	StateMap:   deviceStateMap}

	timer := time.NewTimer(time.Second * time.Duration(duration))
	defer timer.Stop()
	go func() {
		<-timer.C
		deviceStateMap, err := service.stateStore.Get(deviceID)
		if err != nil {
			service.app.Logger.Error(err)
			return
		}
		channels := deviceStateMap.GetChannels()
		channels[channelID] = common.SWITCH_OFF
		deviceStateMap.SetChannels(channels)
		service.stateStore.Put(deviceID, deviceStateMap)
		// service.farmChannels.DeviceStateChangeChan <- model.DeviceStateChange{
		// 	DeviceID:   service.deviceID,
		// 	DeviceType: deviceType,
		// 	StateMap:   deviceStateMap}
	}()

	service.app.Logger.Debugf("DeviceService timed switch event: %+v", event)
	if logMessage == "" {
		logMessage = fmt.Sprintf("Starting %s timer for %d seconds",
			channelConfig.GetName(), duration)
	}
	service.notify(eventType, logMessage)
	service.eventLogService.Create(deviceID, deviceType, eventType, logMessage)
	service.app.Logger.Debug(logMessage)
	return event, nil
}

// Updates the device state with a new metric value.
func (service *IOSwitchDeviceService) SetMetricValue(key string, value float64) error {
	deviceState, err := service.stateStore.Get(service.deviceID)
	if err != nil {
		return err
	}
	metrics := deviceState.GetMetrics()
	metrics[key] = value
	deviceState.SetMetrics(metrics)
	if err := service.stateStore.Put(service.deviceID, deviceState); err != nil {
		service.app.Logger.Errorf("Error: %s", err)
		service.error("Farm.poll", "Farm.poll", err)
	}
	if err := service.deviceStore.Save(service.deviceID, deviceState); err != nil {
		service.app.Logger.Errorf("Error: %s", err)
		service.error("Farm.poll", "Farm.poll", err)
	}
	if service.app.Mode == common.CONFIG_MODE_VIRTUAL {
		err := service.device.(*device.VirtualIOSwitch).WriteState(deviceState)
		if err != nil {
			return err
		}
	}
	service.farmChannels.DeviceStateChangeChan <- common.DeviceStateChange{
		DeviceID:   service.deviceID,
		DeviceType: service.device.GetType(),
		StateMap:   deviceState}
	return nil
}

// Returns a device channel configuration
func (service *IOSwitchDeviceService) ChannelConfig(channelID int) (config.Channel, error) {
	deviceConfig, err := service.deviceDAO.Get(service.farmID,
		service.deviceID, service.consistency)
	if err != nil {
		service.app.Logger.Errorf("Error: ", err)
		return nil, err
	}

	channels := deviceConfig.GetChannels()
	for _, channel := range channels {
		if channel.GetBoardID() == channelID {
			return channel, nil
		}
	}
	return nil, fmt.Errorf("channel ID not found: %d", channelID)
}

// Broadcast a real-time farm push notification message to connected clients
func (service *IOSwitchDeviceService) notify(eventType, message string) {
	config, err := service.Config()
	if err != nil {
		service.error("notify", eventType, err)
		return
	}
	if !config.IsNotify() {
		deviceConfig, err := service.deviceDAO.Get(service.farmID,
			service.deviceID, service.consistency)
		if err != nil {
			service.app.Logger.Errorf("Error: ", err)
			return
		}
		service.app.Logger.Warningf("%s notifications disabled!", deviceConfig.GetType())
		return
	}
	service.farmChannels.FarmNotifyChan <- common.FarmNotification{
		EventType: eventType,
		Message:   message}
}

// Broadcast a real-time farm error via push notification to connected clients
func (service *IOSwitchDeviceService) error(method, eventType string, err error) {
	service.app.Logger.Errorf("Error: %s", method, err)
	service.farmChannels.FarmErrorChan <- common.FarmError{
		Method:    method,
		EventType: eventType,
		Error:     err}
}
