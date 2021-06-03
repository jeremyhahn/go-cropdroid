// +build future

package gorm

// NOT IN USE YET! - intended for key/value compatibility if i end up going the route
// of storing the entire config in etcd store
import (
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type DynamicConfigDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	//ConfigDAO
}

func NewDynamicConfigDAO(logger *logging.Logger, db *gorm.DB) *DynamicConfigDAO { //ConfigDAO {
	return &DynamicConfigDAO{logger: logger, db: db}
}

func (dao *DynamicConfigDAO) Save(config config.ControllerConfigConfig) error {
	return nil
}

func (dao *DynamicConfigDAO) Get(controllerID int, name string) (config.ControllerConfigConfig, error) {
	return &config.ControllerConfigItem{}, nil
}

func (dao *DynamicConfigDAO) GetAll(controllerID int) ([]config.ControllerConfigConfig, error) {
	dao.logger.Debugf("%+v", dao.app.Config)
	entities := make([]config.ControllerConfigConfig, 0)
	return entities, nil
}
