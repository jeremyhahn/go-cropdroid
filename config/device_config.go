package config

// Config represents a configuration value for a given user within a
// a specific organization for a specific device
type DeviceConfigItem struct {
	ID                 uint64 `gorm:"primary_key:auto_increment" json:"id"`
	UserID             uint64 `gorm:"index:user_id" json:"user_id"`
	DeviceID           uint64 `gorm:"index:config_device_id" json:"device_id"`
	Key                string `gorm:"size:255;index:key" json:"key"`
	Value              string `gorm:"size:255" json:"value"`
	DeviceConfigConfig `yaml:"-" json:"-"`
}

func NewDeviceConfigItem() *DeviceConfigItem {
	return &DeviceConfigItem{}
}

func CreateDeviceConfigItem(userID, deviceID uint64, key, value string) *DeviceConfigItem {
	return &DeviceConfigItem{
		UserID:   userID,
		DeviceID: deviceID,
		Key:      key,
		Value:    value}
}

// SetID sets the config unique identifier
func (c *DeviceConfigItem) SetID(id uint64) {
	c.ID = id
}

// GetID returns the config unique identifier
func (c *DeviceConfigItem) GetID() uint64 {
	return c.ID
}

// SetUserID returns the users unique identifier
func (c *DeviceConfigItem) SetUserID(id uint64) {
	c.UserID = id
}

// GetUserID returns the users unique identifier
func (c *DeviceConfigItem) GetUserID() uint64 {
	return c.UserID
}

func (c *DeviceConfigItem) SetDeviceID(id uint64) {
	c.DeviceID = id
}

// GetDeviceID returns the users unique identifier
func (c *DeviceConfigItem) GetDeviceID() uint64 {
	return c.DeviceID
}

// SetKey sets the config key (ex: mydevice.enable)
func (c *DeviceConfigItem) SetKey(key string) {
	c.Key = key
}

// GetKey returns the config key
func (c *DeviceConfigItem) GetKey() string {
	return c.Key
}

// GetValue returns the config value
func (c *DeviceConfigItem) GetValue() string {
	return c.Value
}

// SetValue sets the config value
func (c *DeviceConfigItem) SetValue(value string) {
	c.Value = value
}
