package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type DefaultChangefeedService struct {
	mutex           *sync.Mutex
	app             *app.App
	serviceRegistry ServiceRegistry
	changefeeders   map[string]datastore.Changefeeder
	feeds           map[string]datastore.Changefeeder
	ChangefeedService
}

func NewChangefeedService(app *app.App, serviceRegistry ServiceRegistry, changefeeders map[string]datastore.Changefeeder) ChangefeedService {

	service := &DefaultChangefeedService{
		mutex:           &sync.Mutex{},
		app:             app,
		serviceRegistry: serviceRegistry,
		changefeeders:   changefeeders,
		feeds:           changefeeders}

	return service
}

func (service *DefaultChangefeedService) FeedCount() int {
	return len(service.feeds)
}

func (service *DefaultChangefeedService) Subscribe() {
	go service.changefeeders["_controller_config_items"].Subscribe(service.OnControllerConfigConfigChange)
	go service.changefeeders["_channels"].Subscribe(service.OnChannelConfigChange)
	go service.changefeeders["_metrics"].Subscribe(service.OnMetricConfigChange)
	go service.changefeeders["_conditions"].Subscribe(service.OnConditionConfigChange)
	go service.changefeeders["_schedules"].Subscribe(service.OnScheduleConfigChange)
	for k, v := range service.changefeeders {
		if string(k[0]) != "_" {
			if k == "server" {
				continue
			}
			go v.Subscribe(service.OnControllerStateChange)
		}
	}
}

// OnFarmStateChange handles use cases where the entire farm is updated in one operation, such as in the
// case of a Raft key/value store.
//func (service *DefaultChangefeedService) OnFarmStateChange(farmState state.FarmStateMap) {
//}

// OnControllerStateChange handles use cases where GORM / MetricDatastore is used to update the farm
// state when a change is made to the underlying database (without FarmService knowing). The database
// must support CDC/changefeed style real-time notifications (cockroachdb / rethinkdb).
func (service *DefaultChangefeedService) OnControllerStateChange(changefeed datastore.Changefeed) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	record := changefeed.GetRawMessage()

	var controllerID int
	err := json.Unmarshal(*record["controller_id"], &controllerID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error: %s", err)
		return
	}

	controller, ok := service.app.ControllerIndex.Get(controllerID)
	if !ok {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Controller index missing controller: changefeed=%s, controller_id=%d",
			changefeed.GetTable(), controllerID)

		service.app.Logger.Warningf("controllerIndex: %+v", service.app.ControllerIndex)
		return
	}

	farmID := controller.GetFarmID()
	controllerType := controller.GetType()

	if farmService, ok := service.serviceRegistry.GetFarmService(farmID); ok {

		farmState := farmService.GetState()
		if farmState == nil {
			return
		}

		metricConfigs := controller.GetMetrics()
		sort.SliceStable(metricConfigs, func(i, j int) bool {
			return strings.ToLower(metricConfigs[i].GetName()) < strings.ToLower(metricConfigs[j].GetName())
		})
		metrics := make(map[string]float64, 0)
		for _, metric := range metricConfigs {
			var value float64
			jsonKey := strings.ToLower(metric.GetKey())
			err = json.Unmarshal(*record[jsonKey], &value)
			if err != nil {
				service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Unmarshal error: %s. changefeed=%s, controllerID=%d",
					err, changefeed.GetTable(), controllerID)
				return
			}
			metricValue, err := farmState.GetMetricValue(controllerType, metric.GetKey())
			if err != nil {
				service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error getting metric value: %s. changefeed=%s, controllerID=%d",
					err, changefeed.GetTable(), controllerID)
				return
			}
			if value == metricValue {
				continue
			}
			metrics[metric.GetKey()] = value
		}

		channelConfigs := controller.GetChannels()
		sort.SliceStable(channelConfigs, func(i, j int) bool {
			return strings.ToLower(channelConfigs[i].GetName()) < strings.ToLower(channelConfigs[j].GetName())
		})
		channels := make(map[int]int, 0)
		for i := range channelConfigs {
			var value int
			channel := fmt.Sprintf("c%d", i)
			err = json.Unmarshal(*record[channel], &value)

			channelValue, err := farmState.GetChannelValue(controllerType, i)
			if err != nil {
				service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error retrieving metric from farm state: changefeed=%s, controller_id=%f, error=%s",
					changefeed.GetTable(), controllerID, err)
				return
			}
			if value == channelValue {
				continue
			}

			channels[channelConfigs[i].GetChannelID()] = value
		}

		controllerStateDelta, err := farmState.Diff(controllerType, metrics, channels)
		if err != nil {
			service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error: %s. changefeed=%s, controllerID=%d",
				err, changefeed.GetTable(), controllerID)
			return
		}

		if controllerStateDelta == nil {
			service.app.Logger.Debugf("[ChangefeedService.OnControllerStateChange] No state difference, aborting. changefeed=%s, controller_id=%d",
				changefeed.GetTable(), controllerID)
			return
		}

		param := make(map[string]state.ControllerStateDeltaMap, 1)
		param[controllerType] = controllerStateDelta

		farmService.PublishControllerDelta(param)
		return
	}

	service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Service registry missing farm service: farm.id=%d, changefeed=%s, controller_id=%f",
		farmID, changefeed.GetTable(), controllerID)
}

