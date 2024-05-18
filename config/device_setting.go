package config

// Config represents a configuration value for a given user within a
// a specific organization for a specific device
type DeviceSetting struct {
	ID             uint64 `gorm:"primaryKey" json:"id"`
	UserID         uint64 `gorm:"index:user_id" json:"user_id"`
	DeviceID       uint64 `gorm:"index:config_device_id" json:"device_id"`
	Key            string `gorm:"size:255;index:key" json:"key"`
	Value          string `gorm:"size:255" json:"value"`
	KeyValueEntity `gorm:"-" yaml:"-" json:"-"`
}

func NewDeviceSetting() *DeviceSetting {
	return &DeviceSetting{}
}

func CreateDeviceSetting(userID, deviceID uint64,
	key, value string) *DeviceSetting {

	return &DeviceSetting{
		UserID:   userID,
		DeviceID: deviceID,
		Key:      key,
		Value:    value}
}

// SetID sets the unique identifier
func (ds *DeviceSetting) SetID(id uint64) {
	ds.ID = id
}

// Identifier returns the unique identifier
func (ds *DeviceSetting) Identifier() uint64 {
	return ds.ID
}

// SetUserID returns the users unique identifier
func (ds *DeviceSetting) SetUserID(id uint64) {
	ds.UserID = id
}

// GetUserID returns the users unique identifier
func (ds *DeviceSetting) GetUserID() uint64 {
	return ds.UserID
}

// GetUserID returns the device unique identifier
func (ds *DeviceSetting) SetDeviceID(id uint64) {
	ds.DeviceID = id
}

// GetDeviceID returns the device unique identifier
func (ds *DeviceSetting) GetDeviceID() uint64 {
	return ds.DeviceID
}

// SetKey sets the setting key (ex: mydevice.enable)
func (ds *DeviceSetting) SetKey(key string) {
	ds.Key = key
}

// GetKey returns the setting key
func (ds *DeviceSetting) GetKey() string {
	return ds.Key
}

// GetValue returns the setting value
func (ds *DeviceSetting) GetValue() string {
	return ds.Value
}

// SetValue sets the setting value
func (ds *DeviceSetting) SetValue(value string) {
	ds.Value = value
}
