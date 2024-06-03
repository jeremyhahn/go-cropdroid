package config

type DeviceSetting interface {
	SetUserID(uint64)
	GetUserID() uint64
	SetDeviceID(uint64)
	GetDeviceID() uint64
	SetKey(string)
	GetKey() string
	SetValue(string)
	GetValue() string
	KeyValueEntity
}

// DeviceSetting represents a user device setting
type DeviceSettingStruct struct {
	ID            uint64 `gorm:"primaryKey" json:"id"`
	UserID        uint64 `gorm:"index:settings_user_id" json:"user_id"`
	DeviceID      uint64 `gorm:"index:settings_device_id" json:"device_id"`
	Key           string `gorm:"size:255;index:key" json:"key"`
	Value         string `gorm:"size:255" json:"value"`
	DeviceSetting `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewDeviceSetting() *DeviceSettingStruct {
	return &DeviceSettingStruct{}
}

func CreateDeviceSetting(userID, deviceID uint64,
	key, value string) *DeviceSettingStruct {

	return &DeviceSettingStruct{
		UserID:   userID,
		DeviceID: deviceID,
		Key:      key,
		Value:    value}
}

func (ds *DeviceSettingStruct) TableName() string {
	return "device_settings"
}

// SetID sets the unique identifier
func (ds *DeviceSettingStruct) SetID(id uint64) {
	ds.ID = id
}

// Identifier returns the unique identifier
func (ds *DeviceSettingStruct) Identifier() uint64 {
	return ds.ID
}

// SetUserID returns the users unique identifier
func (ds *DeviceSettingStruct) SetUserID(id uint64) {
	ds.UserID = id
}

// GetUserID returns the users unique identifier
func (ds *DeviceSettingStruct) GetUserID() uint64 {
	return ds.UserID
}

// GetUserID returns the device unique identifier
func (ds *DeviceSettingStruct) SetDeviceID(id uint64) {
	ds.DeviceID = id
}

// GetDeviceID returns the device unique identifier
func (ds *DeviceSettingStruct) GetDeviceID() uint64 {
	return ds.DeviceID
}

// SetKey sets the setting key (ex: mydevice.enable)
func (ds *DeviceSettingStruct) SetKey(key string) {
	ds.Key = key
}

// GetKey returns the setting key
func (ds *DeviceSettingStruct) GetKey() string {
	return ds.Key
}

// GetValue returns the setting value
func (ds *DeviceSettingStruct) GetValue() string {
	return ds.Value
}

// SetValue sets the setting value
func (ds *DeviceSettingStruct) SetValue(value string) {
	ds.Value = value
}
