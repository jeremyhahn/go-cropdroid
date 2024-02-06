package config

import "time"

// Workflow defines an object responsible for executing a series
// of WorkflowStep, which can be triggered manually or based on a
// Schedule or Condition.
type Workflow struct {
	ID            uint64          `gorm:"primaryKey" yaml:"id" json:"id"`
	FarmID        uint64          `yaml:"farm" json:"farm_id"`
	Name          string          `gorm:"name" yaml:"name" json:"name"`
	Conditions    []*Condition    `gorm:"conditions" yaml:"conditions" json:"conditions"`
	Schedules     []*Schedule     `gorm:"schedules" yaml:"schedules" json:"schedules"`
	Steps         []*WorkflowStep `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"steps" json:"steps"`
	LastCompleted *time.Time      `gorm:"type:timestamp" yaml:"lastCompleted" json:"lastCompleted"`
}

func NewWorkflow() *Workflow {
	return &Workflow{
		Conditions: make([]*Condition, 0),
		Schedules:  make([]*Schedule, 0),
		Steps:      make([]*WorkflowStep, 0)}
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

// GetConditions gets the workflow conditions
func (w *Workflow) GetConditions() []*Condition {
	return w.Conditions
}

// SetConditions sets the workflow conditions
func (w *Workflow) SetConditions(conditions []*Condition) {
	w.Conditions = conditions
}

// GetSchedules gets the workflow schedules
func (w *Workflow) GetSchedules() []*Schedule {
	return w.Schedules
}

// SetSchedules sets the workflow schedules
func (w *Workflow) SetSchedules(schedules []*Schedule) {
	w.Schedules = schedules
}

// GetSteps gets the workflow steps
func (w *Workflow) GetSteps() []*WorkflowStep {
	return w.Steps
}

// SetStep updates / sets an existing workflow step
func (w *Workflow) SetStep(step *WorkflowStep) {
	for i, s := range w.GetSteps() {
		if s.GetID() == step.GetID() {
			w.Steps[i] = step
			return
		}
	}
	step.SetOrder(len(w.Steps) + 1)
	w.Steps = append(w.Steps, step)
}

// SetSteps sets the workflow steps. The order of the steps are
// preserved.
func (w *Workflow) SetSteps(steps []*WorkflowStep) {
	for i, step := range steps {
		step.SetOrder(i + 1)
	}
	w.Steps = steps
}

// AddStep adds a new workflow step as the last step in the workflow.
func (w *Workflow) AddStep(step *WorkflowStep) {
	step.SetOrder(len(w.Steps) + 1)
	w.Steps = append(w.Steps, step)
}

// Removes the specified workflow step from the workflow.
// The steps are re-ordered after the step is removed.
func (w *Workflow) RemoveStep(step *WorkflowStep) error {
	for i, s := range w.Steps {
		if s.GetID() == step.GetID() {
			w.Steps = append(w.Steps[:i], w.Steps[i+1:]...)
			return nil
		}
	}
	for i, s := range w.Steps {
		s.SetOrder(i + 1)
	}
	return ErrWorkflowStepNotFound
}

// GetLastCompleted returns the time the workflow was last
// successfully completed.
func (w *Workflow) GetLastCompleted() *time.Time {
	return w.LastCompleted
}

// SetLastCompleted sets the time the workflow was last
// successfully completed.
func (w *Workflow) SetLastCompleted(t *time.Time) {
	w.LastCompleted = t
}
