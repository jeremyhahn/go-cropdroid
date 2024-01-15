package mapper

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type WorkflowMapper interface {
	MapConfigToView(config *config.Workflow) *viewmodel.Workflow
	MapViewToConfig(*viewmodel.Workflow) *config.Workflow
}

type DefaultWorkflowMapper struct {
}

func NewWorkflowMapper() WorkflowMapper {
	return &DefaultWorkflowMapper{}
}

func (mapper *DefaultWorkflowMapper) MapConfigToView(config *config.Workflow) *viewmodel.Workflow {
	steps := make([]viewmodel.WorkflowStep, len(config.GetSteps()))
	for i, step := range config.GetSteps() {
		steps[i] = viewmodel.WorkflowStep{
			ID:         step.GetID(),
			WorkflowID: step.GetWorkflowID(),
			DeviceID:   step.GetDeviceID(),
			ChannelID:  step.GetChannelID(),
			Webhook:    step.GetWebhook(),
			Duration:   step.GetDuration(),
			Wait:       step.GetWait(),
			State:      step.GetState()}
	}
	return &viewmodel.Workflow{
		ID:            config.GetID(),
		FarmID:        config.GetFarmID(),
		Name:          config.GetName(),
		LastCompleted: config.GetLastCompleted(),
		Steps:         steps}
}

func (mapper *DefaultWorkflowMapper) MapViewToConfig(workflow *viewmodel.Workflow) *config.Workflow {
	steps := make([]*config.WorkflowStep, len(workflow.GetSteps()))
	for i, step := range workflow.GetSteps() {
		steps[i] = &config.WorkflowStep{
			ID:         step.GetID(),
			WorkflowID: step.GetWorkflowID(),
			DeviceID:   step.GetDeviceID(),
			ChannelID:  step.GetChannelID(),
			Webhook:    step.GetWebhook(),
			Duration:   step.GetDuration(),
			Wait:       step.GetWait(),
			State:      step.GetState()}
	}
	return &config.Workflow{
		ID:            workflow.GetID(),
		FarmID:        workflow.GetFarmID(),
		Name:          workflow.GetName(),
		LastCompleted: workflow.GetLastCompleted(),
		Steps:         steps}
}
