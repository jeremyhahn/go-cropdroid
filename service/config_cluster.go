// +build cluster

package service

import (
	"fmt"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
)

type ConfigurationService struct {
	app               *app.App
	datastoreRegistry datastore.DatastoreRegistry
	serviceRegistry   ServiceRegistry
	mutex             *sync.RWMutex
	farmServices      map[int]FarmService
	channelIndex      map[int]config.ChannelConfig
	//deviceIndex   map[int]config.DeviceConfig
	ConfigService
}

func NewConfigService(app *app.App, datastoreRegistry datastore.DatastoreRegistry,
	serviceRegistry ServiceRegistry) ConfigService {

	return &ConfigurationService{
		app:               app,
		datastoreRegistry: datastoreRegistry,
		serviceRegistry:   serviceRegistry,
		mutex:             &sync.RWMutex{},
		farmServices:      make(map[int]FarmService, 0),
		channelIndex:      make(map[int]config.ChannelConfig, 0)}
	//deviceIndex:   make(map[int]config.DeviceConfig, 0)}
}

func (service *ConfigurationService) SetValue(session Session, farmID, deviceID uint64, key, value string) error {
	//func (service *ConfigurationService) SetValue(farmID, deviceID int, key, value string) error {

	service.app.Logger.Debugf("[ConfigurationService.Set] Setting config farmID=%d, deviceID=%d, key=%s, value=%s",
		farmID, deviceID, key, value)

	/*
		configDAO := service.datastoreRegistry.GetDeviceConfigDAO()
		configItem, err := configDAO.Get(deviceID, key)
		if err != nil {
			return err
		}
		configItem.SetValue(value)
		configDAO.Save(configItem)
		service.app.Logger.Debugf("[ConfigurationService.Set] Saved configuration item: %+v", configItem)
	*/

	farmService := service.serviceRegistry.GetFarmService(farmID)
	if farmService == nil {
		err := fmt.Errorf("Unable to locate farm service in service registry! farm.id=%d", farmID)
		service.app.Logger.Errorf("[ConfigurationService.Set] Error: %s", err)
		return err
	}
	farmConfig := farmService.GetConfig()
	configSet := false
	for _, device := range farmConfig.GetDevices() {
		if device.GetID() == deviceID {
			for k, _ := range device.ConfigMap {
				if k == key {
					device.ConfigMap[k] = value
					configSet = true
					break
				}
			}
		}
		if configSet {
			break // TODO clean this up with a function return
		} else {
			service.app.Logger.Errorf("unable to set config for device %d", deviceID)
		}
	}
	farmConfig.HydrateConfigs()
	farmService.SetConfig(farmConfig)

	return nil
}

func (service *ConfigurationService) GetServerConfig() config.ServerConfig {
	return service.app.Config
}

/*

func (service *ConfigurationService) OnChannelChange(channel config.ChannelConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.app.Logger.Debugf("OnChannelChange fired! Received channel: %+v", channel)

	device := service.getDevice(channel.GetDeviceID())
	if device == nil {
		return
	}
	device.SetChannel(channel)

	service.farmServices[device.GetFarmID()].PublishConfig()
}

func (service *ConfigurationService) getDevice(id int) config.DeviceConfig {
	device, ok := service.deviceIndex[id]
	if !ok {
		service.Sync()
		device, ok = service.deviceIndex[id]
		if !ok {
			service.app.Logger.Errorf("[ConfigurationService.getDevice] Failed to locate device after syncing indexes! device.id=%d",
				device.GetID())
			return nil
		}
	}
	return device
}
*/
