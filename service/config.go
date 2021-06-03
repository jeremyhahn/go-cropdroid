// +build !cluster

package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/datastore"
)

var (
	ErrConfigKeyNotFound = errors.New("Controller config key not found")
)

type ConfigurationService struct {
	app               *app.App
	datastoreRegistry datastore.DatastoreRegistry
	//mapperRegistry    mapper.MapperRegistry
	serviceRegistry ServiceRegistry
	//farmServices      map[int]FarmService
	//channelIndex      map[int]config.ChannelConfig
	//controllerIndex   map[int]config.ControllerConfig
	mutex                *sync.RWMutex
	farmConfigChangeChan chan config.FarmConfig
	ConfigService
}

func NewConfigService(app *app.App, datastoreRegistry datastore.DatastoreRegistry,
	serviceRegistry ServiceRegistry, farmConfigChangeChan chan config.FarmConfig) ConfigService {

	return &ConfigurationService{
		app:               app,
		datastoreRegistry: datastoreRegistry,
		//mapperRegistry:    mapperRegistry,
		serviceRegistry: serviceRegistry,
		//farmServices:      make(map[int]FarmService, 0),
		//channelIndex:      make(map[int]config.ChannelConfig, 0),
		//controllerIndex:   make(map[int]config.ControllerConfig, 0),
		mutex:                &sync.RWMutex{},
		farmConfigChangeChan: farmConfigChangeChan}
}

func (service *ConfigurationService) SetValue(session Session, farmID, controllerID int, key, value string) error {
	service.app.Logger.Debugf("[ConfigurationService.Set] Setting config farmID=%d, controllerID=%d, key=%s, value=%s",
		farmID, controllerID, key, value)

	configDAO := service.datastoreRegistry.GetControllerConfigDAO()
	configItem, err := configDAO.Get(controllerID, key)
	if err != nil {
		return err
	}
	configItem.SetValue(value)
	configDAO.Save(configItem)
	service.app.Logger.Debugf("[ConfigurationService.Set] Saved configuration item: %+v", configItem)

	farmService, ok := service.serviceRegistry.GetFarmService(farmID)
	if !ok {
		err := fmt.Errorf("Unable to locate farm service in service registry! farm.id=%d", farmID)
		service.app.Logger.Errorf("[ConfigurationService.Set] Error: %s", err)
		return err
	}
	/*
		controllers, err := service.serviceRegistry.GetControllerServices(farmService.GetFarmID())
		if err != nil {
			return err
		}
		for _, controller := range controllers {
			controllerConfig := controller.GetControllerConfig()
			if controllerConfig.GetID() == controllerID {
				if controller.GetControllerType() == common.CONTROLLER_TYPE_SERVER {

				}
				for _, c := range controllerConfig.GetConfigs() {
					if c.GetKey() == key {
						c.SetValue(value)
						controllerConfig.SetConfig(&c)
						farmService.PublishConfig()
						return nil
					}
				}
				return ErrConfigKeyNotFound
			}
		}
		return ErrControllerNotFound
	*/
	farmConfig, err := service.datastoreRegistry.GetFarmDAO().Get(farmService.GetFarmID())
	if err != nil {
		return nil
	}
	service.farmConfigChangeChan <- farmConfig
	return nil
}

/*
func (service *ConfigurationService) OnControllerConfigChange(controllerConfig config.ControllerConfigConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.app.Logger.Debugf("OnControllerConfigChange fired! Received controller config: %+v", controllerConfig)

	controller := service.getController(controllerConfig.GetControllerID())
	if controller == nil {
		return
	}
	controller.SetConfig(controllerConfig)
	service.farmServices[controller.GetFarmID()].GetConfig().ParseConfigs()
	service.farmServices[controller.GetFarmID()].PublishConfig()
}
*/

func (service *ConfigurationService) GetServerConfig() config.ServerConfig {
	return service.app.Config
}

