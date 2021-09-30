package viewmodel

import "github.com/jeremyhahn/go-cropdroid/config"

type WorkflowStep struct {
	ID                        uint64 `yaml:"id" json:"id"`
	WorkflowID                uint64 `yaml:"workflow" json:"workflowId"`
	DeviceID                  uint64 `yaml:"device" json:"deviceId"`
	DeviceType                string `yaml:"deviceType" json:"deviceType"`
	ChannelID                 uint64 `yaml:"channel" json:"channelId"`
	ChannelName               string `yaml:"channelName" json:"channelName"`
	Webhook                   string `yaml:"webhook" json:"webhook"`
	Duration                  int    `yaml:"duration" json:"duration"`
	Wait                      int    `yaml:"wait" json:"wait"`
	Text                      string `yaml:"text" json:"text"`
	State                     int    `yaml:"state" json:"state"`
	config.WorkflowStepConfig `yaml:"-" json:"-"`
}

func NewWorkflowStep() *WorkflowStep {
	return &WorkflowStep{}
}

func (ws *WorkflowStep) SetID(id uint64) {
	ws.ID = id
}

func (ws *WorkflowStep) GetID() uint64 {
	return ws.ID
}

func (ws *WorkflowStep) SetWorkflowID(id uint64) {
	ws.WorkflowID = id
}

func (ws *WorkflowStep) GetWorkflowID() uint64 {
	return ws.WorkflowID
}

func (ws *WorkflowStep) SetDeviceID(id uint64) {
	ws.DeviceID = id
}

func (ws *WorkflowStep) GetDeviceID() uint64 {
	return ws.DeviceID
}

func (ws *WorkflowStep) SetDeviceType(t string) {
	ws.DeviceType = t
}

func (ws *WorkflowStep) GetDeviceType() string {
	return ws.DeviceType
}

func (ws *WorkflowStep) SetChannelID(id uint64) {
	ws.ChannelID = id
}

func (ws *WorkflowStep) GetChannelID() uint64 {
	return ws.ChannelID
}

func (ws *WorkflowStep) SetChannelName(name string) {
	ws.ChannelName = name
}

func (ws *WorkflowStep) GetChannelName() string {
	return ws.ChannelName
}

func (ws *WorkflowStep) SetWebhook(url string) {
	ws.Webhook = url
}

func (ws *WorkflowStep) GetWebhook() string {
	return ws.Webhook
}

func (ws *WorkflowStep) SetDuration(seconds int) {
	ws.Duration = seconds
}

func (ws *WorkflowStep) GetDuration() int {
	return ws.Duration
}

func (ws *WorkflowStep) SetWait(seconds int) {
	ws.Wait = seconds
}

func (ws *WorkflowStep) GetWait() int {
	return ws.Wait
}

func (ws *WorkflowStep) SetText(text string) {
	ws.Text = text
}

func (ws *WorkflowStep) GetText() string {
	return ws.Text
}

// GetStatus returns the current state of the workflow step.
// See common.Constants.WORKFLOW_STATE_* for possible states.
func (ws *WorkflowStep) GetState() int {
	return ws.State
}

// SetStatus sets the current state of the workflow step.
// See common.Constants.WORKFLOW_STATE_* for possible states.
func (ws *WorkflowStep) SetState(state int) {
	ws.State = state
}
