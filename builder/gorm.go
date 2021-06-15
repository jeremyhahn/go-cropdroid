package builder

import (
	"crypto/rsa"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/cockroach"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/store"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"
)

type GormConfigBuilder struct {
	app              *app.App
	farmStateStore   state.FarmStorer
	deviceStateStore state.DeviceStorer
	deviceDataStore  datastore.DeviceDatastore
	consistency      int
	ConfigBuilder
}

func NewGormConfigBuilder(_app *app.App, farmStateStore state.FarmStorer,
	deviceStateStore state.DeviceStorer,
	deviceStore string) ConfigBuilder {

	var deviceDatastore datastore.DeviceDatastore
	if deviceStore == "redis" {
		deviceDatastore = datastore.NewRedisDeviceStore(":6379", "")
	} else {
		deviceDatastore = store.NewGormDeviceStore(_app.Logger, _app.GORM,
			_app.GORMInitParams.Engine, _app.Location)
	}

	return &GormConfigBuilder{
		app:              _app,
		farmStateStore:   farmStateStore,
		deviceStateStore: deviceStateStore,
		deviceDataStore:  deviceDatastore,
		consistency:      common.CONSISTENCY_CACHED}
}

func (builder *GormConfigBuilder) Build() (config.ServerConfig,
	service.ServiceRegistry, []rest.RestService, error) {

	var restServices []rest.RestService

	datastoreRegistry := gorm.NewGormRegistry(builder.app.Logger, builder.app.GORM)
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := service.CreateServiceRegistry(builder.app, datastoreRegistry, mapperRegistry)

	changefeeders := make(map[string]datastore.Changefeeder, 0)

	if builder.app.DatastoreCDC && builder.app.DatastoreType == "cockroach" {
		// Farm and device config tables
		changefeeders["_device_config_items"] = cockroach.NewCockroachChangefeed(builder.app, "device_config_items")
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

	farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)

	farmConfigs, err := farmDAO.GetAll()
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	serverConfig.SetFarms(farmConfigs)

	for _, farmConfig := range farmConfigs {

		// Each farm gets its own DAO / session to the DB
		deviceDAO := datastoreRegistry.GetDeviceDAO()

		gormFarmConfigStore := store.NewGormFarmConfigStore(farmDAO, 1)
		gormDeviceConfigStore := store.NewGormDeviceConfigStore(deviceDAO, 3)

		_, err := service.NewFarmFactory(
			builder.app, datastoreRegistry, serviceRegistry, builder.farmStateStore,
			gormFarmConfigStore, builder.deviceStateStore, gormDeviceConfigStore,
			builder.deviceDataStore, mapperRegistry.GetDeviceMapper(),
			changefeeders, nil, nil).BuildService(&farmConfig)

		if err != nil {
			builder.app.Logger.Fatal(err)
		}
	}

	// Build JWT service
	jsonWriter := rest.NewJsonWriter()
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	jwtService := service.CreateJsonWebTokenService(builder.app, farmDAO, mapperRegistry.GetDeviceMapper(), serviceRegistry, jsonWriter, 525960, rsaKeyPair) // 1 year jwt expiration
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	serviceRegistry.SetJsonWebTokenService(jwtService)

	if builder.app.DatastoreCDC {
		changefeedService := service.NewChangefeedService(builder.app, serviceRegistry,
			changefeeders)
		serviceRegistry.SetChangefeedService(changefeedService)
	}

	restServiceRegistry := rest.NewFreewareRestServiceRegistry(mapperRegistry, serviceRegistry)
	if restServices == nil {
		restServices = restServiceRegistry.GetRestServices()
	}

	//return serverConfig, serviceRegistry, restServices, deviceIndex, channelIndex, err
	return serverConfig, serviceRegistry, restServices, err
}
