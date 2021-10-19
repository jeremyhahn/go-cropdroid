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
	configstore "github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/store"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"
)

type ClusterConfigBuilder struct {
	app               *app.App
	params            *cluster.ClusterParams
	appStateTTL       int
	appStateTick      int
	datastoreRegistry datastore.DatastoreRegistry
	serviceRegistry   service.ClusterServiceRegistry
	mapperRegistry    mapper.MapperRegistry
	changefeeders     map[string]datastore.Changefeeder
	gossipCluster     cluster.GossipCluster
	raftCluster       cluster.RaftCluster
}

func NewClusterConfigBuilder(_app *app.App, params *cluster.ClusterParams,
	gossipCluster cluster.GossipCluster, raftCluster cluster.RaftCluster,
	dataStore string, appStateTTL, appStateTick int) *ClusterConfigBuilder {

	return &ClusterConfigBuilder{
		app:           _app,
		params:        params,
		gossipCluster: gossipCluster,
		raftCluster:   raftCluster}
}

func (builder *ClusterConfigBuilder) Build() (app.KeyPair, config.ServerConfig,
	service.ClusterServiceRegistry, []rest.RestService, chan uint64, error) {

	builder.mapperRegistry = mapper.CreateRegistry()
	builder.datastoreRegistry = builder.params.GetDatastoreRegistry()
	// builder.serviceRegistry = service.CreateClusterServiceRegistry(builder.app,
	// 	builder.datastoreRegistry, builder.mapperRegistry)
	builder.serviceRegistry = service.CreateClusterServiceRegistry(builder.app,
		builder.datastoreRegistry, builder.mapperRegistry,
		builder.gossipCluster, builder.raftCluster)

	gormInitializer := gorm.NewGormInitializer(builder.app.Logger,
		builder.app.GormDB, builder.app.Location, builder.app.Mode)

	farmProvisioner := provisioner.NewRaftFarmProvisioner(
		builder.app.Logger, builder.gossipCluster, builder.app.Location,
		builder.datastoreRegistry.NewFarmDAO(),
		builder.mapperRegistry.GetUserMapper(),
		//		builder.params.GetFarmProvisionerChan(),
		//		builder.params.GetFarmDeprovisionerChan(),
		gormInitializer)

	builder.serviceRegistry.SetFarmProvisioner(farmProvisioner)

	// initializer := gorm.NewGormInitializer(builder.app.Logger,
	// 	builder.app.GormDB, builder.app.Location)

	// farmProvisioner := provisioner.NewRaftFarmProvisioner(
	// 	builder.app.Logger, builder.app.GossipCluster, builder.app.Location,
	// 	builder.datastoreRegistry.NewFarmDAO(),
	// 	builder.mapperRegistry.GetUserMapper(), initializer)

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

	orgDAO := gorm.NewOrganizationDAO(builder.app.Logger, builder.app.GORM)
	farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)
	roleDAO := gorm.NewRoleDAO(builder.app.Logger, builder.app.GORM)
	farmConfigs, err := farmDAO.GetAll()
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	serverConfig.SetFarms(farmConfigs)

	for _, farmConfig := range farmConfigs {

		// Make a copy of the farmConfig
		var conf *config.Farm = &config.Farm{}
		*conf = farmConfig

		farmFactory := builder.BuildFarmFactory(farmConfig.GetStateStore(),
			farmConfig.GetConfigStore(), farmConfig.GetDataStore())

		_, err := farmFactory.BuildClusterService(conf)
		if err != nil {
			builder.app.Logger.Fatalf("Error loading farm config: %s", err)
		}

		builder.app.Logger.Debugf("Farm ID: %s", farmConfig.GetID())
		builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
		builder.app.Logger.Debugf("Mode: %s", serverConfig.GetMode())
		builder.app.Logger.Debugf("Timezone: %s", serverConfig.GetTimezone())
		builder.app.Logger.Debugf("Polling interval: %d", serverConfig.GetInterval())

		builder.serviceRegistry.GetEventLogService().Create("System", "Startup")
	}

	// Listen for new farm provisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-builder.params.GetFarmProvisionerChan():
				builder.app.Logger.Debugf("Processing new farm provisioning request: farmConfig=%+v", farmConfig)
				farmFactory := builder.BuildFarmFactory(farmConfig.GetStateStore(),
					farmConfig.GetConfigStore(), farmConfig.GetDataStore())
				farmService, err := farmFactory.BuildClusterService(farmConfig)
				if err != nil {
					builder.app.Logger.Errorf("Error: %s", err)
				}
				builder.serviceRegistry.AddFarmService(farmService)
				builder.params.GetFarmTickerProvisionerChan() <- farmConfig.GetID()
				go farmService.Run()
			}
		}
	}()

	// Listen for new farm deprovisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-builder.params.GetFarmDeprovisionerChan():
				farmID := farmConfig.GetID()
				builder.app.Logger.Debugf("Processing deprovisioning request for farm %d", farmID)
				farmService := builder.serviceRegistry.GetFarmService(farmID)
				if farmService == nil {
					builder.app.Logger.Errorf("Farm not found: %d", farmID)
					continue
				}
				farmService.Stop()
				if err := farmDAO.Delete(farmConfig); err != nil {
					builder.app.Logger.Error(err)
				}
				builder.params.GetFarmTickerProvisionerChan() <- farmID
			}
		}
	}()

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
	jwtService := service.CreateJsonWebTokenService(builder.app, orgDAO, farmDAO, defaultRole,
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

	// if builder.app.DataStoreCDC {
	// 	changefeedService := service.NewChangefeedService(builder.app,
	// 		builder.serviceRegistry, builder.changefeeders)
	// 	builder.serviceRegistry.SetChangefeedService(changefeedService)
	// }

	publicKey := string(rsaKeyPair.GetPublicBytes())

	restServices := rest.NewRestServiceRegistry(publicKey, builder.mapperRegistry,
		builder.serviceRegistry).GetRestServices()

	return rsaKeyPair, serverConfig, builder.serviceRegistry, restServices,
		builder.params.GetFarmTickerProvisionerChan(), err
}

