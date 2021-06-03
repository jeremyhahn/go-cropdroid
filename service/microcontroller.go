package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/controller"
	"github.com/jeremyhahn/cropdroid/datastore"
	"github.com/jeremyhahn/cropdroid/mapper"
	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/state"
	"github.com/jeremyhahn/cropdroid/util"
	"github.com/jeremyhahn/cropdroid/viewmodel"
)

var (
	ErrNoControllerState = errors.New("No controller state")
)

type MicroControllerService struct {
	app                 *app.App
	config              config.ControllerConfig
	mapper              mapper.ControllerMapper
	controller          controller.Controller
	controllerStateDAO  datastore.ControllerStateDAO
	farmService         FarmService
	eventLogService     EventLogService
	notificationService NotificationService
	conditionService    ConditionService
	scheduleService     ScheduleService
	backoffTable        map[int]time.Time
	//observers           []common.ControllerObserver
	common.ControllerService
	//common.ControllerObserver
}

func NewMicroControllerService(app *app.App, config config.ControllerConfig, controllerStateDAO datastore.ControllerStateDAO,
	controllerMapper mapper.ControllerMapper, controller controller.Controller, controllerIndex state.ControllerIndex,
	farmService FarmService, conditionService ConditionService, scheduleService ScheduleService, eventLogService EventLogService,
	notificationService NotificationService) (common.ControllerService, error) {

	return &MicroControllerService{
		app:                 app,
		config:              config,
		mapper:              controllerMapper,
		controller:          controller,
		controllerStateDAO:  controllerStateDAO,
		farmService:         farmService,
		eventLogService:     eventLogService,
		notificationService: notificationService,
		conditionService:    conditionService,
		scheduleService:     scheduleService,
		//observers:           make([]common.ControllerObserver, 0),
		backoffTable: make(map[int]time.Time, 0)}, nil
}

func (service *MicroControllerService) GetControllerConfig() config.ControllerConfig {
	controllerType := service.controller.GetType()
	config := service.farmService.GetConfig()
	if config != nil {
		for _, controller := range config.GetControllers() {
			if controller.GetType() == controllerType {
				return &controller
			}
		}
	}
	//service.app.Logger.Errorf("Config not found for %s controller", controllerType)
	service.app.Logger.Fatalf("Config not found for %s controller. farm.id=%d, config.id=%d",
		controllerType, service.farmService.GetFarmID(), service.farmService.GetConfigClusterID())
	return nil
}

func (service *MicroControllerService) GetControllerType() string {
	return service.controller.GetType()
}

func (service *MicroControllerService) GetState() (state.ControllerStateMap, error) {
	return service.farmService.GetState().GetController(service.controller.GetType())
}

func (service *MicroControllerService) GetView() (common.ControllerView, error) {
	controller, err := service.GetController()
	if err != nil {
		return nil, err
	}
	metrics := controller.GetMetrics()
	channels := controller.GetChannels()
	sort.SliceStable(metrics, func(i, j int) bool {
		return strings.ToLower(metrics[i].GetName()) < strings.ToLower(metrics[j].GetName())
	})
	sort.SliceStable(channels, func(i, j int) bool {
		return strings.ToLower(channels[i].GetName()) < strings.ToLower(channels[j].GetName())
	})
	return viewmodel.NewControllerView(service.app, metrics, channels), err
}

func (service *MicroControllerService) GetHistory(metric string) ([]float64, error) {
	values, err := service.controllerStateDAO.GetLast30Days(service.config.GetID(), metric)
	if err != nil {
		return nil, err
	}
	return values, nil /*  */
}

