package builder

import (
	"crypto/rsa"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/cockroach"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/store"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"
)

type GormConfigBuilder struct {
	app              *app.App
	farmStateStore   state.FarmStorer
	deviceStateStore state.DeviceStorer
	deviceDataStore  datastore.DeviceDataStore
	consistencyLevel int
	appStateTTL      int
	appStateTick     int
	idGenerator      util.IdGenerator
}

func NewGormConfigBuilder(_app *app.App, dataStore string, appStateTTL int, appStateTick int) *GormConfigBuilder {

	farmStateStore := state.NewMemoryFarmStore(_app.Logger, 1, appStateTTL, time.Duration(appStateTick))
	deviceStateStore := state.NewMemoryDeviceStore(_app.Logger, 3, appStateTTL, time.Duration(appStateTick))

	var deviceDatastore datastore.DeviceDataStore
	if dataStore == "redis" {
		deviceDatastore = datastore.NewRedisDataStore(":6379", "")
	} else {
		deviceDatastore = store.NewGormDataStore(_app.Logger, _app.GORM,
			_app.GORMInitParams.Engine, _app.Location)
	}

	return &GormConfigBuilder{
		app:              _app,
		farmStateStore:   farmStateStore,
		deviceStateStore: deviceStateStore,
		deviceDataStore:  deviceDatastore,
		consistencyLevel: common.CONSISTENCY_CACHED,
		idGenerator:      util.NewIdGenerator(_app.DataStoreEngine)}
}

func (builder *GormConfigBuilder) Build() (app.KeyPair,
	service.ServiceRegistry, []rest.RestService, chan uint64, error) {

	var restServices []rest.RestService
	var farmConfigs []config.FarmConfig

	datastoreRegistry := gorm.NewGormRegistry(builder.app.Logger, builder.app.GormDB)
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := service.CreateServiceRegistry(builder.app, datastoreRegistry, mapperRegistry)

	changefeeders := make(map[string]datastore.Changefeeder, 0)

	if builder.app.Config.DataStoreCDC && builder.app.Config.DataStoreEngine == "cockroach" {
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

	// serverConfig := config.NewServer()
	// serverConfig.SetInterval(builder.app.Interval)
	// serverConfig.SetTimezone(builder.app.Location.String())
	// serverConfig.SetMode(builder.app.Mode)
	//serverConfig.SetSmtp()
	//serverConfig.SetLicense()

	orgDAO := gorm.NewOrganizationDAO(builder.app.Logger, builder.app.GORM)
	farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)
	roleDAO := gorm.NewRoleDAO(builder.app.Logger, builder.app.GORM)

	farmProvisionerChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
	farmDeprovisionerChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
	farmTickerProvisionerChan := make(chan uint64, common.BUFFERED_CHANNEL_SIZE)
	gormInitializer := gorm.NewGormInitializer(builder.app.Logger,
		builder.app.GormDB, builder.idGenerator, builder.app.Location, builder.app.Config.Mode)
	gormFarmConfigStore := store.NewGormFarmConfigStore(datastoreRegistry.NewFarmDAO(), 1)
	gormDeviceConfigStore := store.NewGormDeviceConfigStore(datastoreRegistry.NewDeviceDAO(), 3)
	farmFactory := service.NewFarmFactory(
		builder.app, datastoreRegistry, serviceRegistry, builder.farmStateStore,
		gormFarmConfigStore, builder.deviceStateStore, gormDeviceConfigStore,
		builder.deviceDataStore, mapperRegistry.GetDeviceMapper(),
		changefeeders, farmProvisionerChan, farmTickerProvisionerChan, builder.idGenerator)
	serviceRegistry.SetFarmFactory(farmFactory)

	farmProvisioner := provisioner.NewGormFarmProvisioner(
		builder.app.Logger, builder.app.GORM, builder.app.Location,
		datastoreRegistry.NewFarmDAO(), farmProvisionerChan,
		farmDeprovisionerChan, mapperRegistry.GetUserMapper(),
		gormInitializer)
	serviceRegistry.SetFarmProvisioner(farmProvisioner)

	orgs, err := datastoreRegistry.GetOrganizationDAO().GetAll()
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	//serverConfig.SetOrganizations(orgs)
	//builder.app.Config.SetOrganizations(orgs)

	if len(orgs) > 0 {
		for _, org := range orgs {
			farmConfigs = append(farmConfigs, org.GetFarms()...)
		}
	} else {
		farmConfigs, err = farmDAO.GetAll()
	}
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	//serverConfig.SetFarms(farmConfigs)
	//builder.app.Config.SetFarms(farmConfigs)

	for _, farmConfig := range farmConfigs {
		farmFactory.BuildService(farmConfig)
		//builder.app.Config..AddFarm(&farmConfig)
	}

	// Listen for new farm provisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-farmProvisionerChan:
				builder.app.Logger.Debugf("Processing provisioner request...")
				farmService, err := farmFactory.BuildService(farmConfig)
				if err != nil {
					builder.app.Logger.Errorf("Error: %s", err)
					break
				}
				serviceRegistry.AddFarmService(farmService)
				farmTickerProvisionerChan <- farmConfig.GetID()
				go farmService.Run()
			}
		}
	}()

	// Listen for new farm deprovisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-farmDeprovisionerChan:
				farmID := farmConfig.GetID()
				builder.app.Logger.Debugf("Processing deprovisioning request for farm %d", farmID)
				farmService := serviceRegistry.GetFarmService(farmID)
				if farmService == nil {
					builder.app.Logger.Errorf("Farm not found: %d", farmID)
					continue
				}
				farmService.Stop()
				if err := farmDAO.Delete(farmConfig); err != nil {
					builder.app.Logger.Error(err)
				}
				farmTickerProvisionerChan <- farmID
			}
		}
	}()

	organizationService := service.NewOrganizationService(
		builder.app.Logger, builder.idGenerator,
		datastoreRegistry.GetOrganizationDAO())
	serviceRegistry.SetOrganizationService(organizationService)

	// Build JWT service
	jsonWriter := rest.NewJsonWriter()
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	//farmConfigStore := store.NewGormFarmConfigStore(farmDAO, 1)
	defaultRole, err := roleDAO.GetByName(builder.app.Config.DefaultRole)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	jwtService := service.CreateJsonWebTokenService(builder.app,
		builder.idGenerator, orgDAO, farmDAO, defaultRole,
		mapperRegistry.GetDeviceMapper(), serviceRegistry, jsonWriter,
		525960, rsaKeyPair) // 1 year jwt expiration
	//jwtService := service.CreateJsonWebTokenService(builder.app, farmDAO, mapperRegistry.GetDeviceMapper(), serviceRegistry, jsonWriter, 525960, rsaKeyPair) // 1 year jwt expiration
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	serviceRegistry.SetJsonWebTokenService(jwtService)

	// if builder.app.DataStoreCDC {
	// 	changefeedService := service.NewChangefeedService(builder.app, serviceRegistry,
	// 		changefeeders)
	// 	serviceRegistry.SetChangefeedService(changefeedService)
	// }

	publicKey := string(rsaKeyPair.GetPublicBytes())
	restServiceRegistry := rest.NewRestServiceRegistry(publicKey, mapperRegistry, serviceRegistry)
	if restServices == nil {
		restServices = restServiceRegistry.GetRestServices()
	}

	return rsaKeyPair, serviceRegistry, restServices, farmTickerProvisionerChan, err
}
