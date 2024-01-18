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

type RaftWorkflowDAO struct {
	logger  *logging.Logger
	raft    RaftNode
	farmDAO dao.FarmDAO
	dao.WorkflowDAO
}

func NewRaftWorkflowDAO(logger *logging.Logger,
	raftNode RaftNode, farmDAO dao.FarmDAO) dao.WorkflowDAO {
	return &RaftWorkflowDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftWorkflowDAO) Save(workflow *config.Workflow) error {
	farmID := workflow.GetFarmID()
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	// if workflow.GetID() == 0 {
	// 	idSetter := dao.raft.GetParams().IdSetter
	// 	idSetter.SetWorkflowIds(farmID, []*config.Workflow{workflow})
	// }
	// if workflow.GetID() == 0 {
	// 	key := fmt.Sprintf("%d-%s", farmID, workflow.GetName())
	// 	id := dao.raft.GetParams().IdGenerator.NewID(key)
	// 	workflow.SetID(id)
	// 	steps := workflow.GetSteps()
	// 	for _, step := range steps {
	// 		if step.GetID() == 0 {
	// 			stepKey := fmt.Sprintf("%s-%d-%d-%d-%d", key, step.GetDeviceID(),
	// 				step.GetChannelID(), step.GetDuration(), step.GetState())
	// 			stepID := dao.raft.GetParams().IdGenerator.NewID(stepKey)
	// 			step.SetID(stepID)
	// 		}
	// 		step.SetWorkflowID(id)
	// 	}
	// 	workflow.SetSteps(steps)
	// }
	farmConfig.SetWorkflow(workflow)
	return dao.farmDAO.Save(farmConfig)
}

func (dao *RaftWorkflowDAO) Get(farmID, workflowID uint64,
	CONSISTENCY_LEVEL int) (*config.Workflow, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	for _, workflow := range farmConfig.GetWorkflows() {
		if workflow.GetID() == workflowID {
			return workflow, nil
		}
	}
	return nil, datastore.ErrNotFound
}

func (dao *RaftWorkflowDAO) Delete(workflow *config.Workflow) error {
	dao.logger.Debugf(fmt.Sprintf("Deleting workflow record: %+v", workflow))
	farmConfig, err := dao.farmDAO.Get(workflow.GetFarmID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	newWorkflowList := make([]*config.Workflow, 0)
	for _, wflow := range farmConfig.GetWorkflows() {
		if wflow.GetID() == workflow.GetID() {
			continue
		}
		newWorkflowList = append(newWorkflowList, wflow)
	}
	farmConfig.SetWorkflows(newWorkflowList)
	return dao.farmDAO.Save(farmConfig)
}

func (dao *RaftWorkflowDAO) GetByFarmID(farmID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Workflow, error) {

	farmConfig, err := dao.farmDAO.Get(farmID, CONSISTENCY_LEVEL)
	if err != nil {
		return nil, err
	}
	return farmConfig.GetWorkflows(), nil
}
