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
	admin := &controllers.UsersController{}
	defaultChain := filters.BuildDefaultChain().
		Chain(filters.AuthChain(true))

	parentRouter.HandleFunc("/users",
		defaultChain.
			Resolve(admin, "index"))

	parentRouter.HandleFunc("/users/judge/{id}",
		defaultChain.
			Resolve(admin, "delete")).
		Methods("GET").
		Name("judgeUser")

	return parentRouter, nil
}
