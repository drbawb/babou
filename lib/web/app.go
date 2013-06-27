// The Babou web application core
package web

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

// Contexts which can be chained together
// ApplyContext will actually attach the context to a specific controller
// TestContext can be used to determine if a route supports a given context.
type ChainableContext interface {
	NewInstance() ChainableContext                                                   // returns a clean instance that is safe for a single request/response
	TestContext(Route, []ChainableContext) error                                     // Allows a context to test if a route is properly configured before any requests are serviced.
	ApplyContext(Controller, http.ResponseWriter, *http.Request, []ChainableContext) // Delegate down the chain until somebody answers the request.
}

// A controller handles a request for a given action.
// Such controller must be willing to accept GET/POST parameters from an HTTP request.
// These parameters are passed in the form of a Context object.
type Controller interface {
	HandleRequest(string) *Result
}

// A route is part of a controller that is capable
// of managing instances for a request life-cycle.
type Route interface {
	Process(string) (Controller, error)
	NewInstance() Controller
	IsSafeInstance() bool // Can this handle requests?
}

// An action takes a map of request-parameters from the middleware
// or router and turns it into a servicable HTTP result.
type Action func(map[string]string) *Result

// Represents an HTTP response from a `Controller`
// The middleware or router is responsible for using
// this result appopriately.
type Result struct {
	Body     []byte //HTTP Response Body
	Status   int    //HTTP Status Code
	Redirect *RedirectPath
}

// Requests an HTTP redirect from the middleware or
// router.
type RedirectPath struct {
	NamedRoute string //or:

	ControllerName string
	ActionName     string
}

// Returns a 404 error if a user asks `babou` for the contents of
// a directory. Useful for serving static files.
func DisableDirectoryListing(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Retrieves GET and POST vars from an http Request
func RetrieveAllParams(request *http.Request) map[string]string {
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
