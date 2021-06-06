package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type GormChannelDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.ChannelDAO
}

func NewChannelDAO(logger *logging.Logger, db *gorm.DB) dao.ChannelDAO {
	return &GormChannelDAO{logger: logger, db: db}
}

func (channelDAO *GormChannelDAO) Save(channel config.ChannelConfig) error {
	channelDAO.logger.Debugf("Saving channel record")
	//return channelDAO.db.Save(channel.(*entity.Channel)).Error
	return channelDAO.db.Save(channel).Error
}

func (channelDAO *GormChannelDAO) Get(channelID int) (config.ChannelConfig, error) {
	channelDAO.logger.Debugf("Getting channel id %d", channelID)
	var entity config.Channel
	if err := channelDAO.db.First(&entity, channelID).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (channelDAO *GormChannelDAO) GetByControllerID(controllerID int) ([]config.Channel, error) {
	channelDAO.logger.Debugf("Getting channel record for controller %d", controllerID)
	var entities []config.Channel
	if err := channelDAO.db.Where("controller_id = ?", controllerID).Order("channel_id").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// GetByOrgUserAndControllerID gets a list of channels the user has permission to access
func (channelDAO *GormChannelDAO) GetByOrgUserAndControllerID(orgID, userID, controllerID int) ([]config.Channel, error) {
	channelDAO.logger.Debugf("Getting channel record for organization '%d'", orgID)
	var channels []config.Channel
	if err := channelDAO.db.Table("channels").
		Select("channels.*").
		Joins("JOIN controllers on channels.controller_id = controllers.id").
		Joins("JOIN farms on farms.id = controllers.farm_id AND farms.organization_id = ?", orgID).
		Joins("JOIN permissions on farms.id = permissions.farm_id").
		Where("channels.controller_id = ? and permissions.user_id = ?", controllerID, userID).
		Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

/*
func (channelDAO *GormChannelDAO) Create(channel config.ChannelConfig) error {
	channelDAO.logger.Debugf("Creating channel record")
	return channelDAO.db.Create(channel).Error
}

func (channelDAO *GormChannelDAO) Update(channel config.ChannelConfig) error {
	channelDAO.logger.Debugf("Updating channel record")
	return channelDAO.db.Update(channel).Error
}

func (channelDAO *GormChannelDAO) GetByControllerNameAndID(controllerID int, name string) (config.ChannelConfig, error) {
	channelDAO.logger.Debugf("Getting channel record '%s'", name)
	var channels []config.Channel
	if err := channelDAO.db.Where("controller_id = ? AND name = ?", controllerID, name).Find(&channels).Error; err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		return nil, errors.New(fmt.Sprintf("Channel '%s' not found in database", name))
	}
	return &channels[0], nil
}
*/
