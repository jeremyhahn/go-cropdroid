//go:build cluster && !cloud
// +build cluster,!cloud

package builder

import (
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	clusterutil "github.com/jeremyhahn/go-cropdroid/cluster/util"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	gormds "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/datastore/redis"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/provisioner"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type ClusterConfigBuilder struct {
	app               *app.App
	params            *clusterutil.ClusterParams
	appStateTTL       int
	appStateTick      int
	datastoreRegistry dao.Registry
	serviceRegistry   service.ClusterServiceRegistry
	mapperRegistry    mapper.MapperRegistry
	gossipNode        cluster.GossipNode
	raftNode          cluster.RaftNode
}

func NewClusterConfigBuilder(app *app.App, params *clusterutil.ClusterParams,
	gossipNode cluster.GossipNode, raftNode cluster.RaftNode,
	dataStore string, appStateTTL, appStateTick int) *ClusterConfigBuilder {

	return &ClusterConfigBuilder{
		app:        app,
		params:     params,
		gossipNode: gossipNode,
		raftNode:   raftNode}
}

func (builder *ClusterConfigBuilder) Build() (
	mapper.MapperRegistry,
	service.ClusterServiceRegistry,
	rest.RestServiceRegistry,
	chan uint64,
	error) {

	builder.mapperRegistry = mapper.CreateRegistry()
	builder.datastoreRegistry = raft.NewRaftRegistry(
		builder.app.Logger, builder.app.IdGenerator, builder.raftNode,
	)
	builder.serviceRegistry = service.CreateClusterServiceRegistry(
		builder.app, builder.datastoreRegistry, builder.mapperRegistry,
		builder.gossipNode, builder.raftNode)

	passwordHasher := util.NewPasswordHasher()

	configInitializer := dao.NewConfigInitializer(builder.app.Logger,
		builder.app.IdGenerator, builder.app.Location,
		builder.datastoreRegistry, passwordHasher, builder.app.Mode)

	builder.gossipNode.SetInitializer(configInitializer)

	serverDAO := builder.datastoreRegistry.(*raft.RaftDaoRegistry).GetServerDAO()
	roleDAO := builder.datastoreRegistry.GetRoleDAO()
	orgDAO := builder.datastoreRegistry.GetOrganizationDAO()
	systemEventLogDAO := builder.datastoreRegistry.GetEventLogDAO()

	// Build and add event log service to registry
	raftParams := builder.raftNode.GetParams()
	systemEventLogService := service.NewEventLogService(builder.app, systemEventLogDAO, raftParams.RaftOptions.SystemClusterID)
	builder.serviceRegistry.AddEventLogService(systemEventLogService)

	var serverConfig = config.NewServer()
	var err error
	if builder.params.Initialize && builder.raftNode.IsLeader(builder.params.ClusterID) {

		builder.app.Logger.Info("Initializing cluster database")

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

		serverConfig.SetID(builder.params.ClusterID)
		if serverErr := serverDAO.Save(serverConfig); serverErr != nil {
			builder.app.Logger.Fatal(serverErr)
		}

		if farmConfig != nil {
			serverConfig.AddFarmRef(farmConfig.ID)
			serverDAO.Save(serverConfig)
		}
	} else {
		serverConfig, err = serverDAO.Get(raftParams.RaftOptions.SystemClusterID, common.CONSISTENCY_LOCAL)
		if err != nil {
			builder.app.Logger.Warning(err)
		}
	}

	farmFactory := service.NewFarmFactoryCluster(
		builder.app, builder.datastoreRegistry, builder.serviceRegistry,
		builder.mapperRegistry.GetDeviceMapper(),
		builder.params.GetFarmProvisionerChan(),
		builder.params.GetFarmTickerProvisionerChan())
	builder.serviceRegistry.SetFarmFactory(farmFactory)

	farmProvisioner := provisioner.NewRaftFarmProvisioner(
		builder.app, builder.gossipNode, builder.app.Location,
		builder.datastoreRegistry.NewFarmDAO(),
		builder.datastoreRegistry.GetUserDAO(),
		builder.datastoreRegistry.GetPermissionDAO(),
		builder.mapperRegistry.GetUserMapper(),
		//		builder.params.GetFarmProvisionerChan(),
		//		builder.params.GetFarmDeprovisionerChan(),
		configInitializer)
	builder.serviceRegistry.SetFarmProvisioner(farmProvisioner)

	// Load all farms that belong to an organization
	err = orgDAO.ForEachPage(query.NewPageQuery(), func(entities []*config.OrganizationStruct) error {
		for _, org := range entities {
			for _, farmConfig := range org.GetFarms() {
				builder.app.Logger.Infof("Creating organization clustered farm service. farm.ID: %d, farm.Name: %s, org.Name: %s",
					farmConfig.ID, farmConfig.Name, org.Name)

				builder.createAndRunFarmCluster(farmFactory, farmConfig)
			}
		}
		return nil
	}, common.CONSISTENCY_LOCAL)
	if err != nil {
		builder.app.Logger.Error(err)
	}

	// Load all standalone farms that do NOT belong to an organization
	if serverConfig != nil && serverConfig.FarmRefs != nil {
		farmDAO := builder.createFarmConfigDAO(config.RAFT_DISK_STORE)
		for _, farmID := range serverConfig.FarmRefs {
			farmConfig, err := farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
			if err != nil {
				builder.app.Logger.Fatal(err)
			}
			builder.createAndRunFarmCluster(farmFactory, farmConfig)
		}
	}

	// Listen for new farm provisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-builder.params.GetFarmProvisionerChan():
				builder.app.Logger.Debugf("Processing new farm provisioning request: farmConfig=%+v", farmConfig)
				builder.createAndRunFarmCluster(farmFactory, farmConfig)
				// Rebuild the webservice routes to include new REST and WebSocket endpoints
				builder.params.GetFarmTickerProvisionerChan() <- farmConfig.Identifier()
			}
		}
	}()

	// Listen for new farm deprovisioning requests
	go func() {
		for {
			select {
			case farmConfig := <-builder.params.GetFarmDeprovisionerChan():
				farmID := farmConfig.Identifier()
				builder.app.Logger.Debugf("Processing deprovisioning request for farm %d", farmID)
				farmService := builder.serviceRegistry.GetFarmService(farmID)
				if farmService == nil {
					builder.app.Logger.Errorf("Farm not found: %d", farmID)
					continue
				}
				farmService.Stop()
				farmDAO := builder.createFarmConfigDAO(config.RAFT_DISK_STORE)
				if err := farmDAO.Delete(farmConfig.(*config.FarmStruct)); err != nil {
					builder.app.Logger.Error(err)
				}
				builder.params.GetFarmTickerProvisionerChan() <- farmID
			}
		}
	}()

	// Add organization service to the service registry
	organizationService := service.NewOrganizationService(
		builder.app.Logger, builder.app.IdGenerator,
		builder.datastoreRegistry.GetOrganizationDAO(),
		builder.mapperRegistry.GetUserMapper())
	builder.serviceRegistry.SetOrganizationService(organizationService)

	// Create the application RSA keypair
	// rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.CertDir, rsa.PSSSaltLengthAuto)
	// if err != nil {
	// 	builder.app.Logger.Fatal(err)
	// }

	// Build device and channel cache/indexes (to provide o(n) lookups when searching service registry and farm)
	// deviceIndex := state.CreateDeviceIndex(farmFactory.GetDeviceIndexMap())
	// channelIndex := state.CreateChannelIndex(farmFactory.GetChannelIndexMap())

	restServiceRegistry := rest.NewRestServiceRegistry(
		builder.app,
		roleDAO,
		builder.mapperRegistry,
		builder.serviceRegistry)

	//builder.app.KeyPair = rsaKeyPair

	return builder.mapperRegistry,
		builder.serviceRegistry,
		restServiceRegistry,
		builder.params.GetFarmTickerProvisionerChan(),
		err
}

