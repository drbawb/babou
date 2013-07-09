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
	"strings"

	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
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

// Includes parameters from URLEncoded POST and GET data.
// Also includes multipart form data if its available
type Param struct {
	All map[string]string

	Files map[string][]*multipart.FileHeader
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
func RetrieveAllParams(request *http.Request) *Param {
	param := &Param{}
	param.All = mux.Vars(request)

	contentType := request.Header.Get("content-type")
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))

	switch contentType {
	case "application/x-www-form-urlencoded":
		fmt.Printf("form-url-encoded \n")
		param.Files = make(map[string][]*multipart.FileHeader)
		if err := request.ParseForm(); err != nil {
			fmt.Printf("err parsing form: %s \n", err.Error())
			return param
		}

		for k, v := range request.Form {
			param.All[k] = v[0]
		}
	case "multipart/form-data":
		fmt.Printf("form-multipart \n")
		param.Files = make(map[string][]*multipart.FileHeader)

		if err := request.ParseMultipartForm(8388608); err != nil {
			fmt.Printf("err parsing form: %s \n", err.Error())
			return param
		}

		param.Files = request.MultipartForm.File
	default:
		// Dunno; empty map and do nothing.
		fmt.Printf("da fuq? \n")
		param.Files = make(map[string][]*multipart.FileHeader)
	}

	return param
}

// ordinary post-data
