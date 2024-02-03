//go:build cluster && !cloud
// +build cluster,!cloud

package builder

import (
	"crypto/rsa"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/cluster/util"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	gormds "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/redis"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"

	"github.com/jeremyhahn/go-cropdroid/config/dao"
)

type ClusterConfigBuilder struct {
	app               *app.App
	params            *util.ClusterParams
	appStateTTL       int
	appStateTick      int
	datastoreRegistry dao.Registry
	serviceRegistry   service.ClusterServiceRegistry
	mapperRegistry    mapper.MapperRegistry
	changefeeders     map[string]datastore.Changefeeder
	gossipNode        cluster.GossipNode
	raftNode          cluster.RaftNode
}

func NewClusterConfigBuilder(app *app.App, params *util.ClusterParams,
	gossipNode cluster.GossipNode, raftNode cluster.RaftNode,
	dataStore string, appStateTTL, appStateTick int) *ClusterConfigBuilder {

	return &ClusterConfigBuilder{
		app:        app,
		params:     params,
		gossipNode: gossipNode,
		raftNode:   raftNode}
}

func (builder *ClusterConfigBuilder) Build() (app.KeyPair, service.ClusterServiceRegistry,
	[]rest.RestService, chan uint64, error) {

	var farmConfigs []*config.Farm

	builder.mapperRegistry = mapper.CreateRegistry()
	builder.datastoreRegistry = cluster.NewRaftRegistry(
		builder.app.Logger, builder.app.IdGenerator, builder.raftNode,
	)
	builder.serviceRegistry = service.CreateClusterServiceRegistry(
		builder.app, builder.datastoreRegistry, builder.mapperRegistry,
		builder.gossipNode, builder.raftNode)

	configInitializer := dao.NewConfigInitializer(builder.app.Logger,
		builder.app.IdGenerator, builder.app.Location,
		builder.datastoreRegistry, builder.app.Mode)

	builder.gossipNode.SetInitializer(configInitializer)

	serverDAO := builder.datastoreRegistry.(*cluster.RaftDaoRegistry).GetServerDAO()
	roleDAO := builder.datastoreRegistry.GetRoleDAO()
	orgDAO := builder.datastoreRegistry.GetOrganizationDAO()
	farmDAO := builder.datastoreRegistry.GetFarmDAO()
	systemEventLogDAO := builder.datastoreRegistry.GetEventLogDAO()

	// Build and add event log service to registry
	raftParams := builder.raftNode.GetParams()
	systemEventLogService := service.NewEventLogService(builder.app, systemEventLogDAO, raftParams.RaftOptions.SystemClusterID)
	builder.serviceRegistry.AddEventLogService(systemEventLogService)

	//eventLogDAO.(*cluster.RaftEventLogDAO).StartCluster()
	//serverEventLogClusterID := builder.params.IdGenerator.CreateEventLogClusterID(
	//	builder.params.RaftOptions.SystemClusterID)
	//builder.raftNode.WaitForClusterReady(serverEventLogClusterID)

	if builder.params.Initialize { //&& builder.raftNode.IsLeader(builder.params.ClusterID) {

		//	serverConfig, serverErr := serverDAO.GetConfig(common.CONSISTENCY_LOCAL)

		//if serverErr != nil && serverErr.Error() == "not found" {

		//	if builder.params.Initialize { //&& builder.raftNode.IsLeader(builder.params.ClusterID) {

		builder.app.Logger.Info("Initializing cluster...")

		provParams := &common.ProvisionerParams{
			UserID:           0,
			RoleID:           0,
			OrganizationID:   0,
			FarmName:         common.DEFAULT_CROP_NAME,
			ConfigStoreType:  builder.app.DefaultConfigStoreType,
			StateStoreType:   builder.app.DefaultStateStoreType,
			DataStoreType:    builder.app.DefaultDataStoreType,
			ConsistencyLevel: common.CONSISTENCY_LOCAL}

		farmConfig, err := configInitializer.Initialize(false, provParams)
		if err != nil {
			builder.app.Logger.Fatal(err)
		}

		serverConfig := config.NewServer()
		serverConfig.SetID(builder.params.ClusterID)
		if serverErr := serverDAO.Save(serverConfig); serverErr != nil {
			builder.app.Logger.Fatal(serverErr)
		}
		if farmConfig != nil {
			serverConfig.AddFarmRef(farmConfig.GetID())
			serverDAO.Save(serverConfig)
		}

		// Passing false to initialize does not create the farm
		// if err := farmDAO.Save(farmConfig); err != nil {
		// 	builder.app.Logger.Error(err)
		// }
	}

	// gormInitializer := gormds.NewGormInitializer(builder.app.Logger,
	// 	builder.app.GormDB, builder.app.IdGenerator, builder.app.Location,
	// 	builder.app.Mode)

	farmProvisioner := provisioner.NewRaftFarmProvisioner(
		builder.app, builder.gossipNode, builder.app.Location,
		builder.datastoreRegistry.NewFarmDAO(),
		builder.datastoreRegistry.GetUserDAO(),
		builder.mapperRegistry.GetUserMapper(),
		//		builder.params.GetFarmProvisionerChan(),
		//		builder.params.GetFarmDeprovisionerChan(),
		configInitializer)
	builder.serviceRegistry.SetFarmProvisioner(farmProvisioner)

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

	farmFactory := service.NewFarmFactoryCluster(
		builder.app, builder.datastoreRegistry, builder.serviceRegistry,
		builder.mapperRegistry.GetDeviceMapper(), builder.changefeeders,
		builder.params.GetFarmProvisionerChan(),
		builder.params.GetFarmTickerProvisionerChan(), farmChannels)
	builder.serviceRegistry.SetFarmFactory(farmFactory)

	// initializer := gorm.NewGormInitializer(builder.app.Logger,
	// 	builder.app.GormDB, builder.app.Location)

	// farmProvisioner := provisioner.NewRaftFarmProvisioner(
	// 	builder.app.Logger, builder.app.GossipNode, builder.app.Location,
	// 	builder.datastoreRegistry.NewFarmDAO(),
	// 	builder.mapperRegistry.GetUserMapper(), initializer)

	//eventLogDAO := gormds.NewEventLogDAO(builder.app.Logger, builder.gormDB)

	// serverConfig := config.NewServer()
	// serverConfig.SetInterval(builder.app.Server.Interval)
	// serverConfig.SetTimezone(builder.app.Location.String())
	// serverConfig.SetMode(builder.app.Mode)
	//serverConfig.SetSmtp()
	//serverConfig.SetLicense()
	//serverConfig.SetOrganizations()

	// orgs, err := orgDAO.GetAll()
	// if err != nil {
	// 	builder.app.Logger.Fatal(err)
	// }

	//builder.app.Server = serverConfig.(*config.Server)
	//builder.app.Logger.Debugf("builder.app.Server: %+v", builder.app.Server)

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
		if err.Error() != "not found" {
			// ignore errors here so the cluster starts up with an
			// empty config and begins listening for new provisioning
			// requests
			builder.app.Logger.Fatal(err)
		} else {
			builder.app.Logger.Warning("No farms configured!")
		}
	}

	farmConfigs, err = farmDAO.GetAll(common.CONSISTENCY_LOCAL)
	if err != nil {
		if err.Error() != "not found" {
			// ignore errors here so the cluster starts up with an
			// empty config and begins listening for new provisioning
			// requests
			builder.app.Logger.Fatal(err)
		}
	}

	//serverConfig.SetFarms(farmConfigs)

	for _, farmConfig := range farmConfigs {

		farmID := farmConfig.GetID()
		farmDAO.(*cluster.RaftFarmConfigDAO).StartCluster(farmID)

		stateStoreType := farmConfig.GetStateStore()
		configStoreType := farmConfig.GetConfigStore()
		dataStoreType := farmConfig.GetDataStore()

		farmStateStore := builder.createFarmStateStore(stateStoreType)
		farmConfigDAO := builder.createFarmConfigDAO(configStoreType,
			farmConfig.GetOrganizationID(), farmID)

		deviceDataStore := builder.createDeviceDataStore(dataStoreType)
		deviceStateStore := builder.createDeviceStateStore(stateStoreType, deviceDataStore)

		farmEventLogDAO := cluster.NewRaftEventLogDAO(builder.app.Logger,
			builder.raftNode, farmID)

		farmService, err := farmFactory.BuildClusterService(farmStateStore,
			farmConfigDAO, farmEventLogDAO, deviceDataStore, deviceStateStore, farmConfig)
		if err != nil {
			builder.app.Logger.Fatalf("Error loading farm config: %s", err)
		}
		go farmService.RunCluster()

		builder.app.Logger.Debugf("Farm ID: %s", farmID)
		builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
		builder.app.Logger.Debugf("Mode: %s", builder.app.Mode)
		builder.app.Logger.Debugf("Timezone: %s", builder.app.Timezone)
		builder.app.Logger.Debugf("Polling interval: %d", builder.app.Interval)
		builder.app.Logger.Debugf("Event Log ID: %s", builder.app.IdGenerator.CreateEventLogClusterID(farmID))

		systemEventLogService.Create(0, common.CONTROLLER_TYPE_SERVER, "System", "Startup")
	}

	// Listen for new farm provisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-builder.params.GetFarmProvisionerChan():
				builder.app.Logger.Debugf("Processing new farm provisioning request: farmConfig=%+v", farmConfig)

				orgID := farmConfig.GetOrganizationID()
				farmID := farmConfig.GetID()

				// Default farm mode to app global configuration
				farmConfig.SetMode(builder.app.Mode)

				// TODO: DRY this up with Gossip.Priovision
				// farmStateKey := fmt.Sprintf("%s-%d", farmConfig.GetName(), farmID)
				// farmStateID := builder.app.IdGenerator.NewID(farmStateKey)

				stateStoreType := farmConfig.GetStateStore()
				configStoreType := farmConfig.GetConfigStore()
				dataStoreType := farmConfig.GetDataStore()

				farmStateStore := builder.createFarmStateStore(stateStoreType)
				farmConfigDAO := builder.createFarmConfigDAO(configStoreType, orgID, farmID)

				deviceDataStore := builder.createDeviceDataStore(dataStoreType)
				deviceStateStore := builder.createDeviceStateStore(stateStoreType, deviceDataStore)
				//deviceConfigStore := builder.createDeviceConfigDAO(configStoreType, orgID)

				// farmFactory := service.NewFarmFactoryCluster(
				// 	builder.app, builder.gormDB, builder.datastoreRegistry,
				// 	builder.serviceRegistry, builder.mapperRegistry.GetDeviceMapper(),
				// 	builder.changefeeders, builder.params.GetFarmProvisionerChan(),
				// 	builder.params.GetFarmTickerProvisionerChan(), farmChannels)

				// farmFactory.CreateFarmConfigCluster(builder.raftNode,
				// 	farmID, farmChannels.FarmConfigChangeChan)
				// builder.raftNode.WaitForClusterReady(farmID)

				// farmFactory.CreateFarmStateCluster(builder.raftNode,
				// 	farmID, farmStateID, farmChannels.FarmStateChangeChan)
				// builder.raftNode.WaitForClusterReady(farmStateID)

				eventLogDAO := cluster.NewRaftEventLogDAO(builder.app.Logger,
					builder.raftNode, farmID)

				farmService, err := farmFactory.BuildClusterService(farmStateStore,
					farmConfigDAO, eventLogDAO, deviceDataStore, deviceStateStore, &farmConfig)
				if err != nil {
					builder.app.Logger.Errorf("Error: %s", err)
				}

				builder.params.GetFarmTickerProvisionerChan() <- farmConfig.GetID()
				go farmService.RunCluster()
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
				if err := farmDAO.Delete(&farmConfig); err != nil {
					builder.app.Logger.Error(err)
				}
				builder.params.GetFarmTickerProvisionerChan() <- farmID
			}
		}
	}()

	organizationService := service.NewOrganizationService(
		builder.app.Logger, builder.app.IdGenerator,
		builder.datastoreRegistry.GetOrganizationDAO())
	builder.serviceRegistry.SetOrganizationService(organizationService)

	// Build JWT service
	jsonWriter := rest.NewJsonWriter()
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	//farmConfigStore := store.NewGormFarmConfigStore(farmDAO, 1)
	defaultRole, err := roleDAO.GetByName(builder.app.DefaultRole, common.CONSISTENCY_LOCAL)
	for err != nil {
		defaultRole, err = roleDAO.GetByName(builder.app.DefaultRole, common.CONSISTENCY_LOCAL)
		time.Sleep(1 * time.Second)
		builder.app.Logger.Warning("Waiting for cluster database initialization (missing default role)...")
		//builder.app.Logger.Fatal(err)
	}

	jwtService := service.CreateJsonWebTokenService(builder.app,
		builder.app.IdGenerator, orgDAO, farmDAO, defaultRole,
		builder.mapperRegistry.GetDeviceMapper(), builder.serviceRegistry,
		jsonWriter, 525960, rsaKeyPair) // 1 year jwt expiration
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

	restServices := rest.NewRestServiceRegistry(builder.app, publicKey, builder.mapperRegistry,
		builder.serviceRegistry).GetRestServices()

	return rsaKeyPair, builder.serviceRegistry, restServices,
		builder.params.GetFarmTickerProvisionerChan(), err
}

