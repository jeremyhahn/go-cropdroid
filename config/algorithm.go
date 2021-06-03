package config

type AlgorithmConfig interface {
	GetID() int
	GetName() string
}

type Algorithm struct {
	ID              int    `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name            string `yaml:"name" json:"name"`
	AlgorithmConfig `yaml:"-" json:"-"`
}

func NewAlgorithm() *Algorithm {
	return &Algorithm{}
}

func (algorithm *Algorithm) GetID() int {
	return algorithm.ID
}

func (algorithm *Algorithm) SetID(id int) {
	algorithm.ID = id
}

func (algorithm *Algorithm) GetName() string {
	return algorithm.Name
}

func (algorithm *Algorithm) SetName(name string) {
	algorithm.Name = name
}