// GetController combines ControllerState and Config to return a fully populated domain model
// Controller instance including child Metric and Channel objects with their current values. This
// operation is more costly than working with the "indexed" maps and functions in FarmState.
func (service *MicroControllerService) GetController() (common.Controller, error) {
	state := service.farmService.GetState()
	if state == nil {
		return nil, ErrNoControllerState
	}
	controllerState, err := service.farmService.GetState().GetController(service.controller.GetType())
	if err != nil {
		return nil, err
	}
	controllerConfig, err := service.farmService.GetConfig().GetController(service.controller.GetType())
	if err != nil {
		return nil, err
	}
	controller, err := service.mapper.MapStateToController(controllerState, *controllerConfig)
	if err != nil {
		return nil, err
	}
	return controller, nil
}

func (service *MicroControllerService) Manage() {

	//eventType := "Manage"

	farmConfig := service.farmService.GetConfig()
	controllerConfig := service.GetControllerConfig()
	controllerType := service.controller.GetType()

	if !controllerConfig.IsEnabled() {
		service.app.Logger.Warningf("%s controller disabled...", controllerType)
		return
	}

	if farmConfig.GetMode() == common.CONFIG_MODE_MAINTENANCE {
		service.app.Logger.Warning("Maintenance mode in progres...")
		return
	}

	farmState := service.farmService.GetState()
	if farmState == nil {
		service.app.Logger.Warningf("[Microcontroller.Manage] No farm state, waiting until initialized. farm.id: %d", farmConfig.GetID())
		return
	}

	for _, err := range service.ManageMetrics(farmState) {
		service.app.Logger.Error(err.Error())
	}

	channels := controllerConfig.GetChannels()
	channelConfigs := make([]config.ChannelConfig, len(channels))
	for i := range channels {
		channelConfigs[i] = &channels[i]
	}
	for _, err := range service.ManageChannels(farmState, channelConfigs) {
		service.app.Logger.Error(err.Error())
	}
}

func (service *MicroControllerService) Poll() (state.ControllerStateMap, error) {
	controllerType := service.controller.GetType()
	eventType := "Poll"
	service.app.Logger.Debugf("[MicroControllerService.Poll] Polling %s controller state...", controllerType)
	state, err := service.controller.State()
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	state.SetID(uint64(service.config.GetID()))
	service.app.Logger.Debugf("[MicroControllerService.Poll] %s state: %+v", controllerType, state)
	return state, nil
}

