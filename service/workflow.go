package service

import (
	"fmt"
	"strings"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type WorkflowService interface {
	GetWorkflow(session Session, workflowID uint64) (config.Workflow, error)
	GetWorkflows(session Session) []config.Workflow
	GetListView(session Session, farmID uint64) ([]*viewmodel.Workflow, error)
	Create(session Session, workflow config.Workflow) (config.Workflow, error)
	Update(session Session, workflow config.Workflow) error
	Delete(session Session, workflow config.Workflow) error
	Run(session Session, workflowID uint64) error
}

type DefaultWorkflowService struct {
	app    *app.App
	dao    dao.WorkflowDAO
	mapper mapper.WorkflowMapper
	WorkflowService
}

// NewWorkflowService creates a new default WorkflowService instance
func NewWorkflowService(
	app *app.App, dao dao.WorkflowDAO,
	mapper mapper.WorkflowMapper) WorkflowService {

	return &DefaultWorkflowService{
		app:    app,
		dao:    dao,
		mapper: mapper}
}

// GetWorkflow retrieves a specific workflow entry from the current FarmConfig
func (service *DefaultWorkflowService) GetWorkflow(session Session,
	workflowID uint64) (config.Workflow, error) {

	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.ID == workflowID {
			return workflow, nil
		}
	}
	return nil, ErrWorkflowNotFound
}

// GetWorkflows retrieves a list of workflow entries from the current FarmConfig
func (service *DefaultWorkflowService) GetWorkflows(session Session) []config.Workflow {
	workflows := session.GetFarmService().GetConfig().GetWorkflows()
	_workflows := make([]config.Workflow, len(workflows))
	for i, workflow := range workflows {
		_workflows[i] = workflow
	}
	return _workflows
}

// Create a new workflow entry in the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowService) Create(session Session, workflow config.Workflow) (config.Workflow, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	farmConfig.AddWorkflow(workflow.(*config.WorkflowStruct))
	err := farmService.SetConfig(farmConfig)
	if err != nil {
		service.app.Logger.Errorf("sesion: %+v, error: %s", session, err)
	}
	return workflow, err
}

// Returns a list of workflow viewmodels intended for consumption by a user interface
func (service *DefaultWorkflowService) GetListView(session Session, farmID uint64) ([]*viewmodel.Workflow, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	viewWorkflows := make([]*viewmodel.Workflow, 0)
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.GetFarmID() == farmID {
			viewWorkflows = append(viewWorkflows,
				service.mapper.MapConfigToView(workflow))
		}
	}
	// Set workflow steps device name and channel name
	for _, viewWorkflow := range viewWorkflows {
		for _, step := range viewWorkflow.GetSteps() {
			device, err := farmConfig.GetDeviceById(step.GetDeviceID())
			if err != nil {
				return nil, err
			}
			deviceType := strings.Title(strings.ToLower(device.GetType()))
			step.SetDeviceType(deviceType)
			//viewWorkflow.SetStep(step)
			for _, channel := range device.GetChannels() {
				if channel.ID == step.GetChannelID() {
					channelName := channel.GetName()
					step.SetChannelName(channelName)
					step.SetText(fmt.Sprintf("%s %s ON for %ds, wait %ds",
						deviceType, channelName, step.GetDuration(), step.GetWait()))

					viewWorkflow.SetStep(step)
					break
				}
			}
		}
	}
	return viewWorkflows, nil
}

// Update an existing workflow entry in the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowService) Update(session Session, workflow config.Workflow) error {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	farmConfig.SetWorkflow(workflow.(*config.WorkflowStruct))
	err := farmService.SetConfig(farmConfig)
	if err != nil {
		service.app.Logger.Errorf("sesion: %+v, error: %s", session, err)
	}
	return err
}

// Delete a workflow entry from the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowService) Delete(session Session, workflow config.Workflow) error {
	// GORM Save does not delete associations, delete steps
	// directly via the DAO instead :(
	if err := service.dao.Delete(workflow.(*config.WorkflowStruct)); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	farmConfig.RemoveWorkflow(workflow.(*config.WorkflowStruct))
	err := farmService.SetConfig(farmConfig)
	if err != nil {
		service.app.Logger.Errorf("sesion: %+v, error: %s", session, err)
	}
	return err
}

// Executes a workflow
func (service *DefaultWorkflowService) Run(session Session, workflowID uint64) error {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.ID == workflowID {
			session.GetFarmService().RunWorkflow(workflow)
			return nil
		}
	}
	return ErrWorkflowNotFound
}
