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

// Represents contexts which can be chained together.
type ChainableContext interface {
	web.Context
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

// Executes the request through the context chain
func (cc *contextChain) Execute(route web.Route, action string) http.HandlerFunc {
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
