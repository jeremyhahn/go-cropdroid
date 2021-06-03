package gorm

import (
	"io/ioutil"
)

type ConfigYamlDAO struct {
	Filename string
}

func NewConfigFile(filename string) *ConfigYamlDAO {
	return &ConfigYamlDAO{Filename: filename}
}

func (config *ConfigYamlDAO) Read() ([]byte, error) {
	data, err := ioutil.ReadFile(config.Filename)
	if err != nil {
		return []byte(""), err
	}
	return data, nil
}

/*
func (config *ConfigYamlDAO) Update(configuration common.Config) error {
	data, err := yaml.Marshal(configuration)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(config.Filename, []byte(data), 0644)
}

func (config *ConfigYamlDAO) Write(filename string, configuration common.Config) error {
	data, err := yaml.Marshal(configuration)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(data), 0644)
}
*/
