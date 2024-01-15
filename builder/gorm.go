package builder

import (
	"crypto/rsa"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	//"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	gormds "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/cockroach"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/store"
	"github.com/jeremyhahn/go-cropdroid/datastore/redis"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"

	"github.com/jinzhu/gorm"
)

type GormConfigBuilder struct {
	app              *app.App
	db               *gorm.DB
	gormDB           gormds.GormDB
	farmStateStore   state.FarmStorer
	deviceStateStore state.DeviceStorer
	deviceDataStore  datastore.DeviceDataStore
	consistencyLevel int
	appStateTTL      int
	appStateTick     int
	idGenerator      util.IdGenerator
}

func NewGormConfigBuilder(_app *app.App, dataStore string, appStateTTL int, appStateTick int) *GormConfigBuilder {

	gormDB := gormds.NewGormDB(_app.Logger, _app.GORMInitParams)
	db := gormDB.Connect(false)

	farmStateStore := state.NewMemoryFarmStore(_app.Logger, 1, appStateTTL, time.Duration(appStateTick))
	deviceStateStore := state.NewMemoryDeviceStore(_app.Logger, 3, appStateTTL, time.Duration(appStateTick))

	var deviceDatastore datastore.DeviceDataStore
	if dataStore == "redis" {
		deviceDatastore = redis.NewRedisDataStore(":6379", "")
	} else {
		deviceDatastore = store.NewGormDataStore(_app.Logger, db,
			_app.GORMInitParams.Engine, _app.Location)
	}

	return &GormConfigBuilder{
		app:              _app,
		db:               db,
		gormDB:           gormDB,
		farmStateStore:   farmStateStore,
		deviceStateStore: deviceStateStore,
		deviceDataStore:  deviceDatastore,
		consistencyLevel: common.CONSISTENCY_LOCAL,
		idGenerator:      util.NewIdGenerator(_app.DataStoreEngine)}
}

