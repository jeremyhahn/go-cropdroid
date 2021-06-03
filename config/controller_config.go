package config

// Config represents a configuration value for a given user within a
// a specific organization for a specific controller
type ControllerConfigItem struct {
	ID                     int    `gorm:"primary_key:auto_increment" json:"id"`
	UserID                 int    `gorm:"index:user_id" json:"user_id"`
	ControllerID           int    `gorm:"index:config_controller_id" json:"controller_id"`
	Key                    string `gorm:"size:255;index:key" json:"key"`
	Value                  string `gorm:"size:255" json:"value"`
	ControllerConfigConfig `yaml:"-" json:"-"`
}

func NewControllerConfigItem() *ControllerConfigItem {
	return &ControllerConfigItem{}
}

func CreateControllerConfigItem(userID, controllerID int, key, value string) *ControllerConfigItem {
	return &ControllerConfigItem{
		UserID:       userID,
		ControllerID: controllerID,
		Key:          key,
		Value:        value}
}

// SetID sets the config unique identifier
func (c *ControllerConfigItem) SetID(id int) {
	c.ID = id
}

// GetID returns the config unique identifier
func (c *ControllerConfigItem) GetID() int {
	return c.ID
}

// SetUserID returns the users unique identifier
func (c *ControllerConfigItem) SetUserID(id int) {
	c.UserID = id
}

// GetUserID returns the users unique identifier
func (c *ControllerConfigItem) GetUserID() int {
	return c.UserID
}

func (c *ControllerConfigItem) SetControllerID(id int) {
	c.ControllerID = id
}

// GetControllerID returns the users unique identifier
func (c *ControllerConfigItem) GetControllerID() int {
	return c.ControllerID
}

// SetKey sets the config key (ex: mycontroller.enable)
func (c *ControllerConfigItem) SetKey(key string) {
	c.Key = key
}

// GetKey returns the config key
func (c *ControllerConfigItem) GetKey() string {
	return c.Key
}

// GetValue returns the config value
func (c *ControllerConfigItem) GetValue() string {
	return c.Value
}

// SetValue sets the config value
func (c *ControllerConfigItem) SetValue(value string) {
	c.Value = value
}
