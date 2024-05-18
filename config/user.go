package config

// User represents a user account in the app
type User struct {
	ID               uint64   `gorm:"primaryKey" yaml:"id" json:"id"`
	Email            string   `gorm:"index" yaml:"email" json:"email"`
	Password         string   `yaml:"password" json:"password"`
	Roles            []*Role  `gorm:"many2many:user_role" yaml:"roles" json:"roles"`
	OrganizationRefs []uint64 `gorm:"-" yaml:"organizationRefs" json:"organizationRefs"`
	FarmRefs         []uint64 `gorm:"-" yaml:"farmRefs" json:"farmRefs"`
	KeyValueEntity   `gorm:"-" yaml:"-" json:"-"`
}

func NewUser() *User {
	user := new(User)
	user.Roles = make([]*Role, 0)
	return user
}

// Identifier gets the users ID
func (user *User) Identifier() uint64 {
	return user.ID
}

func (user *User) SetID(id uint64) {
	user.ID = id
}

func (user *User) SetEmail(email string) {
	user.Email = email
}

// GetEmail gets the users email address
func (user *User) GetEmail() string {
	return user.Email
}

func (user *User) SetPassword(pw string) {
	user.Password = pw
}

// GetPassword gets the users encrypted password
func (user *User) GetPassword() string {
	return user.Password
}

func (user *User) GetRoles() []*Role {
	return user.Roles
}

func (user *User) SetRoles(roles []*Role) {
	user.Roles = roles
}

func (user *User) AddRole(role *Role) {
	user.Roles = append(user.Roles, role)
}

func (user *User) RedactPassword() {
	user.Password = "***"
}

func (user *User) SetOrganizationRefs(refs []uint64) {
	user.OrganizationRefs = refs
}

func (user *User) GetOrganizationRefs() []uint64 {
	return user.OrganizationRefs
}

func (user *User) AddOrganizationRef(id uint64) {
	user.OrganizationRefs = append(user.OrganizationRefs, id)
}

func (user *User) SetFarmRefs(refs []uint64) {
	user.FarmRefs = refs
}

func (user *User) GetFarmRefs() []uint64 {
	return user.FarmRefs
}

func (user *User) AddFarmRef(id uint64) {
	user.FarmRefs = append(user.FarmRefs, id)
}

func (user *User) HasFarmRef(id uint64) bool {
	for _, ref := range user.FarmRefs {
		if ref == id {
			return true
		}
	}
	return false
}
