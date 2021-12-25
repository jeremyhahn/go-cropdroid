package rest

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/jeremyhahn/go-cropdroid/common"
	"gopkg.in/yaml.v2"
)

type YamlResponse struct {
	Error             string      `yaml:"error"`
	Success           bool        `yaml:"success"`
	Payload           interface{} `yaml:"payload"`
	common.HttpWriter `yaml:"-"`
}

type YamlWriter struct {
	common.HttpWriter
}

func NewYamlWriter() *YamlWriter {
	return &YamlWriter{}
}

func (writer *YamlWriter) Write(w http.ResponseWriter, status int, response interface{}) {
	yamlResponse, err := yaml.Marshal(response)
	if err != nil {
		errResponse := YamlResponse{Error: fmt.Sprintf("YamlWriter failed to marshal response entity %s %+v", reflect.TypeOf(response), response)}
		errBytes, err := yaml.Marshal(errResponse)
		if err != nil {
			errResponse := YamlResponse{Error: fmt.Sprintf("YamlWriter internal server error: %s", err.Error())}
			errBytes, _ := yaml.Marshal(errResponse)
			http.Error(w, string(errBytes), http.StatusInternalServerError)
		}
		http.Error(w, string(errBytes), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(status)
	w.Write(yamlResponse)
}

func (writer *YamlWriter) Success200(w http.ResponseWriter, response interface{}) {
	writer.Write(w, http.StatusOK, YamlResponse{
		Success: true,
		Payload: response})
}

func (writer *YamlWriter) Error200(w http.ResponseWriter, err error) {
	writer.Write(w, http.StatusOK, YamlResponse{
		Success: false,
		Payload: err.Error()})
}

func (writer *YamlWriter) Error400(w http.ResponseWriter, err error) {
	writer.Write(w, http.StatusBadRequest, YamlResponse{
		Error:   err.Error(),
		Success: false,
		Payload: nil})
}

func (writer *YamlWriter) Error500(w http.ResponseWriter, err error) {
	writer.Write(w, http.StatusInternalServerError, YamlResponse{
		Error:   err.Error(),
		Success: false,
		Payload: nil})
}
