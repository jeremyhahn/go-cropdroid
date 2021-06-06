package service

import (
	"fmt"
	"strings"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/controller"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/cockroach"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type FarmFactory struct {
	app                       *app.App
	farmStore                 state.FarmStorer
	configStore               state.ConfigStorer
	controllerMapper          mapper.ControllerMapper
	changefeeders             map[string]datastore.Changefeeder
	controllerIndexMap        map[int]config.ControllerConfig
	channelIndexMap           map[int]config.ChannelConfig
	datastoreRegistry         datastore.DatastoreRegistry
	serviceRegistry           ServiceRegistry
	farmProvisionerChan       chan config.FarmConfig
	farmTickerProvisionerChan chan int
}

func NewFarmFactory(app *app.App, datastoreRegistry datastore.DatastoreRegistry, serviceRegistry ServiceRegistry,
	farmStore state.FarmStorer, configStore state.ConfigStorer, controllerMapper mapper.ControllerMapper,
	changefeeders map[string]datastore.Changefeeder, farmProvisionerChan chan config.FarmConfig,
	farmTickerProvisionerChan chan int) *FarmFactory {

	return &FarmFactory{
		app:                       app,
		farmStore:                 farmStore,
		configStore:               configStore,
		controllerMapper:          controllerMapper,
		changefeeders:             changefeeders,
		controllerIndexMap:        make(map[int]config.ControllerConfig, 0),
		channelIndexMap:           make(map[int]config.ChannelConfig, 0),
		datastoreRegistry:         datastoreRegistry,
		serviceRegistry:           serviceRegistry,
		farmProvisionerChan:       farmProvisionerChan,
		farmTickerProvisionerChan: farmTickerProvisionerChan}
}

func (fb *FarmFactory) RunProvisionerConsumer() {
	for {
		select {
		case farmConfig := <-fb.farmProvisionerChan:
			fb.app.Logger.Debugf("[FarmFactory.WatchProvisionerChan] Processing provisioner request...")
			farmConfigChangeChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
			farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
			farmService, err := fb.BuildService(farmConfig, farmConfigChangeChan, farmStateChangeChan)
			if err != nil {
				fb.app.Logger.Errorf("[FarmFactory.WatchProvisionerChan] Error: %s", err)
			}
			fb.serviceRegistry.AddFarmService(farmService)
			fb.farmTickerProvisionerChan <- farmConfig.GetID()
		default:
			fb.app.Logger.Error("[FarmFactory.WatchProvisionerChan] Error: Unable to process provisioner request")
		}
	}
}

func (fb *FarmFactory) BuildService(farmConfig config.FarmConfig,
	farmConfigChangeChan chan config.FarmConfig, farmStateChangeChan chan state.FarmStateMap) (FarmService, error) {

	farmID := farmConfig.GetID()
	farmConfigChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
	farmStateChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
	//controllerStateChan := make(chan map[string]state.ControllerStateMap, common.BUFFERED_CHANNEL_SIZE)
	controllerStateDeltaChan := make(chan map[string]state.ControllerStateDeltaMap, common.BUFFERED_CHANNEL_SIZE)

	farmDAO := fb.datastoreRegistry.GetFarmDAO()
	controllerConfigDAO := fb.datastoreRegistry.GetControllerConfigDAO()

	farmService, err := CreateFarmService(fb.app, farmDAO, controllerConfigDAO, fb.farmStore, fb.configStore,
		farmConfig, fb.controllerMapper, fb.serviceRegistry, farmConfigChan, farmConfigChangeChan, farmStateChan,
		/*controllerStateChan,*/ farmStateChangeChan, controllerStateDeltaChan)
	if err != nil {
		return nil, err
	}

	// Build farm controller and channel indexes
	controllers := farmConfig.GetControllers()
	controllerServices := make([]common.ControllerService, 0, len(controllers))
	for i, _controller := range controllers {
		//if controller.GetType() == common.CONTROLLER_TYPE_SERVER {
		//	continue
		//}

		// Build farm's microcontrollers (replace farmService.BuildControllerServices())
		fb.app.Logger.Debugf("Building %s controller service", _controller.GetType())
		var microcontroller controller.Controller
		controllerType := _controller.GetType()
		if !_controller.IsEnabled() {
			fb.app.Logger.Warningf("%s service disabled...", strings.Title(_controller.GetType()))
			continue
		}
		if farmConfig.GetMode() == common.CONFIG_MODE_VIRTUAL {
			microcontroller = controller.NewVirtualController(fb.app, state.NewFarmStateMap(farmID), "", controllerType)
		} else {
			microcontroller = controller.NewHttpController(fb.app, _controller.GetURI(), controllerType)
		}
		controllerConfig, err := farmConfig.GetController(controllerType)
		if err != nil {
			return nil, err
		}
		service, err := NewMicroControllerService(fb.app, controllerConfig, fb.app.MetricDatastore,
			fb.controllerMapper, microcontroller, fb.app.ControllerIndex, farmService,
			fb.serviceRegistry.GetConditionService(), fb.serviceRegistry.GetScheduleService(),
			fb.serviceRegistry.GetEventLogService(), fb.serviceRegistry.GetNotificationService())
		if err != nil {
			fb.app.Logger.Fatalf("Unable to create %s service: ", controllerType, err.Error())
		}
		controllerServices = append(controllerServices, service)

		// Initialize changefeeds
		if len(fb.changefeeders) > 0 && fb.app.DatastoreType == "cockroach" {
			if _, ok := fb.changefeeders[_controller.GetType()]; !ok {
				tableName := fmt.Sprintf("state_%d", _controller.GetID())
				fb.changefeeders[_controller.GetType()] = cockroach.NewCockroachChangefeed(fb.app, tableName)
			}
		}

		// Initialize controller and channel global indexes
		fb.controllerIndexMap[_controller.GetID()] = &controllers[i]
		channels := _controller.GetChannels()
		for _, channel := range channels {
			fb.channelIndexMap[channel.GetID()] = &channels[i]
		}
	}

	fb.serviceRegistry.SetControllerServices(farmID, controllerServices)

	if err := fb.serviceRegistry.AddFarmService(farmService); err != nil {
		return nil, err
	}

	return farmService, nil
}

func (fb *FarmFactory) GetFarmProvisionerChan() chan config.FarmConfig {
	return fb.farmProvisionerChan
}

func (fb *FarmFactory) GetControllerIndexMap() map[int]config.ControllerConfig {
	return fb.controllerIndexMap
}

func (fb *FarmFactory) GetChannelIndexMap() map[int]config.ChannelConfig {
	return fb.channelIndexMap
}
