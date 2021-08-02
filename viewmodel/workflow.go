package viewmodel

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/model"
)

type Workflow struct {
	ID                    uint64           `yaml:"id" json:"id"`
	FarmID                uint64           `yaml:"farm" json:"farmId"`
	Name                  string           `yaml:"name" json:"name"`
	Conditions            []Condition      `yaml:"conditions" json:"conditions"`
	Schedules             []model.Schedule `yaml:"schedules" json:"schedules"`
	Steps                 []WorkflowStep   `yaml:"steps" json:"steps"`
	config.WorkflowConfig `yaml:"-" json:"-"`
}

func NewWorkflow() *Workflow {
	return &Workflow{
		Conditions: make([]Condition, 0),
		Schedules:  make([]model.Schedule, 0),
		Steps:      make([]WorkflowStep, 0)}
}

// GetID gets the workflow ID
func (w *Workflow) GetID() uint64 {
	return w.ID
}

// SetID sets the workflow ID
func (w *Workflow) SetID(id uint64) {
	w.ID = id
}

// GetID gets the workflow farm ID
func (w *Workflow) GetFarmID() uint64 {
	return w.FarmID
}

// SetID sets the workflow farm ID
func (w *Workflow) SetFarmID(id uint64) {
	w.FarmID = id
}

// GetName gets the workflow name
func (w *Workflow) GetName() string {
	return w.Name
}

// SetName sets the workflow name
func (w *Workflow) SetName(name string) {
	w.Name = name
}

// // GetConditions gets the workflow conditions
// func (w *Workflow) GetConditions() []Condition {
// 	return w.Conditions
// }

// // SetConditions sets the workflow conditions
// func (w *Workflow) SetConditions(conditions []Condition) {
// 	w.Conditions = conditions
// }

// // GetSchedules gets the workflow schedules
// func (w *Workflow) GetSchedules() []Schedule {
// 	return w.Schedules
// }

// // SetSchedules sets the workflow schedules
// func (w *Workflow) SetSchedules(schedules []Schedule) {
// 	w.Schedules = schedules
// }

// GetSteps gets the workflow steps
func (w *Workflow) GetSteps() []WorkflowStep {
	return w.Steps
}

// SetSteps sets the workflow steps
func (w *Workflow) SetSteps(steps []WorkflowStep) {
	w.Steps = steps
}

// SetStep updates / sets an existing workflow step
func (w *Workflow) SetStep(step WorkflowStep) error {
	for i, s := range w.GetSteps() {
		if s.GetID() == step.GetID() {
			w.Steps[i] = step
			return nil
		}
	}
	return config.ErrWorkflowStepNotFound
}
