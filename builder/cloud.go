// +build cloud

package builder

import (
	"crypto/rsa"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/cluster"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/datastore/gorm"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/jeremyhahn/cropdroid/state"
	"github.com/jeremyhahn/cropdroid/webservice/rest"
)

type CloudConfigBuilder struct {
	app    *app.App
	params *cluster.ClusterParams
	ConfigBuilder
}

func NewCloudConfigBuilder(_app *app.App, params *cluster.ClusterParams) ConfigBuilder {
	return &CloudConfigBuilder{app: _app, params: params}
}

func (builder *CloudConfigBuilder) Build() (config.ServerConfig, service.ServiceRegistry, []rest.RestService,
	state.ControllerIndex, state.ChannelIndex, error) {

	datastoreRegistry := gorm.NewGormRegistry(builder.app.Logger, builder.app.GORM)
	mapperRegistry := mapper.CreateRegistry()
	serviceRegistry := service.CreateClusterServiceRegistry(builder.app, datastoreRegistry, mapperRegistry)

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

	controllerIndexMap := make(map[int]config.ControllerConfig, 0)
	channelIndexMap := make(map[int]config.ChannelConfig, 0)

	// Build JWT service
	farmDAO := gorm.NewFarmDAO(builder.app.Logger, builder.app.GORM)
	jsonWriter := rest.NewJsonWriter()
	rsaKeyPair, err := app.CreateRsaKeyPair(builder.app.Logger, builder.app.KeyDir, rsa.PSSSaltLengthAuto)
	if err != nil {
		builder.app.Logger.Fatal(err)
	}

	jwtService := service.CreateJsonWebTokenService(builder.app, farmDAO, mapperRegistry.GetControllerMapper(), serviceRegistry, jsonWriter, 525960, rsaKeyPair) // 1 year jwt expiration
	if err != nil {
		builder.app.Logger.Fatal(err)
	}
	serviceRegistry.SetJsonWebTokenService(jwtService)

	configService := service.NewConfigService(builder.app, datastoreRegistry)
	serviceRegistry.SetConfigService(configService)

	controllerIndex := state.CreateControllerIndex(controllerIndexMap)
	channelIndex := state.CreateChannelIndex(channelIndexMap)

	restServiceRegistry := rest.NewClusterRestServiceRegistry(mapperRegistry, serviceRegistry)
	restServices := restServiceRegistry.GetRestServices()

	return serverConfig, serviceRegistry, restServices, controllerIndex, channelIndex, nil
}
