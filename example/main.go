package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/Logiraptor/nap"
)

// UserResource manages users
// +nap /user/{age}
type UserResource struct {
	nap.ResourceStub
}

type user struct {
	Name string
	Age  int
}

// Get returns some users
func (u UserResource) Get(req nap.Request) nap.Response {
	macsAge, err := req.IntValue("age")
	if err != nil {
		return nap.JSONError(400, err.Error())
	}

	return nap.Multi(
		nap.Headers(http.Header{
			"X-Custom-Header": {"Yeah"},
		}),
		nap.StatusCode(200),
		nap.JSON([]user{
			{"Patrick", 21},
			{"Yesmar", 21},
			{"Tobey", 22},
			{"Mac", macsAge},
		}),
	)
}

// Post errors out
func (u UserResource) Post(req nap.Request) nap.Response {
	return nap.JSONError(400, "Invalid user id")
}

func main() {
	r := mux.NewRouter()
	r.Handle("/user/{age:[0-9]+}", nap.Handler(UserResource{}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.ListenAndServe(":"+port, r)
}
