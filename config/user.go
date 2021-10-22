package config

// User represents a user account in the app
type User struct {
	ID         uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Email      string `gorm:"index" yaml:"email" json:"email"`
	Password   string `yaml:"password" json:"password"`
	Roles      []Role `gorm:"many2many:permissions" yaml:"roles" json:"roles"`
	UserConfig `yaml:"-" json:"-"`
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

func (entity *User) GetRoles() []Role {
	return entity.Roles
}

func (entity *User) SetRoles(roles []Role) {
	entity.Roles = roles
}

func (entity *User) AddRole(role Role) {
	entity.Roles = append(entity.Roles, role)
}

func (entity *User) RedactPassword() {
	entity.Password = "***"
}
