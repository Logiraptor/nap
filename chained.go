package nap

import "net/http"

type chainedResponse struct {
	resps []Response
}

func (c chainedResponse) Encode(rw http.ResponseWriter) error {
	for _, r := range c.resps {
		err := r.Encode(rw)
		if err != nil {
			return err
		}
	}
	return nil
}

// Multi chains a series of Responses together.
func Multi(resps ...Response) Response {
	return chainedResponse{resps}
}
