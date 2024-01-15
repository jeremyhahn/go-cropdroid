package config

type Algorithm struct {
	ID   uint64 `gorm:"primaryKey" yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
}

func NewAlgorithm() *Algorithm {
	return &Algorithm{}
}

func (algorithm *Algorithm) GetID() uint64 {
	return algorithm.ID
}

func (algorithm *Algorithm) SetID(id uint64) {
	algorithm.ID = id
}

func (algorithm *Algorithm) GetName() string {
	return algorithm.Name
}

func (algorithm *Algorithm) SetName(name string) {
	algorithm.Name = name
}
