// +build cluster
// +build !cloud

package builder

import (
	"crypto/rsa"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	configstore "github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/cockroach"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/store"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"
)

type ClusterConfigBuilder struct {
	app               *app.App
	params            *cluster.ClusterParams
	farmStateStore    state.FarmStorer
	deviceStateStore  state.DeviceStorer
	deviceDataStore   datastore.DeviceDatastore
	consistencyLevel  int
	appStateTTL       int
	appStateTick      int
	datastoreRegistry datastore.DatastoreRegistry
	serviceRegistry   service.ServiceRegistry
	mapperRegistry    mapper.MapperRegistry
	changefeeders     map[string]datastore.Changefeeder
	ConfigBuilder
}

func NewClusterConfigBuilder(_app *app.App, params *cluster.ClusterParams,
	deviceStore string, appStateTTL, appStateTick int) ConfigBuilder {

	farmStateStore := state.NewMemoryFarmStore(_app.Logger, 1, appStateTTL,
		time.Duration(appStateTick))
	deviceStateStore := state.NewMemoryDeviceStore(_app.Logger, 3, appStateTTL,
		time.Duration(appStateTick))

	// farmStateStore := cluster.NewRaftFarmStateStore(App.Logger, App.RaftCluster)
	// deviceStateStore := cluster.NewRaftDeviceStateStore(App.Logger, App.RaftCluster)

	var deviceDatastore datastore.DeviceDatastore
	if deviceStore == "raft" {
		deviceDatastore = cluster.NewRaftDeviceStateStore(_app.Logger, _app.RaftCluster)
	} else if deviceStore == "redis" {
		deviceDatastore = datastore.NewRedisDeviceStore(":6379", "")
	} else {
		deviceDatastore = store.NewGormDeviceStore(_app.Logger, _app.GORM,
			_app.GORMInitParams.Engine, _app.Location)
	}

	return &ClusterConfigBuilder{
		app:              _app,
		params:           params,
		farmStateStore:   farmStateStore,
		deviceStateStore: deviceStateStore,
		deviceDataStore:  deviceDatastore,
		consistencyLevel: common.CONSISTENCY_CACHED}
}

