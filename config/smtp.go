package config

type Smtp struct {
	Enable     bool   `yaml:"enable" json:"enable"`
	Host       string `yaml:"host" json:"host"`
	Port       int    `yaml:"port" json:"port"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	Recipient  string `yaml:"recipient" json:"recipient"`
	SmtpConfig `yaml:"-" json:"-"`
}

func NewSmtp() SmtpConfig {
	return &Smtp{}
}

func (smtp *Smtp) SetEnable(enabled bool) {
	smtp.Enable = enabled
}

func (smtp *Smtp) IsEnabled() bool {
	return smtp.Enable
}

func (smtp *Smtp) SetHost(host string) {
	smtp.Host = host
}

func (smtp *Smtp) GetHost() string {
	return smtp.Host
}

func (smtp *Smtp) SetPort(port int) {
	smtp.Port = port
}

func (smtp *Smtp) GetPort() int {
	return smtp.Port
}

func (smtp *Smtp) SetUsername(username string) {
	smtp.Username = username
}

func (smtp *Smtp) GetUsername() string {
	return smtp.Username
}

func (smtp *Smtp) SetPassword(password string) {
	smtp.Password = password
}

func (smtp *Smtp) GetPassword() string {
	return smtp.Password
}

func (smtp *Smtp) SetRecipient(recipient string) {
	smtp.Recipient = recipient
}

func (smtp *Smtp) GetRecipient() string {
	return smtp.Recipient
}
