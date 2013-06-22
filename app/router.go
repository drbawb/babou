package app

import (
	mux "github.com/gorilla/mux"
	http "net/http"
)

func LoadRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", testWrap())

	return r
}

func testWrap() http.HandlerFunc {
	fn := func(response http.ResponseWriter, request *http.Request) {
		response.Write([]byte("hello, world."))
	}

	return fn
}