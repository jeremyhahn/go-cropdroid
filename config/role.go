package config

type Role struct {
	ID             uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Name           string `yaml:"name" json:"name"`
	Users          []User `gorm:"many2many:permissions" yaml:"-" json:"-"`
	KeyValueEntity `gorm:"-" yaml:"-" json:"-"`
}

func NewRole() *Role {
	return &Role{}
}

func (role *Role) SetID(id uint64) {
	role.ID = id
}

func (role *Role) Identifier() uint64 {
	return role.ID
}

func (role *Role) SetName(name string) {
	role.Name = name
}

func (role *Role) GetName() string {
	return role.Name
}
