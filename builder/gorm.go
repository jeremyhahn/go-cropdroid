package builder

import (
	"crypto/rsa"
	"fmt"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jeremyhahn/cropdroid/datastore/gorm"
	"github.com/jeremyhahn/cropdroid/datastore/gorm/cockroach"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/jeremyhahn/cropdroid/state"
	"github.com/jeremyhahn/cropdroid/webservice/rest"
)

type GormConfigBuilder struct {
	app *app.App
	ConfigBuilder
}

func NewGormConfigBuilder(_app *app.App) ConfigBuilder {
	return &GormConfigBuilder{app: _app}
}

func (builder *GormConfigBuilder) Build() (config.ServerConfig, service.ServiceRegistry, []rest.RestService,
	state.ControllerIndex, state.ChannelIndex, error) {

	datastoreRegistry := gorm.NewGormRegistry(builder.app.Logger, builder.app.GORM)
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := service.CreateServiceRegistry(builder.app, datastoreRegistry, mapperRegistry)

	controllerIndexMap := make(map[int]config.ControllerConfig, 0)
	channelIndexMap := make(map[int]config.ChannelConfig, 0)

	changefeeders := make(map[string]datastore.Changefeeder, 0)

	if builder.app.DatastoreCDC && builder.app.DatastoreType == "cockroach" {
		// Farm and controller config tables
		changefeeders["_controller_config_items"] = cockroach.NewCockroachChangefeed(builder.app, "controller_config_items")
		changefeeders["_channels"] = cockroach.NewCockroachChangefeed(builder.app, "channels")
		changefeeders["_metrics"] = cockroach.NewCockroachChangefeed(builder.app, "metrics")
		changefeeders["_conditions"] = cockroach.NewCockroachChangefeed(builder.app, "conditions")
		changefeeders["_schedules"] = cockroach.NewCockroachChangefeed(builder.app, "schedules")
	}

	// TODO: Replace with modular backend event log storage
	eventLogDAO := gorm.NewEventLogDAO(builder.app.Logger, builder.app.GORM)
	eventLogService := service.NewEventLogService(builder.app, eventLogDAO, common.CONTROLLER_TYPE_SERVER)
	serviceRegistry.SetEventLogService(eventLogService)

	serverConfig := config.NewServer()
	serverConfig.SetInterval(builder.app.Interval)
	serverConfig.SetTimezone(builder.app.Location.String())
	serverConfig.SetMode(builder.app.Mode)
	//serverConfig.SetSmtp()
	//serverConfig.SetLicense()
	//serverConfig.SetOrganizations()
	/*
	   orgs, err := orgDAO.GetAll()
	   if err != nil {
	     builder.app.Logger.Fatal(err)
	   }
	*/

	//builder.app.Config = serverConfig.(*config.Server)

	//builder.app.Logger.Debugf("builder.app.Config: %+v", builder.app.Config)

	farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)

	farmConfigs, err := farmDAO.GetAll()
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	serverConfig.SetFarms(farmConfigs)

	farmServices := make(map[int]service.FarmService, len(farmConfigs))
	var restServices []rest.RestService

	for _, farmConfig := range farmConfigs {

		//builder.app.Logger.Debugf("Farm: %+v", farmConfig)

		farmID := farmConfig.GetID()

		// Build farm service
		farmDAO := datastoreRegistry.GetFarmDAO()
		controllerConfigDAO := datastoreRegistry.GetControllerConfigDAO()
		farmConfigChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
		farmConfigChangeChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
		farmStateChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
		farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
		//controllerStateChan := make(chan map[string]state.ControllerStateMap, common.BUFFERED_CHANNEL_SIZE)
		controllerStateDeltaChan := make(chan map[string]state.ControllerStateDeltaMap, common.BUFFERED_CHANNEL_SIZE)

		farmService, err := service.CreateFarmService(builder.app, farmDAO, controllerConfigDAO, builder.app.FarmStore,
			builder.app.ConfigStore, &farmConfig, mapperRegistry.GetControllerMapper(), serviceRegistry, farmConfigChan,
			farmConfigChangeChan, farmStateChan /*controllerStateChan,*/, farmStateChangeChan, controllerStateDeltaChan)
		if err != nil {
			builder.app.Logger.Fatal(err)
		}
		farmServices[farmID] = farmService

		// Build farm controller and channel indexes
		controllers := farmConfig.GetControllers()
		for i, controller := range controllers {
			//if controller.GetType() == common.CONTROLLER_TYPE_SERVER {
			//	continue
			//}
			controllerIndexMap[controller.GetID()] = &controllers[i]
			if builder.app.DatastoreCDC && builder.app.DatastoreType == "cockroach" {
				// Controller metric/channel state
				if _, ok := changefeeders[controller.GetType()]; !ok {
					tableName := fmt.Sprintf("state_%d", controller.GetID())
					changefeeders[controller.GetType()] = cockroach.NewCockroachChangefeed(builder.app, tableName)
				}
			}
			channels := controller.GetChannels()
			for _, channel := range channels {
				channelIndexMap[channel.GetID()] = &channels[i]
			}
		}

		// Build controllers
		controllerServices, err := farmService.BuildControllerServices()
		if err != nil {
			builder.app.Logger.Fatal(err)
		}
		serviceRegistry.SetControllerServices(farmID, controllerServices)

		builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
		builder.app.Logger.Debugf("Mode: %s", serverConfig.GetMode())
		builder.app.Logger.Debugf("Timezone: %s", serverConfig.GetTimezone())
		builder.app.Logger.Debugf("Polling interval: %d", serverConfig.GetInterval())

		serviceRegistry.GetEventLogService().Create("System", "Startup")
	}
	serviceRegistry.SetFarmServices(farmServices)

	// Build JWT service
	jsonWriter := rest.NewJsonWriter()
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	jwtService := service.CreateJsonWebTokenService(builder.app, farmDAO, mapperRegistry.GetControllerMapper(), serviceRegistry, jsonWriter, 525960, rsaKeyPair) // 1 year jwt expiration
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	serviceRegistry.SetJsonWebTokenService(jwtService)

	controllerIndex := state.CreateControllerIndex(controllerIndexMap)
	channelIndex := state.CreateChannelIndex(channelIndexMap)

	if builder.app.DatastoreCDC {
		changefeedService := service.NewChangefeedService(builder.app, serviceRegistry, changefeeders)
		serviceRegistry.SetChangefeedService(changefeedService)
	}

	restServiceRegistry := rest.NewFreewareRestServiceRegistry(mapperRegistry, serviceRegistry)
	if restServices == nil {
		restServices = restServiceRegistry.GetRestServices()
	}

	return serverConfig, serviceRegistry, restServices, controllerIndex, channelIndex, err
}
