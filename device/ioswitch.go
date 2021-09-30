package device

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

type SmartSwitch struct {
	app        *app.App
	baseURL    string
	httpClient HttpClient
	deviceType string
	IOSwitcher
}

// TODO: Rename to IOSwitch
func CreateSmartSwitch(httpClient HttpClient, app *app.App, baseURL, deviceType string) IOSwitcher {
	app.Logger.Debugf("[CreateSmartSwitch] Initializing %s device at %s", deviceType, baseURL)
	return &SmartSwitch{
		app:        app,
		baseURL:    baseURL,
		httpClient: httpClient,
		deviceType: deviceType}
}

func NewSmartSwitch(app *app.App, baseURL, deviceType string) IOSwitcher {
	app.Logger.Debugf("[NewSmartSwitch] Initializing %s device. url=%s", deviceType, baseURL)
	return &SmartSwitch{
		app:        app,
		baseURL:    baseURL,
		httpClient: &http.Client{},
		deviceType: deviceType}
}

func (d *SmartSwitch) GetType() string {
	return d.deviceType
}

func (d *SmartSwitch) State() (state.DeviceStateMap, error) {
	response, err := d.httpClient.Get(d.baseURL + "/state")
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	d.app.Logger.Debugf("responseData: %s", responseData)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	var state state.DeviceState
	err = json.Unmarshal(responseData, &state)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	state.Timestamp = time.Now().In(d.app.Location)
	return &state, nil
}

func (d *SmartSwitch) Switch(channel, position int) (*common.Switch, error) {
	endpoint := fmt.Sprintf("%s/%s/%d/%d", d.baseURL, "switch", channel, position)
	d.app.Logger.Debugf("Endpoint: %s", endpoint)
	response, err := d.httpClient.Get(endpoint)
	if err != nil {
		d.app.Logger.Error(err.Error())
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		d.app.Logger.Error(err.Error())
	}
	responseData = bytes.Replace(responseData, []byte(": NAN"), []byte(":null"), -1)
	d.app.Logger.Debugf("responseData: %s", responseData)
	var _switch common.Switch
	err = json.Unmarshal(responseData, &_switch)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	return &_switch, nil
}

func (d *SmartSwitch) TimerSwitch(channel, duration int) (common.TimerEvent, error) {
	endpoint := fmt.Sprintf("%s/%s/%d/%d", d.baseURL, "timer", channel, duration)
	d.app.Logger.Debugf("endpoint=%s", endpoint)
	response, err := d.httpClient.Get(endpoint)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	d.app.Logger.Debugf("responseData: %s", responseData)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	var event common.ChannelTimerEvent
	err = json.Unmarshal(responseData, &event)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	event.Timestamp = time.Now().In(d.app.Location)
	return &event, nil
}

func (d *SmartSwitch) SystemInfo() (DeviceInfo, error) {
	endpoint := fmt.Sprintf("%s/%s", d.baseURL, "system")
	d.app.Logger.Debugf("endpoint=%s", endpoint)
	response, err := d.httpClient.Get(endpoint)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	d.app.Logger.Debugf("responseData: %s", responseData)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	var info DefaultDeviceInfo
	err = json.Unmarshal(responseData, &info)
	if err != nil {
		d.app.Logger.Error(err.Error())
		return nil, err
	}
	return &info, nil
}
