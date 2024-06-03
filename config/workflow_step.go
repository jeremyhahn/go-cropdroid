package config

import "fmt"

type WorkflowStep interface {
	GetWorkflowID() uint64
	SetWorkflowID(workflowID uint64)
	GetDeviceID() uint64
	SetDeviceID(workflowID uint64)
	GetChannelID() uint64
	SetChannelID(channelID uint64)
	GetWebhook() string
	SetWebhook(url string)
	GetDuration() int
	SetDuration(seconds int)
	GetWait() int
	SetWait(seconds int)
	GetState() int
	SetState(state int)
	GetSortOrder() int
	SetSortOrder(position int)
	KeyValueEntity
}

// WorkflowStepStruct defines a single step for a Workflow.
type WorkflowStepStruct struct {
	ID uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	//Name string `gorm:"name" yaml:"name" json:"name"`
	WorkflowID uint64 `yaml:"workflow" json:"workflow_id"`
	DeviceID   uint64 `yaml:"device" json:"device_id"`
	ChannelID  uint64 `yaml:"channel" json:"channel_id"`
	Webhook    string `yaml:"webhook" json:"webhook"`
	//AlgorithmID        int    `yaml:"algorithm" json:"algorithm_id"`
	Duration     int `yaml:"duration" json:"duration"`
	Wait         int `yaml:"wait" json:"wait"`
	State        int `yaml:"state" json:"state"`
	SortOrder    int `yaml:"sortOrder" json:"sort_order"`
	WorkflowStep `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewWorkflowStep() *WorkflowStepStruct {
	return &WorkflowStepStruct{}
}

func (ws *WorkflowStepStruct) TableName() string {
	return "workflow_steps"
}

// Identifier gets the workflow step ID
func (ws *WorkflowStepStruct) Identifier() uint64 {
	return ws.ID
}

// SetID sets the workflow step ID
func (ws *WorkflowStepStruct) SetID(id uint64) {
	ws.ID = id
}

// GetWorkflowID gets the workflow ID
func (ws *WorkflowStepStruct) GetWorkflowID() uint64 {
	return ws.WorkflowID
}

// SetID sets the workflow step ID
func (ws *WorkflowStepStruct) SetWorkflowID(id uint64) {
	ws.WorkflowID = id
}

// // GetName gets the workflow name
// func (ws *WorkflowStepStruct) GetName() string {
// 	return ws.Name
// }

// // SetName sets the workflow name
// func (ws *WorkflowStepStruct) SetName(name string) {
// 	ws.Name = name
// }

// GetDeviceID returns the device identifier to target for execution
func (ws *WorkflowStepStruct) GetDeviceID() uint64 {
	return ws.DeviceID
}

// SetDeviceID sets the device identifier to target for execution
func (ws *WorkflowStepStruct) SetDeviceID(id uint64) {
	ws.DeviceID = id
}

// GetChannelID gets the target channel ID to execute
func (ws *WorkflowStepStruct) GetChannelID() uint64 {
	return ws.ChannelID
}

// SetChannelID sets the target channel ID to execute
func (ws *WorkflowStepStruct) SetChannelID(id uint64) {
	ws.ChannelID = id
}

// GetWebhook gets the target webhook URL to execute
func (ws *WorkflowStepStruct) GetWebhook() string {
	return ws.Webhook
}

// SetWebhook sets the target webhook URL to execute
func (ws *WorkflowStepStruct) SetWebhook(url string) {
	ws.Webhook = url
}

// GetDuration gets the workflow step duration
func (ws *WorkflowStepStruct) GetDuration() int {
	return ws.Duration
}

// SetDuration sets the workflow step duration
func (ws *WorkflowStepStruct) SetDuration(duration int) {
	ws.Duration = duration
}

// GetWait gets the number of seconds to wait for the workflow step
// before proceeding to the next step.
func (ws *WorkflowStepStruct) GetWait() int {
	return ws.Wait
}

// SetWait sets the number of seconds to wait for the workflow step
// before proceeding to the next step.
func (ws *WorkflowStepStruct) SetWait(seconds int) {
	ws.Wait = seconds
}

// GetStatus returns the current state of the workflow step.
// See common.Constants.WORKFLOW_STATE_* for possible states.
func (ws *WorkflowStepStruct) GetState() int {
	return ws.State
}

// SetStatus sets the current state of the workflow step.
// See common.Constants.WORKFLOW_STATE_* for possible states.
func (ws *WorkflowStepStruct) SetState(state int) {
	ws.State = state
}

// GetOrder returns the order this step comes in the workflow
func (ws *WorkflowStepStruct) GetSortOrder() int {
	return ws.SortOrder
}

// SetOrder sets the order of execution for the workflow step
func (ws *WorkflowStepStruct) SetOrder(sortOrder int) {
	ws.SortOrder = sortOrder
}

func (ws *WorkflowStepStruct) String() string {
	return fmt.Sprintf("%d-%d-%d-%s-%d-%d-%d-%d",
		ws.WorkflowID, ws.DeviceID, ws.ChannelID,
		ws.Webhook, ws.Duration, ws.Wait, ws.State,
		ws.SortOrder)
}
