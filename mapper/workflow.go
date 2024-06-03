package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type WorkflowMapper interface {
	MapConfigToView(config *config.WorkflowStruct) *viewmodel.Workflow
	MapViewToConfig(*viewmodel.Workflow) *config.WorkflowStruct
}

type DefaultWorkflowMapper struct {
}

func NewWorkflowMapper() WorkflowMapper {
	return &DefaultWorkflowMapper{}
}

func (mapper *DefaultWorkflowMapper) MapConfigToView(config *config.WorkflowStruct) *viewmodel.Workflow {
	steps := make([]viewmodel.WorkflowStep, len(config.GetSteps()))
	for i, step := range config.GetSteps() {
		steps[i] = viewmodel.WorkflowStep{
			ID:         step.ID,
			WorkflowID: step.GetWorkflowID(),
			DeviceID:   step.GetDeviceID(),
			ChannelID:  step.GetChannelID(),
			Webhook:    step.GetWebhook(),
			Duration:   step.GetDuration(),
			Wait:       step.GetWait(),
			State:      step.GetState()}
	}
	return &viewmodel.Workflow{
		ID:            config.ID,
		FarmID:        config.GetFarmID(),
		Name:          config.GetName(),
		LastCompleted: config.GetLastCompleted(),
		Steps:         steps}
}

func (mapper *DefaultWorkflowMapper) MapViewToConfig(workflow *viewmodel.Workflow) *config.WorkflowStruct {
	steps := make([]*config.WorkflowStepStruct, len(workflow.GetSteps()))
	for i, step := range workflow.GetSteps() {
		steps[i] = &config.WorkflowStepStruct{
			ID:         step.ID,
			WorkflowID: step.GetWorkflowID(),
			DeviceID:   step.GetDeviceID(),
			ChannelID:  step.GetChannelID(),
			Webhook:    step.GetWebhook(),
			Duration:   step.GetDuration(),
			Wait:       step.GetWait(),
			State:      step.GetState()}
	}
	return &config.WorkflowStruct{
		ID:            workflow.ID,
		FarmID:        workflow.GetFarmID(),
		Name:          workflow.GetName(),
		LastCompleted: workflow.GetLastCompleted(),
		Steps:         steps}
}
