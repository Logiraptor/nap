package nap

import (
	"net/http"
)

// A Response can encode itself onto the wire
type Response interface {
	Encode(http.ResponseWriter) error
}

// A Resource defines methods on a REST endpoint
type Resource interface {
	Get(Request) Response
	Post(Request) Response
	Put(Request) Response
	Delete(Request) Response
}

// ResourceStub implements 405 methods. This is meant to be embedded.
type ResourceStub struct{}

// Get implements the Resource interface as a simple 405 Method Not Allowed Handler
func (r ResourceStub) Get(Request) Response {
	return Multi(StatusCode(http.StatusMethodNotAllowed), PlainText("METHOD NOT ALLOWED"))
}

// Post implements the Resource interface as a simple 405 Method Not Allowed Handler
func (r ResourceStub) Post(Request) Response {
	return Multi(StatusCode(http.StatusMethodNotAllowed), PlainText("METHOD NOT ALLOWED"))
}

// Put implements the Resource interface as a simple 405 Method Not Allowed Handler
func (r ResourceStub) Put(Request) Response {
	return Multi(StatusCode(http.StatusMethodNotAllowed), PlainText("METHOD NOT ALLOWED"))
}

// Delete implements the Resource interface as a simple 405 Method Not Allowed Handler
func (r ResourceStub) Delete(Request) Response {
	return Multi(StatusCode(http.StatusMethodNotAllowed), PlainText("METHOD NOT ALLOWED"))
}

// Handler turns a Resource into an http.Handler
func Handler(r Resource) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		method := req.Method
		var resp Response
		switch method {
		case "GET":
			resp = r.Get(Request{Request: req})
		case "POST":
			resp = r.Post(Request{Request: req})
		case "PUT":
			resp = r.Put(Request{Request: req})
		case "DELETE":
			resp = r.Delete(Request{Request: req})
		}
		err := resp.Encode(rw)
		if err != nil {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
		}
	})
}
