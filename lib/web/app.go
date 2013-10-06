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

// A dispatcher returns a request-facing version of itself and
// action which is considered safe to infoke ONLY IF the context
// satisfies the dispatcher with no errors.
type Controller interface {
	Dispatch(string) (Controller, Action)
	TestContext([]ChainableContext) error
}

// Contexts which can be chained together
// ApplyContext will actually attach the context to a specific controller
// TestContext can be used to determine if a route supports a given context.
type ChainableContext interface {
	NewInstance() ChainableContext // returns a clean instance that is safe for a single request/response

	// Resolve
	TestContext(Controller, []ChainableContext) error // Allows a context to test if a route is properly configured before any requests are serviced.

	// Attach
	ApplyContext(Controller, http.ResponseWriter, *http.Request, []ChainableContext) // Delegate down the chain until somebody answers the request.

	// After
	CloseContext()
}

// Contexts which provide a callback for the
// "AFTER-ATTACH" phase of route resolution.
type AfterPhaseContext interface {
	ChainableContext
	AfterAttach(http.ResponseWriter, *http.Request) error
}

// Contexts which have view helpers associated with them
// If they are passed to a RenderWith method their view helpers will be added.
type ViewableContext interface {
	ChainableContext
	GetViewHelpers() []interface{}
}

// An action takes a map of request-parameters from the middleware
// or router and turns it into a servicable HTTP result.
type Action func() *Result

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
	All   map[string]string
	Form  map[string]string
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
		// TODO: Empty response. For now initialize an empty response.
		_ = request.ParseForm()
		param.Files = make(map[string][]*multipart.FileHeader)
	}

	for k, v := range request.Form {
		param.All[k] = v[0]
	}

	return param
}
