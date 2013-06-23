// The Babou web application core
package web

import (
	"net/http"
	"strings"
)

// data

type Result struct {
	Body     []byte //HTTP Response Body
	Status   int    //HTTP Status Code
	Redirect *RedirectPath
}

type RedirectPath struct {
	NamedRoute string //or:

	ControllerName string
	ActionName     string
}

// A controller must take an action and map it to a Result
// Results can be created manually or by the app/View helpers.
type Controller interface {
	HandleRequest(string, map[string]string) *Result
}

func DisableDirectoryListing(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

type Action func(map[string]string) *Result

// functions
