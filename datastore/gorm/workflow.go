package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormWorkflowDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.WorkflowDAO
}

func NewWorkflowDAO(logger *logging.Logger, db *gorm.DB) dao.WorkflowDAO {
	return &GormWorkflowDAO{logger: logger, db: db}
}

func (dao *GormWorkflowDAO) Create(workflow config.WorkflowConfig) error {
	return dao.db.Create(workflow).Error
}

func (dao *GormWorkflowDAO) Save(workflow config.WorkflowConfig) error {
	return dao.db.Save(workflow).Error
}

func (dao *GormWorkflowDAO) Delete(workflow config.WorkflowConfig) error {
	step := &config.WorkflowStep{}
	step.SetWorkflowID(workflow.GetID())
	if err := dao.db.Model(step).Delete(step).Error; err != nil {
		return err
	}
	return dao.db.Delete(workflow).Error
}

func (dao *GormWorkflowDAO) Get(id uint64) (config.WorkflowConfig, error) {
	var workflow config.Workflow
	if err := dao.db.First(&workflow, id).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (dao *GormWorkflowDAO) GetByFarmID(id uint64) ([]config.Workflow, error) {
	var workflows []config.Workflow
	if err := dao.db.
		Preload("Steps").
		Where("farm_id = ?", id).
		Find(&workflows).Error; err != nil {
		return nil, err
	}
	return workflows, nil
}
