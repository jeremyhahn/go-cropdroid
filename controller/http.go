package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type HttpController struct {
	app            *app.App
	baseURL        string
	httpClient     *http.Client
	controllerType string
	Controller
}

func NewHttpController(app *app.App, baseURL, controllerType string) Controller {
	app.Logger.Debugf("[HttpController] Initializing %s controller at %s", controllerType, baseURL)
	return &HttpController{
		app:            app,
		baseURL:        baseURL,
		controllerType: controllerType}
}

func (c *HttpController) GetType() string {
	return c.controllerType
}

func (c *HttpController) State() (state.ControllerStateMap, error) {
	response, err := http.Get(c.baseURL + "/state")
	if err != nil {
		c.app.Logger.Error(err.Error())
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	c.app.Logger.Debugf("[HttpController.State] responseData: %s", responseData)
	if err != nil {
		c.app.Logger.Error(err.Error())
		return nil, err
	}
	var state state.ControllerState
	err = json.Unmarshal(responseData, &state)
	if err != nil {
		c.app.Logger.Error(err.Error())
		return nil, err
	}
	state.Timestamp = time.Now().In(c.app.Location)
	return &state, nil
}

func (c *HttpController) Switch(channel, position int) (*common.Switch, error) {
	endpoint := fmt.Sprintf("%s/%s/%d/%d", c.baseURL, "switch", channel, position)
	c.app.Logger.Debugf("[HttpController.Switch] Endpoint: %s", endpoint)
	response, err := http.Get(endpoint)
	if err != nil {
		c.app.Logger.Error(err.Error())
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.app.Logger.Error(err.Error())
	}
	responseData = bytes.Replace(responseData, []byte(": NAN"), []byte(":null"), -1)
	c.app.Logger.Debugf("[HttpController.Switch] responseData: %s", responseData)
	var _switch common.Switch
	err = json.Unmarshal(responseData, &_switch)
	if err != nil {
		c.app.Logger.Error(err.Error())
		return nil, err
	}
	return &_switch, nil
}

func (c *HttpController) TimerSwitch(channel, duration int) (common.TimerEvent, error) {
	endpoint := fmt.Sprintf("%s/%s/%d/%d", c.baseURL, "timer", channel, duration)
	c.app.Logger.Debugf("endpoint=%s", endpoint)
	response, err := http.Get(endpoint)
	if err != nil {
		c.app.Logger.Error(err.Error())
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	c.app.Logger.Debugf("[HttpController.TimedSwitch] responseData: %s", responseData)
	if err != nil {
		c.app.Logger.Error(err.Error())
		return nil, err
	}
	var event common.ChannelTimerEvent
	err = json.Unmarshal(responseData, &event)
	if err != nil {
		c.app.Logger.Error(err.Error())
		return nil, err
	}
	event.Timestamp = time.Now().In(c.app.Location)
	return &event, nil
}
