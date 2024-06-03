package config

type Smtp interface {
	SetEnable(bool)
	IstEnabled() bool
	GetHost() string
	SetHost(host string)
	GetPort() int
	SetPort(port int)
	GetUsername() string
	SetUsername(username string)
	GetPassword() string
	SetPassword(password string)
	GetRecipient() string
	SetRecipient(recipient string)
}

type SmtpStruct struct {
	Enable    bool   `yaml:"enable" json:"enable"`
	Host      string `yaml:"host" json:"host"`
	Port      int    `yaml:"port" json:"port"`
	Username  string `yaml:"username" json:"username"`
	Password  string `yaml:"password" json:"password"`
	Recipient string `yaml:"recipient" json:"recipient"`
	Smtp
}

func NewSmtp() *SmtpStruct {
	return &SmtpStruct{}
}

func (smtp *SmtpStruct) SetEnable(enabled bool) {
	smtp.Enable = enabled
}

func (smtp *SmtpStruct) IsEnabled() bool {
	return smtp.Enable
}

func (smtp *SmtpStruct) SetHost(host string) {
	smtp.Host = host
}

func (smtp *SmtpStruct) GetHost() string {
	return smtp.Host
}

func (smtp *SmtpStruct) SetPort(port int) {
	smtp.Port = port
}

func (smtp *SmtpStruct) GetPort() int {
	return smtp.Port
}

func (smtp *SmtpStruct) SetUsername(username string) {
	smtp.Username = username
}

func (smtp *SmtpStruct) GetUsername() string {
	return smtp.Username
}

func (smtp *SmtpStruct) SetPassword(password string) {
	smtp.Password = password
}

func (smtp *SmtpStruct) GetPassword() string {
	return smtp.Password
}

func (smtp *SmtpStruct) SetRecipient(recipient string) {
	smtp.Recipient = recipient
}

func (smtp *SmtpStruct) GetRecipient() string {
	return smtp.Recipient
}
