package builder

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"

	gormds "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/redis"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"

	"gorm.io/gorm"
)

type GormConfigBuilder struct {
	app               *app.App
	db                *gorm.DB
	gormDB            gormds.GormDB
	datastoreRegistry dao.Registry
	serviceRegistry   service.ServiceRegistry
	farmStateStore    state.FarmStateStorer
	deviceStateStore  state.DeviceStateStorer
	deviceDataStore   datastore.DeviceDataStore
	consistencyLevel  int
	databaseInit      bool
	idGenerator       util.IdGenerator
}

func NewGormConfigBuilder(app *app.App) *GormConfigBuilder {

	gormDB := gormds.NewGormDB(app.Logger, app.GORMInitParams)
	db := gormDB.Connect(false)

	farmStateStore := state.NewMemoryFarmStore(app.Logger, 1, app.StateTTL,
		time.Duration(app.StateTick))

	deviceStateStore := state.NewMemoryDeviceStore(app.Logger, 3, app.StateTTL,
		time.Duration(app.StateTick))

	var deviceDatastore datastore.DeviceDataStore
	if app.DataStoreEngine == "redis" {
		deviceDatastore = redis.NewRedisDataStore(":6379", "")
	} else {
		deviceDatastore = gormds.NewGormDeviceDataStore(app.Logger, db,
			app.GORMInitParams.Engine, app.Location)
	}

	return &GormConfigBuilder{
		app:              app,
		db:               db,
		gormDB:           gormDB,
		farmStateStore:   farmStateStore,
		deviceStateStore: deviceStateStore,
		deviceDataStore:  deviceDatastore,
		databaseInit:     app.DatabaseInit,
		consistencyLevel: common.CONSISTENCY_LOCAL,
		idGenerator:      util.NewIdGenerator(app.DataStoreEngine)}
}

func (builder *GormConfigBuilder) Build() (mapper.MapperRegistry,
	service.ServiceRegistry, rest.RestServiceRegistry, chan uint64, error) {

	builder.datastoreRegistry = gormds.NewGormRegistry(builder.app.Logger, builder.gormDB)
	mapperRegistry := mapper.CreateRegistry()
	builder.serviceRegistry = service.CreateServiceRegistry(builder.app,
		builder.datastoreRegistry, mapperRegistry)

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
	deviceConfigDAO := gormds.NewDeviceSettingDAO(builder.app.Logger, builder.db)

	farmProvisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
	farmDeprovisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
	farmTickerProvisionerChan := make(chan uint64, common.BUFFERED_CHANNEL_SIZE)

	passwordHasher := util.CreatePasswordHasher(builder.app.PasswordHasherParams)

	configInitializer := dao.NewConfigInitializer(builder.app.Logger,
		builder.app.IdGenerator, builder.app.Location, builder.datastoreRegistry,
		passwordHasher, builder.app.Mode)

	farmFactory := service.NewFarmFactory(
		builder.app, farmDAO, builder.datastoreRegistry, deviceConfigDAO, builder.serviceRegistry,
		mapperRegistry.GetDeviceMapper(), farmProvisionerChan, farmTickerProvisionerChan, builder.idGenerator)
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
	err := orgDAO.ForEachPage(query.NewPageQuery(), func(entities []*config.OrganizationStruct) error {
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
	err = farmDAO.ForEachPage(query.NewPageQuery(), func(entities []*config.FarmStruct) error {
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

	// Listen for new farm provisioning requests
	go func() {
		for {
			farmConfig := <-farmProvisionerChan
			builder.app.Logger.Debugf("Processing provisioner request...")
			farmDAO.Save(farmConfig.(*config.FarmStruct))
			builder.createAndRunFarm(farmDAO, farmFactory, farmConfig)
			farmTickerProvisionerChan <- farmConfig.Identifier()
		}
	}()

	// Listen for new farm deprovisioning requests
	go func() {
		for {
			farmConfig := <-farmDeprovisionerChan
			farmID := farmConfig.Identifier()
			builder.app.Logger.Debugf("Processing deprovisioning request for farm %d", farmID)
			farmService := builder.serviceRegistry.GetFarmService(farmID)
			if farmService == nil {
				builder.app.Logger.Errorf("Farm not found: %d", farmID)
				continue
			}
			farmService.Stop()
			if err := farmDAO.Delete(farmConfig.(*config.FarmStruct)); err != nil {
				builder.app.Logger.Error(err)
			}
			farmTickerProvisionerChan <- farmID
		}
	}()

	organizationService := service.NewOrganizationService(
		builder.app.Logger, builder.idGenerator,
		builder.datastoreRegistry.GetOrganizationDAO(),
		mapperRegistry.GetUserMapper())
	builder.serviceRegistry.SetOrganizationService(organizationService)

	restServiceRegistry := rest.NewRestServiceRegistry(
		builder.app,
		roleDAO,
		mapperRegistry,
		builder.serviceRegistry)

	return mapperRegistry, builder.serviceRegistry, restServiceRegistry, farmTickerProvisionerChan, err
}

func (builder *GormConfigBuilder) createAndRunFarm(farmDAO dao.FarmDAO,
	farmFactory service.FarmFactory, farmConfig config.Farm) {

	farmID := farmConfig.Identifier()

	farmChannels := &service.FarmChannels{
		FarmConfigChan:       make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
		FarmConfigChangeChan: make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
		// FarmStateChan:        make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChangeChan: make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmErrorChan:       make(chan common.FarmError, common.BUFFERED_CHANNEL_SIZE),
		FarmNotifyChan:      make(chan common.FarmNotification, common.BUFFERED_CHANNEL_SIZE),
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

	farmEventLogService := builder.serviceRegistry.GetEventLogService(farmID)
	farmEventLogService.Create(farmID, common.CONTROLLER_TYPE_SERVER, "Startup", "Starting farm service")
}

func (builder *GormConfigBuilder) initDatabase() {

	builder.app.Logger.Info("Initializing database")

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
}
