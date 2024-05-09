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
	KeyValueEntity     `gorm:"-" yaml:"-" json:"-"`
}

func NewCustomer() *Customer {
	return new(Customer)
}

func (customer *Customer) SetID(id uint64) {
	customer.ID = id
}

func (customer *Customer) Identifier() uint64 {
	return customer.ID
}

func (customer *Customer) KVKey() []byte {
	return []byte(customer.Email)
}
