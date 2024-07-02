package webservice

import "github.com/jeremyhahn/go-trusted-platform/pki/ca"

type Config struct {
	JWTExpiration int         `yaml:"jwt-expiration" json:"jwt_expiration" mapstructure:"jwt-expiration"`
	Port          int         `yaml:"port" json:"port" mapstructure:"port"`
	TLSPort       int         `yaml:"tls-port" json:"tls_port" mapstructure:"tls-port"`
	TLSCA         string      `yaml:"tls-ca" json:"tls_ca" mapstructure:"tls-ca"`
	TLSKey        string      `yaml:"tls-key" json:"tls_key" mapstructure:"tls-key"`
	TLSCRT        string      `yaml:"tls-crt" json:"tls_crt" mapstructure:"tls-crt"`
	X509          ca.Identity `yaml:"x509" json:"x509" mapstructure:"x509"`
}

// type Identity struct {
// 	KeySize int                      `yaml:"key-size" json:"key_size" mapstructure:"key-size"`
// 	Valid   int                      `yaml:"valid" json:"valid" mapstructure:"valid"`
// 	Subject Subject                  `yaml:"subject" json:"subject" mapstructure:"subject"`
// 	SANS    *SubjectAlternativeNames `yaml:"sans" json:"sans" mapstructure:"sans"`
// }

// type Subject struct {
// 	CommonName         string `yaml:"cn" json:"cn" mapstructure:"cn"`
// 	Organization       string `yaml:"organization" json:"organization" mapstructure:"organization"`
// 	OrganizationalUnit string `yaml:"organizational-unit" json:"organizational_unit" mapstructure:"organizational-unit"`
// 	Country            string `yaml:"country" json:"country" mapstructure:"country"`
// 	Province           string `yaml:"province" json:"province" mapstructure:"province"`
// 	Locality           string `yaml:"locality" json:"locality" mapstructure:"locality"`
// 	Address            string `yaml:"address" json:"address" mapstructure:"address"`
// 	PostalCode         string `yaml:"postal-code" json:"postal_code" mapstructure:"postal-code"`
// }

// type SubjectAlternativeNames struct {
// 	DNS   []string `yaml:"dns" json:"dns" mapstructure:"dns"`
// 	IPs   []string `yaml:"ips" json:"ips" mapstructure:"ips"`
// 	Email []string `yaml:"email" json:"email" mapstructure:"email"`
// }
