package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/jeremyhahn/go-cropdroid/common"
)

type JsonResponse struct {
	Code              int         `json:"code"`
	Error             string      `json:"error"`
	Success           bool        `json:"success"`
	Payload           interface{} `json:"payload"`
	common.HttpWriter `json:"-"`
}

type JsonWriter struct {
	common.HttpWriter
}

func NewJsonWriter() *JsonWriter {
	return &JsonWriter{}
}

func (writer *JsonWriter) Write(w http.ResponseWriter, status int, response interface{}) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		errResponse := JsonResponse{Error: fmt.Sprintf("JsonWriter failed to marshal response entity %s %+v", reflect.TypeOf(response), response)}
		errBytes, err := json.Marshal(errResponse)
		if err != nil {
			errResponse := JsonResponse{Error: fmt.Sprintf("JsonWriter internal server error: %s", err.Error())}
			errBytes, _ := json.Marshal(errResponse)
			http.Error(w, string(errBytes), http.StatusInternalServerError)
		}
		http.Error(w, string(errBytes), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}

func (writer *JsonWriter) Success200(w http.ResponseWriter, response interface{}) {
	writer.Write(w, http.StatusOK, JsonResponse{
		Code:    200,
		Success: true,
		Payload: response})
}

func (writer *JsonWriter) Error200(w http.ResponseWriter, err error) {
	writer.Write(w, http.StatusOK, JsonResponse{
		Code:    200,
		Success: false,
		Payload: err.Error()})
}

func (writer *JsonWriter) Error400(w http.ResponseWriter, err error) {
	writer.Write(w, http.StatusBadRequest, JsonResponse{
		Code:    400,
		Error:   err.Error(),
		Success: false,
		Payload: nil})
}

func (writer *JsonWriter) Error500(w http.ResponseWriter, err error) {
	writer.Write(w, http.StatusInternalServerError, JsonResponse{
		Code:    500,
		Error:   err.Error(),
		Success: false,
		Payload: nil})
}
