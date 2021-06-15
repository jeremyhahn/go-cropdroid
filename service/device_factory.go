package service

import (
	"strings"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
)

// "Global" device service used to manage all devices used by the platform
type DeviceFactory interface {
	BuildServices(deviceConfigs []config.Device, datastore datastore.DeviceDatastore,
		mode string) ([]DeviceService, error)
	BuildService(datastore datastore.DeviceDatastore,
		deviceConfig config.DeviceConfig, mode string) (DeviceService, error)
	GetAll(session Session) ([]config.Device, error)
	GetDevices(session Session) ([]common.Device, error)
}

type SmartSwitchFactory struct {
	app              *app.App
	farmID           uint64
	farmName         string
	stateStore       state.DeviceStorer
	configStore      store.DeviceConfigStorer
	consistency      int
	deviceDAO    dao.DeviceDAO
	deviceMapper mapper.DeviceMapper
	serviceRegistry  ServiceRegistry
	farmChannels     *FarmChannels
	DeviceFactory
}

func NewDeviceFactory(app *app.App, farmID uint64, farmName string, deviceDAO dao.DeviceDAO,
	stateStore state.DeviceStorer, configStore store.DeviceConfigStorer, consistency int,
	deviceMapper mapper.DeviceMapper, serviceRegistry ServiceRegistry,
	farmChannels *FarmChannels) DeviceFactory {

	return &SmartSwitchFactory{
		app:              app,
		farmID:           farmID,
		farmName:         farmName,
		stateStore:       stateStore,
		configStore:      configStore,
		consistency:      consistency,
		deviceDAO:    deviceDAO,
		deviceMapper: deviceMapper,
		serviceRegistry:  serviceRegistry,
		farmChannels:     farmChannels}
}

// Builds all device services for a given farm
func (factory *SmartSwitchFactory) BuildServices(deviceConfigs []config.Device,
	datastore datastore.DeviceDatastore, mode string) ([]DeviceService, error) {

	services := make([]DeviceService, 0, len(deviceConfigs))
	for _, deviceConfig := range deviceConfigs {
		if deviceConfig.GetType() == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		service, err := factory.BuildService(datastore, &deviceConfig, mode)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	factory.serviceRegistry.SetDeviceServices(factory.farmID, services)
	return services, nil
}

// Builds a new device service
func (factory *SmartSwitchFactory) BuildService(datastore datastore.DeviceDatastore,
	deviceConfig config.DeviceConfig, mode string) (DeviceService, error) {

	var _device device.SmartSwitcher
	deviceID := deviceConfig.GetID()
	deviceType := deviceConfig.GetType()

	factory.app.Logger.Debugf("Building %s service", deviceType)

	if !deviceConfig.IsEnabled() {
		factory.app.Logger.Warningf("%s service disabled...", strings.Title(deviceType))
	}

	if mode == common.CONFIG_MODE_VIRTUAL {
		farmStateMap := state.NewFarmStateMap(factory.farmID)
		_device = device.NewVirtualSmartSwitch(factory.app, farmStateMap, "", deviceType)
	} else {
		_device = device.NewSmartSwitch(factory.app, deviceConfig.GetURI(), deviceType)
	}

	service, err := NewDeviceService(factory.app, deviceID,
		factory.farmName, factory.stateStore, factory.configStore, datastore,
		factory.deviceMapper, _device, factory.serviceRegistry.GetEventLogService(),
		factory.farmChannels, factory.consistency)

	if err != nil {
		factory.app.Logger.Error(err.Error())
		return nil, ErrCreateService
	}

	factory.configStore.Cache(deviceID, deviceConfig)

	return service, nil
}

// Returns all device configs for a given farm session
func (factory *SmartSwitchFactory) GetAll(session Session) ([]config.Device, error) {
	deviceEntities, err := factory.deviceDAO.GetByFarmId(
		session.GetFarmService().GetConfig().GetID())
	if err != nil {
		return nil, err
	}
	return deviceEntities, nil
}

// Returns all common.device objects for a given farm session
func (factory *SmartSwitchFactory) GetDevices(session Session) ([]common.Device, error) {
	var devices []common.Device
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	deviceEntities, err := factory.deviceDAO.GetByFarmId(farmConfig.GetID())
	if err != nil {
		return nil, err
	}
	//devices := make([]common.Device, len(deviceEntities)-1) // -1 for server device
	for _, entity := range deviceEntities {
		if entity.GetType() == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		deviceState, err := farmService.GetState().GetDevice(entity.GetType())
		if err != nil {
			return nil, err
		}
		deviceConfig, err := farmConfig.GetDevice(entity.GetType())
		if err != nil {
			return nil, err
		}
		device, err := factory.deviceMapper.MapStateToDevice(deviceState, *deviceConfig)
		if err != nil {
			return nil, err
		}
		//devices[i] = device
		devices = append(devices, device) // dynamically add - not sure how many server devices there will be
	}
	return devices, nil
}
