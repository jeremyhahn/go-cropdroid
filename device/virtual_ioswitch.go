package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type VirtualIOSwitch struct {
	SmartSwitch
	farmState state.FarmStateMap
	stateFile string
	startTime time.Time
}

func CreateVirtualIOSwitch(httpClient HttpClient, app *app.App,
	farmState state.FarmStateMap, baseURL, deviceType,
	stateFile string) VirtualIOSwitcher {

	app.Logger.Debugf("[CreateSmartSwitch] Initializing %s device at %s", deviceType, baseURL)
	return &VirtualIOSwitch{
		SmartSwitch: SmartSwitch{
			app:        app,
			baseURL:    baseURL,
			httpClient: httpClient,
			deviceType: deviceType},
		farmState: farmState,
		stateFile: stateFile,
		startTime: time.Now()}
}

func NewVirtualIOSwitch(app *app.App, farmState state.FarmStateMap,
	baseURL, deviceType string) VirtualIOSwitcher {

	return &VirtualIOSwitch{
		SmartSwitch: SmartSwitch{
			app:        app,
			baseURL:    baseURL,
			deviceType: deviceType},
		farmState: farmState,
		stateFile: fmt.Sprintf("%s/%s/v%s.json", app.HomeDir, common.HTTP_PUBLIC_HTML, deviceType)}
}

func (c *VirtualIOSwitch) GetType() string {
	return c.deviceType
}

func (c *VirtualIOSwitch) State() (state.DeviceStateMap, error) {
	if _, err := os.Stat(c.stateFile); err == nil {
		state := state.NewDeviceStateMap()
		data, err := os.ReadFile(c.stateFile)
		if err != nil {
			c.app.Logger.Error(err.Error())
			return nil, err
		}
		err = json.Unmarshal(data, state)
		if err != nil {
			c.app.Logger.Errorf("Error: %s", err.Error())
			return nil, err
		}
		state.SetTimestamp(time.Now().In(c.app.Location))
		c.app.Logger.Debugf("%s state: %+v", c.deviceType, state)
		return state, nil
	}
	return nil, errors.New(fmt.Sprintf("Virtual device data file not found: %s", c.stateFile))
}

func (c *VirtualIOSwitch) Switch(channel, position int) (*common.Switch, error) {
	state, err := c.State()
	if err != nil {
		return nil, err
	}
	state.GetChannels()[channel] = position
	c.farmState.SetDevice(c.deviceType, state)
	stateJson, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		c.app.Logger.Errorf("Error marshalling virtual device state: %s", err.Error())
		return nil, err
	}
	err = os.WriteFile(c.stateFile, stateJson, 0644)
	if err != nil {
		c.app.Logger.Errorf("Error writing virtual device file: %s", err.Error())
		return nil, err
	}
	return &common.Switch{
		Channel: channel,
		State:   position}, nil
}

func (c *VirtualIOSwitch) TimerSwitch(channel, duration int) (common.TimerEvent, error) {
	state, err := c.State()
	if err != nil {
		return nil, err
	}
	state.GetChannels()[channel] = common.SWITCH_ON
	c.farmState.SetDevice(c.deviceType, state)

	durationTimer := time.NewTimer(time.Duration(duration) * time.Second)
	go func() {
		<-durationTimer.C
		state.GetChannels()[channel] = common.SWITCH_OFF
		c.farmState.SetDevice(c.deviceType, state)
		c.WriteState(state)
	}()

	c.WriteState(state)

	return &common.ChannelTimerEvent{
		Channel:   channel,
		Duration:  duration,
		Timestamp: time.Now()}, nil
}

func (c *VirtualIOSwitch) SystemInfo() (DeviceInfo, error) {
	return &DefaultDeviceInfo{
		FirmwareVersion: "virtfw-v0.0.1a",
		HardwareVersion: "virthw-v0.0.1a",
		Uptime:          int64(time.Since(c.startTime).Seconds())}, nil
}

func (c *VirtualIOSwitch) WriteState(state state.DeviceStateMap) error {
	stateJson, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		c.app.Logger.Errorf("Error marshalling virtual state: %s", err.Error())
		return err
	}
	c.app.Logger.Debugf("Writing virtual state: %s", stateJson)
	err = os.WriteFile(c.stateFile, stateJson, 0644)
	if err != nil {
		c.app.Logger.Errorf("Error writing virtual state file: %s", err.Error())
		return err
	}
	c.app.Logger.Debugf("Wrote virtual state to: %s", c.stateFile)
	err = os.WriteFile(c.stateFile, stateJson, 0644)
	if err != nil {
		c.app.Logger.Errorf("Error writing virtual state file: %s", err.Error())
		return err
	}
	return nil
}
