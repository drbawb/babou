package app

import (
	controllers "babou/app/controllers"

	filters "babou/app/filters"
	web "babou/lib/web"

	mux "github.com/gorilla/mux"
	http "net/http"

	fmt "fmt"
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

	// Testing method
	r.HandleFunc("/session/create/{name}",
		filters.AuthWrap(session, "create")).Methods("GET").Name("sessionCreate")

	// Catch-All: Displays all public assets.
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		web.DisableDirectoryListing(http.FileServer(http.Dir("assets/")))))

	return r
}

func devWrap(route web.Route, action string) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		context := &web.DevContext{Params: retrieveAllParams(request)}

		controller, err := route.Process(action, context)
		if err != nil {
			fmt.Printf("error from devWrap, getting request-instance: %s \n", err.Error())
		}

		result := controller.HandleRequest(action)

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

// Helper function wrap gpto a controller#action pair into a http.HandlerFunc
func wrap(controller web.Controller, action string) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		params := retrieveAllParams(request)

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

// Retrieves GET and POST vars
func retrieveAllParams(request *http.Request) map[string]string {
	vars := mux.Vars(request)
	if err := request.ParseForm(); err != nil {
		return vars // could not parse form
	}

	var postVars map[string][]string
	postVars = map[string][]string(request.Form)
	for k, v := range postVars {
		// Ignore duplicate arguments taking the first.
		// POST will supersede any GET data in the event of collisions.
		vars[k] = v[0]
	}

	return vars
}