func (builder *ClusterConfigBuilder) createFarmStateStore(storeType int) state.FarmStorer {
	var farmStateStore state.FarmStorer
	switch storeType {
	case state.MEMORY_STORE:
		farmStateStore = builder.newMemoryFarmStateStore()
	case state.RAFT_STORE:
		farmStateStore = cluster.NewRaftFarmStateStore(builder.app.Logger, builder.raftNode)
	default:
		farmStateStore = builder.newMemoryFarmStateStore()
	}
	return farmStateStore
	//return builder.newMemoryFarmStateStore()
}

func (builder *ClusterConfigBuilder) createFarmConfigDAO(storeType int,
	orgID, farmID uint64) dao.FarmDAO {
	var farmDAO dao.FarmDAO
	switch storeType {
	case config.MEMORY_STORE, config.GORM_STORE:
		gormDB := gormds.NewGormDB(builder.app.Logger, builder.app.GORMInitParams).Connect(false)
		farmDAO = gormds.NewFarmDAO(builder.app.Logger, gormDB,
			builder.app.IdGenerator)
		//farmConfigStore = store.NewGormFarmConfigStore(farmDAO, 1)
	case config.RAFT_MEMORY_STORE, config.RAFT_DISK_STORE:
		serverDAO := cluster.NewRaftServerDAO(builder.app.Logger,
			builder.raftNode, builder.app.ClusterID)
		userDAO := cluster.NewRaftUserDAO(builder.app.Logger,
			builder.raftNode, builder.params.RaftOptions.UserClusterID)
		farmDAO = cluster.NewRaftFarmConfigDAO(builder.app.Logger,
			builder.raftNode, serverDAO, userDAO)
	}
	return farmDAO
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
		deviceStateStore = cluster.NewRaftDeviceStateStore(builder.app.Logger, builder.raftNode)
	default:
		deviceStateStore = builder.newMemoryDeviceStateStore()
	}
	return deviceStateStore
	//return builder.newMemoryDeviceStateStore()
}

//func (builder *ClusterConfigBuilder) createDeviceConfigStore(storeType int, clusterID uint64) dao.DeviceDAO {
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

//deviceDAO := builder.datastoreRegistry.GetDeviceDAO()
//return store.NewGormDeviceConfigStore(deviceDAO, 3)
//}

func (builder *ClusterConfigBuilder) createDeviceDataStore(storeType int) datastore.DeviceDataStore {
	var deviceDataStore datastore.DeviceDataStore
	switch storeType {
	case datastore.GORM_STORE:
		gormDB := gormds.NewGormDB(builder.app.Logger, builder.app.GORMInitParams).Connect(false)
		deviceDataStore = gormds.NewGormDeviceDataStore(builder.app.Logger,
			gormDB, builder.app.GORMInitParams.Engine,
			builder.app.Location)
	case datastore.RAFT_STORE:
		deviceDataStore = cluster.NewRaftDeviceDataDAO(builder.app.Logger, builder.raftNode)
	case datastore.REDIS_TS:
		deviceDataStore = redis.NewRedisDataStore(":6379", "")
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
