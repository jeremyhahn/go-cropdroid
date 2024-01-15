package service

import (
	"strings"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
)

// "Global" device service used to manage all devices used by the platform
type DeviceFactory interface {
	BuildServices(deviceConfigs []*config.Device, datastore datastore.DeviceDataStore,
		mode string) ([]DeviceService, error)
	BuildService(datastore datastore.DeviceDataStore,
		deviceConfig *config.Device, mode string) (DeviceService, error)
	GetAll(session Session) ([]*config.Device, error)
	GetDevices(session Session) ([]common.Device, error)
}

type DefaultDeviceFactory struct {
	app             *app.App
	farmID          uint64
	farmName        string
	deviceDAO       dao.DeviceDAO
	stateStore      state.DeviceStorer
	consistency     int
	deviceMapper    mapper.DeviceMapper
	serviceRegistry ServiceRegistry
	farmChannels    *FarmChannels
	DeviceFactory
}

func NewDeviceFactory(app *app.App, farmID uint64, farmName string,
	deviceDAO dao.DeviceDAO, configStoreType int, consistency int,
	stateStore state.DeviceStorer, deviceMapper mapper.DeviceMapper,
	serviceRegistry ServiceRegistry, farmChannels *FarmChannels) DeviceFactory {

	return &DefaultDeviceFactory{
		app:             app,
		farmID:          farmID,
		farmName:        farmName,
		deviceDAO:       deviceDAO,
		stateStore:      stateStore,
		consistency:     consistency,
		deviceMapper:    deviceMapper,
		serviceRegistry: serviceRegistry,
		farmChannels:    farmChannels}
}

// Builds all device services for a given farm
func (factory *DefaultDeviceFactory) BuildServices(deviceConfigs []*config.Device,
	datastore datastore.DeviceDataStore, mode string) ([]DeviceService, error) {

	services := make([]DeviceService, 0, len(deviceConfigs))
	for _, deviceConfig := range deviceConfigs {
		if deviceConfig.GetType() == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		service, err := factory.BuildService(datastore, deviceConfig, mode)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	factory.serviceRegistry.SetDeviceServices(factory.farmID, services)
	return services, nil
}

// Builds a new device service
func (factory *DefaultDeviceFactory) BuildService(datastore datastore.DeviceDataStore,
	deviceConfig *config.Device, mode string) (DeviceService, error) {

	// gormDB := factory.app.GormDB.CloneConnection()
	// deviceDAO := gorm.NewDeviceDAO(factory.app.Logger, gormDB)

	var _device device.IOSwitcher
	deviceID := deviceConfig.GetID()
	deviceType := deviceConfig.GetType()

	factory.app.Logger.Debugf("Building %s service", deviceType)

	if !deviceConfig.IsEnabled() {
		factory.app.Logger.Warningf("%s service disabled...", strings.Title(deviceType))
	}

	if mode == common.CONFIG_MODE_VIRTUAL {
		farmStateMap := state.NewFarmStateMap(factory.farmID)
		_device = device.NewVirtualIOSwitch(factory.app, farmStateMap, "", deviceType)
	} else {
		_device = device.NewSmartSwitch(factory.app, deviceConfig.GetURI(), deviceType)
	}

	// -- Disabling as part of new raft package work --
	// // HACK: This populates the Raft config store from GORM dao
	// //       so NewDeviceService can look it up
	// factory.deviceDAO.Save(deviceConfig)

	service, err := NewDeviceService(factory.app, factory.farmID, deviceID,
		factory.farmName, factory.stateStore, factory.deviceDAO, datastore,
		factory.deviceMapper, _device, factory.serviceRegistry.GetEventLogService(),
		factory.farmChannels, factory.consistency)

	if err != nil {
		factory.app.Logger.Error(err.Error())
		return nil, ErrCreateService
	}

	return service, nil
}

// Should the following 2 methods be refactored into the device service?

// Returns all device configs for a given farm session
func (factory *DefaultDeviceFactory) GetAll(session Session) ([]*config.Device, error) {
	return session.GetFarmService().GetConfig().GetDevices(), nil
}

// Returns all common.device objects for a given farm session
func (factory *DefaultDeviceFactory) GetDevices(session Session) ([]common.Device, error) {
	var devices []common.Device
	farmService := session.GetFarmService()
	deviceConfigs := farmService.GetConfig().GetDevices()
	for _, deviceConfig := range deviceConfigs {
		if deviceConfig.GetType() == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		deviceState, err := factory.stateStore.Get(deviceConfig.GetID())
		if err != nil {
			return nil, err
		}
		device, err := factory.deviceMapper.MapStateToDevice(deviceState, deviceConfig)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, nil
}
