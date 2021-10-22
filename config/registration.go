package config

import "time"

type Registration struct {
	ID        uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Email     string `yaml:"email" json:"email"`
	Password  string `yaml:"password" json:"password"`
	CreatedAt int64  `yaml:"created" json:"created"`
	OrgID     uint64 `yaml:"org_id" json:"org_id"`
	OrgName   string `yaml:"org_name" json:"org_name"`
	RegistrationConfig
}

func NewRegistration() *Registration {
	return &Registration{
		CreatedAt: time.Now().Unix()}
}

func CreateRegistration(id uint64) *Registration {
	return &Registration{
		ID:        id,
		CreatedAt: time.Now().Unix()}
}

// GetToken gets the registration token
func (r *Registration) GetID() uint64 {
	return r.ID
}

// Sets the registration email address
func (r *Registration) SetEmail(email string) {
	r.Email = email
}

// Gets the registration email address
func (r *Registration) GetEmail() string {
	return r.Email
}

// Set the registration password
func (r *Registration) SetPassword(pw string) {
	r.Password = pw
}

// Gets the registration password
func (r *Registration) GetPassword() string {
	return r.Password
}

// Redacts / obfuscates the password to keep it secure whie in transit or memory
func (r *Registration) RedactPassword() {
	r.Password = "***"
}

func (r *Registration) GetCreatedAt() int64 {
	return r.CreatedAt
}

func (r *Registration) SetOrganizationID(id uint64) {
	r.OrgID = id
}

func (r *Registration) GetOrganizationID() uint64 {
	return r.OrgID
}

func (r *Registration) SetOrganizationName(name string) {
	r.OrgName = name
}

func (r *Registration) GetOrganizationName() string {
	return r.OrgName
}
