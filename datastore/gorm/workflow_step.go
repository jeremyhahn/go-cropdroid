package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
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
	workflowStep *config.WorkflowStepStruct) error {

	return dao.db.Save(workflowStep).Error
}

func (dao *GormWorkflowStepDAO) Delete(farmID uint64,
	workflowStep *config.WorkflowStepStruct) error {

	step := &config.WorkflowStepStruct{}
	step.SetWorkflowID(workflowStep.GetWorkflowID())
	// if err := dao.db.Model(step).Delete(step).Error; err != nil {
	// 	return err
	// }
	// return dao.db.Delete(workflowStep).Error
	return dao.db.Model(step).Delete(workflowStep).Error
}

func (dao *GormWorkflowStepDAO) Get(farmID, workflowID,
	workflowStepID uint64, CONSISTENCY_LEVEL int) (*config.WorkflowStepStruct, error) {

	var step *config.WorkflowStepStruct
	if err := dao.db.First(step, workflowStepID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			dao.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		dao.logger.Error(err)
		return nil, err
	}
	return step, nil
}

func (dao *GormWorkflowStepDAO) GetByWorkflowID(farmID,
	workflowID uint64, CONSISTENCY_LEVEL int) ([]*config.WorkflowStepStruct, error) {

	var steps []*config.WorkflowStepStruct
	if err := dao.db.
		Where("workflow_id = ?", workflowID).
		Find(&steps).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			dao.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		dao.logger.Error(err)
		return nil, err
	}
	return steps, nil
}
