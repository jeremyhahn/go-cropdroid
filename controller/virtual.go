package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/jeremyhahn/cropdroid/app"
	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/state"
)

type VirtualHttpController struct {
	HttpController
	farmState state.FarmStateMap
	stateFile string
}

func NewVirtualController(app *app.App, farmState state.FarmStateMap, baseURL, controllerType string) VirtualController {
	return &VirtualHttpController{
		HttpController: HttpController{
			app:            app,
			baseURL:        baseURL,
			controllerType: controllerType},
		farmState: farmState,
		stateFile: fmt.Sprintf("%s/%s/v%s.json", app.HomeDir, common.HTTP_PUBLIC_HTML, controllerType)}
}

func (c *VirtualHttpController) GetType() string {
	return c.controllerType
}

func (c *VirtualHttpController) State() (state.ControllerStateMap, error) {
	if _, err := os.Stat(c.stateFile); err == nil {
		state := state.NewControllerStateMap()
		data, err := ioutil.ReadFile(c.stateFile)
		if err != nil {
			c.app.Logger.Error(err.Error())
			return nil, err
		}
		err = json.Unmarshal(data, state)
		if err != nil {
			c.app.Logger.Errorf("[VirtualController.State] Error: %s", err.Error())
			return nil, err
		}
		state.SetTimestamp(time.Now().In(c.app.Location))
		c.app.Logger.Debugf("%s state: %+v", c.controllerType, state)
		return state, nil
	}
	return nil, errors.New(fmt.Sprintf("Virtual controller data file not found: %s", c.stateFile))
}

func (c *VirtualHttpController) Switch(channel, position int) (*common.Switch, error) {
	state, err := c.State()
	if err != nil {
		return nil, err
	}
	state.GetChannels()[channel] = position
	c.farmState.SetController(c.controllerType, state)
	stateJson, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		c.app.Logger.Errorf("Error marshalling virtual controller state: %s", err.Error())
		return nil, err
	}
	err = ioutil.WriteFile(c.stateFile, stateJson, 0644)
	if err != nil {
		c.app.Logger.Errorf("Error writing virtual controller file: %s", err.Error())
		return nil, err
	}
	return &common.Switch{
		Channel: channel,
		State:   position}, nil
}

func (c *VirtualHttpController) TimerSwitch(channel, duration int) (common.TimerEvent, error) {
	state, err := c.State()
	if err != nil {
		return nil, err
	}
	state.GetChannels()[channel] = common.SWITCH_ON
	c.farmState.SetController(c.controllerType, state)

	durationTimer := time.NewTimer(time.Duration(duration) * time.Second)
	go func() {
		<-durationTimer.C
		state.GetChannels()[channel] = common.SWITCH_OFF
		c.farmState.SetController(c.controllerType, state)
		c.WriteState(state)
	}()

	c.WriteState(state)

	return &common.ChannelTimerEvent{
		Channel:   channel,
		Duration:  duration,
		Timestamp: time.Now()}, nil
}

func (c *VirtualHttpController) WriteState(state state.ControllerStateMap) error {
	stateJson, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		c.app.Logger.Errorf("Error marshalling virtual state: %s", err.Error())
		return err
	}
	c.app.Logger.Debugf("Writing virtual state: %s", stateJson)
	err = ioutil.WriteFile(c.stateFile, stateJson, 0644)
	if err != nil {
		c.app.Logger.Errorf("Error writing virtual state file: %s", err.Error())
		return err
	}
	c.app.Logger.Debugf("Wrote virtual state to: %s", c.stateFile)
	err = ioutil.WriteFile(c.stateFile, stateJson, 0644)
	if err != nil {
		c.app.Logger.Errorf("Error writing virtual state file: %s", err.Error())
		return err
	}
	return nil
}
