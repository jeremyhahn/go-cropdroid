package config

type WorkflowStep struct {
	ID uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	//Name string `gorm:"name" yaml:"name" json:"name"`
	WorkflowID uint64 `yaml:"workflow" json:"workflow_id"`
	DeviceID   uint64 `yaml:"device" json:"device_id"`
	ChannelID  uint64 `yaml:"channel" json:"channel_id"`
	Webhook    string `yaml:"webhook" json:"webhook"`
	//AlgorithmID        int    `yaml:"algorithm" json:"algorithm_id"`
	Duration           int `yaml:"duration" json:"duration"`
	Wait               int `yaml:"wait" json:"wait"`
	State              int `yaml:"state" json:"state"`
	WorkflowStepConfig `yaml:"-" json:"-"`
}

func NewWorkflowStep() *WorkflowStep {
	return &WorkflowStep{}
}

// GetID gets the workflow step ID
func (ws *WorkflowStep) GetID() uint64 {
	return ws.ID
}

// SetID sets the workflow step ID
func (ws *WorkflowStep) SetID(id uint64) {
	ws.ID = id
}

// GetWorkflowID gets the workflow ID
func (ws *WorkflowStep) GetWorkflowID() uint64 {
	return ws.WorkflowID
}

// SetID sets the workflow step ID
func (ws *WorkflowStep) SetWorkflowID(id uint64) {
	ws.WorkflowID = id
}

// // GetName gets the workflow name
// func (ws *WorkflowStep) GetName() string {
// 	return ws.Name
// }

// // SetName sets the workflow name
// func (ws *WorkflowStep) SetName(name string) {
// 	ws.Name = name
// }

// GetDeviceID returns the device identifier to target for execution
func (ws *WorkflowStep) GetDeviceID() uint64 {
	return ws.DeviceID
}

// SetDeviceID sets the device identifier to target for execution
func (ws *WorkflowStep) SetDeviceID(id uint64) {
	ws.DeviceID = id
}

// GetChannelID gets the target channel ID to execute
func (ws *WorkflowStep) GetChannelID() uint64 {
	return ws.ChannelID
}

// SetChannelID sets the target channel ID to execute
func (ws *WorkflowStep) SetChannelID(id uint64) {
	ws.ChannelID = id
}

// GetWebhook gets the target webhook URL to execute
func (ws *WorkflowStep) GetWebhook() string {
	return ws.Webhook
}

// SetWebhook sets the target webhook URL to execute
func (ws *WorkflowStep) SetWebhook(url string) {
	ws.Webhook = url
}

// GetDuration gets the workflow step duration
func (ws *WorkflowStep) GetDuration() int {
	return ws.Duration
}

// SetDuration sets the workflow step duration
func (ws *WorkflowStep) SetDuration(duration int) {
	ws.Duration = duration
}

// GetWait gets the number of seconds to wait for the workflow step
// before proceeding to the next step.
func (ws *WorkflowStep) GetWait() int {
	return ws.Wait
}

// SetWait sets the number of seconds to wait for the workflow step
// before proceeding to the next step.
func (ws *WorkflowStep) SetWait(seconds int) {
	ws.Wait = seconds
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
