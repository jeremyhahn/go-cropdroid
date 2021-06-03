// +build config

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestYamlConfigReadWrite(t *testing.T) {

	testConfig := "test-config.yaml"

	// read
	configFile := common.NewConfigFile("config.yaml")
	data, err := configFile.Read()
	assert.Nil(t, err)

	// unmarshal
	conf := &common.Config{}
	err = yaml.Unmarshal(data, &conf)
	assert.Nil(t, err)

	fmt.Printf("%#v\n", conf)

	// update
	conf.Interval = 2

	// write
	err = configFile.Write(testConfig, conf)
	assert.Nil(t, err)

	// re-read the file ane make sure interval is 2
	configFile = common.NewConfigFile(testConfig)
	data, err = configFile.Read()
	assert.Nil(t, err)

	// unmarshal
	conf = &common.Config{}
	err = yaml.Unmarshal(data, &conf)
	assert.Nil(t, err)

	assert.Equal(t, conf.Interval, 2)

	// Overwrite modified config.yaml with original
	err = os.Remove(testConfig)
	assert.Nil(t, err)
}
