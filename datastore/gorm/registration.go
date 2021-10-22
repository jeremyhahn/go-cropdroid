package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormRegistrationDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.RegistrationDAO
}

func NewRegistrationDAO(logger *logging.Logger, db *gorm.DB) dao.RegistrationDAO {
	return &GormRegistrationDAO{logger: logger, db: db}
}

// Saves a new registraion record to the database. The uint64 ID is shifted left
// because sqlite max int is a signed integer.
func (registrationDAO *GormRegistrationDAO) Save(registration config.RegistrationConfig) error {
	reg := *registration.(*config.Registration)
	reg.ID &^= (1 << 63)
	registrationDAO.logger.Debugf(fmt.Sprintf("Saving registration record: %+v", registration))
	return registrationDAO.db.Save(reg).Error
}

// Gets a new registration record from the database. The uint64 ID is shifted left
// because sqlite max int is a signed integer.
func (registrationDAO *GormRegistrationDAO) Get(id uint64) (config.RegistrationConfig, error) {
	id &^= (1 << 63)
	registrationDAO.logger.Debugf("Getting registration id: %d", id)
	var entity config.Registration
	if err := registrationDAO.db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (registrationDAO *GormRegistrationDAO) Delete(registration config.RegistrationConfig) error {
	registrationDAO.logger.Debugf(fmt.Sprintf("Deleting registration record: %+v", registration))
	return registrationDAO.db.Delete(registration).Error
}