// Creates and new FarmFactory and FarmService initialized based on farmConfig
// configuration values.
func (builder *ClusterConfigBuilder) BuildFarmFactory(stateStoreType,
	configStoreType, dataStoreType int) service.FarmFactory {

	farmStateStore := builder.createFarmStateStore(stateStoreType)
	farmConfigStore := builder.createFarmConfigStore(configStoreType)

	deviceDataStore := builder.createDeviceDataStore(dataStoreType)
	deviceStateStore := builder.createDeviceStateStore(stateStoreType, deviceDataStore)
	deviceConfigStore := builder.createDeviceConfigStore()

	// LEAVING OFF HERE = NEED TO FIGURE OUT HOW TO STORE FARM FACTORY
	// FOR MULTIPLE FARMS / ORGANIZATION

	return service.NewFarmFactory(
		builder.app, builder.datastoreRegistry, builder.serviceRegistry, farmStateStore,
		farmConfigStore, deviceStateStore, deviceConfigStore,
		deviceDataStore, builder.mapperRegistry.GetDeviceMapper(),
		builder.changefeeders, builder.params.GetFarmProvisionerChan(),
		builder.params.GetFarmTickerProvisionerChan())
}

func (builder *ClusterConfigBuilder) createFarmStateStore(storeType int) state.FarmStorer {
	var farmStateStore state.FarmStorer
	switch storeType {
	case state.MEMORY_STORE:
		farmStateStore = builder.newMemoryFarmStateStore()
	case state.RAFT_STORE:
		farmStateStore = cluster.NewRaftFarmStateStore(builder.app.Logger, builder.raftCluster)
	default:
		farmStateStore = builder.newMemoryFarmStateStore()
	}
	return farmStateStore
	//return builder.newMemoryFarmStateStore()
}

