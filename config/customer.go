package config

type Customer struct {
	ID uint64 `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	// A reference to the customer id created and stored by the credit card processor
	ProcessorID        string           `yaml:"processor_id" json:"processor_id"`
	Description        string           `yaml:"description" json:"description"`
	Name               string           `yaml:"name" json:"name"`
	Email              string           `gorm:"index:idx_email,unique" yaml:"email" json:"email"`
	Phone              string           `yaml:"phone" json:"phone"`
	Address            *Address         `gorm:"foreignKey:AddressID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"address" json:"address"`
	Shipping           *ShippingAddress `gorm:"foreignKey:ShippingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"shipping" json:"shipping"`
	PaymentMethodLast4 string           `json:"payment_method_last4"`
	PaymentMethodID    string           `json:"payment_method_id"`
	AddressID          uint64           `yaml:"address_id" json:"address_id"`
	ShippingID         uint64           `yaml:"shipping_id" json:"shipping_id"`
}

type Address struct {
	ID         uint64 `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Line1      string `yaml:"line1" json:"line1"`
	Line2      string `yaml:"line2" json:"line2"`
	City       string `yaml:"city" json:"city"`
	State      string `yaml:"state" json:"state"`
	PostalCode string `yaml:"postal_code" json:"postal_code"`
	Country    string `yaml:"country" json:"country"`
}

type ShippingAddress struct {
	ID        uint64   `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name      string   `yaml:"name" json:"name"`
	Phone     string   `yaml:"phone" json:"phone"`
	Address   *Address `gorm:"foreignKey:AddressID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"address" json:"address"`
	AddressID uint64   `yaml:"address_id" json:"address_id"`
}
