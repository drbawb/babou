package filters

import (
	"errors"
	"fmt"
	"net/http"

	web "github.com/drbawb/babou/lib/web"
)

// A chain which can store and retrieve GET/POST request variables.
type ParamterChainLink interface {
	SetParams(*web.Param)
	GetParams() *web.Param
	SetResponsePair(http.ResponseWriter, *http.Request)
}

// A controller which accepts GET/POST request variables.
type ParameterizedController interface {
	SetContext(*DevContext) error
}

// Test impl. of Context interface.
type DevContext struct {
	Params *web.Param

	Response http.ResponseWriter
	Request  *http.Request

	isInit bool
}

// Returns an instance of ParameterChain that is not initialized for request handling.
//
// This can be used for checking runtime-type dependencies as well as creating clean instances
// on a per request basis.
func ParameterChain() *DevContext {
	context := &DevContext{isInit: false}

	return context
}

func (dc *DevContext) CloseContext() {}

// The parameter context requires that a route implements ParamterizedController
func (dc *DevContext) TestContext(route web.Route, chain []web.ChainableContext) error {
	_, ok := route.(ParameterizedController)
	if !ok {
		return errors.New(fmt.Sprintf("Route :: %T :: does not support the paramter context", route))
	}

	return nil
}

// Returns an uninitialized paramter context that is suitable for creating per-request instances
// and checking runtime type dependencies.
func (dc *DevContext) NewInstance() web.ChainableContext {
	newDc := &DevContext{isInit: false}

	return newDc
}

// Applies the context to a ParamterizedController
func (dc *DevContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request, chain []web.ChainableContext) {
	dc.SetParams(web.RetrieveAllParams(request))
	dc.SetResponsePair(response, request)

	dc.isInit = true

	v, ok := controller.(ParameterizedController)
	if ok {
		if err := v.SetContext(dc); err != nil {
			fmt.Printf("Error setting paramter context: %s \n", err.Error())
		}
	} else {
		fmt.Printf("Tried to wrap a controller that is not request parameter aware \n")
	}
}

// Sets the get/post variables for this request.
func (dc *DevContext) SetParams(params *web.Param) {
	dc.Params = params
}

// Can retrieve a map of get/post vars for the current request being processed.
//
// Note that in the case of name conflicts - the POST variables take precedence and replace
// any conflict GET variables.
func (dc *DevContext) GetParams() *web.Param {
	return dc.Params
}

// Sets the current HTTP Response & Request pair
func (dc *DevContext) SetResponsePair(w http.ResponseWriter, r *http.Request) {
	dc.Request = r
	dc.Response = w
}
