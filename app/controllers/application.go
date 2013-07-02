package controllers

import (
	errors "errors"

	filters "github.com/drbawb/babou/app/filters"
	web "github.com/drbawb/babou/lib/web"
)

// A generic routine that will implement `Process` for any `Route` interface
func process(route web.Route, action string) (web.Controller, error) {
	if !route.IsSafeInstance() {
		controller := route.NewInstance() // get a controller

		return controller, nil
	}

	return nil, errors.New("This controller is not equipped to service public facing requests")
}

// A generic routine that will test for the 'default' chain which
// includes the stock implementations of:
//   ParameterChainLink, SessionChainLink, and FlashChainLink
func testContext(chain []web.ChainableContext) error {
	var paramFound, sessionFound, flashFound = false, false, false

checkDefault:
	for i := 0; i < len(chain); i++ {
		if !paramFound {
			_, ok := chain[i].(filters.ParamterChainLink)
			if ok {
				paramFound = true
				continue checkDefault
			}

		}

		if !flashFound {
			_, ok := chain[i].(filters.FlashChainLink)
			if ok {
				flashFound = true
				sessionFound = true // (implied by default flashcontext's impl. of TestContext)
				continue checkDefault
			}

		}

		if !sessionFound {
			_, ok := chain[i].(filters.SessionChainLink)
			if ok {
				sessionFound = true
				continue checkDefault
			}
		}

		if paramFound && sessionFound && flashFound {
			break checkDefault
		}
	}

	if paramFound && sessionFound && flashFound {
		return nil
	} else {
		return errors.New(`This controller requires the default chain: which includes an implementation
			of ParamterContext, SessionContext, and FlashContext`)
	}

}