func (service *MicroControllerService) SetMetricValue(key string, value float64) error {
	//eventType := "SetMetricValue"
	controllerType := service.controller.GetType()
	service.farmService.SetMetricValue(controllerType, key, value)
	/*
		controllerState, err := service.GetState()
		if err != nil {
			return err
		}
		service.farmService.SetControllerState(controllerType, controllerState)
	*/
	/*
		// This works with raft cluster
		if service.app.MetricDatastore != nil {
			if err := service.app.MetricDatastore.Save(controllerType, service.config.GetID(), service.GetState()); err != nil {
				service.app.Logger.Errorf("[FarmService.poll] Error: %s", err)
				service.error("Farm.poll", "Farm.poll", err)
				return
			}
		}*/

	// This works with sqlite memory / disk
	controllerState, err := service.GetState()
	if err != nil {
		return err
	}

	if service.app.Mode == common.CONFIG_MODE_VIRTUAL || service.app.Mode == common.MODE_STANDALONE { // TODO: consolidate mode
		//virtualController := service.controller.NewVirtualController(server.app, farmState, "", service.config.GetType())
		err := service.controller.(controller.VirtualController).WriteState(controllerState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (service *MicroControllerService) Switch(channelID, position int, logMessage string) (*common.Switch, error) {
	eventType := "Switch"
	switchPosition := util.NewSwitchPosition(position)
	channelConfig, err := service.getChannelConfig(channelID)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	if logMessage == "" {
		logMessage = fmt.Sprintf("Switching %s %s", strings.ToLower(channelConfig.GetName()), switchPosition.ToString())
	}
	service.notify(eventType, logMessage)
	service.eventLogService.Create(eventType, logMessage)
	service.app.Logger.Debug(fmt.Sprintf("Switching controller %s (channel=%d), %s", channelConfig.GetName(), channelID, switchPosition.ToString()))
	_switch, err := service.controller.Switch(channelConfig.GetChannelID(), position)
	if err != nil {
		return _switch, err
	}
	controllerType := service.controller.GetType()
	return _switch, service.farmService.SetSwitchValue(controllerType, channelID, position)
}

func (service *MicroControllerService) TimerSwitch(channelID, duration int, logMessage string) (common.TimerEvent, error) {
	eventType := "TimerSwitch"
	channelConfig, err := service.getChannelConfig(channelID)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	event, err := service.controller.TimerSwitch(channelID, duration)
	if err != nil {
		service.error(eventType, eventType, err)
		return nil, err
	}
	service.app.Logger.Debugf("MicroControllerService timed switch event: %+v", event)
	if logMessage == "" {
		logMessage = fmt.Sprintf("Starting %s timer for %d seconds", channelConfig.GetName(), duration)
	}
	service.notify(eventType, logMessage)
	service.eventLogService.Create(eventType, logMessage)
	service.app.Logger.Debug(logMessage)
	return event, nil
}

func (service *MicroControllerService) ManageMetrics(farmState state.FarmStateMap) []error {
	var errors []error
	eventType := "ALARM"

	service.app.Logger.Debugf("[MicroControllerService.ManageMetrics] Processing configured %s metrics...", service.controller.GetType())

	metricConfigs := service.GetControllerConfig().GetMetrics()

	for _, metric := range metricConfigs {

		if !metric.IsEnabled() {
			continue
		}

		metricValue, err := farmState.GetMetricValue(service.controller.GetType(), metric.GetKey())
		if err != nil {
			errors = append(errors, err)
			continue
		}

		service.app.Logger.Debugf("[MicroControllerService.ManageMetrics] notify=%t, metric=%s, value=%.2f, alarmLow=%.2f, alarmHigh=%.2f",
			metric.IsNotify(), metric.GetKey(), metricValue, metric.GetAlarmLow(), metric.GetAlarmHigh())

		if metric.IsNotify() && metricValue <= metric.GetAlarmLow() {
			message := fmt.Sprintf("%s LOW: %.2f", metric.GetName(), metricValue)
			service.notify(eventType, message)
		}

		if metric.IsNotify() && metricValue >= metric.GetAlarmHigh() {
			message := fmt.Sprintf("%s HIGH: %.2f", metric.GetName(), metricValue)
			service.notify(eventType, message)
		}
	}
	return errors
}

func (service *MicroControllerService) ManageChannels(farmState state.FarmStateMap, channels []config.ChannelConfig) []error {

	service.app.Logger.Debugf("[MicroControllerService.ManageChannels] Processing configured %s channels...", service.controller.GetType())

	var errors []error
	for _, channel := range channels {

		if len(channel.GetName()) <= 0 {
			channel.SetName(fmt.Sprintf("channel%d", channel.GetChannelID()))
		}

		if !channel.IsEnabled() {
			continue
		}

		service.app.Logger.Debugf("[MicroControllerService.ManageChannels] Processing channel %+v", channel)

		backoff := channel.GetBackoff()
		if backoff > 0 {
			if timer, ok := service.backoffTable[channel.GetID()]; ok {
				if time.Since(timer).Minutes() < float64(backoff) {
					elapsed := time.Since(timer).Minutes()
					service.app.Logger.Debugf("[MicroControllerService.ManageChannels] Waiting for %s backoff timer to expire. timer=%s, now=%s, elapsed=%.2f, backoff=%d",
						channel.GetName(), timer.String(), time.Now().String(), elapsed, backoff)
					return nil
				} else {
					delete(service.backoffTable, channel.GetID())
				}
			}
		}

		if len(channel.GetConditions()) > 0 {
			handled, err := service.handleChannelConditions(farmState, channel)
			if err != nil {
				service.app.Logger.Debugf("[MicroControllerService.ManageChannels] Error processing %s conditions: %s", channel.GetName(), err)
				errors = append(errors, err)
				continue
			}
			if handled {
				service.app.Logger.Debugf("[MicroControllerService.ManageChannels] Channel already handled by conditional, aborting schedule processing...")
				continue
			}
		}

		if len(channel.GetSchedule()) > 0 {
			if err := service.handleChannelSchedule(farmState, channel); err != nil {
				service.app.Logger.Debugf("[MicroControllerService.ManageChannels] Error processing %s schedules: %s", channel.GetName(), err)
				errors = append(errors, err)
				continue
			}
		}

	}
	return errors
}

func (service *MicroControllerService) getChannelConfig(channelID int) (config.ChannelConfig, error) {
	controllerConfig, err := service.farmService.GetConfig().GetController(service.controller.GetType())
	if err != nil {
		return nil, err
	}
	/*
		channels := controllerConfig.GetChannels()
		if channelID < 0 || channelID > len(channels) {
			return nil, fmt.Errorf("Channel ID not found: %d", channelID)
		}
		return channels[channelID], nil
	*/
	channels := controllerConfig.GetChannels()
	//channelConfigs := make([]config.ChannelConfig, len(channels))
	for _, channel := range channels {
		if channel.GetChannelID() == channelID {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("Channel ID not found: %d", channelID)
}

func (service *MicroControllerService) handleChannelConditions(state state.FarmStateMap, channel config.ChannelConfig) (bool, error) {

	var conditionController config.ControllerConfig
	var conditionMetric config.MetricConfig
	var condition config.ConditionConfig
	var value float64
	result := false
	backoff := channel.GetBackoff()
	debounce := channel.GetDebounce()

	parse := func(condition config.ConditionConfig) (config.ControllerConfig, config.MetricConfig, error) {
		for _, controller := range service.farmService.GetConfig().GetControllers() {
			for _, metric := range controller.GetMetrics() {
				if metric.GetID() == condition.GetMetricID() {
					return &controller, &metric, nil
				}
			}
		}
		errmsg := fmt.Sprintf("Orphaned condition: %+v", condition)
		service.app.Logger.Error(errmsg)
		return nil, nil, errors.New(errmsg)
	}

	for _, _condition := range channel.GetConditions() {

		service.app.Logger.Debugf("[MicroControllerService.handleChannelConditions] Parsing %s condition: %+v", channel.GetName(), _condition)

		_conditionController, _conditionMetric, err := parse(&_condition)
		if err != nil {
			return false, err
		}
		conditionController = _conditionController
		conditionMetric = _conditionMetric
		condition = &_condition

		_value, err := state.GetMetricValue(conditionController.GetType(), conditionMetric.GetKey())
		if err != nil {
			return false, err
		}
		value = _value

		_result, err := service.conditionService.IsTrue(&_condition, _value)
		if err != nil {
			return false, err
		}
		result = _result

		if _result {
			break
		}
	}

	position, err := state.GetChannelValue(service.controller.GetType(), channel.GetChannelID())
	if err != nil {
		return false, err
	}

	if channel.GetAlgorithmID() > 0 {
		// Dont continue processing channels managed by algorithms
		handled, err := service.handleChannelAlgorithm(channel, conditionMetric, value, condition.GetThreshold())
		if err != nil {
			return false, err
		}
		return handled, nil
	}

	if result {

		service.app.Logger.Debugf("[MicroControllerService.handleChannelConditions] %s conditional evaluated true, current position: %d", channel.GetName(), position)

		if position == common.SWITCH_OFF {

			service.app.Logger.Debugf("[MicroControllerService.handleChannelConditions] Switching ON channel: id=%d, name=%s, metric.key=%s, metric.value=%.2f",
				channel.GetID(), channel.GetName(), conditionMetric.GetKey(), value)

			if channel.GetDuration() > 0 {
				message := fmt.Sprintf("Switching ON %s for %d seconds. %s %.2f %s",
					channel.GetName(), channel.GetDuration(), conditionMetric.GetName(), value, conditionMetric.GetUnit())
				_, err := service.TimerSwitch(channel.GetChannelID(), channel.GetDuration(), message)
				if err != nil {
					return false, err
				}
			} else {
				message := fmt.Sprintf("Switching ON %s. %s %.2f %s", channel.GetName(), conditionMetric.GetName(), value, conditionMetric.GetUnit())
				_, err := service.Switch(channel.GetChannelID(), common.SWITCH_ON, message)
				if err != nil {
					return false, err
				}
			}

			if backoff > 0 {
				service.backoffTable[channel.GetID()] = time.Now()
			}

			return true, nil
		}

	} else {

		service.app.Logger.Debugf("[MicroControllerService.handleChannelConditions] %s conditional evaluated false, current position: %d", channel.GetName(), position)

		if debounce > 0 {
			service.app.Logger.Debugf("[MicroControllerService.handleChannelConditions] Inspecting debounce window: debounce=%d, threshold=%.2f, value=%.2f",
				debounce, condition.GetThreshold(), value)

			if value >= condition.GetThreshold()-float64(debounce) {
				return false, nil
			}
		}

		if position == common.SWITCH_ON {

			service.app.Logger.Debugf("[MicroControllerService.handleChannelConditions] Switching OFF channel: id=%d, name=%s, metric.key=%s, value=%.2f. debounce=%d",
				channel.GetID(), channel.GetName(), conditionMetric.GetKey(), value, debounce)

			message := fmt.Sprintf("Switching OFF %s. %s %.2f %s", channel.GetName(), conditionMetric.GetName(), value, conditionMetric.GetUnit())
			_, err := service.Switch(channel.GetChannelID(), common.SWITCH_OFF, message)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

func (service *MicroControllerService) handleChannelSchedule(state state.FarmStateMap, channel config.ChannelConfig) error {

	//eventType := "Scheduled Channel"

	var activeSchedule config.ScheduleConfig

	if !channel.IsEnabled() {
		return nil
	}

	controllerState, err := state.GetController(service.controller.GetType())
	if err != nil {
		return err
	}

	position := controllerState.GetChannels()[channel.GetChannelID()]
	service.app.Logger.Debugf("[MicroControllerService.handleChannelSchedule] %s switch position: %d", channel.GetName(), position)

	for _, schedule := range channel.GetSchedule() {

		executionCount := schedule.GetExecutionCount()
		if schedule.GetCount() > 0 && executionCount >= schedule.GetCount() {
			service.app.Logger.Debugf("[MicroControllerService.handleChannelSchedule] Reached max execution count: %d", executionCount)
			continue
		}

		if service.scheduleService.IsScheduled(&schedule, channel.GetDuration()) {
			activeSchedule = &schedule
			break
		}
	}

	if activeSchedule != nil {
		service.app.Logger.Debugf("[MicroControllerService.handleChannelSchedule] %s scheduled ON condition met. Current position: %d", channel.GetName(), position)
		if position == common.SWITCH_OFF {

			message := fmt.Sprintf("Switching ON scheduled %s.", channel.GetName())
			_, err := service.Switch(channel.GetChannelID(), common.SWITCH_ON, message)
			if err != nil {
				return err
			}

			executionCount := activeSchedule.GetExecutionCount() + 1
			activeSchedule.SetLastExecuted(time.Now())
			activeSchedule.SetExecutionCount(executionCount)
			session := CreateSystemSession(service.app.Logger, service.farmService)
			if err := service.scheduleService.Update(session, activeSchedule); err != nil {
				return err
			}
		}

	} else {

		service.app.Logger.Debugf("[MicroControllerService.handleChannelSchedule] %s scheduled OFF condition met. Current position: %d", channel.GetName(), position)
		if position == common.SWITCH_ON {
			message := fmt.Sprintf("Switching OFF scheduled %s.", channel.GetName())
			_, err := service.Switch(channel.GetChannelID(), common.SWITCH_OFF, message)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (service *MicroControllerService) handleChannelAlgorithm(channel config.ChannelConfig, metric config.MetricConfig, value, threshold float64) (bool, error) {
	controllerType := service.controller.GetType()
	service.app.Logger.Debugf("[MicroControllerService.handleChannelAlgorithm] Processing %s %s algorithm", controllerType, channel.GetName())
	if channel.GetAlgorithmID() == common.ALGORITHM_PH_ID {
		configs := service.GetControllerConfig().GetConfigs()
		gallons := 0
		gallonsConfigKey := fmt.Sprintf("%s.gallons", controllerType)
		for _, config := range configs {
			if config.GetKey() == gallonsConfigKey {
				g, err := strconv.Atoi(config.GetValue())
				if err != nil {
					return false, err
				}
				gallons = g
			}
		}
		if gallons <= 0 {
			return false, fmt.Errorf("%s configuration value must be greater than 0. value: %d", gallonsConfigKey, gallons)
		}
		diff := value - threshold
		dose := int(math.Round(diff * float64(gallons/2)))
		if dose <= 0 {
			return false, nil
		}
		service.app.Logger.Debugf("[MicroControllerService.handleChannelAlgorithm] Autodosing using pH algorithm: diff=%.2f, dose=%d", diff, dose)
		message := fmt.Sprintf("%s: %.2f, auto-dosing %s for %d seconds", metric.GetName(), value, channel.GetName(), dose)
		_, err := service.TimerSwitch(channel.GetChannelID(), dose, message)
		if err != nil {
			return false, err
		}
		if channel.GetBackoff() > 0 {
			service.backoffTable[channel.GetID()] = time.Now()
		}
		return true, nil
	}
	return false, nil
}

func (service *MicroControllerService) notify(eventType, message string) error {
	if !service.GetControllerConfig().IsNotify() {
		service.app.Logger.Warningf("[MicroControllerService.notify] MicroControllerService notifications disabled!")
		return nil
	}
	return service.notificationService.Enqueue(&model.Notification{
		Controller: service.farmService.GetConfig().GetName(),
		Priority:   common.NOTIFICATION_PRIORITY_LOW,
		Title:      service.controller.GetType(),
		Type:       eventType,
		Message:    message,
		Timestamp:  time.Now()})
}

func (service *MicroControllerService) error(method, eventType string, err error) {
	service.app.Logger.Errorf("[MicroControllerService.%s] Error: %s", method, err)
	service.notificationService.Enqueue(&model.Notification{
		Controller: service.farmService.GetConfig().GetName(),
		Priority:   common.NOTIFICATION_PRIORITY_HIGH,
		Title:      service.controller.GetType(),
		Type:       eventType,
		Message:    err.Error(),
		Timestamp:  time.Now()})
}

/*
func (service *MicroControllerService) NotifyObservers() {
	for _, observer := range service.observers {
		controller, err := service.farmService.GetState().GetController(service.controller.GetType())
		if err != nil {
			service.app.Logger.Errorf("[MicroControllerService.NotifyObservers] Error: %s", err)
		}
		observer.OnControllerStateChange(controller)
	}
}

func (service *MicroControllerService) RegisterObserver(observer common.ControllerObserver) {
	state.observers = append(state.observers, observer)
}

func (service *MicroControllerService) NotifyObservers() {
	for _, observer := range state.observers {
		observer.OnControllerStateChange(state)
	}
}

func (service *MicroControllerService) getMetricConfig(metricID int) (config.MetricConfig, error) {
	controllerConfig, err := service.farmService.GetConfig().GetController(service.controller.GetType())
	if err != nil {
		return nil, err
	}
	metrics := controllerConfig.GetMetrics()
	for _, metric := range metrics {
		if metric.GetID() == metricID {
			return metric, nil
		}
	}
	return nil, fmt.Errorf("Metric ID not found: %d", metricID)
}

*/
