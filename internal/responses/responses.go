package responses

import (
	"encoding/json"
	"net/http"

	e "github.com/taylorsmcclure/kube-server/internal/errors"
)

// Easily creates a HTTP JSON response with response code and message
// TODO: figure out a better way to validate resMessage is an object that can be marshalled to JSON
func ReturnJsonResponse(w http.ResponseWriter, httpStatus int, resMessage interface{}) http.ResponseWriter {
	defer e.NonFatal()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resMessage)

	return w
}
