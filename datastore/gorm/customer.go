package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormCustomerDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.CustomerDAO
}

func NewCustomerDAO(logger *logging.Logger, db *gorm.DB) dao.CustomerDAO {
	return &GormCustomerDAO{logger: logger, db: db}
}

func (dao *GormCustomerDAO) Save(customer *config.Customer) error {
	return dao.db.Save(&customer).Error
}

func (dao *GormCustomerDAO) Update(customer *config.Customer) error {
	return dao.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&customer).Error
}

func (dao *GormCustomerDAO) Delete(customer *config.Customer) error {
	return dao.db.Delete(customer).Error
}

// This method is only here for sake of completeness for the interface. This method
// is only used by the Raft datastore.
func (dao *GormCustomerDAO) Get(customerID uint64, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	dao.logger.Debugf("Getting customer %s", customerID)
	var customer *config.Customer
	if err := dao.db.
		First(&customer, customerID).Error; err != nil {
		dao.logger.Errorf("[CustomerDAO.Get] %s", err.Error())
		return nil, err
	}
	return customer, nil
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
		// if err != gorm.ErrRecordNotFound {
		// 	return nil, err
		// }
		return nil, err
	}
	return customer, nil
}

func (dao *GormCustomerDAO) GetByProcessorID(id string, CONSISTENCY_LEVEL int) (*config.Customer, error) {
	dao.logger.Debugf("Getting customer by processor_id: %s", id)
	var customer *config.Customer
	if err := dao.db.
		Preload("Address").
		Preload("Shipping").
		Preload("Shipping.Address").
		Table("customers").
		First(&customer, "processor_id = ?", id).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return customer, nil
}

func (dao *GormCustomerDAO) GetAll(CONSISTENCY_LEVEL int) ([]*config.Customer, error) {
	dao.logger.Debug("Getting all customers")
	var customers []*config.Customer
	if err := dao.db.
		Preload("Address").
		Preload("Shipping").
		Preload("Shipping.Address").
		Order("name asc").
		Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}
