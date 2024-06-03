package config

type Algorithm interface {
	GetName() string
	SetName(string)
	KeyValueEntity
}

type AlgorithmStruct struct {
	ID        uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Name      string `yaml:"name" json:"name"`
	Algorithm `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewAlgorithm() *AlgorithmStruct {
	return new(AlgorithmStruct)
}

func (algorithm *AlgorithmStruct) TableName() string {
	return "algorithms"
}

func (algorithm *AlgorithmStruct) SetID(id uint64) {
	algorithm.ID = id
}

func (algorithm *AlgorithmStruct) Identifier() uint64 {
	return algorithm.ID
}

func (algorithm *AlgorithmStruct) SetName(name string) {
	algorithm.Name = name
}

func (algorithm *AlgorithmStruct) GetName() string {
	return algorithm.Name
}
