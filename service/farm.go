package service

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/config/dao"
	"github.com/jeremyhahn/cropdroid/config/observer"
	"github.com/jeremyhahn/cropdroid/controller"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/state"
)

type DefaultFarmService struct {
	configMutex          *sync.RWMutex
	stateMutex           *sync.RWMutex
	app                  *app.App
	dao                  dao.FarmDAO
	stateStore           state.FarmStorer
	configStore          state.ConfigStorer
	configClusterID      uint64
	config               config.FarmConfig
	state                state.FarmStateMap
	controllerMapper     mapper.ControllerMapper
	controllerConfigDAO  dao.ControllerConfigDAO
	serviceRegistry      ServiceRegistry
	notificationService  NotificationService
	controllerServices   []common.ControllerService
	farmConfigChan       chan config.FarmConfig
	farmConfigChangeChan chan config.FarmConfig
	farmStateChan        chan state.FarmStateMap
	farmStateChangeChan  chan state.FarmStateMap
	//controllerStateChan      chan map[string]state.ControllerStateMap
	controllerStateDeltaChan chan map[string]state.ControllerStateDeltaMap
	running                  bool
	FarmService
	observer.FarmConfigObserver
}

func NewFarmService(_app *app.App, farmDAO dao.FarmDAO, controllerConfigDAO dao.ControllerConfigDAO,
	farmConfig config.FarmConfig, controllerMapper mapper.ControllerMapper,
	serviceRegistry ServiceRegistry, farmConfigChan chan config.FarmConfig,
	farmConfigChangeChan chan config.FarmConfig, farmStateChan chan state.FarmStateMap,
	farmStateChangeChan chan state.FarmStateMap, controllerStateDeltaChan chan map[string]state.ControllerStateDeltaMap) (FarmService, error) {

	farmState := state.NewFarmStateMap(farmConfig.GetID())
	_app.FarmStore.Put(farmConfig.GetID(), farmState)

	return CreateFarmService(_app, farmDAO, controllerConfigDAO, _app.FarmStore, _app.ConfigStore, farmConfig, controllerMapper, serviceRegistry,
		farmConfigChan, farmConfigChangeChan, farmStateChan, farmStateChangeChan, controllerStateDeltaChan)
}

func CreateFarmService(app *app.App, farmDAO dao.FarmDAO, controllerConfigDAO dao.ControllerConfigDAO,
	stateStore state.FarmStorer, configStore state.ConfigStorer,
	farmConfig config.FarmConfig, controllerMapper mapper.ControllerMapper, serviceRegistry ServiceRegistry,
	farmConfigChan chan config.FarmConfig, farmConfigChangeChan chan config.FarmConfig,
	farmStateChan chan state.FarmStateMap, farmStateChangeChan chan state.FarmStateMap,
	/*controllerStateChan chan map[string]state.ControllerStateMap,*/
	controllerStateDeltaChan chan map[string]state.ControllerStateDeltaMap) (FarmService, error) {

	farmID := farmConfig.GetID()

	// Only used when clustering enabled
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(fmt.Sprintf("%d-%d", farmConfig.GetOrganizationID(), farmConfig.GetID())))
	configClusterID := clusterHash.Sum64()

	_state, err := stateStore.Get(farmID)
	if err != nil {
		app.Logger.Warningf("[CreateFarmState] Error (farmID=%d): %s", farmID, err)
		_state = state.NewFarmStateMap(farmID)
	}
	if _state == nil {
		_state = state.NewFarmStateMap(farmID)
	}

	return &DefaultFarmService{
		configMutex:          &sync.RWMutex{},
		stateMutex:           &sync.RWMutex{},
		app:                  app,
		dao:                  farmDAO,
		stateStore:           stateStore,
		configStore:          configStore,
		config:               farmConfig,
		configClusterID:      configClusterID,
		state:                _state,
		controllerMapper:     controllerMapper,
		controllerConfigDAO:  controllerConfigDAO,
		serviceRegistry:      serviceRegistry,
		notificationService:  serviceRegistry.GetNotificationService(),
		farmConfigChan:       farmConfigChan,
		farmConfigChangeChan: farmConfigChangeChan,
		farmStateChan:        farmStateChan,
		farmStateChangeChan:  farmStateChangeChan,
		//controllerStateChan:      controllerStateChan,
		controllerStateDeltaChan: controllerStateDeltaChan,
		running:                  false}, nil
}

