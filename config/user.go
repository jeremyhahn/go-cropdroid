package config

type CommonUser interface {
	GetEmail() string
	SetEmail(string)
	GetPassword() string
	SetPassword(string)
	HasRole(name string) bool
	SetOrganizationRefs(ids []uint64)
	GetOrganizationRefs() []uint64
	SetFarmRefs(ids []uint64)
	GetFarmRefs() []uint64
	KeyValueEntity
}

type User interface {
	GetRoles() []*RoleStruct
	SetRoles([]*RoleStruct)
	AddRole(*RoleStruct)
	CommonUser
}

// User represents a user account in the app
type UserStruct struct {
	ID               uint64        `gorm:"primaryKey" yaml:"id" json:"id"`
	Email            string        `gorm:"index" yaml:"email" json:"email"`
	Password         string        `yaml:"password" json:"password"`
	Roles            []*RoleStruct `gorm:"many2many:user_role" yaml:"roles" json:"roles"`
	OrganizationRefs []uint64      `gorm:"-" yaml:"organizationRefs" json:"organizationRefs"`
	FarmRefs         []uint64      `gorm:"-" yaml:"farmRefs" json:"farmRefs"`
	User             `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewUser() *UserStruct {
	user := new(UserStruct)
	user.Roles = make([]*RoleStruct, 0)
	return user
}

func (user *UserStruct) TableName() string {
	return "users"
}

// Identifier gets the users ID
func (user *UserStruct) Identifier() uint64 {
	return user.ID
}

func (user *UserStruct) SetID(id uint64) {
	user.ID = id
}

func (user *UserStruct) SetEmail(email string) {
	user.Email = email
}

// GetEmail gets the users email address
func (user *UserStruct) GetEmail() string {
	return user.Email
}

func (user *UserStruct) SetPassword(pw string) {
	user.Password = pw
}

// GetPassword gets the users encrypted password
func (user *UserStruct) GetPassword() string {
	return user.Password
}

func (user *UserStruct) GetRoles() []*RoleStruct {
	return user.Roles
}

func (user *UserStruct) SetRoles(roles []*RoleStruct) {
	user.Roles = roles
}

func (user *UserStruct) AddRole(role *RoleStruct) {
	user.Roles = append(user.Roles, role)
}

func (user *UserStruct) RedactPassword() {
	user.Password = "***"
}

func (user *UserStruct) SetOrganizationRefs(refs []uint64) {
	user.OrganizationRefs = refs
}

func (user *UserStruct) GetOrganizationRefs() []uint64 {
	return user.OrganizationRefs
}

func (user *UserStruct) AddOrganizationRef(id uint64) {
	user.OrganizationRefs = append(user.OrganizationRefs, id)
}

func (user *UserStruct) SetFarmRefs(refs []uint64) {
	user.FarmRefs = refs
}

func (user *UserStruct) GetFarmRefs() []uint64 {
	return user.FarmRefs
}

func (user *UserStruct) AddFarmRef(id uint64) {
	user.FarmRefs = append(user.FarmRefs, id)
}

func (user *UserStruct) HasFarmRef(id uint64) bool {
	for _, ref := range user.FarmRefs {
		if ref == id {
			return true
		}
	}
	return false
}
