// +build broken

package test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonSerialization(t *testing.T) {

	type tdata struct {
		Field1 int    `json:"field1"`
		Field2 string `json:"field2"`
	}

	jsonData, err := json.Marshal(tdata)

	println(jsonData)

	assert.Nil(t, err)
	assert.NotNil(t, jsonData)
	//assert.Equal(t, tdata.Field1, jsonData.)
}