func (farm *DefaultFarmService) GetFarmID() int {
	return farm.config.GetID()
}

func (farm *DefaultFarmService) GetConfigClusterID() uint64 {
	return farm.configClusterID
}

func (farm *DefaultFarmService) GetControllers() []common.ControllerService {
	return farm.controllerServices
}

func (farm *DefaultFarmService) GetConfig() config.FarmConfig {

	farm.app.Logger.Debugf("[FarmService.GetConfig] Getting farm configuration. farm.id=%d, farm.configClusterID=%d",
		farm.config.GetID(), farm.configClusterID)

	if farm.app.Mode == common.MODE_CLUSTER {

		conf, err := farm.configStore.Get(farm.configClusterID)
		if err != nil {
			farm.app.Logger.Errorf("[FarmService.GetConfig] Error: %s", err)
			return nil
		}
		return conf

	} else {

		farm.configMutex.RLock()
		defer farm.configMutex.RUnlock()
		return farm.config
	}
}

func (farm *DefaultFarmService) SetConfig(farmConfig config.FarmConfig) error {

	// Called by ConditionService and ScheduleService

	if err := farm.configStore.Put(farm.configClusterID, farmConfig); err != nil {
		farm.app.Logger.Errorf("[FarmService.SetConfig] Error: %s", err)
		return err
	}

	// Don't modify Raft state!
	//if farm.app.Mode == common.CONFIG_MODE_VIRTUAL || farm.app.Mode == common.MODE_STANDALONE {
	farm.configMutex.Lock()
	farm.config = farmConfig
	farm.configMutex.Unlock()
	//}

	farm.PublishConfig()
	return nil
}

func (farm *DefaultFarmService) GetState() state.FarmStateMap {
	//farm.stateMutex.RLock()
	//defer farm.stateMutex.RUnlock()
	//return farm.state

	// Reading from the stateStore vs cached farm.state avoids having to use
	// locks to protect it. It also guarantees consistent read across all
	// nodes in the Raft cluster as well as provides the "last state" while
	// inside poll().
	state, err := farm.stateStore.Get(farm.config.GetID())
	if err != nil {
		farm.app.Logger.Errorf("[FarmService.GetState] Error: %s", err)
		return nil
	}
	return state
}

/*
func (farm *DefaultFarmService) SetState(state state.FarmStateMap) {
	farm.stateMutex.Lock()
	defer farm.stateMutex.Unlock()
	farm.state = state
}*/

