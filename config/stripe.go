package config

type StripeKey struct {
	Secret      string `yaml:"secret" json:"secret" mapstructure:"secret"`
	Publishable string `yaml:"publishable" json:"publishable" mapstructure:"publishable"`
	Webook      string `yaml:"webhook" json:"webhook" mapstructure:"webhook"`
}

type Stripe struct {
	Key *StripeKey `yaml:"key" json:"key" mapstructure:"key"`
	Tax *StripeTax `yaml:"tax" json:"tax" mapstructure:"tax"`
}

type StripeTax struct {
	Fixed   []*string `yaml:"fixed" json:"fixed" mapstructure:"fixed"`
	Dynamic []*string `yaml:"dynamic" json:"dynamic" mapstructure:"dynamic"`
}

// type StripeTaxList struct {
// 	Rates []*StripeTaxRate `yaml:"rates" json:"rates" mapstructure:"rates"`
// }

// type StripeTaxRate struct {
// 	// Required
// 	DisplayName string `yaml:"display_name" json:"display_name" mapstructure:"display_name"`
// 	Inclusive   string `yaml:"inclusive" json:"inclusive" mapstructure:"inclusive"`
// 	Percentage  string `yaml:"percentage" json:"percentage" mapstructure:"percentage"`
// 	// Optional
// 	Active       string `yaml:"active" json:"active" mapstructure:"active"`
// 	Country      string `yaml:"country" json:"country" mapstructure:"country"`
// 	Description  string `yaml:"description" json:"description" mapstructure:"description"`
// 	Jurisdiction string `yaml:"jurisdiction" json:"jurisdiction" mapstructure:"jurisdiction"`
// 	State        string `yaml:"state" json:"state" mapstructure:"state"`
// 	TaxType      string `yaml:"tax_type" json:"tax_type" mapstructure:"tax_type"`
// }
