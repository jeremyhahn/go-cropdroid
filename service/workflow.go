package service

import (
	"fmt"
	"strings"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/viewmodel"
)

type WorkflowService interface {
	GetWorkflow(session Session, workflowID uint64) (config.WorkflowConfig, error)
	GetWorkflows(session Session) []config.Workflow
	GetListView(session Session, farmID uint64) ([]*viewmodel.Workflow, error)
	Create(session Session, workflow config.WorkflowConfig) (config.WorkflowConfig, error)
	Update(session Session, workflow config.WorkflowConfig) error
	Delete(session Session, workflow config.WorkflowConfig) error
}

type DefaultWorkflowService struct {
	app    *app.App
	dao    dao.WorkflowDAO
	mapper mapper.WorkflowMapper
	WorkflowService
}

// NewWorkflowService creates a new default WorkflowService instance
func NewWorkflowService(app *app.App, dao dao.WorkflowDAO, mapper mapper.WorkflowMapper) WorkflowService {
	return &DefaultWorkflowService{
		app:    app,
		dao:    dao,
		mapper: mapper}
}

// GetWorkflow retrieves a specific workflow entry from the current FarmConfig
func (service *DefaultWorkflowService) GetWorkflow(session Session, workflowID uint64) (config.WorkflowConfig, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.GetID() == workflowID {
			return &workflow, nil
		}
	}
	return nil, ErrWorkflowNotFound
}

// GetWorkflows retrieves a list of workflow entries from the current FarmConfig
func (service *DefaultWorkflowService) GetWorkflows(session Session) []config.Workflow {
	return session.GetFarmService().GetConfig().GetWorkflows()
}

// Create a new workflow entry in the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowService) Create(session Session, workflow config.WorkflowConfig) (config.WorkflowConfig, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	farmConfig.AddWorkflow(workflow)
	err := farmService.SetConfig(farmConfig)
	if err != nil {
		service.app.Logger.Errorf("sesion: %+v, error: %s", session, err)
	}
	return workflow, err
}

func (service *DefaultWorkflowService) GetListView(session Session, farmID uint64) ([]*viewmodel.Workflow, error) {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	viewWorkflows := make([]*viewmodel.Workflow, 0)
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.GetFarmID() == farmID {
			viewWorkflows = append(viewWorkflows,
				service.mapper.MapConfigToView(&workflow))
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
				if channel.GetID() == step.GetChannelID() {
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

	// for _, device := range farmConfig.GetDevices() {
	// 	for _, channel := range device.GetChannels() {
	// 		if channel.GetID() == channelID {
	// 			channelConditions := channel.GetConditions()
	// 			viewConditions := make([]*viewmodel.Condition, 0, len(channelConditions))
	// 			for _, condition := range channelConditions {
	// 				// Look up the metric for this condition
	// 				for _, metric := range device.GetMetrics() {
	// 					if metric.GetID() == condition.GetMetricID() {
	// 						viewConditions = append(viewConditions,
	// 							service.mapper.MapEntityToView(
	// 								&condition, device.GetType(), &metric, channelID))
	// 						break
	// 					}
	// 				}

	// 			}
	// 			return viewConditions, nil
	// 		}
	// 	}
	// }

	return nil, ErrChannelNotFound
}

// Update an existing workflow entry in the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowService) Update(session Session, workflow config.WorkflowConfig) error {
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	farmConfig.SetWorkflow(workflow)
	err := farmService.SetConfig(farmConfig)
	if err != nil {
		service.app.Logger.Errorf("sesion: %+v, error: %s", session, err)
	}
	return err
}

// Delete a workflow entry from the FarmConfig and datastore and publish
// the new FarmConfig to connected clients.
func (service *DefaultWorkflowService) Delete(session Session, workflow config.WorkflowConfig) error {
	// GORM Save does not delete associations, delete it
	// directly via the DAO instead :(
	if err := service.dao.Delete(workflow); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	farmConfig.RemoveWorkflow(workflow)
	err := farmService.SetConfig(farmConfig)
	if err != nil {
		service.app.Logger.Errorf("sesion: %+v, error: %s", session, err)
	}
	return err
}
