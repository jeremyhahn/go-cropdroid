package config

type Algorithm struct {
	ID             uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Name           string `yaml:"name" json:"name"`
	KeyValueEntity `gorm:"-" yaml:"-" json:"-"`
}

func NewAlgorithm() *Algorithm {
	return new(Algorithm)
}

func (algorithm *Algorithm) SetID(id uint64) {
	algorithm.ID = id
}

func (algorithm *Algorithm) Identifier() uint64 {
	return algorithm.ID
}
