package builder

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	//"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	gormds "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/cockroach"
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
	app               *app.App
	db                *gorm.DB
	gormDB            gormds.GormDB
	datastoreRegistry dao.Registry
	farmStateStore    state.FarmStorer
	deviceStateStore  state.DeviceStorer
	deviceDataStore   datastore.DeviceDataStore
	consistencyLevel  int
	appStateTTL       int
	appStateTick      int
	databaseInit      bool
	idGenerator       util.IdGenerator
}

func NewGormConfigBuilder(_app *app.App, dataStore string,
	appStateTTL int, appStateTick int, databaseInit bool) *GormConfigBuilder {

	gormDB := gormds.NewGormDB(_app.Logger, _app.GORMInitParams)
	db := gormDB.Connect(false)

	farmStateStore := state.NewMemoryFarmStore(_app.Logger, 1, appStateTTL,
		time.Duration(appStateTick))

	deviceStateStore := state.NewMemoryDeviceStore(_app.Logger, 3, appStateTTL,
		time.Duration(appStateTick))

	var deviceDatastore datastore.DeviceDataStore
	if dataStore == "redis" {
		deviceDatastore = redis.NewRedisDataStore(":6379", "")
	} else {
		deviceDatastore = gormds.NewGormDeviceDataStore(_app.Logger, db,
			_app.GORMInitParams.Engine, _app.Location)
	}

	return &GormConfigBuilder{
		app:              _app,
		db:               db,
		gormDB:           gormDB,
		farmStateStore:   farmStateStore,
		deviceStateStore: deviceStateStore,
		deviceDataStore:  deviceDatastore,
		databaseInit:     databaseInit,
		consistencyLevel: common.CONSISTENCY_LOCAL,
		idGenerator:      util.NewIdGenerator(_app.DataStoreEngine)}
}

func (builder *GormConfigBuilder) Build() (app.KeyPair,
	service.ServiceRegistry, []rest.RestService, chan uint64, error) {

	var restServices []rest.RestService
	var farmConfigs []*config.Farm

	builder.datastoreRegistry = gormds.NewGormRegistry(builder.app.Logger, builder.gormDB)
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := service.CreateServiceRegistry(builder.app,
		builder.datastoreRegistry, mapperRegistry)

	changefeeders := make(map[string]datastore.Changefeeder, 0)

	if builder.app.DataStoreCDC && builder.app.DataStoreEngine == "cockroach" {
		// Farm and device config tables
		changefeeders["_device_config_items"] = cockroach.NewCockroachChangefeed(builder.app, "device_config_items")
		changefeeders["_channels"] = cockroach.NewCockroachChangefeed(builder.app, "channels")
		changefeeders["_metrics"] = cockroach.NewCockroachChangefeed(builder.app, "metrics")
		changefeeders["_conditions"] = cockroach.NewCockroachChangefeed(builder.app, "conditions")
		changefeeders["_schedules"] = cockroach.NewCockroachChangefeed(builder.app, "schedules")
	}

	eventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.db, 0)
	eventLogService := service.NewEventLogService(builder.app, eventLogDAO, 0)
	serviceRegistry.AddEventLogService(eventLogService)

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

	configInitializer := dao.NewConfigInitializer(builder.app.Logger,
		builder.app.IdGenerator, builder.app.Location, builder.datastoreRegistry,
		builder.app.Mode)

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
		builder.app.Logger, builder.db, builder.app.Location, builder.datastoreRegistry.NewFarmDAO(),
		builder.datastoreRegistry.GetPermissionDAO(), farmProvisionerChan, farmDeprovisionerChan,
		mapperRegistry.GetUserMapper(), configInitializer)
	serviceRegistry.SetFarmProvisioner(farmProvisioner)

	// Initialize the database with a default farm
	if builder.databaseInit {
		builder.initDatabase()
	}

	// Load all the organizations
	orgs, err := orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	if err != nil {
		if err.Error() == "no such table: organizations" {
			// Assume this is a first start and the database
			// needs to be initialized
			builder.initDatabase()
			orgs, err = orgDAO.GetAll(common.CONSISTENCY_LOCAL)
			if err != nil {
				builder.app.Logger.Fatal(err)
			}
		} else {
			builder.app.Logger.Fatal(err)
		}
	}

	// Load all the farms
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

	// Build a FarmService for each farm in the database
	for _, farmConfig := range farmConfigs {

		farmID := farmConfig.GetID()

		farmEventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.db, int(farmID))

		farmService, err := farmFactory.BuildService(builder.farmStateStore, farmDAO, farmEventLogDAO,
			builder.deviceDataStore, builder.deviceStateStore,
			farmConfig)
		if err != nil {
			builder.app.Logger.Errorf("Error: %s", err)
			continue
		}
		farmService.RefreshHardwareVersions()

		builder.app.Logger.Debugf("Farm ID: %s", farmID)
		builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
		builder.app.Logger.Debugf("Mode: %s", builder.app.Mode)
		builder.app.Logger.Debugf("Timezone: %s", builder.app.Timezone)
		builder.app.Logger.Debugf("Polling interval: %d", builder.app.Interval)

		eventLogService.Create(0, common.CONTROLLER_TYPE_SERVER, "System", "Startup")
	}

	// Listen for new farm provisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-farmProvisionerChan:
				builder.app.Logger.Debugf("Processing provisioner request...")

				farmKey := fmt.Sprintf("%d-%s", farmConfig.OrganizationID, farmConfig.Name)
				farmID := builder.idGenerator.NewID32(farmKey)

				farmEventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.db, farmID)
				farmService, err := farmFactory.BuildService(
					builder.farmStateStore, farmDAO, farmEventLogDAO, builder.deviceDataStore,
					builder.deviceStateStore, &farmConfig)
				if err != nil {
					builder.app.Logger.Errorf("Error: %s", err)
					break
				}
				farmService.InitializeState(true)
				farmService.RefreshHardwareVersions()
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
		builder.datastoreRegistry.GetOrganizationDAO())
	serviceRegistry.SetOrganizationService(organizationService)

	// Build JWT service
	jsonWriter := rest.NewJsonWriter()
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	defaultRole, err := roleDAO.GetByName(builder.app.DefaultRole, common.CONSISTENCY_LOCAL)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	jwtService := service.CreateJsonWebTokenService(builder.app,
		builder.idGenerator, orgDAO, farmDAO, defaultRole,
		mapperRegistry.GetDeviceMapper(), serviceRegistry, jsonWriter,
		525960, rsaKeyPair) // 1 year jwt expiration
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

func (builder *GormConfigBuilder) initDatabase() {

	builder.app.Logger.Info("Initializing database...")

	builder.gormDB.Create()
	builder.gormDB.Migrate()

	provParams := &common.ProvisionerParams{
		UserID:           0,
		RoleID:           0,
		OrganizationID:   0,
		FarmName:         common.DEFAULT_CROP_NAME,
		ConfigStoreType:  builder.app.DefaultConfigStoreType,
		StateStoreType:   builder.app.DefaultStateStoreType,
		DataStoreType:    builder.app.DefaultDataStoreType,
		ConsistencyLevel: common.CONSISTENCY_LOCAL}

	configInitializer := dao.NewConfigInitializer(builder.app.Logger,
		builder.app.IdGenerator, builder.app.Location,
		builder.datastoreRegistry, builder.app.Mode)

	_, err := configInitializer.Initialize(builder.app.EnableDefaultFarm, provParams)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	// if err := farmDAO.Save(farmConfig); err != nil {
	// 	builder.app.Logger.Fatal(err)
	// }
}
