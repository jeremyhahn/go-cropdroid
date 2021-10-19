package config

type Role struct {
	ID   uint64 `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
	//Users []User `yaml:"users" json:"users"`
	//Users      []User `gorm:"many2many:permissions"`
	RoleConfig `yaml:"-" json:"-"`
}

func NewRole() *Role {
	return &Role{}
}

func (role *Role) SetID(id uint64) {
	role.ID = id
}

func (role *Role) GetID() uint64 {
	return role.ID
}

func (role *Role) SetName(name string) {
	role.Name = name
}

func (role *Role) GetName() string {
	return role.Name
}
