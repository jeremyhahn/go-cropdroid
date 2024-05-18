package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
)

type WorkflowStepService interface {
	GetStep(session Session, workflowID, stepID uint64) (*config.WorkflowStep, error)
	GetSteps(session Session, workflowID uint64) ([]*config.WorkflowStep, error)
	Create(session Session, step *config.WorkflowStep) (*config.WorkflowStep, error)
	Update(session Session, step *config.WorkflowStep) error
	Delete(session Session, step *config.WorkflowStep) error
}

type DefaultWorkflowStepService struct {
	app *app.App
	dao dao.WorkflowStepDAO
	WorkflowStepService
}

// NewWorkflowService creates a new default WorkflowService instance
func NewWorkflowStepService(app *app.App, dao dao.WorkflowStepDAO) WorkflowStepService {
	return &DefaultWorkflowStepService{
		app: app,
		dao: dao}
}

// GetWorkflow retrieves a specific workflow entry from the current FarmConfig
func (service *DefaultWorkflowStepService) GetStep(session Session, workflowID, stepID uint64) (*config.WorkflowStep, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.ID == workflowID {
			for _, step := range workflow.GetSteps() {
				if step.ID == stepID {
					return step, nil
				}
			}
			return nil, ErrWorkflowStepNotFound
		}
	}
	return nil, ErrWorkflowNotFound
}

// GetWorkflows retrieves a list of workflow entries from the current FarmConfig
func (service *DefaultWorkflowStepService) GetSteps(session Session, workflowID uint64) ([]*config.WorkflowStep, error) {
	for _, workflow := range session.GetFarmService().GetConfig().GetWorkflows() {
		if workflow.ID == workflowID {
			return workflow.GetSteps(), nil
		}
	}
	return nil, ErrWorkflowNotFound
}

// Create a new workflow entry in the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowStepService) Create(session Session, step *config.WorkflowStep) (*config.WorkflowStep, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.ID == step.GetWorkflowID() {
			workflow.AddStep(step)
			farmConfig.SetWorkflow(workflow)
			return step, farmService.SetConfig(farmConfig)
		}
	}
	return nil, ErrWorkflowNotFound
}

// Update an existing workflow entry in the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowStepService) Update(session Session, step *config.WorkflowStep) error {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.ID == step.GetWorkflowID() {
			workflow.SetStep(step)
			return farmService.SetConfig(farmConfig)
		}
	}
	return ErrWorkflowNotFound
}

// Delete a workflow entry from the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowStepService) Delete(session Session, step *config.WorkflowStep) error {
	// farmService.SetConfig doesnt delete the workflow :(
	farmID := session.GetRequestedFarmID()
	if err := service.dao.Delete(farmID, step); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.ID == step.GetWorkflowID() {
			workflow.RemoveStep(step)
			farmConfig.SetWorkflow(workflow)
			return farmService.SetConfig(farmConfig)
		}
	}
	return ErrWorkflowNotFound
}
