package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	ErrControllerNotFound = errors.New("Controller not found")
)

// FarmStateMap stores a map of real-time controller states for a farm
type FarmStateMap interface {
	SetController(controllerType string, controllerState ControllerStateMap)
	GetController(controllerType string) (ControllerStateMap, error)
	GetControllers() map[string]ControllerStateMap
	GetMetrics(controller string) (map[string]float64, error)
	GetMetricValue(controller, key string) (float64, error)
	SetMetricValue(controller, key string, value float64) error
	GetChannels(controller string) ([]int, error)
	GetChannelValue(controller string, channelID int) (int, error)
	SetChannelValue(controller string, channelID int, value int) error
	Diff(controller string, metrics map[string]float64, channels map[int]int) (ControllerStateDeltaMap, error)
	GetFarmID() int
	SetFarmID(int)
	GetTimestamp() int64
	//UnmarshalJSON(data []byte) error
	String() string
}

type FarmState struct {
	mutex        *sync.RWMutex                 `json:"-"`
	Id           int                           `json:"id"`
	Controllers  map[string]ControllerStateMap `json:"controllers"`
	Timestamp    int64                         `json:"timestamp"`
	FarmStateMap `json:"-"`
}

func NewFarmStateMap(id int) FarmStateMap {
	return &FarmState{
		mutex:       &sync.RWMutex{},
		Id:          id,
		Controllers: make(map[string]ControllerStateMap, 0),
		Timestamp:   time.Now().Unix()}
}

func CreateFarmState(id int, state map[string]ControllerStateMap) FarmStateMap {
	return &FarmState{
		mutex:       &sync.RWMutex{},
		Id:          id,
		Controllers: state,
		Timestamp:   time.Now().Unix()}
}

func (farm *FarmState) GetFarmID() int {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	return farm.Id
}

func (farm *FarmState) SetFarmID(id int) {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	farm.Id = id
}

func (farm *FarmState) SetController(controllerType string, controllerState ControllerStateMap) {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	farm.Controllers[controllerType] = controllerState
}

func (farm *FarmState) GetController(controllerType string) (ControllerStateMap, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if controller, ok := farm.Controllers[controllerType]; ok {
		return controller, nil
	}
	return nil, fmt.Errorf("Controller not found in farm state: %s", controllerType)
}

func (farm *FarmState) GetControllers() map[string]ControllerStateMap {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	return farm.Controllers
}

func (farm *FarmState) GetMetricValue(controller, key string) (float64, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _controller, ok := farm.Controllers[controller]; ok {
		if cachedMetric, ok := _controller.GetMetrics()[key]; ok {
			return cachedMetric, nil
		} else {
			return 0.0, fmt.Errorf("Metric not found in farm state: %s.%s", controller, key)
		}
	}
	return 0.0, fmt.Errorf("Controller not found in farm state: %s", controller)
}

func (farm *FarmState) GetMetrics(controller string) (map[string]float64, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _controller, ok := farm.Controllers[controller]; ok {
		return _controller.GetMetrics(), nil
	}
	return nil, fmt.Errorf("Controller not found in farm state: %s", controller)
}

func (farm *FarmState) SetMetricValue(controller, key string, value float64) error {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	if _controller, ok := farm.Controllers[controller]; ok {
		_controller.GetMetrics()[key] = value
		return nil
	}
	return fmt.Errorf("Controller not found in farm state: %s", controller)
}

func (farm *FarmState) GetChannelValue(controller string, channelID int) (int, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _controller, ok := farm.Controllers[controller]; ok {
		//channels := _controller.GetChannels()
		//if channelID > len(channels) || channelID < 0 {
		//	return 0, fmt.Errorf("Invalid channel ID %d (%s controller)", channelID, controller)
		//}
		return _controller.GetChannels()[channelID], nil
	}
	return 0.0, fmt.Errorf("Controller not found in farm state: %s", controller)
}

func (farm *FarmState) SetChannelValue(controller string, channelID int, value int) error {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _controller, ok := farm.Controllers[controller]; ok {
		_controller.GetChannels()[channelID] = value
		return nil
	}
	return fmt.Errorf("Controller not found in farm state: %s", controller)
}

func (farm *FarmState) GetChannels(controller string) ([]int, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _controller, ok := farm.Controllers[controller]; ok {
		return _controller.GetChannels(), nil
	}
	return nil, fmt.Errorf("Controller not found in farm state: %s", controller)
}

// Diff takes a map of metric and channels that represent a proposed state about to be merged into the state store
// and compares/diffs it against the current stored state. A ControllerStateDeltaMap is returned that contains only the metrics
// and channels that've changed. Metrics take the form map["metric.key"] = float64 and channels map["channel.channelId"] = int.
func (farm *FarmState) Diff(controller string, metrics map[string]float64, channels map[int]int) (ControllerStateDeltaMap, error) {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	newMetrics := make(map[string]float64, 0)
	newChannels := make(map[int]int, 0)
	if _controller, ok := farm.Controllers[controller]; ok {
		_metrics := _controller.GetMetrics()
		_channels := _controller.GetChannels()
		for k, v := range metrics {
			if metrics[k] == _metrics[k] {
				continue
			}
			_metrics[k] = v
			newMetrics[k] = v
		}
		for k, v := range channels {
			if channels[k] == _channels[k] {
				continue
			}
			_channels[k] = v
			newChannels[k] = v
		}
		if len(newMetrics) == 0 && len(newChannels) == 0 {
			return nil, nil
		}
		return CreateControllerStateDeltaMap(newMetrics, newChannels), nil
	}
	log.Printf("Controller not found in farm state: %s\n", controller)
	return nil, ErrControllerNotFound
	//return CreateControllerStateDeltaMap(metrics, channels), nil
}

func (farm *FarmState) GetTimestamp() int64 {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	return farm.Timestamp
}

func (farm *FarmState) UnmarshalJSON(data []byte) error {

	if farm.mutex == nil {
		farm.mutex = &sync.RWMutex{}
	}
	farm.mutex.Lock()
	defer farm.mutex.Unlock()

	var message map[string]*json.RawMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return err
	}

	var farmID int
	err = json.Unmarshal(*message["id"], &farmID)
	if err != nil {
		return err
	}
	farm.Id = farmID

	var controllersRawMessage map[string]ControllerState
	err = json.Unmarshal(*message["controllers"], &controllersRawMessage)
	if err != nil {
		return err
	}

	farm.Controllers = make(map[string]ControllerStateMap)
	for k, v := range controllersRawMessage {
		controllerState := v
		farm.Controllers[k] = &controllerState
	}

	var timestamp int64
	err = json.Unmarshal(*message["timestamp"], &timestamp)
	if err != nil {
		return err
	}
	farm.Timestamp = timestamp

	return nil
}

func (farm *FarmState) String() string {
	bytes, _ := json.Marshal(farm)
	return string(bytes)
}
