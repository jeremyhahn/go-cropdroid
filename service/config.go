// +build ignore

package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
)

var (
	ErrConfigKeyNotFound = errors.New("Device config key not found")
)

type ConfigurationService struct {
	app               *app.App
	datastoreRegistry datastore.DatastoreRegistry
	//mapperRegistry    mapper.MapperRegistry
	serviceRegistry ServiceRegistry
	//farmServices      map[int]FarmService
	//channelIndex      map[int]config.ChannelConfig
	//deviceIndex   map[int]config.DeviceConfig
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
		//deviceIndex:   make(map[int]config.DeviceConfig, 0),
		mutex:                &sync.RWMutex{},
		farmConfigChangeChan: farmConfigChangeChan}
}

func (service *ConfigurationService) SetValue(session Session, farmID, deviceID uint64, key, value string) error {
	service.app.Logger.Debugf("Setting config farmID=%d, deviceID=%d, key=%s, value=%s",
		farmID, deviceID, key, value)

	configDAO := service.datastoreRegistry.GetDeviceConfigDAO()
	configItem, err := configDAO.Get(deviceID, key)
	if err != nil {
		return err
	}
	configItem.SetValue(value)
	configDAO.Save(configItem)
	service.app.Logger.Debugf("Saved configuration item: %+v", configItem)

	farmService := service.serviceRegistry.GetFarmService(farmID)
	if farmService == nil {
		err := fmt.Errorf("Unable to locate farm service in service registry! farm.id=%d", farmID)
		service.app.Logger.Errorf("Error: %s", err)
		return err
	}
	/*
		devices, err := service.serviceRegistry.GetDeviceServices(farmService.GetFarmID())
		if err != nil {
			return err
		}
		for _, device := range devices {
			deviceConfig := device.GetDeviceConfig()
			if deviceConfig.GetID() == deviceID {
				if device.GetDeviceType() == common.CONTROLLER_TYPE_SERVER {

				}
				for _, c := range deviceConfig.GetConfigs() {
					if c.GetKey() == key {
						c.SetValue(value)
						deviceConfig.SetConfig(&c)
						farmService.PublishConfig()
						return nil
					}
				}
				return ErrConfigKeyNotFound
			}
		}
		return ErrDeviceNotFound
	*/
	farmConfig, err := service.datastoreRegistry.GetFarmDAO().Get(farmService.GetFarmID())
	if err != nil {
		return nil
	}
	service.farmConfigChangeChan <- farmConfig
	return nil
}