func (service *DefaultChangefeedService) OnControllerConfigConfigChange(changefeed datastore.Changefeed) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	record := changefeed.GetRawMessage()

	var id int
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.id: %s", err)
		return
	}

	var controllerID int
	err = json.Unmarshal(*record["controller_id"], &controllerID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.controllerID: %s", err)
		return
	}

	var userID int
	err = json.Unmarshal(*record["user_id"], &userID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.userID: %s", err)
		return
	}

	var controllerConfigItem config.ControllerConfigItem
	err = json.Unmarshal(changefeed.GetBytes(), &controllerConfigItem)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling changfeed: %s", err)
	}
	controllerConfigItem.SetID(id)
	controllerConfigItem.SetUserID(userID)
	controllerConfigItem.SetControllerID(controllerID)

	service.app.Logger.Errorf("OnControllerConfigConfigChange fired! Received controller config: %+v", &controllerConfigItem)

	controller, ok := service.app.ControllerIndex.Get(controllerID)
	if !ok {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Controller index missing controller: changefeed=%s, controller_id=%d, index=%+v",
			changefeed.GetTable(), controllerID, service.app.ControllerIndex.GetAll())
		return
	}
	controller.SetConfig(&controllerConfigItem)

	if farmService, ok := service.serviceRegistry.GetFarmService(controller.GetFarmID()); ok {
		farmService.GetConfig().ParseConfigs()
		farmService.PublishConfig()
	}
}

func (service *DefaultChangefeedService) OnMetricConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id int
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error unmarshaling channel.id: %s", err)
		return
	}

	var controllerID int
	err = json.Unmarshal(*record["controller_id"], &controllerID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error unmarshaling channel.controllerID: %s", err)
		return
	}

	var metric config.Metric
	err = json.Unmarshal(changefeed.GetBytes(), &metric)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnMetricConfigChange] Error unmarshaling changfeed: %s", err)
	}
	metric.SetID(id)
	metric.SetControllerID(controllerID)

	service.app.Logger.Errorf("OnMetricChange fired! Received metric config: %+v", metric)

	controller, ok := service.app.ControllerIndex.Get(metric.GetControllerID())
	if !ok {
		return
	}
	controller.SetMetric(&metric)

	if farmService, ok := service.serviceRegistry.GetFarmService(controller.GetFarmID()); ok {
		farmService.PublishConfig()
	}
}

func (service *DefaultChangefeedService) OnChannelConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id int
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error unmarshaling channel.id: %s", err)
		return
	}

	var controllerID int
	err = json.Unmarshal(*record["controller_id"], &controllerID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error unmarshaling channel.controllerID: %s", err)
		return
	}

	var algorithmID int
	err = json.Unmarshal(*record["algorithm_id"], &controllerID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerStateChange] Error unmarshaling channel.algorithmID: %s", err)
		return
	}

	var channel config.Channel
	err = json.Unmarshal(changefeed.GetBytes(), &channel)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigChange] Error unmarshaling changfeed: %s", err)
	}
	channel.SetID(id)
	channel.SetControllerID(controllerID)
	channel.SetAlgorithmID(algorithmID)

	service.app.Logger.Errorf("OnChannelConfigChange fired! Received channel config: %+v", channel)

	controller, ok := service.app.ControllerIndex.Get(controllerID)
	if !ok {
		service.app.Logger.Errorf("OnChannelConfigChange: Failed to retrieve controller from index: %+v", service.app.ControllerIndex)
		return
	}
	controller.SetChannel(&channel)

	if farmService, ok := service.serviceRegistry.GetFarmService(controller.GetFarmID()); ok {
		farmService.PublishConfig()
	}
}

func (service *DefaultChangefeedService) OnConditionConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id uint64
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.id: %s", err)
		return
	}

	var channelID int
	err = json.Unmarshal(*record["channel_id"], &channelID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.channelID: %s", err)
		return
	}

	var metricID int
	err = json.Unmarshal(*record["metric_id"], &metricID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.metricID: %s", err)
		return
	}

	var condition config.Condition
	err = json.Unmarshal(changefeed.GetBytes(), &condition)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnChannelConfigChange] Error unmarshaling changfeed: %s", err)
	}
	condition.SetID(id)
	condition.SetChannelID(channelID)
	condition.SetMetricID(metricID)

	service.app.Logger.Errorf("OnConditionConfigChange fired! Received condition config: %+v", condition)

	channel, ok := service.app.ChannelIndex.Get(channelID)
	if !ok {
		return
	}

	channel.SetCondition(&condition)

	controller, ok := service.app.ControllerIndex.Get(channel.GetControllerID())
	if !ok {
		return
	}

	if farmService, ok := service.serviceRegistry.GetFarmService(controller.GetFarmID()); ok {
		farmService.PublishConfig()
	}
}

func (service *DefaultChangefeedService) OnScheduleConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id uint64
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.id: %s", err)
		return
	}

	var channelID int
	err = json.Unmarshal(*record["channel_id"], &channelID)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnControllerConfigConfigChange] Error unmarshaling channel.channelID: %s", err)
		return
	}

	var schedule config.Schedule
	err = json.Unmarshal(changefeed.GetBytes(), &schedule)
	if err != nil {
		service.app.Logger.Errorf("[ChangefeedService.OnChannelConfigChange] Error unmarshaling changfeed: %s", err)
	}
	schedule.SetID(id)
	schedule.SetChannelID(channelID)

	service.app.Logger.Errorf("OnScheduleConfigChange fired! Received condition config: %+v", schedule)

	channel, ok := service.app.ChannelIndex.Get(channelID)
	if !ok {
		return
	}

	channel.SetScheduleItem(&schedule)

	controller, ok := service.app.ControllerIndex.Get(channel.GetControllerID())
	if !ok {
		return
	}

	if farmService, ok := service.serviceRegistry.GetFarmService(controller.GetFarmID()); ok {
		farmService.PublishConfig()
	}
}
