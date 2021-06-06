package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormAlgorithmDAO struct {
	db     *gorm.DB
	logger *logging.Logger
	dao.AlgorithmDAO
}

func NewAlgorithmDAO(logger *logging.Logger, db *gorm.DB) dao.AlgorithmDAO {
	return &GormAlgorithmDAO{logger: logger, db: db}
}

func (dao *GormAlgorithmDAO) Create(algorithm config.AlgorithmConfig) error {
	return dao.db.Create(algorithm).Error
}

/*
func (dao *GormAlgorithmDAO) Save(res entity.AlgorithmEntity) error {
	return dao.db.Save(res).Error
}

func (dao *GormAlgorithmDAO) Update(res entity.AlgorithmEntity) error {
	return dao.db.Update(res).Error
}

func (dao *GormAlgorithmDAO) Get(name string) (entity.AlgorithmEntity, error) {
	var Algorithms []entity.Algorithm
	if err := dao.db.Where("name = ?", name).Find(&Algorithms).Error; err != nil {
		return nil, err
	}
	if len(Algorithms) == 0 {
		return nil, errors.New(fmt.Sprintf("Algorithm '%s' not found in database", name))
	}
	return &Algorithms[0], nil
}
*/

func (dao *GormAlgorithmDAO) GetAll() ([]config.Algorithm, error) {
	var algorithms []config.Algorithm
	if err := dao.db.Find(&algorithms).Error; err != nil {
		return nil, err
	}
	return algorithms, nil
}