func (builder *ClusterConfigBuilder) createAndRunFarmCluster(farmFactory service.FarmFactoryCluster, farmConfig config.Farm) {

	farmID := farmConfig.Identifier()
	farmName := farmConfig.GetName()

	stateStoreType := farmConfig.GetStateStore()
	configStoreType := farmConfig.GetConfigStore()
	dataStoreType := farmConfig.GetDataStore()

	farmChannels := &service.FarmChannels{
		FarmConfigChan:        make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
		FarmConfigChangeChan:  make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE),
		FarmStateChangeChan:   make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE),
		FarmErrorChan:         make(chan common.FarmError, common.BUFFERED_CHANNEL_SIZE),
		FarmNotifyChan:        make(chan common.FarmNotification, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateChangeChan: make(chan common.DeviceStateChange, common.BUFFERED_CHANNEL_SIZE),
		DeviceStateDeltaChan:  make(chan map[string]state.DeviceStateDeltaMap, common.BUFFERED_CHANNEL_SIZE)}

	farmStateStore := builder.createFarmStateStore(stateStoreType, farmID, farmChannels.FarmStateChangeChan)
	farmConfigDAO := builder.createFarmConfigDAO(configStoreType)

	deviceDataStore := builder.createDeviceDataStore(dataStoreType)
	deviceStateStore := builder.createDeviceStateStore(stateStoreType)

	farmEventLogDAO := raft.NewRaftEventLogDAO(builder.app.Logger,
		builder.raftNode, farmID)

	builder.app.Logger.Debugf("Farm ID: %s", farmID)
	builder.app.Logger.Debugf("Farm Name: %s", farmName)
	builder.app.Logger.Debugf("Mode: %s", builder.app.Mode)
	builder.app.Logger.Debugf("Timezone: %s", builder.app.Timezone)
	builder.app.Logger.Debugf("Polling interval: %d", builder.app.Interval)

	farmService, err := farmFactory.BuildClusterService(farmEventLogDAO,
		farmConfigDAO, farmConfig, farmStateStore, deviceStateStore, deviceDataStore, farmChannels)
	if err != nil {
		builder.app.Logger.Fatalf("Error loading farm config: %s", err)
	}
	go farmService.RunCluster()

	farmEventLogService := builder.serviceRegistry.GetEventLogService(farmID)
	startupEventLogMsg := fmt.Sprintf("Starting farm on node %d", builder.raftNode.GetParams().NodeID)
	farmEventLogService.Create(0, common.CONTROLLER_TYPE_SERVER, "System", startupEventLogMsg)
}

