package app

import (
	controllers "babou/app/controllers"

	filters "babou/app/filters"
	web "babou/lib/web"

	mux "github.com/gorilla/mux"
	http "net/http"
)

// Describe your routes and apply before/after filters to a context here.
func LoadRoutes() *mux.Router {
	r := mux.NewRouter()
	web.Router = r

	// Shorthand for controllers
	home := controllers.NewHomeController()
	login := controllers.NewLoginController()
	session := controllers.NewSessionController()

	// Shows public homepage, redirects to private site if valid session can be found.
	r.HandleFunc("/", wrap(home, "index")).Name("homeIndex")

	// Displays a login form.
	r.HandleFunc("/login", wrap(login,
		"index")).Name("loginIndex")
	// Displays a registration form
	r.HandleFunc("/register", wrap(login,
		"new")).Methods("GET").Name("loginNew")
	// Handles a new user's registration request.
	r.HandleFunc("/register", wrap(login,
		"create")).Methods("POST").Name("loginCreate")

	/* Initializes a session for the user and sets aside 4KiB backend storage
	// for any stateful information.
	r.HandleFunc("/session/create",
		wrap(session, "create")).Methods("POST").Name("sessionCreate")
	*/

	// BuildChain() will auto-wrap a DevContext; further filters will apply their own context(s).
	r.HandleFunc("/session/create/{name}",
		filters.BuildChain().Chain(filters.AuthChain()).Execute(session, "create")).Methods("GET").Name("sessionCreate")

	// Catch-All: Displays all public assets.
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		web.DisableDirectoryListing(http.FileServer(http.Dir("assets/")))))

	return r
}

// Helper function wrap gpto a controller#action pair into a http.HandlerFunc
func wrap(controller web.Controller, action string) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		params := web.RetrieveAllParams(request)

		result := controller.HandleRequest(action, params)

		if result.Status >= 300 && result.Status <= 399 {
			handleRedirect(result.Redirect, response, request)
		} else if result.Status == 404 {
			http.NotFound(response, request)
		} else if result.Status == 500 {
			http.Error(response, string(result.Body), 500)
		} else {
			// Assume 200
			response.Write(result.Body)
		}

	}
}

//TODO: should move some of this to a library package.
func handleRedirect(redirect *web.RedirectPath, response http.ResponseWriter, request *http.Request) {
	if redirect.NamedRoute != "" {
		url, err := web.Router.Get(redirect.NamedRoute).URL()
		if err != nil {
			http.Error(response, string("While trying to redirect you to another page the server encountered an error. Please reload the homepage"),
				500)
		}

		http.Redirect(response, request, url.Path, 302)
	}
}
