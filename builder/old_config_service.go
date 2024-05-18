//go:build ignore
// +build ignore

package service

import (
	"fmt"
	"strconv"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type ConfigService interface {
	GetConfiguration() config.ServerConfig
	SetValue(deviceID int, key, value string) error
	Build() (config.ServerConfig, error)
	BuildCloud() (config.ServerConfig, error)
	BuildOrganization(serverConfig config.ServerConfig, orgID int) (config.ServerConfig, error)
	Reload() error
}

type ConfigurationService struct {
	app             *app.App
	organizationDAO dao.OrganizationDAO
	farmDAO         dao.FarmDAO
	userDAO         dao.UserDAO
	deviceDAO       dao.DeviceDAO
	configDAO       dao.ConfigDAO
	metricDAO       dao.MetricDAO
	channelDAO      dao.ChannelDAO
	conditionDAO    dao.ConditionDAO
	scheduleDAO     dao.ScheduleDAO
	deviceMapper    mapper.DeviceMapper
	conditionMapper mapper.ConditionMapper
	scheduleMapper  mapper.ScheduleMapper
	supportsReload  bool
	ConfigService
}

func NewConfigService(app *app.App, organizationDAO dao.OrganizationDAO, farmDAO dao.FarmDAO, userDAO dao.UserDAO,
	deviceDAO dao.DeviceDAO, configDAO dao.ConfigDAO, metricDAO dao.MetricDAO, channelDAO dao.ChannelDAO,
	conditionDAO dao.ConditionDAO, scheduleDAO dao.ScheduleDAO, deviceMapper mapper.DeviceMapper,
	conditionMapper mapper.ConditionMapper, scheduleMapper mapper.ScheduleMapper, supportsReload bool) ConfigService {

	return &ConfigurationService{
		app:             app,
		organizationDAO: organizationDAO,
		farmDAO:         farmDAO,
		userDAO:         userDAO,
		deviceDAO:       deviceDAO,
		configDAO:       configDAO,
		metricDAO:       metricDAO,
		channelDAO:      channelDAO,
		conditionDAO:    conditionDAO,
		scheduleDAO:     scheduleDAO,
		deviceMapper:    deviceMapper,
		conditionMapper: conditionMapper,
		scheduleMapper:  scheduleMapper,
		supportsReload:  supportsReload}
}

func (service *ConfigurationService) GetConfiguration() config.ServerConfig {
	/*
		bytes, _ := json.Marshal(service.scope.GetConfig())
		service.app.Logger.Debugf("Config json: %s", string(bytes))
		return service.scope.GetConfig()
	*/
	serverConfig, err := service.Build()
	if err != nil {
		service.app.Logger.Error(err)
		return &config.Server{}
	}
	return serverConfig
}

func (service *ConfigurationService) SetValue(deviceID int, key, value string) error {
	service.app.Logger.Debugf("[ConfigurationService.Set] Setting config deviceID=%s, key=%s, value=%s", deviceID, key, value)
	configItem, err := service.configDAO.Get(deviceID, key)
	if err != nil {
		return err
	}
	configItem.SetValue(value)
	service.configDAO.Save(configItem)
	/*
		if err := service.scope.GetState().Notify(deviceID, key, value); err != nil {
			return err
		}*/

	service.app.Logger.Debugf("[ConfigurationService.Set] Saved configuration item: %+v", configItem)
	service.Reload()
	return nil
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

/*
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
}*/

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
		Devices:        make([]config.Device, 0)}

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
					farm.SetName(configEntity.GetValue())
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
