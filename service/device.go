package service

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/store"

	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type SmartSwitcherDeviceService struct {
	app             *app.App
	deviceID        uint64
	farmName        string
	consistency     int
	stateStore      state.DeviceStorer
	configStore     store.DeviceConfigStorer
	deviceStore     datastore.DeviceDatastore
	device          device.SmartSwitcher
	deviceMutex     *sync.RWMutex
	mapper          mapper.DeviceMapper
	eventLogService EventLogService
	farmChannels    *FarmChannels
	DeviceService
}

func NewDeviceService(app *app.App, deviceID uint64, farmName string,
	stateStore state.DeviceStorer, configStore store.DeviceConfigStorer,
	deviceDatastore datastore.DeviceDatastore, deviceMapper mapper.DeviceMapper,
	device device.SmartSwitcher, eventLogService EventLogService,
	farmChannels *FarmChannels, consistency int) (DeviceService, error) {

	return &SmartSwitcherDeviceService{
		app:             app,
		deviceID:        deviceID,
		farmName:        farmName,
		stateStore:      stateStore,
		configStore:     configStore,
		deviceStore:     deviceDatastore,
		mapper:          deviceMapper,
		device:          device,
		deviceMutex:     &sync.RWMutex{},
		eventLogService: eventLogService,
		farmChannels:    farmChannels,
		consistency:     consistency}, nil
}

func (service *SmartSwitcherDeviceService) GetDeviceType() string {
	return service.device.GetType()
}

func (service *SmartSwitcherDeviceService) GetConfig() (config.DeviceConfig, error) {
	return service.configStore.Get(service.deviceID, service.consistency)
}

func (service *SmartSwitcherDeviceService) GetState() (state.DeviceStateMap, error) {
	return service.stateStore.Get(service.deviceID)
}

func (service *SmartSwitcherDeviceService) SetMode(mode string, d device.SmartSwitcher) {
	//var unsafeDevice = (*unsafe.Pointer)(unsafe.Pointer(&service.device))
	//atomic.StorePointer(unsafeDevice, unsafe.Pointer(&d))
	service.deviceMutex.Lock()
	defer service.deviceMutex.Unlock()
	service.device = d
}

