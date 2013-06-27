package filters

import (
	lib "babou/lib"
	web "babou/lib/web"
	"net/http"
)

// An executable list of ChainableContext's
type contextChain struct {
	request  http.Request
	response *http.ResponseWriter

	list []ChainableContext
}

// Contexts which can be chained together
// ApplyContext will actually attach the context to a specific controller
// TestContext can be used to determine if a route supports a given context.
type ChainableContext interface {
	TestContext(web.Route, []ChainableContext) error                 // Allows a context to test if a route is properly configured before any requests are serviced.
	ApplyContext(web.Controller, http.ResponseWriter, *http.Request) // Delegate down the chain until somebody answers the request.
}

// Prepares a route with no contexts.
// Will simply call a bare metal controller by default.
func BuildChain() *contextChain {
	chain := &contextChain{list: make([]ChainableContext, 0)}

	chain.Chain(&DevContext{})

	return chain
}

// Appends a context to end of the chain.
func (cc *contextChain) Chain(context ChainableContext) *contextChain {
	cc.list = append(cc.list, context)

	return cc
}

// Executes the request through the context chain on a route#action pairing.
// `panics` if #TestContext() of any context in the chain fails for the given route.
//	  Note that this method wraps the chain with a DevContext; so calling this method
//    will ensure that the `ParamterizedChain` interface is fulfilled for all other chainlinks.
func (cc *contextChain) Execute(route web.Route, action string) http.HandlerFunc {
	panicMessages := make([]string, 0)

	if route == nil {
		panicMessages = append(panicMessages, "A context chain was executed on a nullroute. Babou is not happy.\n")
	}

	for i := 0; i < len(cc.list); i++ {
		if err := cc.list[i].TestContext(route, cc.list); err != nil {
			panicMessages = append(panicMessages, err.Error())
		}
	}

	if len(panicMessages) > 0 {
		// Panic if this chain does not pass runtime-type checks.
		panic(panicMessages)
	}

	// Will create an HttpHandler that encapsulates the chain's current state.
	return func(response http.ResponseWriter, request *http.Request) {
		//Generate a controller for the route
		controller, err := route.Process(action)
		if err != nil {
			lib.Println("Could not open controller with first context")
			response.Write([]byte("server error."))
		}

		for i := 0; i < len(cc.list); i++ {
			cc.list[i].ApplyContext(controller, response, request)
		}

		result := controller.HandleRequest(action)

		if result.Status >= 300 && result.Status <= 399 {
			//handleRedirect(result.Redirect, response, request)
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