/*
// Sync the indexes with current farm configs
func (service *ConfigurationService) Sync() {
	service.app.Logger.Debug("[ConfigurationService.Sync] Syncing configuration")
	service.farmServices = service.serviceRegistry.GetFarmServices()
	for _, farmService := range service.farmServices {
		controllers := farmService.GetConfig().GetControllers()
		for i, controller := range controllers {
			service.controllerIndex[controller.GetID()] = &controllers[i]
			channels := controller.GetChannels()
			for _, channel := range channels {
				service.channelIndex[channel.GetID()] = &channels[i]
			}
		}
	}
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

func (service *ConfigurationService) getChannel(id int) config.ChannelConfig {
	channel, ok := service.channelIndex[id]
	if !ok {
		service.Sync()
		channel, ok = service.channelIndex[id]
		if !ok {
			service.app.Logger.Errorf("[ConfigurationService.getChannel] Failed to locate channel after syncing indexes! channel.id=%d", id)
			return nil
		}
	}
	return channel
}

func (service *ConfigurationService) OnControllerConfigChange(controllerConfig config.ControllerConfigConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.app.Logger.Debugf("OnControllerConfigChange fired! Received controller config: %+v", controllerConfig)

	controller := service.getController(controllerConfig.GetControllerID())
	if controller == nil {
		return
	}
	controller.SetConfig(controllerConfig)
	service.farmServices[controller.GetFarmID()].GetConfig().ParseConfigs()
	service.farmServices[controller.GetFarmID()].PublishConfig()
}

func (service *ConfigurationService) OnMetricChange(metric config.MetricConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.app.Logger.Debugf("OnMetricChange fired! Received metric: %+v", metric)

	controller, ok := service.controllerIndex[metric.GetControllerID()]
	if !ok {
		service.controllerIndex[metric.GetControllerID()] = controller
	}
	controller.SetMetric(metric)

	service.farmServices[controller.GetFarmID()].PublishConfig()
}

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

func (service *ConfigurationService) OnConditionChange(condition config.ConditionConfig) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	channel := service.getChannel(condition.GetChannelID())
	if channel == nil {
		return
	}
	channel.SetCondition(condition)

	controller := service.getController(channel.GetControllerID())
	if controller == nil {
		return
	}

	service.farmServices[controller.GetFarmID()].PublishConfig()
}

func (service *ConfigurationService) OnScheduleChange(schedule config.ScheduleConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	channel := service.getChannel(schedule.GetChannelID())
	if channel == nil {
		return
	}
	channel.SetScheduleItem(schedule)

	controller := service.getController(channel.GetControllerID())
	if controller == nil {
		return
	}

	service.farmServices[controller.GetFarmID()].PublishConfig()
}

func (service *ConfigurationService) GetServerConfig() config.ServerConfig {
	return service.app.Config
}

func (service *ConfigurationService) Reload() error {
	if !service.supportsReload {
		service.app.Logger.Warning("Skipping configuration service reload! service.supportsReload=false")
		return nil
	}

	newConfig, err := service.Build()
	if err != nil {
		return err
	}

	service.app.Config = newConfig.(*config.Server)

	service.app.Logger.Debug("Configuration service reloaded! service.supportsReload=true")
	return nil
}

func (service *ConfigurationService) Build() (config.ServerConfig, error) {
	farmID := 1
	serverConfig := config.NewServer()
	serverConfig.SetID(service.app.ServerID)
	serverConfig.SetMode(service.app.Mode)
	serverConfig.SetTimezone(service.app.Location.String())
	return service.BuildOrganization(serverConfig, farmID)
}

func (service *ConfigurationService) BuildCloud() (config.ServerConfig, error) {
	serverConfig := config.NewServer()
	serverConfig.SetMode(service.app.Mode)
	serverConfig.SetTimezone(service.app.Location.String())
	serverConfig.SetSmtp(&config.Smtp{})
	orgs, err := service.organizationDAO.GetAll()
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		farmConfig, err := service.farmDAO.First()
		if err != nil {
			return nil, err
		}
		org := config.CreateOrganization([]config.Farm{*farmConfig.(*config.Farm)}, nil)
		serverConfig.SetID(farmConfig.GetID())
		serverConfig.SetOrganizations([]config.Organization{*org})
	} else {
		serverConfig.SetID(service.app.ServerID)
		serverConfig.SetOrganizations(orgs) //return service.BuildOrganization(serverConfig, orgID)
	}
	return serverConfig, nil
}

func (service *ConfigurationService) BuildOrganization(serverConfig config.ServerConfig, orgID int) (config.ServerConfig, error) {
	service.app.Logger.Debugf("Building config for orgID: %d", orgID)
	orgs, err := service.organizationDAO.Find(orgID)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		service.app.Logger.Debugf("org length: 0")
		return serverConfig, nil
	}
	serverConfig.SetOrganizations(orgs)
	return serverConfig, nil
}

func (service *ConfigurationService) BuildOrganization(serverConfig config.ServerConfig, farmID int) (config.ServerConfig, error) {

	service.app.Logger.Debugf("Building config for farmID: %d", farmID)

	controllerEntities, err := service.controllerDAO.GetByFarmId(farmID)
	if err != nil {
		return nil, err
	}
	if len(controllerEntities) <= 0 {
		return nil, fmt.Errorf("No controllers found for farmID %d", farmID)
	}

	farm := &config.Farm{
		ID:             0,
		OrganizationID: 0,
		Controllers:    make([]config.Controller, 0)}

	for _, controller := range controllerEntities {

		service.app.Logger.Debugf("Loading controller: %s", controller.GetType())

		if controller.GetID() == common.CONTROLLER_TYPE_ID_SERVER {
			configEntities, err := service.configDAO.GetAll(controller.GetID())
			if err != nil {
				return nil, err
			}
			for _, configEntity := range configEntities {
				service.app.Logger.Debugf("[ConfigService.BuildOrganization] Setting config: %+v", configEntity)
				name := configEntity.GetKey()
				switch name {
				case "name":
					service.SetName(configEntity.GetValue())
				case "interval":
					interval, err := strconv.Atoi(configEntity.GetValue())
					if err != nil {
						service.app.Logger.Fatal(err)
					}
					serverConfig.SetInterval(int(interval))
				case "timezone":
					serverConfig.SetTimezone(configEntity.GetValue())
				case "mode":
					serverConfig.SetMode(configEntity.GetValue())
				}
			}
			_, err = service.buildSMTP(controller.GetID(), serverConfig)
			if err != nil {
				return nil, err
			}
			continue
		}
		service.app.Logger.Debugf("Building microcontroller configuration: %s", controller.GetType())
		_, err := service.buildController(controller, farm)
		if err != nil {
			return nil, err
		}
	}
	org := config.CreateOrganization([]config.Farm{*farm}, nil)
	serverConfig.SetOrganizations([]config.Organization{*org})
	return serverConfig, nil
}

func (service *ConfigurationService) buildSMTP(controllerID int, c config.ServerConfig) (config.ServerConfig, error) {
	smtpEnable, err := service.configDAO.Get(controllerID, common.CONFIG_SMTP_ENABLE_KEY)
	if err != nil {
		return nil, err
	}
	smtpHost, err := service.configDAO.Get(controllerID, common.CONFIG_SMTP_HOST_KEY)
	if err != nil {
		return nil, err
	}
	smtpPort, err := service.configDAO.Get(controllerID, common.CONFIG_SMTP_PORT_KEY)
	if err != nil {
		return nil, err
	}
	smtpPortInt, err := strconv.ParseInt(smtpPort.GetValue(), 10, 0)
	if err != nil {
		return nil, err
	}
	smtpUsername, err := service.configDAO.Get(controllerID, common.CONFIG_SMTP_USERNAME_KEY)
	if err != nil {
		return nil, err
	}
	smtpPassword, err := service.configDAO.Get(controllerID, common.CONFIG_SMTP_PASSWORD_KEY)
	if err != nil {
		return nil, err
	}
	smtpTo, err := service.configDAO.Get(controllerID, common.CONFIG_SMTP_RECIPIENT_KEY)
	if err != nil {
		return nil, err
	}
	bEnable, err := strconv.ParseBool(smtpEnable.GetValue())
	if err != nil {
		return nil, err
	}
	smtpConfig := &config.Smtp{
		Enable:    bEnable,
		Host:      smtpHost.GetValue(),
		Port:      int(smtpPortInt),
		Username:  smtpUsername.GetValue(),
		Password:  smtpPassword.GetValue(),
		Recipient: smtpTo.GetValue()}
	c.SetSmtp(smtpConfig)
	return c, nil
}

//func (service *ConfigurationService) buildFarm(farm entity.FarmEntity, config config.ServerConfig) (config.ServerConfig, error) {
//}

func (service *ConfigurationService) buildController(controller config.Controller, farmConfig config.FarmConfig) (config.FarmConfig, error) {
	//configs := make(map[string]string, len(configEntities))
	//for _, entity := range configEntities {
	//	service.app.Logger.Debugf("[ConfigService.buildController] Setting config: %+v", entity)
	//	configs[entity.GetKey()] = entity.GetValue()
	//}

	//MapEntityToConfig(controllerEntity entity.ControllerEntity, configEntities []entity.ConfigEntity) (common.ControllerConfig, error)

	//	controllerConfig, err := service.controllerMapper.MapConfigToModel(controller, configEntities)
	//	if err != nil {
	//		return nil, err
	//	}

	metrics, err := service.buildMetrics(controller)
	if err != nil {
		return nil, err
	}
	channels, err := service.buildChannels(controller)
	if err != nil {
		return nil, err
	}
	//controller.SetConfigs(configEntities)
	controller.SetMetrics(metrics)
	controller.SetChannels(channels)
	farmConfig.AddController(controller)
	return farmConfig, nil
}

func (service *ConfigurationService) buildMetrics(controller config.Controller) ([]config.Metric, error) {
	metrics, err := service.metricDAO.GetByControllerID(controller.GetID())
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (service *ConfigurationService) buildChannels(controller config.Controller) ([]config.Channel, error) {
	channels, err := service.channelDAO.GetByControllerID(controller.GetID())
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		service.app.Logger.Debugf("[ConfigService.buildChannels] channel=%+v", channel)
		schedule, err := service.buildSchedule(channel)
		if err != nil {
			return nil, err
		}
		conditions, err := service.buildConditions(channel)
		if err != nil {
			return nil, err
		}
		channel.SetSchedule(schedule)
		channel.SetConditions(conditions)
	}
	return channels, nil
}

func (service *ConfigurationService) buildSchedule(channel config.Channel) ([]config.Schedule, error) {
	schedules, err := service.scheduleDAO.GetByChannelID(channel.GetID())
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (service *ConfigurationService) buildConditions(channel config.Channel) ([]config.Condition, error) {
	conditions, err := service.conditionDAO.GetByChannelID(channel.GetID())
	if err != nil {
		return nil, err
	}
	return conditions, nil
}
*/
