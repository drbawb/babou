// The Babou web application core
package web


// data

type Result struct {
	Body []byte 	//HTTP Response Body
	Status int 	//HTTP Status Code
}

// A controller must take an action and map it to a Result
// Results can be created manually or by the app/View helpers.
type Controller interface {
	HandleRequest(string, map[string]string) *Result
}

type Action func(map[string]string) *Result
// functions

