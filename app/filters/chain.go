// The filters package provides information that is contextual to a given request/response
// By chaining contexts together you can safely provide a controller access to request/response
// data with a minimal duplication of logic.
//
// This package's chain builder _may panic_ if runtime type dependencies are not satisfied.
// This is considered an unrecoverable error; and it is caught by the web-server while loading routes.
//
// Continuing to use a context-chain which has panic'd will likely result in nil-pointers
// or calls to improperly initialized contexts.
package filters

import (
	"net/http"

	lib "github.com/drbawb/babou/lib"
	web "github.com/drbawb/babou/lib/web"

	"fmt"
)

// An executable list of ChainableContext's
type contextChain struct {
	request  http.Request
	response *http.ResponseWriter

	list []web.ChainableContext
}

// Prepares a route with the default contexts:
// this includes a ParamterChainLink, FlashChainLink, and SessionChainLink.
// The latter are lazily loaded when requested by a controller.
// (This implements the chain tested by the default application controller's `testContext` method.)
func BuildDefaultChain() *contextChain {
	chain := &contextChain{list: make([]web.ChainableContext, 0, 3)}
	chain.Chain(&DevContext{}, &SessionContext{}, &FlashContext{})

	return chain
}

// Prepares a route with no contexts.
// Will simply call a bare metal controller by default.
func BuildChain() *contextChain {
	chain := &contextChain{list: make([]web.ChainableContext, 0, 1)}

	chain.Chain(&DevContext{})

	return chain
}

// Appends a context to end of the chain.
func (cc *contextChain) Chain(context ...web.ChainableContext) *contextChain {
	cc.list = append(cc.list, context...)

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

	if err := route.TestContext(cc.list); err != nil {
		panicMessages = append(panicMessages, err.Error())
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

		for i := 0; i < len(cc.list); i++ {
			currentChain[i].CloseContext()
		}

		if result.Status >= 300 && result.Status <= 399 {
			handleRedirect(result.Redirect, response, request)
		} else if result.Status == 404 {
			http.NotFound(response, request)
		} else if result.Status == 500 {
			http.Error(response, string(result.Body), 500)
		} else {
			// Assume 200
			if result.IsFile && result.Filename != "" {
				response.Header().Add("content-disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
			}
			response.Write(result.Body)
		}
	}
}

//TODO: still needs to be in a library >_>
func handleRedirect(redirect *web.RedirectPath, response http.ResponseWriter, request *http.Request) {
	if redirect.NamedRoute != "" {
		url, err := web.Router.Get(redirect.NamedRoute).URL()
		if err != nil {
			http.Error(response, string("While trying to redirect you to another page the server encountered an error. Please reload the homepage"),
				500)
		}

		http.Redirect(response, request, url.Path, 302)
	}
}
