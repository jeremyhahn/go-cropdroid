package service

import (
	"strings"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/device"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
)

// Device factory / service used to manage all devices across all farms
type DeviceFactory interface {
	BuildServices(
		deviceConfigs []*config.DeviceStruct,
		datastore datastore.DeviceDataStore,
		mode string) ([]DeviceServicer, error)
	BuildService(
		datastore datastore.DeviceDataStore,
		deviceConfig *config.DeviceStruct,
		mode string) (DeviceServicer, error)
}

type DefaultDeviceFactory struct {
	app               *app.App
	farmID            uint64
	farmName          string
	datastoreRegistry dao.Registry
	eventLogDAO       dao.EventLogDAO
	stateStore        state.DeviceStateStorer
	consistency       int
	deviceMapper      mapper.DeviceMapper
	serviceRegistry   ServiceRegistry
	farmChannels      *FarmChannels
	DeviceFactory
}

func NewDeviceFactory(
	app *app.App,
	farmID uint64,
	farmName string,
	datastoreRegistry dao.Registry,
	eventLogDAO dao.EventLogDAO,
	configStoreType, consistency int,
	stateStore state.DeviceStateStorer,
	deviceMapper mapper.DeviceMapper,
	serviceRegistry ServiceRegistry,
	farmChannels *FarmChannels) DeviceFactory {

	return &DefaultDeviceFactory{
		app:               app,
		farmID:            farmID,
		farmName:          farmName,
		datastoreRegistry: datastoreRegistry,
		eventLogDAO:       eventLogDAO,
		stateStore:        stateStore,
		consistency:       consistency,
		deviceMapper:      deviceMapper,
		serviceRegistry:   serviceRegistry,
		farmChannels:      farmChannels}
}

// Builds all device services for a given farm
func (factory *DefaultDeviceFactory) BuildServices(
	deviceConfigs []*config.DeviceStruct,
	datastore datastore.DeviceDataStore,
	mode string) ([]DeviceServicer, error) {

	services := make([]DeviceServicer, 0, len(deviceConfigs))
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
func (factory *DefaultDeviceFactory) BuildService(
	datastore datastore.DeviceDataStore,
	deviceConfig *config.DeviceStruct,
	mode string) (DeviceServicer, error) {

	var _device device.IOSwitcher
	deviceID := deviceConfig.ID
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

	service, err := NewDeviceService(factory.app, factory.farmID, deviceID,
		factory.farmName, factory.stateStore, factory.datastoreRegistry.NewDeviceDAO(),
		factory.eventLogDAO, datastore, factory.deviceMapper, _device,
		factory.farmChannels, factory.consistency)

	if err != nil {
		factory.app.Logger.Error(err.Error())
		return nil, ErrCreateService
	}

	return service, nil
}
