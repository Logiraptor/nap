package nap

import "net/http"

type stringResponse struct {
	text string
}

func (s stringResponse) Encode(rw http.ResponseWriter) error {
	rw.Header().Set("Content-Type", "text/plain")
	_, err := rw.Write([]byte(s.text))
	return err
}

// PlainText returns a response which will encode a plain text string
func PlainText(s string) Response {
	return stringResponse{s}
}
