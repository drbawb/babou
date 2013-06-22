package app

import (
	controllers "babou/app/controllers"

	web "babou/lib/web"

	mux "github.com/gorilla/mux"
	http "net/http"
)

// Babou will load these routes.
func LoadRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", wrap(controllers.NewHomeController(), "index"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		web.DisableDirectoryListing(http.FileServer(http.Dir("assets/")))))

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
