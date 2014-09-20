package nap

import (
	"encoding/json"
	"fmt"

	"net/http"
)

type jsonResponse struct {
	data interface{}
}

type jsonError struct {
	Code  int
	Error string
}

func (j jsonResponse) Encode(rw http.ResponseWriter) error {
	rw.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(j.data)
}

// JSON wraps data to be written out as JSON
func JSON(data interface{}) Response {
	return jsonResponse{data}
}

// JSONError returns a standard error message
func JSONError(code int, message string) Response {
	return Multi(StatusCode(code), JSON(jsonError{Code: code, Error: message}))
}

// JSONErrorf returns a standard error message
func JSONErrorf(code int, message string, args ...interface{}) Response {
	return Multi(StatusCode(code), JSON(jsonError{Code: code, Error: fmt.Sprintf(message, args...)}))
}

// JSONSuccess returns a generic json success message
func JSONSuccess() Response {
	return JSON(struct{ Status string }{"Success"})
}
