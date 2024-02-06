package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormWorkflowStepDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.WorkflowStepDAO
}

func NewWorkflowStepDAO(logger *logging.Logger, db *gorm.DB) dao.WorkflowStepDAO {
	return &GormWorkflowStepDAO{logger: logger, db: db}
}

func (dao *GormWorkflowStepDAO) Save(farmID uint64,
	workflowStep *config.WorkflowStep) error {

	return dao.db.Save(workflowStep).Error
}

func (dao *GormWorkflowStepDAO) Delete(farmID uint64,
	workflowStep *config.WorkflowStep) error {

	step := &config.WorkflowStep{}
	step.SetWorkflowID(workflowStep.GetWorkflowID())
	// if err := dao.db.Model(step).Delete(step).Error; err != nil {
	// 	return err
	// }
	// return dao.db.Delete(workflowStep).Error
	return dao.db.Model(step).Delete(workflowStep).Error
}

func (dao *GormWorkflowStepDAO) Get(farmID, workflowID,
	workflowStepID uint64, CONSISTENCY_LEVEL int) (*config.WorkflowStep, error) {

	var step *config.WorkflowStep
	if err := dao.db.First(step, workflowStepID).Error; err != nil {
		return nil, err
	}
	return step, nil
}

func (dao *GormWorkflowStepDAO) GetByWorkflowID(farmID,
	workflowID uint64, CONSISTENCY_LEVEL int) ([]*config.WorkflowStep, error) {

	var steps []*config.WorkflowStep
	if err := dao.db.
		Where("workflow_id = ?", workflowID).
		Find(&steps).Error; err != nil {
		return nil, err
	}
	return steps, nil
}