func (service *SmartSwitcherDeviceService) GetView() (common.DeviceView, error) {
	device, err := service.GetDevice()
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

func (service *SmartSwitcherDeviceService) GetHistory(metric string) ([]float64, error) {
	values, err := service.deviceStore.GetLast30Days(service.deviceID, metric)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// GetDevice combines DeviceState and Config to return a fully populated domain model
// Device instance including child Metric and Channel objects with their current values. This
// operation is more costly than working with the "indexed" maps and functions in FarmState.
func (service *SmartSwitcherDeviceService) GetDevice() (common.Device, error) {
	deviceState, err := service.stateStore.Get(service.deviceID)
	if err != nil {
		return nil, err
	}
	deviceConfig, err := service.configStore.Get(service.deviceID, service.consistency)
	if err != nil {
		return nil, err
	}
	device, err := service.mapper.MapStateToDevice(deviceState, *deviceConfig.(*config.Device))
	if err != nil {
		return nil, err
	}
	return device, nil
}

// Poll gets the current state from the device and sends it to the FarmService asynchronously
// on the deviceStateChangeChan channel. The FarmService#WatchDeviceStateChange watches
// for the DeviceStateChange to update the farm state and publish the new state to connected
// websocket clients.
func (service *SmartSwitcherDeviceService) Poll(deviceStateChangeChan chan<- common.DeviceStateChange) error {
	deviceID := service.deviceID
	deviceType := service.device.GetType()
	eventType := "Poll"
	service.app.Logger.Debugf("Polling %s state...", deviceType)
	state, err := service.device.State()
	if err != nil {
		service.error(eventType, eventType, err)
		return err
	}
	state.SetID(deviceID)
	service.stateStore.Put(service.deviceID, state)
	deviceStateChangeChan <- common.DeviceStateChange{
		DeviceID:   deviceID,
		DeviceType: deviceType,
		StateMap:   state}
	return nil
}

// Sends a command to the device to turn a switch on or off
func (service *SmartSwitcherDeviceService) Switch(channelID, position int, logMessage string) (*common.Switch, error) {
	eventType := "Switch"
	switchPosition := util.NewSwitchPosition(position)
	channelConfig, err := service.getChannelConfig(channelID)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	if logMessage == "" {
		logMessage = fmt.Sprintf("Switching %s %s", strings.ToLower(channelConfig.GetName()), switchPosition.ToString())
	}
	service.notify(eventType, logMessage)
	service.eventLogService.Create(eventType, logMessage)
	service.app.Logger.Debug(fmt.Sprintf("Switching device %s (channel=%d), %s", channelConfig.GetName(), channelID, switchPosition.ToString()))
	_switch, err := service.device.Switch(channelConfig.GetChannelID(), position)
	if err != nil {
		return _switch, err
	}
	service.farmChannels.SwitchChangedChan <- common.SwitchValueChanged{
		DeviceType: service.device.GetType(),
		ChannelID:  channelID,
		Value:      position}
	return _switch, nil
}

// Sends a command to the device to activate a timed switch. The switch will be on for duration
// specified and then turned off.
func (service *SmartSwitcherDeviceService) TimerSwitch(channelID, duration int, logMessage string) (common.TimerEvent, error) {
	eventType := "TimerSwitch"
	channelConfig, err := service.getChannelConfig(channelID)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	event, err := service.device.TimerSwitch(channelID, duration)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	service.app.Logger.Debugf("DeviceService timed switch event: %+v", event)
	if logMessage == "" {
		logMessage = fmt.Sprintf("Starting %s timer for %d seconds", channelConfig.GetName(), duration)
	}
	service.notify(eventType, logMessage)
	service.eventLogService.Create(eventType, logMessage)
	service.app.Logger.Debug(logMessage)
	return event, nil
}

// Sets a metric value.
func (service *SmartSwitcherDeviceService) SetMetricValue(key string, value float64) error {

	/*
		deviceState, err := service.GetState()
		if err != nil {
			return err
		}
		service.farmService.SetDeviceState(deviceType, deviceState)
	*/
	/*
		// This works with raft cluster
		if service.app.MetricDatastore != nil {
			if err := service.app.MetricDatastore.Save(deviceType, service.config.GetID(), service.GetState()); err != nil {
				service.app.Logger.Errorf("Error: %s", err)
				service.error("Farm.poll", "Farm.poll", err)
				return
			}
		}*/

	deviceState, err := service.stateStore.Get(service.deviceID)
	if err != nil {
		return err
	}

	deviceConfig, err := service.configStore.Get(service.deviceID, service.consistency)
	if err != nil {
		return err
	}

	deviceType := deviceConfig.GetType()

	if service.app.Mode == common.CONFIG_MODE_VIRTUAL || service.app.Mode == common.MODE_STANDALONE { // TODO: consolidate mode
		//virtualDevice := service.device.NewVirtualDevice(server.app, farmState, "", service.config.GetType())
		err := service.device.(*device.VirtualSmartSwitch).WriteState(deviceState)
		if err != nil {
			return err
		}
	} else {
		if err := service.deviceStore.Save(service.deviceID, deviceState); err != nil {
			service.app.Logger.Errorf("Error: %s", err)
			service.error("Farm.poll", "Farm.poll", err)
		}
	}

	service.farmChannels.MetricChangedChan <- common.MetricValueChanged{
		DeviceType: deviceType,
		Key:        key,
		Value:      value}

	return nil
}

func (service *SmartSwitcherDeviceService) getChannelConfig(channelID int) (config.ChannelConfig, error) {
	deviceConfig, err := service.configStore.Get(service.deviceID, service.consistency)
	if err != nil {
		service.app.Logger.Errorf("Error: ", err)
		return nil, err
	}

	channels := deviceConfig.GetChannels()
	for _, channel := range channels {
		if channel.GetChannelID() == channelID {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("Channel ID not found: %d", channelID)
}

func (service *SmartSwitcherDeviceService) notify(eventType, message string) {
	config, err := service.GetConfig()
	if err != nil {
		service.error("notify", eventType, err)
		return
	}
	if !config.IsNotify() {
		deviceConfig, err := service.configStore.Get(service.deviceID, service.consistency)
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

func (service *SmartSwitcherDeviceService) error(method, eventType string, err error) {
	service.app.Logger.Errorf("Error: %s", method, err)
	service.farmChannels.FarmErrorChan <- common.FarmError{
		Method:    method,
		EventType: eventType,
		Error:     err}
}
