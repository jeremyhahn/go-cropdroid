package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormChannelDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.ChannelDAO
}

func NewChannelDAO(logger *logging.Logger, db *gorm.DB) dao.ChannelDAO {
	return &GormChannelDAO{logger: logger, db: db}
}

// FarmID is accepted to support key/value database compatibility.
func (channelDAO *GormChannelDAO) Save(farmID uint64, channel *config.ChannelStruct) error {
	channelDAO.logger.Debugf("Saving channel record")
	//return channelDAO.db.Save(channel.(*entity.Channel)).Error
	return channelDAO.db.Save(channel).Error
}

// Looks up a device by its ID. Organization and Farm ID are accepted
// to support key/value database compatibility.
func (channelDAO *GormChannelDAO) GetByDevice(orgID, farmID,
	deviceID uint64, CONSISTENCY_LEVEL int) ([]*config.ChannelStruct, error) {

	var channels []*config.ChannelStruct
	if err := channelDAO.db.Table("channels").
		Select("channels.*").
		Joins("JOIN devices on channels.device_id = devices.id").
		Joins("JOIN farms on farms.id = devices.farm_id AND farms.organization_id = ?", orgID).
		Joins("JOIN permissions on farms.id = permissions.farm_id").
		Where("channels.device_id = ? and permissions.farm_id = ?", deviceID, farmID).
		Find(&channels).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			channelDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		channelDAO.logger.Error(err)
		return nil, err
	}
	return channels, nil
}

// The orgID and farmID are accepted to support
// key/value database compatibility.
func (channelDAO *GormChannelDAO) Get(orgID, farmID,
	channelID uint64, CONSISTENCY_LEVEL int) (*config.ChannelStruct, error) {

	channelDAO.logger.Debugf("Getting channel id %d", channelID)
	var entity config.ChannelStruct
	if err := channelDAO.db.First(&entity, channelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			channelDAO.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		channelDAO.logger.Error(err)
		return nil, err
	}
	return &entity, nil
}

// func (channelDAO *GormChannelDAO) GetByDeviceID(deviceID uint64) ([]config.Channel, error) {
// 	channelDAO.logger.Debugf("Getting channel record for device %d", deviceID)
// 	var entities []config.Channel
// 	if err := channelDAO.db.Where("device_id = ?", deviceID).Order("channel_id").Find(&entities).Error; err != nil {
// 		return nil, err
// 	}
// 	return entities, nil
// }

// // GetByOrgUserAndDeviceID gets a list of channels the user has permission to access
// func (channelDAO *GormChannelDAO) GetByOrgUserAndDeviceID(orgID, userID, deviceID uint64) ([]config.Channel, error) {
// 	channelDAO.logger.Debugf("Getting channel record for organization '%d'", orgID)
// 	var channels []config.Channel
// 	if err := channelDAO.db.Table("channels").
// 		Select("channels.*").
// 		Joins("JOIN devices on channels.device_id = devices.id").
// 		Joins("JOIN farms on farms.id = devices.farm_id AND farms.organization_id = ?", orgID).
// 		Joins("JOIN permissions on farms.id = permissions.farm_id").
// 		Where("channels.device_id = ? and permissions.user_id = ?", deviceID, userID).
// 		Find(&channels).Error; err != nil {
// 		return nil, err
// 	}
// 	return channels, nil
// }

/*
func (channelDAO *GormChannelDAO) Create(channel config.Channel) error {
	channelDAO.logger.Debugf("Creating channel record")
	return channelDAO.db.Create(channel).Error
}

func (channelDAO *GormChannelDAO) Update(channel config.Channel) error {
	channelDAO.logger.Debugf("Updating channel record")
	return channelDAO.db.Update(channel).Error
}

func (channelDAO *GormChannelDAO) GetByDeviceNameAndID(deviceID int, name string) (config.Channel, error) {
	channelDAO.logger.Debugf("Getting channel record '%s'", name)
	var channels []config.Channel
	if err := channelDAO.db.Where("device_id = ? AND name = ?", deviceID, name).Find(&channels).Error; err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		return nil, errors.New(fmt.Sprintf("Channel '%s' not found in database", name))
	}
	return &channels[0], nil
}
*/
