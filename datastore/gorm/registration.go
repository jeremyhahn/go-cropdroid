package gorm

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
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
func (registrationDAO *GormRegistrationDAO) Save(registration *config.RegistrationStruct) error {
	registrationDAO.logger.Debugf(fmt.Sprintf("Saving registration record: %+v", registration))
	return registrationDAO.db.Save(registration).Error
}

// Gets a new registration record from the database. The uint64 ID is shifted left
// because sqlite max int is a signed integer.
func (registrationDAO *GormRegistrationDAO) Get(registrationID uint64, CONSISTENCY_LEVEL int) (*config.RegistrationStruct, error) {
	registrationDAO.logger.Debugf("Getting registration id: %d", registrationID)
	var entity config.RegistrationStruct
	if err := registrationDAO.db.First(&entity, registrationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			registrationDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		registrationDAO.logger.Error(err)
		return nil, err
	}
	return &entity, nil
}

func (registrationDAO *GormRegistrationDAO) Delete(registration *config.RegistrationStruct) error {
	registrationDAO.logger.Debugf(fmt.Sprintf("Deleting registration record: %+v", registration))
	return registrationDAO.db.Delete(registration).Error
}
