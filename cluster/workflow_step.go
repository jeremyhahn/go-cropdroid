//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"

	logging "github.com/op/go-logging"
)

type RaftWorkflowStepDAO struct {
	logger  *logging.Logger
	raft    RaftNode
	farmDAO dao.FarmDAO
	dao.WorkflowStepDAO
}

func NewRaftWorkflowStepDAO(logger *logging.Logger,
	raftNode RaftNode, farmDAO dao.FarmDAO) dao.WorkflowStepDAO {
	return &RaftWorkflowStepDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftWorkflowStepDAO) Save(farmID uint64, workflowStep *config.WorkflowStep) error {
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.GetID() == workflowStep.GetWorkflowID() {
			workflow.SetStep(workflowStep)
			farmConfig.SetWorkflow(workflow)
			return dao.farmDAO.Save(farmConfig)
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftWorkflowStepDAO) Get(farmID, workflowID, workflowStepID uint64,
	CONSISTENCY_LEVEL int) (*config.WorkflowStep, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	for _, workflow := range farmConfig.GetWorkflows() {
		for _, workflowStep := range workflow.GetSteps() {
			if workflowStep.GetID() == workflowStepID {
				return workflowStep, nil
			}
		}
	}
	return nil, datastore.ErrNotFound
}

func (dao *RaftWorkflowStepDAO) Delete(farmID uint64, workflowStep *config.WorkflowStep) error {
	dao.logger.Debugf(fmt.Sprintf("Deleting workflowStep record: %+v", workflowStep))
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	newWorkflowStepList := make([]*config.WorkflowStep, 0)
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.GetID() == workflowStep.GetWorkflowID() {
			for _, step := range workflow.GetSteps() {
				if step.GetID() == workflowStep.GetID() {
					continue
				}
			}
			newWorkflowStepList = append(newWorkflowStepList, workflowStep)
			workflow.SetSteps(newWorkflowStepList)
			farmConfig.SetWorkflow(workflow)
			return dao.farmDAO.Save(farmConfig)
		}
	}
	return datastore.ErrNotFound
}

func (dao *RaftWorkflowStepDAO) GetByWorkflowID(farmID, workflowID uint64,
	CONSISTENCY_LEVEL int) ([]*config.WorkflowStep, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, CONSISTENCY_LEVEL)
	if err != nil {
		return nil, err
	}
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.GetID() == workflowID {
			return workflow.GetSteps(), nil
		}
	}
	return nil, datastore.ErrNotFound
}