func (builder *ClusterConfigBuilder) createFarmStateStore(storeType int,
	farmID uint64, farmStateChangeChan chan state.FarmStateMap) state.FarmStateStorer {
	var farmStateStore state.FarmStateStorer
	switch storeType {
	case state.MEMORY_STORE:
		farmStateStore = builder.newMemoryFarmStateStore()
	case state.RAFT_STORE:
		farmStateStore = raft.NewRaftFarmStateStore(builder.app.Logger, builder.raftNode,
			farmID, farmStateChangeChan)
	default:
		farmStateStore = builder.newMemoryFarmStateStore()
	}
	return farmStateStore
}

func (builder *ClusterConfigBuilder) createFarmConfigDAO(storeType int) dao.FarmDAO {
	var farmDAO dao.FarmDAO
	switch storeType {
	case config.MEMORY_STORE, config.GORM_STORE:
		gormDB := gormds.NewGormDB(builder.app.Logger, builder.app.GORMInitParams).Connect(false)
		farmDAO = gormds.NewFarmDAO(builder.app.Logger, gormDB,
			builder.app.IdGenerator)
	case config.RAFT_MEMORY_STORE, config.RAFT_DISK_STORE:
		serverDAO := raft.NewRaftServerDAO(builder.app.Logger,
			builder.raftNode, builder.app.ClusterID)
		userDAO := raft.NewGenericRaftDAO[*config.User](builder.app.Logger,
			builder.raftNode, builder.params.RaftOptions.UserClusterID).(dao.UserDAO)
		farmDAO = raft.NewRaftFarmConfigDAO(builder.app.Logger,
			builder.raftNode, serverDAO, userDAO)
	}
	return farmDAO
}

func (builder *ClusterConfigBuilder) createDeviceStateStore(storeType int) state.DeviceStateStorer {
	var deviceStateStore state.DeviceStateStorer
	switch storeType {
	case state.MEMORY_STORE:
		deviceStateStore = builder.newMemoryDeviceStateStore()
	case state.RAFT_STORE:
		deviceStateStore = raft.NewRaftDeviceStateStore(builder.app.Logger, builder.raftNode)
	default:
		deviceStateStore = builder.newMemoryDeviceStateStore()
	}
	return deviceStateStore
}

func (builder *ClusterConfigBuilder) createDeviceDataStore(storeType int) datastore.DeviceDataStore {
	var deviceDataStore datastore.DeviceDataStore
	switch storeType {
	case datastore.GORM_STORE:
		gormDB := gormds.NewGormDB(builder.app.Logger, builder.app.GORMInitParams).Connect(false)
		deviceDataStore = gormds.NewGormDeviceDataStore(builder.app.Logger,
			gormDB, builder.app.GORMInitParams.Engine,
			builder.app.Location)
	case datastore.RAFT_STORE:
		deviceDataStore = raft.NewRaftDeviceDataDAO(builder.app.Logger, builder.raftNode, 0)
	case datastore.REDIS_TS:
		deviceDataStore = redis.NewRedisDataStore(":6379", "")
	}
	return deviceDataStore
}

// Creates a new MemoryFarmStore instance
func (builder *ClusterConfigBuilder) newMemoryFarmStateStore() state.FarmStateStorer {
	return state.NewMemoryFarmStore(builder.app.Logger, 1,
		builder.appStateTTL, time.Duration(builder.appStateTick))
}

// Creates a new MemoryDeviceStore instance
func (builder *ClusterConfigBuilder) newMemoryDeviceStateStore() state.DeviceStateStorer {
	return state.NewMemoryDeviceStore(builder.app.Logger, 1,
		builder.appStateTTL, time.Duration(builder.appStateTick))
}
