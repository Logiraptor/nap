package nap

import (
	"net/http"
)

type headerResponse struct {
	headers http.Header
}

func (h headerResponse) Encode(rw http.ResponseWriter) error {
	for k, a := range h.headers {
		for _, v := range a {
			rw.Header().Add(k, v)
		}
	}

	return nil
}

// Headers wraps an existing response to add headers
func Headers(headers http.Header) Response {
	return headerResponse{
		headers: headers,
	}
}
