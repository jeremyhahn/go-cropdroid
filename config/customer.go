package config

type Customer interface {
	GetProcessorID() uint64
	SetProcessorID(id uint64)
	GetDescription() string
	SetDescription(description string)
	GetName() string
	SetName(name string)
	GetPhone() string
	SetPhone(number string)
	GetAddress() *AddressStruct
	SetAddress(address *AddressStruct)
	GetShippingAddress() *ShippingAddressStruct
	SetShippingAddress(address *ShippingAddressStruct)
	GetPaymentMethodID() uint64
	SetPaymentMethodID(id uint64)
	KeyValueEntity
}

type CustomerStruct struct {
	ID uint64 `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	// A reference to the customer id created and stored by the credit card processor
	ProcessorID        string                 `yaml:"processor_id" json:"processor_id"`
	Description        string                 `yaml:"description" json:"description"`
	Name               string                 `yaml:"name" json:"name"`
	Email              string                 `gorm:"index:idx_email,unique" yaml:"email" json:"email"`
	Phone              string                 `yaml:"phone" json:"phone"`
	Address            *AddressStruct         `gorm:"foreignKey:AddressID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"address" json:"address"`
	Shipping           *ShippingAddressStruct `gorm:"foreignKey:ShippingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"shipping" json:"shipping"`
	PaymentMethodLast4 string                 `json:"payment_method_last4"`
	PaymentMethodID    string                 `json:"payment_method_id"`
	AddressID          uint64                 `yaml:"address_id" json:"address_id"`
	ShippingID         uint64                 `yaml:"shipping_id" json:"shipping_id"`
	KeyValueEntity     `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewCustomer() *CustomerStruct {
	return new(CustomerStruct)
}

func (condition *CustomerStruct) TableName() string {
	return "customers"
}

func (customer *CustomerStruct) SetID(id uint64) {
	customer.ID = id
}

func (customer *CustomerStruct) Identifier() uint64 {
	return customer.ID
}

func (customer *CustomerStruct) SetProcessorID(id string) {
	customer.ProcessorID = id
}

func (customer *CustomerStruct) GetProcessorID() string {
	return customer.ProcessorID
}

func (customer *CustomerStruct) SetDesscription(text string) {
	customer.Description = text
}

func (customer *CustomerStruct) GetDescription() string {
	return customer.Description
}

func (customer *CustomerStruct) SetName(name string) {
	customer.Name = name
}

func (customer *CustomerStruct) GetName() string {
	return customer.Name
}

func (customer *CustomerStruct) SetEmail(email string) {
	customer.Email = email
}

func (customer *CustomerStruct) GetEmail() string {
	return customer.Email
}

func (customer *CustomerStruct) SetPhone(number string) {
	customer.Phone = number
}

func (customer *CustomerStruct) GetPhone() string {
	return customer.Phone
}

func (customer *CustomerStruct) SetAddress(address *AddressStruct) {
	customer.Address = address
}

func (customer *CustomerStruct) GetAddress() *AddressStruct {
	return customer.Address
}

func (customer *CustomerStruct) SetShippingAddress(address *ShippingAddressStruct) {
	customer.Shipping = address
}

func (customer *CustomerStruct) GetShippingAddress() *ShippingAddressStruct {
	return customer.Shipping
}

func (customer *CustomerStruct) SetPaymentMethodLast4(last4 string) {
	customer.PaymentMethodLast4 = last4
}

func (customer *CustomerStruct) GetPaymentMethodLast4() string {
	return customer.PaymentMethodLast4
}
