package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormCustomerDAO struct {
	logger         *logging.Logger
	db             *gorm.DB
	GenericGormDAO dao.GenericDAO[*config.Customer]
	dao.CustomerDAO
}

func NewCustomerDAO(logger *logging.Logger, db *gorm.DB) dao.CustomerDAO {
	return &GormCustomerDAO{
		logger:         logger,
		db:             db,
		GenericGormDAO: NewGenericGormDAO[*config.Customer](logger, db)}
}

func (dao *GormCustomerDAO) GetByEmail(email string, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	dao.logger.Debugf("Getting customer by email: %s", email)
	var customer *config.Customer
	if err := dao.db.
		Preload("Address").
		Preload("Shipping").
		Preload("Shipping.Address").
		Table("customers").
		First(&customer, "email = ?", email).Error; err != nil {
		return customer, err
	}
	return customer, nil
}

func (dao *GormCustomerDAO) Save(customer *config.Customer) error {
	return dao.GenericGormDAO.Save(customer)
}

func (dao *GormCustomerDAO) Get(id uint64, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	return dao.GenericGormDAO.Get(id, CONSISTENCY_LEVEL)
}

func (dao *GormCustomerDAO) GetPage(CONSISTENCY_LEVEL, page, pageSize int) ([]*config.Customer, error) {
	return dao.GenericGormDAO.GetPage(CONSISTENCY_LEVEL, page, pageSize)
}

func (dao *GormCustomerDAO) Update(customer *config.Customer) error {
	return dao.GenericGormDAO.Update(customer)
}

func (dao *GormCustomerDAO) Delete(customer *config.Customer) error {
	return dao.GenericGormDAO.Delete(customer)
}

// func (dao *GenericGormDAO[E]) Save(entity *E) error {
// 	dao.logger.Infof("Save entity: %+v", entity)
// 	return dao.db.Save(entity).Error
// }

// func (dao *GenericGormDAO[E]) Get(id uint64, CONSISTENCY_LEVEL int) (*E, error) {
// 	dao.logger.Infof("Get entity with id: %d", id)
// 	var entity E
// 	if err := dao.db.
// 		First(&entity, id).Error; err != nil {
// 		dao.logger.Warningf("GenericGormDAO.Get error: %s", err.Error())
// 		return &entity, err
// 	}
// 	return &entity, nil
// }

// // This method is only here for sake of completeness for the interface. This method
// // is only used by the Raft datastore.
// func (dao *GormCustomerDAO) Get(customerID uint64, CONSISTENCY_LEVEL int) (*config.Customer, error) {
// 	dao.logger.Debugf("Getting customer %s", customerID)
// 	var customer *config.Customer
// 	if err := dao.db.
// 		First(&customer, customerID).Error; err != nil {
// 		dao.logger.Errorf("[CustomerDAO.Get] %s", err.Error())
// 		return nil, err
// 	}
// 	return customer, nil
// }

// func (dao *GormCustomerDAO) GetByProcessorID(id string, CONSISTENCY_LEVEL int) (*config.Customer, error) {
// 	dao.logger.Debugf("Getting customer by processor_id: %s", id)
// 	var customer *config.Customer
// 	if err := dao.db.
// 		Preload("Address").
// 		Preload("Shipping").
// 		Preload("Shipping.Address").
// 		Table("customers").
// 		First(&customer, "processor_id = ?", id).Error; err != nil {
// 		if err != gorm.ErrRecordNotFound {
// 			return nil, err
// 		}
// 	}
// 	return customer, nil
// }

// func (dao *GormCustomerDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Customer, error) {
// 	dao.logger.Debug("Getting all customers")
// 	var customers []*config.Customer
// 	if err := dao.db.
// 		Preload("Address").
// 		Preload("Shipping").
// 		Preload("Shipping.Address").
// 		Order("name asc").
// 		Find(&customers).Error; err != nil {
// 		return nil, err
// 	}
// 	return customers, nil
// }
