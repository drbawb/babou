// The Babou web application core
package web

import (
	"net/http"
	"strings"
)

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

// A `Controller` handles a request by taking an action-name and
// a map of request parameters from the middleware or router.
// These results are usually passed to an Actio or otherwise
// turned into a servicable `Result` object.
type Controller interface {
	HandleRequest(string, map[string]string) *Result
}

// An action takes a map of request-parameters from the middleware
// or router and turns it into a servicable HTTP result.
type Action func(map[string]string) *Result

// Returns a 404 error if a user asks `babou` for the contents of
// a directory.
func DisableDirectoryListing(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}
