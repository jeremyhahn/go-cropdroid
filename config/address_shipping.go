package config

type ShippingAddress struct {
	ID             uint64   `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name           string   `yaml:"name" json:"name"`
	Phone          string   `yaml:"phone" json:"phone"`
	Address        *Address `gorm:"foreignKey:AddressID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"address" json:"address"`
	AddressID      uint64   `yaml:"address_id" json:"address_id"`
	KeyValueEntity `gorm:"-" yaml:"-" json:"-"`
}

func NewShippingAddress() *ShippingAddress {
	return new(ShippingAddress)
}

func (shippingAddress *ShippingAddress) SetID(id uint64) {
	shippingAddress.ID = id
}

func (shippingAddress *ShippingAddress) Identifier() uint64 {
	return shippingAddress.ID
}

func (shippingAddress *ShippingAddress) KVKey() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, []byte(shippingAddress.Name)...)
	bytes = append(bytes, []byte(shippingAddress.Phone)...)
	bytes = append(bytes, []byte(shippingAddress.Address.KVKey())...)
	return []byte(bytes)
}
