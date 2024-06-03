package gorm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Tester interface {
	GetField1() string
	SetField1(v string)
}

type Test1 struct {
	Field1 string `json:"field1"`
}

func NewTest1() Tester {
	return &Test1{Field1: "Test 1 field 1"}
}

func (t *Test1) GetField1() string {
	return t.Field1
}

func (t *Test1) SetField1(v string) {
	t.Field1 = v
}

/////

type Tester2 interface {
	Tester
	GetTest1() Test1
}

type Test2 struct {
	Test1   Test1  `json:"Test1"`
	Field2  string `json:"field2"`
	Tester  `json:"-"`
	Tester2 `json:"-"`
}

func NewTest2() Tester2 {
	return &Test2{
		Test1: Test1{
			Field1: "Test 2 field 1",
		},
		Field2: "Test 2 field 2"}
}

func (t *Test2) GetTest1() Test1 {
	return t.Test1
}

func (t *Test2) GetField1() string {
	return t.Test1.Field1
}

func (t *Test2) SetField1(v string) {
	t.Test1.Field1 = v
}

func (t *Test2) GetField2() string {
	return t.Field2
}

func (t *Test2) SetField2(v string) {
	t.Field2 = v
}

func TestThatItWorks(t *testing.T) {

	tester := NewTest2()
	tester.SetField1("Value 2")
	field1 := tester.GetField1()
	assert.Equal(t, "Value 2", field1)

	jsonData, err := json.Marshal(tester)
	assert.Nil(t, err)

	fmt.Println(string(jsonData))

	jsonData, err = json.Marshal(tester.GetTest1())
	assert.Nil(t, err)

	fmt.Println(string(jsonData))
}
