package config

type Address interface {
	SetLine1(line1 string)
	GetLine1() string
	SetLine2(line2 string)
	GetLine2() string
	SetCity(city string)
	GetCity() string
	SetState(state string)
	GetState() Stripe
	SetPostalCode(postalCode string)
	GetPostalCode()
	SetCountry(country string)
	GetCountry() string
	KeyValueEntity
}
type AddressStruct struct {
	ID             uint64 `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Line1          string `yaml:"line1" json:"line1"`
	Line2          string `yaml:"line2" json:"line2"`
	City           string `yaml:"city" json:"city"`
	State          string `yaml:"state" json:"state"`
	PostalCode     string `yaml:"postal_code" json:"postal_code"`
	Country        string `yaml:"country" json:"country"`
	KeyValueEntity `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewAddress() *AddressStruct {
	return new(AddressStruct)
}

func (address *AddressStruct) TableName() string {
	return "addresses"
}

func (address *AddressStruct) SetID(id uint64) {
	address.ID = id
}

func (address *AddressStruct) Identifier() uint64 {
	return address.ID
}

func (address *AddressStruct) SetLine1(line1 string) {
	address.Line1 = line1
}

func (address *AddressStruct) GetLine1() string {
	return address.Line1
}

func (address *AddressStruct) SetLine2(line2 string) {
	address.Line2 = line2
}

func (address *AddressStruct) GetLine2() string {
	return address.Line2
}

func (address *AddressStruct) SetCity(city string) {
	address.City = city
}

func (address *AddressStruct) GetCity() string {
	return address.City
}

func (address *AddressStruct) SetState(state string) {
	address.State = state
}

func (address *AddressStruct) GetState() string {
	return address.State
}

func (address *AddressStruct) SetPostalCode(postalCode string) {
	address.PostalCode = postalCode
}

func (address *AddressStruct) GetPostalCode() string {
	return address.PostalCode
}

func (address *AddressStruct) SetCountry(country string) {
	address.Country = country
}

func (address *AddressStruct) GetCountry() string {
	return address.Country
}

// func (address *Address) KVKey() []byte {
// 	bytes := make([]byte, 0)
// 	bytes = append(bytes, []byte(address.Line1)...)
// 	bytes = append(bytes, []byte(address.Line2)...)
// 	bytes = append(bytes, []byte(address.PostalCode)...)
// 	return []byte(bytes)
// }
