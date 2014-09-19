package nap

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Request is a wrapper around http.Request for convenience
// gorilla mux vars are considered higher precedence to FormValues
type Request struct {
	*http.Request
	muxVars map[string]string
}

// Int64Value parses the named parameter as an int64 in base 10
func (r *Request) Int64Value(name string) (int64, error) {
	return strconv.ParseInt(r.getValue(name), 10, 64)
}

// IntValue parses the named parameter as an int in base 10
func (r *Request) IntValue(name string) (int, error) {
	return strconv.Atoi(r.getValue(name))
}

// StringValue exists for symmetry with Int64Value and IntValue
func (r *Request) StringValue(name string) string {
	return r.getValue(name)
}

func (r *Request) getValue(name string) string {
	if r.muxVars == nil {
		r.muxVars = mux.Vars(r.Request)
	}

	if v, ok := r.muxVars[name]; ok {
		return v
	}

	return r.FormValue(name)
}
