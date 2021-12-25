package config

// User represents a user account in the app
type User struct {
	ID            uint64         `gorm:"primaryKey" yaml:"id" json:"id"`
	Email         string         `gorm:"index" yaml:"email" json:"email"`
	Password      string         `yaml:"password" json:"password"`
	Roles         []Role         `gorm:"many2many:permissions" yaml:"roles" json:"roles"`
	Organizations []Organization `gorm:"many2many:permissions" yaml:"orgs" json:"orgs"`
	UserConfig    `yaml:"-" json:"-"`
}

func NewUser() *User {
	return &User{Roles: make([]Role, 0)}
}

// GetID gets the users ID
func (entity *User) GetID() uint64 {
	return entity.ID
}

func (entity *User) SetEmail(email string) {
	entity.Email = email
}

// GetEmail gets the users email address
func (entity *User) GetEmail() string {
	return entity.Email
}

func (entity *User) SetPassword(pw string) {
	entity.Password = pw
}

// GetPassword gets the users encrypted password
func (entity *User) GetPassword() string {
	return entity.Password
}

func (entity *User) GetRoles() []RoleConfig {
	roleConfigs := make([]RoleConfig, len(entity.Roles))
	for i, role := range entity.Roles {
		roleConfigs[i] = &role
	}
	return roleConfigs
}

func (entity *User) SetRoles(roles []RoleConfig) {
	roleStructs := make([]Role, len(roles))
	for i, role := range roles {
		roleStructs[i] = *role.(*Role)
	}
	entity.Roles = roleStructs
}

func (entity *User) AddRole(role RoleConfig) {
	entity.Roles = append(entity.Roles, *role.(*Role))
}

func (entity *User) RedactPassword() {
	entity.Password = "***"
}
