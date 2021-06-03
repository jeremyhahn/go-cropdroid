// +build cluster

package service

import (
	"fmt"
	"sync"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/datastore"
)

type ConfigurationService struct {
	app               *app.App
	datastoreRegistry datastore.DatastoreRegistry
	serviceRegistry   ServiceRegistry
	mutex             *sync.RWMutex
	farmServices      map[int]FarmService
	channelIndex      map[int]config.ChannelConfig
	controllerIndex   map[int]config.ControllerConfig
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
		channelIndex:      make(map[int]config.ChannelConfig, 0),
		controllerIndex:   make(map[int]config.ControllerConfig, 0)}
}

func (service *ConfigurationService) SetValue(session Session, farmID, controllerID int, key, value string) error {
	//func (service *ConfigurationService) SetValue(farmID, controllerID int, key, value string) error {

	service.app.Logger.Debugf("[ConfigurationService.Set] Setting config farmID=%d, controllerID=%d, key=%s, value=%s",
		farmID, controllerID, key, value)

	/*
		configDAO := service.datastoreRegistry.GetControllerConfigDAO()
		configItem, err := configDAO.Get(controllerID, key)
		if err != nil {
			return err
		}
		configItem.SetValue(value)
		configDAO.Save(configItem)
		service.app.Logger.Debugf("[ConfigurationService.Set] Saved configuration item: %+v", configItem)
	*/

	farmService, ok := service.serviceRegistry.GetFarmService(farmID)
	if !ok {
		err := fmt.Errorf("Unable to locate farm service in service registry! farm.id=%d", farmID)
		service.app.Logger.Errorf("[ConfigurationService.Set] Error: %s", err)
		return err
	}
	farmConfig := farmService.GetConfig()
	configSet := false
	for _, controller := range farmConfig.GetControllers() {
		if controller.GetID() == controllerID {
			for k, _ := range controller.ConfigMap {
				if k == key {
					controller.ConfigMap[k] = value
					configSet = true
					break
				}
			}
		}
		if configSet {
			break // TODO clean this up with a function return
		} else {
			service.app.Logger.Errorf("unable to set config for controller %d", controllerID)
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

	controller := service.getController(channel.GetControllerID())
	if controller == nil {
		return
	}
	controller.SetChannel(channel)

	service.farmServices[controller.GetFarmID()].PublishConfig()
}

func (service *ConfigurationService) getController(id int) config.ControllerConfig {
	controller, ok := service.controllerIndex[id]
	if !ok {
		service.Sync()
		controller, ok = service.controllerIndex[id]
		if !ok {
			service.app.Logger.Errorf("[ConfigurationService.getController] Failed to locate controller after syncing indexes! controller.id=%d",
				controller.GetID())
			return nil
		}
	}
	return controller
}
*/
