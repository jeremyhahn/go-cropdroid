//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"fmt"
	"sort"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"

	logging "github.com/op/go-logging"
)

type RaftWorkflowDAO struct {
	logger  *logging.Logger
	raft    cluster.RaftNode
	farmDAO dao.FarmDAO
	dao.WorkflowDAO
}

func NewRaftWorkflowDAO(logger *logging.Logger,
	raftNode cluster.RaftNode, farmDAO dao.FarmDAO) dao.WorkflowDAO {
	return &RaftWorkflowDAO{
		logger:  logger,
		raft:    raftNode,
		farmDAO: farmDAO}
}

func (dao *RaftWorkflowDAO) Save(workflow *config.Workflow) error {
	idSetter := dao.raft.GetParams().IdSetter
	farmID := workflow.GetFarmID()
	farmConfig, err := dao.farmDAO.Get(farmID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}
	idSetter.SetWorkflowIds(farmID, []*config.Workflow{workflow})
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
		if workflow.ID == workflowID {
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
		if wflow.ID == workflow.ID {
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
	workflows := farmConfig.GetWorkflows()
	for _, workflow := range workflows {
		workflowSteps := workflow.GetSteps()
		sort.SliceStable(workflowSteps, func(i, j int) bool {
			return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
		})
	}
	return workflows, nil
}
