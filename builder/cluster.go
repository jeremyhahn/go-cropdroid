// +build cluster
// +build !cloud

package builder

import (
	"crypto/rsa"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/cockroach"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"
)

type ClusterConfigBuilder struct {
	app    *app.App
	params *cluster.ClusterParams
	ConfigBuilder
}

func NewClusterConfigBuilder(_app *app.App, params *cluster.ClusterParams) ConfigBuilder {
	return &ClusterConfigBuilder{app: _app, params: params}
}

func (builder *ClusterConfigBuilder) Build() (config.ServerConfig, service.ServiceRegistry, []rest.RestService,
	state.DeviceIndex, state.ChannelIndex, error) {

	//datastoreRegistry := gorm.NewGormRegistry(builder.app.Logger, builder.app.GORM)
	datastoreRegistry := builder.params.GetDatastoreRegistry()
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := service.CreateClusterServiceRegistry(builder.app, datastoreRegistry, mapperRegistry)

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
	//builder.app.Config = serverConfig.(*config.Server)
	//builder.app.Logger.Debugf("builder.app.Config: %+v", builder.app.Config)

	farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)
	farmConfigChangeChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
	farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)

	farmConfigs, err := farmDAO.GetAll()
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	serverConfig.SetFarms(farmConfigs)

	//farmServices := make(map[int]service.FarmService, len(farmConfigs))
	var restServices []rest.RestService

	raftConfigStore := cluster.NewRaftFarmConfigStore(builder.app.Logger, builder.app.RaftCluster)

	farmFactory := service.NewFarmFactory(builder.app, datastoreRegistry, serviceRegistry, builder.app.FarmStore,
		raftConfigStore, mapperRegistry.GetDeviceMapper(), changefeeders,
		builder.params.GetFarmProvisionerChan(), builder.params.GetFarmTickerProvisionerChan())

	go farmFactory.RunClusterProvisionerConsumer()

	for _, farmConfig := range farmConfigs {

		var conf *config.Farm = &config.Farm{}
		*conf = farmConfig
		_, err := farmFactory.BuildClusterService(conf, farmConfigChangeChan, farmStateChangeChan)
		if err != nil {
			builder.app.Logger.Fatal(err)
		}

		//builder.app.Logger.Debugf("Farm: %+v", farmConfig)

		builder.app.Logger.Debugf("Farm ID: %s", farmConfig.GetID())
		builder.app.Logger.Debugf("Farm Name: %s", farmConfig.GetName())
		builder.app.Logger.Debugf("Mode: %s", serverConfig.GetMode())
		builder.app.Logger.Debugf("Timezone: %s", serverConfig.GetTimezone())
		builder.app.Logger.Debugf("Polling interval: %d", serverConfig.GetInterval())

		serviceRegistry.GetEventLogService().Create("System", "Startup")
	}
	//serviceRegistry.SetFarmServices(farmServices)
	serviceRegistry.SetFarmFactory(farmFactory)

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

	configService := service.NewConfigService(builder.app, datastoreRegistry, serviceRegistry)
	serviceRegistry.SetConfigService(configService)

	// Build device and channel cache/indexes (to provide o(n) lookups when searching service registry and farm)
	deviceIndex := state.CreateDeviceIndex(farmFactory.GetDeviceIndexMap())
	channelIndex := state.CreateChannelIndex(farmFactory.GetChannelIndexMap())

	if builder.app.DatastoreCDC {
		changefeedService := service.NewChangefeedService(builder.app, serviceRegistry, changefeeders)
		serviceRegistry.SetChangefeedService(changefeedService)
	}

	restServiceRegistry := rest.NewClusterRestServiceRegistry(mapperRegistry, serviceRegistry)
	if restServices == nil {
		restServices = restServiceRegistry.GetRestServices()
	}

	return serverConfig, serviceRegistry, restServices, deviceIndex, channelIndex, err
}