// Publishes the entire farm configuration (all controllers)
func (farm *DefaultFarmService) PublishConfig() error {
	select {
	case farm.farmConfigChan <- farm.config:
		farm.app.Logger.Error("PublishConfig fired! %+v", farm.config)
	default:
		errmsg := "Farm config channel buffer full, discarding update!"
		farm.app.Logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

// Publishes the entire farm state (all controllers)
func (farm *DefaultFarmService) PublishState() error {
	select {
	case farm.farmStateChan <- farm.state:
		farm.app.Logger.Error("PublishState fired!")
	default:
		errmsg := "Farm state channel buffer full, discarding update!"
		farm.app.Logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

/*
// Publishes a full ControllerStateMap that contains ALL metrics and/or channels
func (farm *DefaultFarmService) PublishControllerState(controllerState map[string]state.ControllerStateMap) error {
	select {
	case farm.controllerStateChan <- controllerState:
		farm.app.Logger.Error("PublishControllerState fired!")
	default:
		errmsg := "Controller state channel buffer full, discarding update!"
		farm.app.Logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}*/

// Publishes a ControllerStateMap that contains ONLY the metrics and/or channels that've changed
func (farm *DefaultFarmService) PublishControllerDelta(controllerState map[string]state.ControllerStateDeltaMap) error {
	select {
	case farm.controllerStateDeltaChan <- controllerState:
		farm.app.Logger.Error("PublishControllerDelta fired!")
	default:
		errmsg := "Controller delta state channel buffer full, discarding update!"
		farm.app.Logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

func (farm *DefaultFarmService) WatchConfig() <-chan config.FarmConfig {
	return farm.farmConfigChan
}

func (farm *DefaultFarmService) WatchState() <-chan state.FarmStateMap {
	return farm.farmStateChan
}

/*func (farm *DefaultFarmService) WatchControllerState() <-chan map[string]state.ControllerStateMap {
	return farm.controllerStateChan
}*/

func (farm *DefaultFarmService) WatchControllerDeltas() <-chan map[string]state.ControllerStateDeltaMap {
	return farm.controllerStateDeltaChan
}

// WatchFarmStateChange is intended to be run within a goroutine that watches the farmStateChangeChan
// and creates a delta that gets published to connected clients anytime the farm state changes.
func (farm *DefaultFarmService) WatchFarmStateChange() {
	farm.app.Logger.Debugf("[Farm.WatchFarmStateChange] Watching for farm state changes (farm.id=%d)", farm.GetFarmID())
	for {
		select {
		case newFarmState := <-farm.farmStateChangeChan:
			lastState := farm.GetState()
			newControllerStates := newFarmState.GetControllers()

			farm.app.Logger.Debugf("[Farm.WatchFarmStateChange] New farm state published. last=%+v, new=%+v",
				lastState, newControllerStates)

			if lastState == nil {
				for controllerType, stateMap := range newControllerStates {
					farm.SetControllerState(controllerType, stateMap)
				}
				continue
			}
			controllerServices, err := farm.serviceRegistry.GetControllerServices(farm.GetFarmID())
			if err != nil {
				farm.app.Logger.Errorf("[Farm.WatchFarmStateChange] Error: %s", err)
			}
			for _, controller := range controllerServices {
				controllerType := controller.GetControllerType()
				newControllerState := newControllerStates[controllerType]
				_, err := farm.OnControllerStateChange(lastState, controllerType, newControllerState)
				if err == state.ErrControllerNotFound {
					farm.SetControllerState(controllerType, newControllerState)
				}
				if err != nil {
					farm.app.Logger.Errorf("[Farm.WatchFarmStateChange] Error: %s", err)
				}
			}
		}
	}
}

func (farm *DefaultFarmService) WatchFarmConfigChange() {
	farm.app.Logger.Debugf("[Farm.WatchFarmConfigChange] Watching for farm config changes (configClusterID=%d)", farm.GetConfigClusterID())
	for {
		select {
		case newConfig := <-farm.farmConfigChangeChan:
			farm.app.Logger.Debugf("[Farm.WatchFarmConfigChange] New config change for farm %d (configClusterID=%d)",
				farm.GetFarmID(), farm.GetConfigClusterID())
			farm.configMutex.Lock()
			farm.config = newConfig
			farm.configMutex.Unlock()
			farm.PublishConfig()
		}
	}
}

func (farm *DefaultFarmService) OnControllerStateChange(lastState state.FarmStateMap,
	controllerType string, newControllerState state.ControllerStateMap) (state.ControllerStateDeltaMap, error) {

	farm.app.Logger.Debugf("[farm.OnControllerStateChange] newControllerState=%+v", newControllerState)

	var delta state.ControllerStateDeltaMap
	newChannelMap := make(map[int]int, len(newControllerState.GetChannels()))
	for i, channel := range newControllerState.GetChannels() {
		newChannelMap[i] = channel
	}
	if lastState == nil {
		delta = state.CreateControllerStateDeltaMap(newControllerState.GetMetrics(), newChannelMap)
	} else {
		d, err := farm.state.Diff(controllerType, newControllerState.GetMetrics(), newChannelMap)
		if err != nil {
			farm.app.Logger.Errorf("[farm.OnControllerStateChange] Error: %s", err)
			return nil, err
		}
		delta = d
	}
	if delta == nil {
		return nil, nil
	}
	if err := farm.PublishControllerDelta(map[string]state.ControllerStateDeltaMap{controllerType: delta}); err != nil {
		return nil, err
	}
	return delta, nil
}

func (farm *DefaultFarmService) SetControllerConfig(controllerConfig config.ControllerConfig) {
	farm.config.SetController(controllerConfig)
	farm.config.ParseConfigs()
}

func (farm *DefaultFarmService) SetControllerState(controllerType string, controllerState state.ControllerStateMap) {
	farm.app.Logger.Debugf("[FarmService.SetControllerState] controllerType: %s, controllerState: %+v",
		controllerType, controllerState)

	// Synchronization handled by farm.state.SetController
	farm.state.SetController(controllerType, controllerState)

	// Putting on every controller state change corrupts raft statemachine (other controller states are saved with empty state)
	//farm.stateStore.Put(farm.config.GetID(), farm.state)
}

func (farm *DefaultFarmService) SetConfigValue(session Session, farmID, controllerID int, key, value string) error {

	farm.app.Logger.Debugf("[FarmService.SetConfigValue] Setting config farmID=%d, controllerID=%d, key=%s, value=%s",
		farmID, controllerID, key, value)

	configItem, err := farm.controllerConfigDAO.Get(controllerID, key)
	if err != nil {
		return err
	}
	configItem.SetValue(value)
	farm.controllerConfigDAO.Save(configItem)
	farm.app.Logger.Debugf("[FarmService.SetConfigValue] Saved configuration item: %+v", configItem)

	farmService, ok := farm.serviceRegistry.GetFarmService(farmID)
	if !ok {
		err := fmt.Errorf("Unable to locate farm service in service registry! farm.id=%d", farmID)
		farm.app.Logger.Errorf("[FarmService.SetConfigValue] Error: %s", err)
		return err
	}

	farmConfig, err := farm.dao.Get(farmService.GetFarmID())
	if err != nil {
		return nil
	}

	farm.farmConfigChangeChan <- farmConfig
	return nil
}

func (farm *DefaultFarmService) SetMetricValue(controllerType string, key string, value float64) error {

	farmState := farm.GetState()

	if err := farmState.SetMetricValue(controllerType, key, value); err != nil {
		farm.app.Logger.Errorf("[Farm.SetMetricValue] Error: %s", err)
		return err
	}

	farm.stateStore.Put(farm.config.GetID(), farmState)

	//metricDelta := make(map[string]float64, 1)
	//metricDelta[key] = value
	metricDelta := map[string]float64{key: value}
	channelDelta := make(map[int]int, 0)
	delta := state.CreateControllerStateDeltaMap(metricDelta, channelDelta)
	if err := farm.PublishControllerDelta(map[string]state.ControllerStateDeltaMap{controllerType: delta}); err != nil {
		return err
	}

	return nil
}

func (farm *DefaultFarmService) SetSwitchValue(controllerType string, channelID int, value int) error {

	farmState := farm.GetState()

	if err := farmState.SetChannelValue(controllerType, channelID, value); err != nil {
		farm.app.Logger.Errorf("[Farm.SetChannelValue] Error: %s", err)
		return err
	}

	farm.stateStore.Put(farm.config.GetID(), farmState)

	metricDelta := make(map[string]float64, 0)
	channelDelta := map[int]int{channelID: value}
	delta := state.CreateControllerStateDeltaMap(metricDelta, channelDelta)
	if err := farm.PublishControllerDelta(map[string]state.ControllerStateDeltaMap{controllerType: delta}); err != nil {
		return err
	}

	return nil
}

func (farm *DefaultFarmService) Run() {
	if farm.running == true {
		farm.app.Logger.Errorf("[Farm.Run] Farm %d already running!", farm.GetFarmID())
		return
	}
	farm.running = true

	go farm.WatchFarmStateChange()
	go farm.WatchFarmConfigChange()

	if !farm.app.DebugFlag {
		// Wait for top of the minute
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				_, _, secs := time.Now().Clock()
				if secs == 0 {
					ticker.Stop()
					return
				}
				farm.app.Logger.Infof("Waiting for top of the minute... %d sec left", 60-secs)
			}
		}
	}

	farm.poll()
	farm.Poll()
}

func (farm *DefaultFarmService) Poll() {
	if farm.config.GetInterval() > 0 {
		ticker := time.NewTicker(time.Duration(farm.config.GetInterval()) * time.Second)
		quit := make(chan struct{})
		for {
			select {
			case <-ticker.C:
				farm.poll()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}
}

func (farm *DefaultFarmService) BuildControllerServices() ([]common.ControllerService, error) {

	controllers := farm.config.GetControllers()
	controllerServices := make([]common.ControllerService, 0)
	for _, _controller := range controllers {

		farm.app.Logger.Debugf("Building %s controller service", _controller.GetType())

		var microcontroller controller.Controller
		controllerType := _controller.GetType()

		if !_controller.IsEnabled() {
			farm.app.Logger.Warningf("%s service disabled...", strings.Title(_controller.GetType()))
			continue
		}

		if farm.config.GetMode() == common.CONFIG_MODE_VIRTUAL {
			microcontroller = controller.NewVirtualController(farm.app, farm.state, "", controllerType)
		} else {
			microcontroller = controller.NewHttpController(farm.app, _controller.GetURI(), controllerType)
		}

		controllerConfig, err := farm.config.GetController(controllerType)
		if err != nil {
			return nil, err
		}

		service, err := NewMicroControllerService(farm.app, controllerConfig, farm.app.MetricDatastore,
			farm.controllerMapper, microcontroller, farm.app.ControllerIndex, farm,
			farm.serviceRegistry.GetConditionService(), farm.serviceRegistry.GetScheduleService(), farm.serviceRegistry.GetEventLogService(),
			farm.serviceRegistry.GetNotificationService())

		if err != nil {
			farm.app.Logger.Fatalf("Unable to create %s service: ", controllerType, err.Error())
		}

		controllerServices = append(controllerServices, service)
	}
	farm.controllerServices = controllerServices
	return controllerServices, nil
}

func (farm *DefaultFarmService) poll() {

	farmID := farm.GetConfig().GetID()

	farm.app.Logger.Debugf("Polling farm: %d", farmID)

	controllerServices, err := farm.serviceRegistry.GetControllerServices(farmID)
	if err != nil {
		farm.app.Logger.Error(err)
		return
	}

	lastState := farm.GetState()

	for _, controller := range controllerServices {
		controllerType := controller.GetControllerType()
		newControllerState, err := controller.Poll()
		if err != nil {
			farm.app.Logger.Errorf("[Farm.poll] Error polling controller state: %s", err)
		}
		if newControllerState == nil {
			farm.app.Logger.Errorf("[Farm.poll] Empty %s controller state", controllerType)
			continue
		}
		_, err = farm.OnControllerStateChange(lastState, controllerType, newControllerState)
		if err != nil {
			farm.error("farm.poll", "farm.poll", err)
			return
		}
		farm.SetControllerState(controllerType, newControllerState)
		if farm.app.MetricDatastore != nil {
			if err := farm.app.MetricDatastore.Save(controller.GetControllerConfig().GetID(), newControllerState); err != nil {
				farm.app.Logger.Errorf("[FarmService.poll] Error: %s", err)
				farm.error("Farm.poll", "Farm.poll", err)
				return
			}
		}
		controller.Manage()
	}
	farm.app.Logger.Debugf("[Farm.poll] storing farm state: %s", farm.state)
	if err := farm.stateStore.Put(farmID, farm.state); err != nil {
		farm.app.Logger.Errorf("[Farm.poll] Error storing farm state: %s", err)
		return
	}
}

func (farm *DefaultFarmService) error(method, eventType string, err error) {
	farm.app.Logger.Errorf("[FarmService.%s] Error: %s", method, err)
	farm.notificationService.Enqueue(&model.Notification{
		Controller: farm.config.GetName(),
		Priority:   common.NOTIFICATION_PRIORITY_HIGH,
		Title:      farm.config.GetName(),
		Type:       eventType,
		Message:    err.Error(),
		Timestamp:  time.Now()})
}
