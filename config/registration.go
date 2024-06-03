package config

type Registration interface {
	GetEmail() string
	SetEmail(email string)
	GetPassword() string
	SetPassowrd(password string)
	RedactPassword()
	SetOrganizationID(id uint64)
	GetOrganizationID() uint64
	SetOrganizationName(name string)
	GetOrganizationName() string
	KeyValueEntity
}

type RegistrationStruct struct {
	ID           uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Email        string `yaml:"email" json:"email"`
	Password     string `yaml:"password" json:"password"`
	OrgID        uint64 `yaml:"org_id" json:"org_id"`
	OrgName      string `yaml:"org_name" json:"org_name"`
	Registration `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewRegistration() *RegistrationStruct {
	return &RegistrationStruct{}
}

func CreateRegistration(id uint64) *RegistrationStruct {
	return &RegistrationStruct{ID: id}
}

func (r *RegistrationStruct) TableName() string {
	return "registrations"
}

// Identifier gets the unique ID
func (r *RegistrationStruct) Identifier() uint64 {
	return r.ID
}

func (r *RegistrationStruct) SetID(id uint64) {
	r.ID = id
}

// Sets the registration email address
func (r *RegistrationStruct) SetEmail(email string) {
	r.Email = email
}

// Gets the registration email address
func (r *RegistrationStruct) GetEmail() string {
	return r.Email
}

// Set the registration password
func (r *RegistrationStruct) SetPassword(pw string) {
	r.Password = pw
}

// Gets the registration password
func (r *RegistrationStruct) GetPassword() string {
	return r.Password
}

// Redacts / obfuscates the password to keep it secure whie in transit or memory
func (r *RegistrationStruct) RedactPassword() {
	r.Password = "***"
}

func (r *RegistrationStruct) SetOrganizationID(id uint64) {
	r.OrgID = id
}

func (r *RegistrationStruct) GetOrganizationID() uint64 {
	return r.OrgID
}

func (r *RegistrationStruct) SetOrganizationName(name string) {
	r.OrgName = name
}

func (r *RegistrationStruct) GetOrganizationName() string {
	return r.OrgName
}
