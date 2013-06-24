package app

import (
	controllers "babou/app/controllers"

	web "babou/lib/web"

	mux "github.com/gorilla/mux"
	http "net/http"

	dbStore "babou/lib/session"
	sessions "github.com/gorilla/sessions"

	fmt "fmt"
)

var store sessions.Store

// Babou will load these routes.
func LoadRoutes() *mux.Router {
	r := mux.NewRouter()
	web.Router = r
	store = dbStore.NewDatabaseStore([]byte("3d1fd34f389d799a2539ff554d922683"))

	// Shows public homepage, redirects to private site if valid session can be found.
	r.HandleFunc("/", wrap(controllers.NewHomeController(), "index")).Name("homeIndex")

	// Displays a login form.
	r.HandleFunc("/login", wrap(controllers.NewLoginController(),
		"index")).Name("loginIndex")
	// Displays a registration form
	r.HandleFunc("/register", wrap(controllers.NewLoginController(),
		"new")).Methods("GET").Name("loginNew")
	// Handles a new user's registration request.
	r.HandleFunc("/register", wrap(controllers.NewLoginController(),
		"create")).Methods("POST").Name("loginCreate")

	r.HandleFunc("/session/create",
		wrap(controllers.NewSessionController(), "create")).Methods("POST").Name("sessionCreate")

	// Testing method
	r.HandleFunc("/session/create",
		wrap(controllers.NewSessionController(), "create")).Methods("GET").Name("sessionCreate")

	// Displays all public assets.
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		web.DisableDirectoryListing(http.FileServer(http.Dir("assets/")))))

	return r
}

// Helper function wrap gpto a controller#action pair into a http.HandlerFunc
func wrap(controller web.Controller, action string) http.HandlerFunc {
	fn := func(response http.ResponseWriter, request *http.Request) {
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

			if true {
				//TODO: need some way to pass sessions in/out of controllers!
				session1, err := store.Get(request, "user")
				fmt.Printf("\n \nsession value was: %s \n \n", session1.Values["foo"])
				session1.Values["foo"] = "hoopaloo"

				if err != nil {
					fmt.Printf("error is: %s \n", err.Error())
				}

				sessions.Save(request, response)

			}

			response.Write(result.Body)
		}

	}

	return fn
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
