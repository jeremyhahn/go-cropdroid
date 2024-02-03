package rest

import (
	"net/http"

	"github.com/jeremyhahn/go-cropdroid/common"
)

func BadRequestError(w http.ResponseWriter, r *http.Request, err error, jsonWriter common.HttpWriter) {
	jsonWriter.Write(w, http.StatusBadRequest, JsonResponse{
		Code:    400,
		Error:   err.Error(),
		Success: false,
		Payload: nil})
}
