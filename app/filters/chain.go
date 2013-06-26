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

	chainHead *contextChainElem
	chainTail *contextChainElem
}

// Doubly linked list of ChainableContext's
type contextChainElem struct {
	next *contextChainElem
	me   ChainableContext
	prev *contextChainElem
}

// Contexts which can be chained together
// ApplyContext will actually attach the context to a specific controller
// TestContext can be used to determine if a route supports a given context.
type ChainableContext interface {
	web.Context
	TestContext(web.Route) error                                        // Allows a context to test if a route is properly configured before any requests are serviced.
	ApplyContext(web.DevController, http.ResponseWriter, *http.Request) // Delegate down the chain until somebody answers the request.
}

// Prepares a route with no contexts.
// Will simply call a bare metal controller by default.
func BuildChain() *contextChain {
	chain := &contextChain{}

	return chain
}

// Appends a context to end of the chain.
func (cc *contextChain) Chain(context ChainableContext) *contextChain {
	chainLink := &contextChainElem{me: context}

	if cc.chainHead == nil {
		// Insert at head of list.
		cc.chainHead = chainLink
		cc.chainTail = chainLink
	} else {
		cc.chainTail.next = chainLink
		chainLink.prev = cc.chainTail

		// Update w/ new tail pointer
		cc.chainTail = chainLink
	}

	return cc
}

// Executes the request through the context chain on a route#action pairing.
// `panics` if #TestContext() of any context in the chain fails for the given route.
func (cc *contextChain) Execute(route web.Route, action string) http.HandlerFunc {
	panicMessages := make([]string, 0)

	if route == nil {
		panicMessages = append(panicMessages, "A context chain was executed on a nullroute. Babou is not happy.\n")
	}

	for e := cc.chainHead; e != nil; e = e.next {
		if err := e.me.TestContext(route); err != nil {
			panicMessages = append(panicMessages, err.Error())
		}
	}

	if len(panicMessages) > 0 {
		panic(panicMessages)
	}

	return func(response http.ResponseWriter, request *http.Request) {
		//Generate a controller for the route
		context := &web.DevContext{Params: web.RetrieveAllParams(request)}
		controller, err := route.Process(action, context)
		if err != nil {
			lib.Println("Could not open controller with first context")
			response.Write([]byte("server error."))
		}

		for e := cc.chainHead; e != nil; e = e.next {
			e.me.ApplyContext(controller, response, request)
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
