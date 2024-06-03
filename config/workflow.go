package config

import (
	"time"
)

type CommonWorkflow interface {
	GetFarmID() uint64
	SetFarmID(farmID uint64)
	GetName() string
	SetName(name string)
	GetLastCompleted() *time.Time
	SetLastCompleted(time *time.Time)
	KeyValueEntity
}

type Workflow interface {
	GetConditions() []*ConditionStruct
	SetConditions(conditions []*ConditionStruct)
	GetSchedules() []*ScheduleStruct
	SetSchedules(schedules []*ScheduleStruct)
	GetSteps() []*WorkflowStepStruct
	SetSteps(steps []*WorkflowStepStruct)
	SetStep(step *WorkflowStepStruct)
	AddStep(step *WorkflowStepStruct)
	RemoveStep(step *WorkflowStepStruct) error
	CommonWorkflow
}

// WorkflowStruct defines a series of workflow steps that can
// be triggered manually or based on a Schedule or Condition.
type WorkflowStruct struct {
	ID            uint64                `gorm:"primaryKey" yaml:"id" json:"id"`
	FarmID        uint64                `yaml:"farm" json:"farm_id"`
	Name          string                `gorm:"name" yaml:"name" json:"name"`
	Conditions    []*ConditionStruct    `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"conditions" json:"conditions"`
	Schedules     []*ScheduleStruct     `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"schedules" json:"schedules"`
	Steps         []*WorkflowStepStruct `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"steps" json:"steps"`
	LastCompleted *time.Time            `gorm:"type:timestamp" yaml:"lastCompleted" json:"lastCompleted"`
	Workflow      `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewWorkflow() *WorkflowStruct {
	return &WorkflowStruct{
		Conditions: make([]*ConditionStruct, 0),
		Schedules:  make([]*ScheduleStruct, 0),
		Steps:      make([]*WorkflowStepStruct, 0)}
}

func (w *WorkflowStruct) TableName() string {
	return "workflows"
}

// Identifier gets the workflow ID
func (w *WorkflowStruct) Identifier() uint64 {
	return w.ID
}

// SetID sets the workflow ID
func (w *WorkflowStruct) SetID(id uint64) {
	w.ID = id
}

// GetID gets the workflow farm ID
func (w *WorkflowStruct) GetFarmID() uint64 {
	return w.FarmID
}

// SetID sets the workflow farm ID
func (w *WorkflowStruct) SetFarmID(id uint64) {
	w.FarmID = id
}

// GetName gets the workflow name
func (w *WorkflowStruct) GetName() string {
	return w.Name
}

// SetName sets the workflow name
func (w *WorkflowStruct) SetName(name string) {
	w.Name = name
}

// GetConditions gets the workflow conditions
func (w *WorkflowStruct) GetConditions() []*ConditionStruct {
	return w.Conditions
}

// SetConditions sets the workflow conditions
func (w *WorkflowStruct) SetConditions(conditions []*ConditionStruct) {
	w.Conditions = conditions
}

// GetSchedules gets the workflow schedules
func (w *WorkflowStruct) GetSchedules() []*ScheduleStruct {
	return w.Schedules
}

// SetSchedules sets the workflow schedules
func (w *WorkflowStruct) SetSchedules(schedules []*ScheduleStruct) {
	w.Schedules = schedules
}

// GetSteps gets the workflow steps
func (w *WorkflowStruct) GetSteps() []*WorkflowStepStruct {
	return w.Steps
}

// SetStep updates / sets an existing workflow step
func (w *WorkflowStruct) SetStep(step *WorkflowStepStruct) {
	for i, s := range w.GetSteps() {
		if s.ID == step.ID {
			w.Steps[i] = step
			return
		}
	}
	step.SetOrder(len(w.Steps) + 1)
	w.Steps = append(w.Steps, step)
}

// SetSteps sets the workflow steps. The order of the steps are
// preserved.
func (w *WorkflowStruct) SetSteps(steps []*WorkflowStepStruct) {
	for i, step := range steps {
		step.SetOrder(i + 1)
	}
	w.Steps = steps
}

// AddStep adds a new workflow step as the last step in the workflow.
func (w *WorkflowStruct) AddStep(step *WorkflowStepStruct) {
	step.SetOrder(len(w.Steps) + 1)
	w.Steps = append(w.Steps, step)
}

// Removes the specified workflow step from the workflow.
// The steps are re-ordered after the step is removed.
func (w *WorkflowStruct) RemoveStep(step *WorkflowStepStruct) error {
	for i, s := range w.Steps {
		if s.ID == step.ID {
			w.Steps = append(w.Steps[:i], w.Steps[i+1:]...)
			return nil
		}
	}
	for i, s := range w.Steps {
		s.SetOrder(i + 1)
	}
	return ErrWorkflowNotFound
}

// GetLastCompleted returns the time the workflow was last
// successfully completed.
func (w *WorkflowStruct) GetLastCompleted() *time.Time {
	return w.LastCompleted
}

// SetLastCompleted sets the time the workflow was last
// successfully completed.
func (w *WorkflowStruct) SetLastCompleted(t *time.Time) {
	w.LastCompleted = t
}
