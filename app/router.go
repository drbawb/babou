package app

import (
	controllers "babou/app/controllers"

	web "babou/lib/web"

	mux "github.com/gorilla/mux"	
	http "net/http"

	fmt "fmt"
)

// Babou will load these routes.
func LoadRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", wrap(controllers.NewHomeController(), "index"))
	r.HandleFunc("/{name}", testVars())

	return r
}

// Helper function wrap gpto a controller#action pair into a http.HandlerFunc
func wrap(controller web.Controller, action string) http.HandlerFunc {
	fn := func(response http.ResponseWriter, request *http.Request) {		
		params := mux.Vars(request)
		result := controller.HandleRequest(action, params)

		response.Write(result.Body)
	}

	return fn
}

func testWrap() http.HandlerFunc {
	fn := func(response http.ResponseWriter, request *http.Request) {
		response.Write([]byte("hello, world."))
	}

	return fn
}

func testVars() http.HandlerFunc {
	fn := func(response http.ResponseWriter, request *http.Request) {
		params := mux.Vars(request)
		name := params["name"]

		response.Write([]byte(fmt.Sprintf("hello, %s.", name)))
	}

	return fn
}