func (builder *ClusterConfigBuilder) Build() (app.KeyPair, config.ServerConfig,
	service.ServiceRegistry, []rest.RestService, error) {

	builder.mapperRegistry = mapper.CreateRegistry()
	builder.datastoreRegistry = builder.params.GetDatastoreRegistry()
	builder.serviceRegistry = service.CreateClusterServiceRegistry(builder.app,
		builder.datastoreRegistry, builder.mapperRegistry)

	builder.changefeeders = make(map[string]datastore.Changefeeder, 0)

	if builder.app.DatastoreCDC && builder.app.DatastoreType == "cockroach" {
		// Farm and device config tables
		builder.changefeeders["_device_config_items"] =
			cockroach.NewCockroachChangefeed(builder.app, "device_config_items")
		builder.changefeeders["_channels"] =
			cockroach.NewCockroachChangefeed(builder.app, "channels")
		builder.changefeeders["_metrics"] =
			cockroach.NewCockroachChangefeed(builder.app, "metrics")
		builder.changefeeders["_conditions"] =
			cockroach.NewCockroachChangefeed(builder.app, "conditions")
		builder.changefeeders["_schedules"] =
			cockroach.NewCockroachChangefeed(builder.app, "schedules")
	}

	// TODO: Replace with modular backend event log storage
	eventLogDAO := gorm.NewEventLogDAO(builder.app.Logger, builder.app.GORM)
	eventLogService := service.NewEventLogService(builder.app, eventLogDAO,
		common.CONTROLLER_TYPE_SERVER)
	builder.serviceRegistry.SetEventLogService(eventLogService)

	serverConfig := config.NewServer()
	serverConfig.SetInterval(builder.app.Interval)
	serverConfig.SetTimezone(builder.app.Location.String())
	serverConfig.SetMode(builder.app.Mode)
	//serverConfig.SetSmtp()
	//serverConfig.SetLicense()
	//serverConfig.SetOrganizations()

	// orgs, err := orgDAO.GetAll()
	// if err != nil {
	// 	builder.app.Logger.Fatal(err)
	// }

	//builder.app.Config = serverConfig.(*config.Server)
	//builder.app.Logger.Debugf("builder.app.Config: %+v", builder.app.Config)

	farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)

	farmConfigs, err := farmDAO.GetAll()
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	serverConfig.SetFarms(farmConfigs)

	var restServices []rest.RestService

	deviceDAO := builder.datastoreRegistry.GetDeviceDAO()
	gormFarmConfigStore := store.NewGormFarmConfigStore(farmDAO, 1)
	gormDeviceConfigStore := store.NewGormDeviceConfigStore(deviceDAO, 3)

	defaultFarmFactory := service.NewFarmFactory(
		builder.app, builder.datastoreRegistry, builder.serviceRegistry, builder.farmStateStore,
		gormFarmConfigStore, builder.deviceStateStore, gormDeviceConfigStore,
		builder.deviceDataStore, builder.mapperRegistry.GetDeviceMapper(),
		builder.changefeeders, builder.params.GetFarmProvisionerChan(),
		builder.params.GetFarmTickerProvisionerChan())

	builder.serviceRegistry.SetFarmFactory(defaultFarmFactory)

	for _, farmConfig := range farmConfigs {

		// Make a copy of the farmConfig
		var conf *config.Farm = &config.Farm{}
		*conf = farmConfig
		builder.BuildFarmService(farmDAO, conf)

		builder.app.Logger.Debugf("Farm ID: %s", farmConfig.GetID())
		builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
		builder.app.Logger.Debugf("Mode: %s", serverConfig.GetMode())
		builder.app.Logger.Debugf("Timezone: %s", serverConfig.GetTimezone())
		builder.app.Logger.Debugf("Polling interval: %d", serverConfig.GetInterval())

		builder.serviceRegistry.GetEventLogService().Create("System", "Startup")
	}

	// Listen for new farm provisioning requests
	// go func() {
	// 	for {
	// 		select {
	// 		case farmConfig := <-builder.params.GetFarmProvisionerChan():
	// 			builder.app.Logger.Debugf("Processing new farm provisioning request: farmConfig=%+v", farmConfig)
	// 			farmService, err := builder.BuildFarmService(farmDAO, farmConfig)
	// 			if err != nil {
	// 				builder.app.Logger.Errorf("Error: %s", err)
	// 			}
	// 			builder.serviceRegistry.AddFarmService(farmService)
	// 			builder.params.GetFarmTickerProvisionerChan() <- farmConfig.GetID()
	// 		default:
	// 			builder.app.Logger.Error("Error: Unable to process provisioner request")
	// 		}
	// 	}
	// }()

	// Build JWT service
	jsonWriter := rest.NewJsonWriter()
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	jwtService := service.CreateJsonWebTokenService(builder.app, farmDAO,
		builder.mapperRegistry.GetDeviceMapper(), builder.serviceRegistry, jsonWriter,
		525960, rsaKeyPair) // 1 year jwt expiration
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	builder.serviceRegistry.SetJsonWebTokenService(jwtService)

	// v0.0.3a: Removing ConfigService in favor of FarmService
	// configService := service.NewConfigService(builder.app, builder.datastoreRegistry,
	// 	builder.serviceRegistry)
	// builder.serviceRegistry.SetConfigService(configService)

	// Build device and channel cache/indexes (to provide o(n) lookups when searching service registry and farm)
	// deviceIndex := state.CreateDeviceIndex(farmFactory.GetDeviceIndexMap())
	// channelIndex := state.CreateChannelIndex(farmFactory.GetChannelIndexMap())

	if builder.app.DatastoreCDC {
		changefeedService := service.NewChangefeedService(builder.app,
			builder.serviceRegistry, builder.changefeeders)
		builder.serviceRegistry.SetChangefeedService(changefeedService)
	}

	publicKey := string(rsaKeyPair.GetPublicBytes())
	restServiceRegistry := rest.NewClusterRestServiceRegistry(publicKey, builder.mapperRegistry,
		builder.serviceRegistry)
	if restServices == nil {
		restServices = restServiceRegistry.GetRestServices()
	}

	return rsaKeyPair, serverConfig, builder.serviceRegistry, restServices, err
}

// Creates and new FarmFactory and FarmService initialized based on farmConfig
// configuration values.
func (builder *ClusterConfigBuilder) BuildFarmService(farmDAO dao.FarmDAO,
	farmConfig config.FarmConfig) (service.FarmService, error) {

	var farmStateStore state.FarmStorer
	switch farmConfig.GetStateStore() {
	case state.MEMORY_STORE:
		farmStateStore = builder.newMemoryFarmStore()
	case state.RAFT_STORE:
		farmStateStore = cluster.NewRaftFarmStateStore(builder.app.Logger,
			builder.app.RaftCluster)
	default:
		farmStateStore = builder.newMemoryFarmStore()
	}

	var farmConfigStore configstore.FarmConfigStorer
	switch farmConfig.GetConfigStore() {
	case configstore.MEMORY_STORE, configstore.GORM_STORE:
		farmConfigStore = store.NewGormFarmConfigStore(farmDAO, 1)
	case configstore.RAFT_STORE:
		farmConfigStore = cluster.NewRaftFarmConfigStore(builder.app.Logger,
			builder.app.RaftCluster)
	}

	deviceDAO := builder.datastoreRegistry.GetDeviceDAO()
	gormDeviceConfigStore := store.NewGormDeviceConfigStore(deviceDAO, 3)

	return service.NewFarmFactory(
		builder.app, builder.datastoreRegistry, builder.serviceRegistry, farmStateStore,
		farmConfigStore, builder.deviceStateStore, gormDeviceConfigStore,
		builder.deviceDataStore, builder.mapperRegistry.GetDeviceMapper(),
		builder.changefeeders, builder.params.GetFarmProvisionerChan(),
		builder.params.GetFarmTickerProvisionerChan()).BuildClusterService(farmConfig)
}

// Creates a new MemoryFarmStore instance
func (builder *ClusterConfigBuilder) newMemoryFarmStore() state.FarmStorer {
	return state.NewMemoryFarmStore(builder.app.Logger, 1,
		builder.appStateTTL, time.Duration(builder.appStateTick))
}
