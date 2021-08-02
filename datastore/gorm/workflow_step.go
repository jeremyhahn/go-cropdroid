package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormWorkflowStepDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.WorkflowStepDAO
}

func NewWorkflowStepDAO(logger *logging.Logger, db *gorm.DB) dao.WorkflowStepDAO {
	return &GormWorkflowStepDAO{logger: logger, db: db}
}

func (dao *GormWorkflowStepDAO) Create(workflow config.WorkflowStepConfig) error {
	return dao.db.Create(workflow).Error
}

func (dao *GormWorkflowStepDAO) Save(workflow config.WorkflowStepConfig) error {
	return dao.db.Save(workflow).Error
}

func (dao *GormWorkflowStepDAO) Delete(workflow config.WorkflowStepConfig) error {
	step := &config.WorkflowStep{}
	step.SetWorkflowID(workflow.GetID())
	if err := dao.db.Model(step).Delete(step).Error; err != nil {
		return err
	}
	return dao.db.Delete(workflow).Error
}

func (dao *GormWorkflowStepDAO) Get(id uint64) (config.WorkflowStepConfig, error) {
	var step config.WorkflowStep
	if err := dao.db.First(&step, id).Error; err != nil {
		return nil, err
	}
	return &step, nil
}

func (dao *GormWorkflowStepDAO) GetByWorkflowID(id uint64) ([]config.WorkflowStep, error) {
	var steps []config.WorkflowStep
	if err := dao.db.
		Where("workflow_id = ?", id).
		Find(&steps).Error; err != nil {
		return nil, err
	}
	return steps, nil
}
