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

	list []web.ChainableContext
}

// Prepares a route with no contexts.
// Will simply call a bare metal controller by default.
func BuildChain() *contextChain {
	chain := &contextChain{list: make([]web.ChainableContext, 0)}

	chain.Chain(&DevContext{})

	return chain
}

// Appends a context to end of the chain.
func (cc *contextChain) Chain(context web.ChainableContext) *contextChain {
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

		currentChain := make([]web.ChainableContext, len(cc.list))

		for i := 0; i < len(cc.list); i++ {
			currentChain[i] = cc.list[i].NewInstance()

			currentChain[i].ApplyContext(controller, response, request, currentChain)
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