func (builder *GormConfigBuilder) Build() (app.KeyPair,
	service.ServiceRegistry, []rest.RestService, chan uint64, error) {

	var restServices []rest.RestService
	var farmConfigs []*config.Farm

	datastoreRegistry := gormds.NewGormRegistry(builder.app.Logger, builder.gormDB)
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := service.CreateServiceRegistry(builder.app, datastoreRegistry, mapperRegistry)

	changefeeders := make(map[string]datastore.Changefeeder, 0)

	if builder.app.DataStoreCDC && builder.app.DataStoreEngine == "cockroach" {
		// Farm and device config tables
		changefeeders["_device_config_items"] = cockroach.NewCockroachChangefeed(builder.app, "device_config_items")
		changefeeders["_channels"] = cockroach.NewCockroachChangefeed(builder.app, "channels")
		changefeeders["_metrics"] = cockroach.NewCockroachChangefeed(builder.app, "metrics")
		changefeeders["_conditions"] = cockroach.NewCockroachChangefeed(builder.app, "conditions")
		changefeeders["_schedules"] = cockroach.NewCockroachChangefeed(builder.app, "schedules")
	}

	// TODO: Replace with modular backend event log storage
	eventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.db)
	eventLogService := service.NewEventLogService(builder.app, eventLogDAO, common.CONTROLLER_TYPE_SERVER)
	serviceRegistry.SetEventLogService(eventLogService)

	// serverConfig := config.NewServer()
	// serverConfig.SetInterval(builder.app.Interval)
	// serverConfig.SetTimezone(builder.app.Location.String())
	// serverConfig.SetMode(builder.app.Mode)
	//serverConfig.SetSmtp()
	//serverConfig.SetLicense()

	orgDAO := gormds.NewOrganizationDAO(builder.app.Logger, builder.db, builder.idGenerator)
	farmDAO := gormds.NewFarmDAO(builder.app.Logger, builder.db, builder.idGenerator)
	roleDAO := gormds.NewRoleDAO(builder.app.Logger, builder.db)
	deviceDAO := gormds.NewDeviceDAO(builder.app.Logger, builder.db)
	deviceConfigDAO := gormds.NewDeviceSettingDAO(builder.app.Logger, builder.db)

	farmProvisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
	farmDeprovisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
	farmTickerProvisionerChan := make(chan uint64, common.BUFFERED_CHANNEL_SIZE)
	gormInitializer := gormds.NewGormInitializer(builder.app.Logger,
		builder.gormDB, builder.idGenerator, builder.app.Location, builder.app.Mode)

	//gormFarmStore := store.NewGormFarmStore(datastoreRegistry.NewFarmDAO(), 1)
	//gormDeviceConfigStore := store.NewGormDeviceConfigStore(datastoreRegistry.NewDeviceDAO(), 3)

	farmChannels := &service.FarmChannels{
		FarmConfigChan:       make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
		FarmConfigChangeChan: make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChan:        make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChangeChan:  make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmErrorChan:        make(chan common.FarmError, common.BUFFERED_CHANNEL_SIZE),
		FarmNotifyChan:       make(chan common.FarmNotification, common.BUFFERED_CHANNEL_SIZE),
		//MetricChangedChan:     make(chan common.MetricValueChanged, common.BUFFERED_CHANNEL_SIZE),
		//SwitchChangedChan:     make(chan common.SwitchValueChanged, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateChangeChan: make(chan common.DeviceStateChange, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateDeltaChan:  make(chan map[string]state.DeviceStateDeltaMap, common.BUFFERED_CHANNEL_SIZE)}

	farmFactory := service.NewFarmFactory(
		builder.app, farmDAO, deviceDAO, deviceConfigDAO, serviceRegistry,
		mapperRegistry.GetDeviceMapper(), changefeeders, farmProvisionerChan,
		farmTickerProvisionerChan, farmChannels, builder.idGenerator)
	serviceRegistry.SetFarmFactory(farmFactory)

	farmProvisioner := provisioner.NewGormFarmProvisioner(
		builder.app.Logger, builder.db, builder.app.Location,
		datastoreRegistry.NewFarmDAO(), farmProvisionerChan,
		farmDeprovisionerChan, mapperRegistry.GetUserMapper(),
		gormInitializer)
	serviceRegistry.SetFarmProvisioner(farmProvisioner)

	orgs, err := orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	//serverConfig.SetOrganizations(orgs)
	//builder.app.Server.SetOrganizations(orgs)

	if len(orgs) > 0 {
		for _, org := range orgs {
			farmConfigs = append(farmConfigs, org.GetFarms()...)
		}
	} else {
		farmConfigs, err = farmDAO.GetAll(common.CONSISTENCY_LOCAL)
	}
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	//serverConfig.SetFarms(farmConfigs)
	//builder.app.Server.SetFarms(farmConfigs)

	for _, farmConfig := range farmConfigs {

		// gormDB := gormds.NewGormDB(builder.app.Logger, builder.app.GORMInitParams)
		// db := gormDB.Connect(false)
		// deviceDAO := gormds.NewDeviceDAO(builder.app.Logger, db)
		// farmDAO := gormds.NewFarmDAO(builder.app.Logger, db)

		farmFactory.BuildService(builder.farmStateStore, farmDAO,
			builder.deviceDataStore, builder.deviceStateStore,
			farmConfig)
		//builder.app.Server..AddFarm(&farmConfig)
	}

	// Listen for new farm provisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-farmProvisionerChan:
				builder.app.Logger.Debugf("Processing provisioner request...")
				farmService, err := farmFactory.BuildService(
					builder.farmStateStore, farmDAO, builder.deviceDataStore,
					builder.deviceStateStore, &farmConfig)
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
				if err := farmDAO.Delete(&farmConfig); err != nil {
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
	//farmConfigStore := store.NewGormFarmStore(farmDAO, 1)
	defaultRole, err := roleDAO.GetByName(builder.app.DefaultRole, common.CONSISTENCY_LOCAL)
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
	restServiceRegistry := rest.NewRestServiceRegistry(builder.app,
		publicKey, mapperRegistry, serviceRegistry)
	if restServices == nil {
		restServices = restServiceRegistry.GetRestServices()
	}

	return rsaKeyPair, serviceRegistry, restServices, farmTickerProvisionerChan, err
}
