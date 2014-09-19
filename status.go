package nap

import "net/http"

type statusCodeResponse struct {
	code int
}

func (s statusCodeResponse) Encode(rw http.ResponseWriter) error {
	rw.WriteHeader(s.code)
	return nil
}

// StatusCode responds with a custom status code before sending the response through.
func StatusCode(code int) Response {
	return statusCodeResponse{
		code: code,
	}
}