func (builder *ClusterConfigBuilder) createFarmConfigStore(storeType int) configstore.FarmConfigStorer {
	var farmConfigStore configstore.FarmConfigStorer
	switch storeType {
	case config.MEMORY_STORE, config.GORM_STORE:
		farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)
		farmConfigStore = store.NewGormFarmConfigStore(farmDAO, 1)
	case config.RAFT_MEMORY_STORE, config.RAFT_DISK_STORE:
		farmConfigStore = cluster.NewRaftFarmConfigStore(builder.app.Logger,
			builder.raftCluster)
	}
	return farmConfigStore
}

func (builder *ClusterConfigBuilder) createDeviceStateStore(storeType int, deviceDataStore datastore.DeviceDataStore) state.DeviceStorer {
	var deviceStateStore state.DeviceStorer
	switch storeType {
	case state.MEMORY_STORE:
		deviceStateStore = builder.newMemoryDeviceStateStore()
	case state.RAFT_STORE:
		// if farmConfig.GetStateStore() == state.RAFT_STORE {
		// 	deviceStateStore = deviceDataStore.(state.DeviceStorer)
		// } else {
		// 	deviceStateStore = cluster.NewRaftDeviceStore(builder.app.Logger, builder.app.RaftCluster)
		// }
		deviceStateStore = cluster.NewRaftDeviceStateStore(builder.app.Logger, builder.raftCluster)
	default:
		deviceStateStore = builder.newMemoryDeviceStateStore()
	}
	return deviceStateStore
	//return builder.newMemoryDeviceStateStore()
}

func (builder *ClusterConfigBuilder) createDeviceConfigStore() configstore.DeviceConfigStorer {
	// var deviceConfigStore configstore.DeviceConfigStorer
	// switch farmConfig.GetConfigStore() {
	// case config.MEMORY_STORE, config.GORM_STORE:
	// 	deviceDAO := builder.datastoreRegistry.GetDeviceDAO()
	// 	deviceConfigStore = store.NewGormDeviceConfigStore(deviceDAO, 3)
	// case config.RAFT_MEMORY_STORE, config.RAFT_DISK_STORE:
	// 	deviceConfigStore = cluster.NewRaftDeviceConfigStore(
	// 		builder.app.Logger, builder.app.RaftCluster)
	// }
	// return deviceConfigStore
	deviceDAO := builder.datastoreRegistry.GetDeviceDAO()
	return store.NewGormDeviceConfigStore(deviceDAO, 3)
}

func (builder *ClusterConfigBuilder) createDeviceDataStore(storeType int) datastore.DeviceDataStore {
	var deviceDataStore datastore.DeviceDataStore
	switch storeType {
	case datastore.GORM_STORE:
		deviceDataStore = store.NewGormDataStore(builder.app.Logger,
			builder.app.GORM, builder.app.GORMInitParams.Engine,
			builder.app.Location)
	case datastore.RAFT_STORE:
		deviceDataStore = cluster.NewRaftDeviceStateStore(builder.app.Logger, builder.raftCluster)
	case datastore.REDIS_TS:
		deviceDataStore = datastore.NewRedisDataStore(":6379", "")
	}
	return deviceDataStore
}

// Creates a new MemoryFarmStore instance
func (builder *ClusterConfigBuilder) newMemoryFarmStateStore() state.FarmStorer {
	return state.NewMemoryFarmStore(builder.app.Logger, 1,
		builder.appStateTTL, time.Duration(builder.appStateTick))
}

// Creates a new MemoryDeviceStore instance
func (builder *ClusterConfigBuilder) newMemoryDeviceStateStore() state.DeviceStorer {
	return state.NewMemoryDeviceStore(builder.app.Logger, 1,
		builder.appStateTTL, time.Duration(builder.appStateTick))
}
