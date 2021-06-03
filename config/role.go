package config

type Role struct {
	ID    int    `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name  string `yaml:"name" json:"name"`
	Users []User `yaml:"users" json:"users"`
	//Users      []User `gorm:"many2many:permissions"`
	RoleConfig `yaml:"-" json:"-"`
}

func NewRole() *Role {
	return &Role{}
}

func (role *Role) SetID(id int) {
	role.ID = id
}

func (role *Role) GetID() int {
	return role.ID
}

func (role *Role) SetName(name string) {
	role.Name = name
}

func (role *Role) GetName() string {
	return role.Name
}