/*
func (service *ConfigurationService) OnDeviceConfigChange(deviceConfig config.DeviceConfigConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.app.Logger.Debugf("OnDeviceConfigChange fired! Received device config: %+v", deviceConfig)

	device := service.getDevice(deviceConfig.GetDeviceID())
	if device == nil {
		return
	}
	device.SetConfig(deviceConfig)
	service.farmServices[device.GetFarmID()].GetConfig().ParseConfigs()
	service.farmServices[device.GetFarmID()].PublishConfig()
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
		devices := farmService.GetConfig().GetDevices()
		for i, device := range devices {
			service.deviceIndex[device.GetID()] = &devices[i]
			channels := device.GetChannels()
			for _, channel := range channels {
				service.channelIndex[channel.GetID()] = &channels[i]
			}
		}
	}
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

func (service *ConfigurationService) OnDeviceConfigChange(deviceConfig config.DeviceConfigConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.app.Logger.Debugf("OnDeviceConfigChange fired! Received device config: %+v", deviceConfig)

	device := service.getDevice(deviceConfig.GetDeviceID())
	if device == nil {
		return
	}
	device.SetConfig(deviceConfig)
	service.farmServices[device.GetFarmID()].GetConfig().ParseConfigs()
	service.farmServices[device.GetFarmID()].PublishConfig()
}

func (service *ConfigurationService) OnMetricChange(metric config.MetricConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.app.Logger.Debugf("OnMetricChange fired! Received metric: %+v", metric)

	device, ok := service.deviceIndex[metric.GetDeviceID()]
	if !ok {
		service.deviceIndex[metric.GetDeviceID()] = device
	}
	device.SetMetric(metric)

	service.farmServices[device.GetFarmID()].PublishConfig()
}

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

func (service *ConfigurationService) OnConditionChange(condition config.ConditionConfig) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	channel := service.getChannel(condition.GetChannelID())
	if channel == nil {
		return
	}
	channel.SetCondition(condition)

	device := service.getDevice(channel.GetDeviceID())
	if device == nil {
		return
	}

	service.farmServices[device.GetFarmID()].PublishConfig()
}

func (service *ConfigurationService) OnScheduleChange(schedule config.ScheduleConfig) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	channel := service.getChannel(schedule.GetChannelID())
	if channel == nil {
		return
	}
	channel.SetScheduleItem(schedule)

	device := service.getDevice(channel.GetDeviceID())
	if device == nil {
		return
	}

	service.farmServices[device.GetFarmID()].PublishConfig()
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

	deviceEntities, err := service.deviceDAO.GetByFarmId(farmID)
	if err != nil {
		return nil, err
	}
	if len(deviceEntities) <= 0 {
		return nil, fmt.Errorf("No devices found for farmID %d", farmID)
	}

	farm := &config.Farm{
		ID:             0,
		OrganizationID: 0,
		Devices:    make([]config.Device, 0)}

	for _, device := range deviceEntities {

		service.app.Logger.Debugf("Loading device: %s", device.GetType())

		if device.GetID() == common.CONTROLLER_TYPE_ID_SERVER {
			configEntities, err := service.configDAO.GetAll(device.GetID())
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
			_, err = service.buildSMTP(device.GetID(), serverConfig)
			if err != nil {
				return nil, err
			}
			continue
		}
		service.app.Logger.Debugf("Building microdevice configuration: %s", device.GetType())
		_, err := service.buildDevice(device, farm)
		if err != nil {
			return nil, err
		}
	}
	org := config.CreateOrganization([]config.Farm{*farm}, nil)
	serverConfig.SetOrganizations([]config.Organization{*org})
	return serverConfig, nil
}

func (service *ConfigurationService) buildSMTP(deviceID int, c config.ServerConfig) (config.ServerConfig, error) {
	smtpEnable, err := service.configDAO.Get(deviceID, common.CONFIG_SMTP_ENABLE_KEY)
	if err != nil {
		return nil, err
	}
	smtpHost, err := service.configDAO.Get(deviceID, common.CONFIG_SMTP_HOST_KEY)
	if err != nil {
		return nil, err
	}
	smtpPort, err := service.configDAO.Get(deviceID, common.CONFIG_SMTP_PORT_KEY)
	if err != nil {
		return nil, err
	}
	smtpPortInt, err := strconv.ParseInt(smtpPort.GetValue(), 10, 0)
	if err != nil {
		return nil, err
	}
	smtpUsername, err := service.configDAO.Get(deviceID, common.CONFIG_SMTP_USERNAME_KEY)
	if err != nil {
		return nil, err
	}
	smtpPassword, err := service.configDAO.Get(deviceID, common.CONFIG_SMTP_PASSWORD_KEY)
	if err != nil {
		return nil, err
	}
	smtpTo, err := service.configDAO.Get(deviceID, common.CONFIG_SMTP_RECIPIENT_KEY)
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

func (service *ConfigurationService) buildDevice(device config.Device, farmConfig config.FarmConfig) (config.FarmConfig, error) {
	//configs := make(map[string]string, len(configEntities))
	//for _, entity := range configEntities {
	//	service.app.Logger.Debugf("[ConfigService.buildDevice] Setting config: %+v", entity)
	//	configs[entity.GetKey()] = entity.GetValue()
	//}

	//MapEntityToConfig(deviceEntity entity.DeviceEntity, configEntities []entity.ConfigEntity) (common.DeviceConfig, error)

	//	deviceConfig, err := service.deviceMapper.MapConfigToModel(device, configEntities)
	//	if err != nil {
	//		return nil, err
	//	}

	metrics, err := service.buildMetrics(device)
	if err != nil {
		return nil, err
	}
	channels, err := service.buildChannels(device)
	if err != nil {
		return nil, err
	}
	//device.SetConfigs(configEntities)
	device.SetMetrics(metrics)
	device.SetChannels(channels)
	farmConfig.AddDevice(device)
	return farmConfig, nil
}

func (service *ConfigurationService) buildMetrics(device config.Device) ([]config.Metric, error) {
	metrics, err := service.metricDAO.GetByDeviceID(device.GetID())
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (service *ConfigurationService) buildChannels(device config.Device) ([]config.Channel, error) {
	channels, err := service.channelDAO.GetByDeviceID(device.GetID())
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
