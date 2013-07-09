// The `babou` web application core
//
// Provides library methods that are useful from many different layers
// in the `app` package.
//
// If an appropriate solution exists in this package it should be preferred
// over a hand-rolled implementation. As they are designed to be fairly generic
// solutions that are fairly composable.
package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"mime/multipart"
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
	CloseContext()
}

// Contexts which have view helpers associated with them
// If they are passed to a RenderWith method their view helpers will be added.
type ViewableContext interface {
	ChainableContext
	GetViewHelpers() []interface{}
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
	Process(string) (Controller, error) // creates and returns a controller for a route.
	TestContext([]ChainableContext) error
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
	Body   []byte //HTTP Response Body
	Status int    //HTTP Status Code

	IsFile   bool //Indicates that the body should be served as a file.
	Filename string

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

// Attempts to retrieve multipart formdata; will return an empty map otherwise
func RetrieveMultipart(request *http.Request) map[string][]*multipart.FileHeader {
	vars := mux.Vars(request)
	outMap := make(map[string][]*multipart.FileHeader, 0)

	// Deal with multipart form data; again: POST data takes precedence over GET data.
	// A separate map of file-info will be returned.
	if request.MultipartForm != nil {
		postMP := request.MultipartForm
		for k, v := range postMP.Value {
			vars[k] = v[0]
		}

		outMap = postMP.File
	}

	return outMap
}

// Retrieves GET and POST vars from an http Request
func RetrieveAllParams(request *http.Request) map[string]string {
	vars := mux.Vars(request)
	if err := request.ParseMultipartForm(8388608); err != nil {
		fmt.Printf("err parsing form: %s \n", err.Error())
		return vars // could not parse form
	}

	// Deal with ordinary data.
	var postVars map[string][]string
	postVars = map[string][]string(request.Form)
	for k, v := range postVars {
		// Ignore duplicate arguments taking the first.
		// POST will supersede any GET data in the event of collisions.
		fmt.Printf("form var: %s, %s", k, v)
		vars[k] = v[0]
	}

	return vars
}
