package builder

import (
	"crypto/rsa"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"

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

	"gorm.io/gorm"
)

type GormConfigBuilder struct {
	app               *app.App
	db                *gorm.DB
	gormDB            gormds.GormDB
	datastoreRegistry dao.Registry
	serviceRegistry   service.ServiceRegistry
	farmStateStore    state.FarmStorer
	deviceStateStore  state.DeviceStorer
	deviceDataStore   datastore.DeviceDataStore
	consistencyLevel  int
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

	builder.datastoreRegistry = gormds.NewGormRegistry(builder.app.Logger, builder.gormDB)
	mapperRegistry := mapper.CreateRegistry()
	builder.serviceRegistry = service.CreateServiceRegistry(builder.app,
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
	builder.serviceRegistry.AddEventLogService(eventLogService)

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

	passwordHasher := util.CreatePasswordHasher(builder.app.PasswordHasherParams)

	configInitializer := dao.NewConfigInitializer(builder.app.Logger,
		builder.app.IdGenerator, builder.app.Location, builder.datastoreRegistry,
		passwordHasher, builder.app.Mode)

	farmFactory := service.NewFarmFactory(
		builder.app, farmDAO, deviceDAO, deviceConfigDAO, builder.serviceRegistry,
		mapperRegistry.GetDeviceMapper(), changefeeders, farmProvisionerChan,
		farmTickerProvisionerChan, builder.idGenerator)
	builder.serviceRegistry.SetFarmFactory(farmFactory)

	farmProvisioner := provisioner.NewGormFarmProvisioner(
		builder.app.Logger, builder.db, builder.app.Location, builder.datastoreRegistry.NewFarmDAO(),
		builder.datastoreRegistry.GetPermissionDAO(), farmProvisionerChan, farmDeprovisionerChan,
		mapperRegistry.GetUserMapper(), configInitializer)
	builder.serviceRegistry.SetFarmProvisioner(farmProvisioner)

	// Initialize the database with a default farm
	if builder.databaseInit {
		builder.initDatabase()
	}

	// Load all farms that belong to an organization
	err := orgDAO.ForEachPage(query.NewPageQuery(), func(entities []*config.Organization) error {
		for _, org := range entities {
			for _, farmConfig := range org.GetFarms() {
				builder.app.Logger.Infof("Creating organization clustered farm service. farm.ID: %d, farm.Name: %s, org.Name: %s",
					farmConfig.ID, farmConfig.Name, org.Name)
				builder.createAndRunFarm(farmDAO, farmFactory, farmConfig)
			}
		}
		return nil
	}, common.CONSISTENCY_LOCAL)
	if err != nil {
		builder.app.Logger.Error(err)
	}

	// Load all standalone farms that do NOT belong to an organization
	err = farmDAO.ForEachPage(query.NewPageQuery(), func(entities []*config.Farm) error {
		for _, farmConfig := range entities {
			builder.app.Logger.Infof("Creating independent clustered farm service: ID: %d, Name: ",
				farmConfig.ID, farmConfig.Name)
			builder.createAndRunFarm(farmDAO, farmFactory, farmConfig)
		}
		return nil
	}, common.CONSISTENCY_LOCAL)
	if err != nil {
		builder.app.Logger.Error(err)
	}

	// // Load all the organizations
	// orgs, err := orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	// if err != nil {
	// 	if err.Error() == "no such table: organizations" {
	// 		// Assume this is a first start and the database
	// 		// needs to be initialized
	// 		builder.initDatabase()
	// 		orgs, err = orgDAO.GetAll(common.CONSISTENCY_LOCAL)
	// 		if err != nil {
	// 			builder.app.Logger.Fatal(err)
	// 		}
	// 	} else {
	// 		builder.app.Logger.Fatal(err)
	// 	}
	// }

	// Load all the farms
	// if len(orgs) > 0 {
	// 	for _, org := range orgs {
	// 		farmConfigs = append(farmConfigs, org.GetFarms()...)
	// 	}
	// } else {
	// 	farmConfigs, err = farmDAO.GetAll(common.CONSISTENCY_LOCAL)
	// }
	// if err != nil {
	// 	builder.app.Logger.Fatal(err)
	// }

	// Build a FarmService for each farm in the database
	// for _, farmConfig := range farmConfigs {

	// 	farmID := farmConfig.ID

	// 	farmEventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.db, int(farmID))

	// 	farmService, err := farmFactory.BuildService(builder.farmStateStore, farmDAO, farmEventLogDAO,
	// 		builder.deviceDataStore, builder.deviceStateStore,
	// 		farmConfig)
	// 	if err != nil {
	// 		builder.app.Logger.Errorf("Error: %s", err)
	// 		continue
	// 	}
	// 	farmService.RefreshHardwareVersions()

	// 	builder.app.Logger.Debugf("Farm ID: %s", farmID)
	// 	builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
	// 	builder.app.Logger.Debugf("Mode: %s", builder.app.Mode)
	// 	builder.app.Logger.Debugf("Timezone: %s", builder.app.Timezone)
	// 	builder.app.Logger.Debugf("Polling interval: %d", builder.app.Interval)

	// 	eventLogService.Create(0, common.CONTROLLER_TYPE_SERVER, "System", "Startup")
	// }

	// Listen for new farm provisioning requests
	go func() {
		for {
			farmConfig := <-farmProvisionerChan
			builder.app.Logger.Debugf("Processing provisioner request...")

			// farmKey := fmt.Sprintf("%d-%s", farmConfig.OrganizationID, farmConfig.Name)
			// farmID := builder.idGenerator.NewID32(farmKey)

			// farmChannels := &service.FarmChannels{
			// 	FarmConfigChan:       make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
			// 	FarmConfigChangeChan: make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
			// 	FarmStateChan:        make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
			// 	FarmStateChangeChan:  make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
			// 	FarmErrorChan:        make(chan common.FarmError, common.BUFFERED_CHANNEL_SIZE),
			// 	FarmNotifyChan:       make(chan common.FarmNotification, common.BUFFERED_CHANNEL_SIZE),
			// 	//MetricChangedChan:     make(chan common.MetricValueChanged, common.BUFFERED_CHANNEL_SIZE),
			// 	//SwitchChangedChan:     make(chan common.SwitchValueChanged, common.BUFFERED_CHANNEL_SIZE),
			// 	DeviceStateChangeChan: make(chan common.DeviceStateChange, common.BUFFERED_CHANNEL_SIZE),
			// 	DeviceStateDeltaChan:  make(chan map[string]state.DeviceStateDeltaMap, common.BUFFERED_CHANNEL_SIZE)}

			// farmEventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.db, farmID)
			// farmService, err := farmFactory.BuildService(
			// 	builder.farmStateStore, farmDAO, farmEventLogDAO, builder.deviceDataStore,
			// 	builder.deviceStateStore, &farmConfig, farmChannels)
			// if err != nil {
			// 	builder.app.Logger.Errorf("Error: %s", err)
			// 	break
			// }
			// farmService.InitializeState(true)
			// farmService.RefreshHardwareVersions()
			// builder.serviceRegistry.AddFarmService(farmService)
			// farmTickerProvisionerChan <- farmConfig.ID
			// go farmService.Run()

			builder.createAndRunFarm(farmDAO, farmFactory, &farmConfig)

			farmTickerProvisionerChan <- farmConfig.ID
		}
	}()

	// Listen for new farm deprovisioning requests
	go func() {
		for {
			farmConfig := <-farmDeprovisionerChan
			farmID := farmConfig.ID
			builder.app.Logger.Debugf("Processing deprovisioning request for farm %d", farmID)
			farmService := builder.serviceRegistry.GetFarmService(farmID)
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
	}()

	organizationService := service.NewOrganizationService(
		builder.app.Logger, builder.idGenerator,
		builder.datastoreRegistry.GetOrganizationDAO())
	builder.serviceRegistry.SetOrganizationService(organizationService)

	// Build JWT service
	jsonWriter := rest.NewJsonWriter(builder.app.Logger)
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	defaultRole, err := roleDAO.GetByName(builder.app.DefaultRole, common.CONSISTENCY_LOCAL)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	jwtService := service.CreateJsonWebTokenService(builder.app,
		builder.idGenerator, defaultRole, mapperRegistry.GetDeviceMapper(),
		builder.serviceRegistry, jsonWriter, 525960, rsaKeyPair) // 1 year jwt expiration
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	builder.serviceRegistry.SetJsonWebTokenService(jwtService)

	// if builder.app.DataStoreCDC {
	// 	changefeedService := service.NewChangefeedService(builder.app, serviceRegistry,
	// 		changefeeders)
	// 	serviceRegistry.SetChangefeedService(changefeedService)
	// }

	publicKey := string(rsaKeyPair.GetPublicBytes())
	restServiceRegistry := rest.NewRestServiceRegistry(builder.app,
		publicKey, mapperRegistry, builder.serviceRegistry)
	restServices = restServiceRegistry.GetRestServices()

	return rsaKeyPair, builder.serviceRegistry, restServices, farmTickerProvisionerChan, err
}

func (builder *GormConfigBuilder) createAndRunFarm(farmDAO dao.FarmDAO,
	farmFactory service.FarmFactory, farmConfig *config.Farm) {

	farmID := farmConfig.ID

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

	farmEventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.db, int(farmID))
	farmService, err := farmFactory.BuildService(
		builder.farmStateStore, farmDAO, farmEventLogDAO, builder.deviceDataStore,
		builder.deviceStateStore, farmConfig, farmChannels)
	if err != nil {
		builder.app.Logger.Errorf("createAndRunFarm error: %s", err)
		return
	}
	farmService.InitializeState(true)
	farmService.RefreshHardwareVersions()
	builder.serviceRegistry.AddFarmService(farmService)
	go farmService.Run()

	builder.app.Logger.Debugf("Farm ID: %s", farmID)
	builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
	builder.app.Logger.Debugf("Mode: %s", builder.app.Mode)
	builder.app.Logger.Debugf("Timezone: %s", builder.app.Timezone)
	builder.app.Logger.Debugf("Polling interval: %d", builder.app.Interval)

	systemEventLogService := builder.serviceRegistry.GetEventLogService(0)
	systemEventLogService.Create(0, common.CONTROLLER_TYPE_SERVER, "System", "Startup")
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

	passwordHasher := util.CreatePasswordHasher(builder.app.PasswordHasherParams)

	configInitializer := dao.NewConfigInitializer(builder.app.Logger,
		builder.app.IdGenerator, builder.app.Location,
		builder.datastoreRegistry, passwordHasher, builder.app.Mode)

	_, err := configInitializer.Initialize(builder.app.EnableDefaultFarm, provParams)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	// if err := farmDAO.Save(farmConfig); err != nil {
	// 	builder.app.Logger.Fatal(err)
	// }
}
