package service

// import (
// 	"github.com/jeremyhahn/go-cropdroid/common"
// 	"github.com/jeremyhahn/go-cropdroid/config"
// 	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
// 	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
// )

// type GenericService[D gorm.GenericDAO, E config.GenericEntity] interface {
// 	Save(entity M) error
// 	Get(id uint64, CONSISTENCY_LEVEL int) (E, error)
// 	GetPage(CONSISTENCY_LEVEL, page, pageSize int) ([]E, error)
// 	Update(model M) error
// 	Delete(model M) error
// }

// type BaseService[D gorm.GenericDAO, E config.GenericEntity] struct {
// 	dao dao.GenericDAO
// 	GenericService
// }

// func NewGenericCrudService(dao dao.GenericDAO) GenericService {
// 	return &DefaultGenericService{dao: dao}
// }

// func (service *DefaultGenericService) GetAll() ([]*config.Generic, error) {
// 	entities, err := service.dao.GetAll(common.CONSISTENCY_LOCAL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return entities, nil
// }
