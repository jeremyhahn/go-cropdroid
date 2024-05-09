package config

type Address struct {
	ID             uint64 `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Line1          string `yaml:"line1" json:"line1"`
	Line2          string `yaml:"line2" json:"line2"`
	City           string `yaml:"city" json:"city"`
	State          string `yaml:"state" json:"state"`
	PostalCode     string `yaml:"postal_code" json:"postal_code"`
	Country        string `yaml:"country" json:"country"`
	KeyValueEntity `gorm:"-" yaml:"-" json:"-"`
}

func NewAddress() *Address {
	return new(Address)
}

func (address *Address) SetID(id uint64) {
	address.ID = id
}

func (address *Address) GetID() uint64 {
	return address.ID
}

func (address *Address) KVKey() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, []byte(address.Line1)...)
	bytes = append(bytes, []byte(address.Line2)...)
	bytes = append(bytes, []byte(address.PostalCode)...)
	return []byte(bytes)
}
