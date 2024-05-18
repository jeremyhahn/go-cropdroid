package gorm

import (
	"sort"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormWorkflowDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.WorkflowDAO
}

func NewWorkflowDAO(logger *logging.Logger, db *gorm.DB) dao.WorkflowDAO {
	return &GormWorkflowDAO{logger: logger, db: db}
}

func (dao *GormWorkflowDAO) Save(workflow *config.Workflow) error {
	return dao.db.Save(workflow).Error
}

func (dao *GormWorkflowDAO) Delete(workflow *config.Workflow) error {
	if err := dao.db.
		Where("workflow_id = ?", workflow.ID).
		Delete(&config.WorkflowStep{}).
		Error; err != nil {
		return err
	}
	return dao.db.Delete(workflow).Error
}

func (dao *GormWorkflowDAO) Get(farmID, workflowID uint64,
	CONSISTENCY_LEVEL int) (*config.Workflow, error) {
	var workflow *config.Workflow
	if err := dao.db.First(workflow, workflowID).Error; err != nil {
		return nil, err
	}
	workflowSteps := workflow.GetSteps()
	sort.SliceStable(workflowSteps, func(i, j int) bool {
		return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
	})
	return workflow, nil
}

func (dao *GormWorkflowDAO) GetByFarmID(farmID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Workflow, error) {
	var workflows []*config.Workflow
	if err := dao.db.
		Preload("Steps").
		Where("farm_id = ?", farmID).
		Find(&workflows).Error; err != nil {
		return nil, err
	}
	// TODO: Replace with order by workflow_steps.sort_order
	for _, workflow := range workflows {
		workflowSteps := workflow.GetSteps()
		sort.SliceStable(workflowSteps, func(i, j int) bool {
			return workflowSteps[i].GetSortOrder() < workflowSteps[j].GetSortOrder()
		})
	}
	return workflows, nil
}
