package admin

import (
	controllers "github.com/drbawb/babou/app/admin/controllers"
	filters "github.com/drbawb/babou/app/filters"

	_ "github.com/drbawb/babou/lib/web"
	mux "github.com/gorilla/mux"
	_ "net/http"
)

// Attaches routes to the parentRouter and returns it.
func LoadRoutes(parentRouter *mux.Router) (*mux.Router, error) {
	// Shorthand for controllers
	admin := &controllers.AdminTestController{}

	// TODO: SUBROUTER
	parentRouter.HandleFunc("/",
		filters.BuildDefaultChain().
			Resolve(admin, "index")).
		Methods("GET").
		Name("admin")

	return parentRouter, nil
}
