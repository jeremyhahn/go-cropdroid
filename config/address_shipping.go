package config

type ShippingAddress interface {
	SetName(name string)
	GetName() string
	SetPhone(number string)
	GetPhone() string
	SetAddress(address *AddressStruct)
	GetAddress() *AddressStruct
	KeyValueEntity
}

type ShippingAddressStruct struct {
	ID              uint64         `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name            string         `yaml:"name" json:"name"`
	Phone           string         `yaml:"phone" json:"phone"`
	Address         *AddressStruct `gorm:"foreignKey:AddressID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"address" json:"address"`
	AddressID       uint64         `yaml:"address_id" json:"address_id"`
	ShippingAddress `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewShippingAddress() *ShippingAddressStruct {
	return new(ShippingAddressStruct)
}

func (shippingAddress *ShippingAddressStruct) TableName() string {
	return "shipping_addresses"
}

func (shippingAddress *ShippingAddressStruct) SetID(id uint64) {
	shippingAddress.ID = id
}

func (shippingAddress *ShippingAddressStruct) Identifier() uint64 {
	return shippingAddress.ID
}

func (shippingAddress *ShippingAddressStruct) SetName(name string) {
	shippingAddress.Name = name
}

func (shippingAddress *ShippingAddressStruct) GetName() string {
	return shippingAddress.Name
}

func (shippingAddress *ShippingAddressStruct) SetPhone(number string) {
	shippingAddress.Phone = number
}

func (shippingAddress *ShippingAddressStruct) GetPhone() string {
	return shippingAddress.Phone
}

func (shippingAddress *ShippingAddressStruct) SetAddress(address *AddressStruct) {
	shippingAddress.Address = address
}

func (shippingAddress *ShippingAddressStruct) GetAddress() *AddressStruct {
	return shippingAddress.Address
}

// func (shippingAddress *ShippingAddressStruct) KVKey() []byte {
// 	bytes := make([]byte, 0)
// 	bytes = append(bytes, []byte(shippingAddress.Name)...)
// 	bytes = append(bytes, []byte(shippingAddress.Phone)...)
// 	bytes = append(bytes, []byte(shippingAddress.Address.KVKey())...)
// 	return []byte(bytes)
// }
