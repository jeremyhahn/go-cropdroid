package config

type Role interface {
	GetName() string
	SetName(name string)
	KeyValueEntity
}

type RoleStruct struct {
	ID    uint64       `gorm:"primaryKey" yaml:"id" json:"id"`
	Name  string       `yaml:"name" json:"name"`
	Users []UserStruct `gorm:"many2many:user_role" yaml:"-" json:"-"`
	Role  `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

// gorm:"many2many:permissions"

func NewRole() *RoleStruct {
	return &RoleStruct{}
}

func (role *RoleStruct) TableName() string {
	return "roles"
}

func (role *RoleStruct) SetID(id uint64) {
	role.ID = id
}

func (role *RoleStruct) Identifier() uint64 {
	return role.ID
}

func (role *RoleStruct) SetName(name string) {
	role.Name = name
}

func (role *RoleStruct) GetName() string {
	return role.Name
}
