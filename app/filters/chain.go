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

func (cc *contextChain) Resolve(route web.Controller, action string) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if v, ok := route.(web.Controller); ok {
			responder, dispatchedAction := v.Dispatch(action)

			// Create fresh chainlink.
			currentChain := make([]web.ChainableContext, len(cc.list))
			for i := 0; i < len(cc.list); i++ {
				currentChain[i] = cc.list[i].NewInstance() // clean filters.

				// applies context, with req/resp, to the fresh controller.
				currentChain[i].ApplyContext(responder, response, request, currentChain)
			}

			// Dispatch action
			result := dispatchedAction() // call when ready

			// Have chains clean up.
			for i := 0; i < len(cc.list); i++ {
				currentChain[i].CloseContext()
			}

			// Have middleware deal with current result.
			handleResult(result, response, request)
		} else {
			response.Write([]byte("ng resolver requires a dispatcher"))
		}

	}
}

func handleResult(result *web.Result, response http.ResponseWriter, request *http.Request) {
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
