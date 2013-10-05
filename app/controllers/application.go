package controllers

import (
	errors "errors"

	filters "github.com/drbawb/babou/app/filters"
	web "github.com/drbawb/babou/lib/web"
)

// An embeddable controller which implements the default context.
// This controller is capable of handling:
// (DevContext), (SessionContext), (FlashContext)
type App struct {
	Dev     *filters.DevContext
	Session *filters.SessionContext
	Flash   *filters.FlashContext
}

// Sets the ParameterContext which contains GET/POST data.
func (ac *App) SetContext(context *filters.DevContext) error {
	if context == nil {
		return errors.New("No context was supplied to this controller!")
	}

	ac.Dev = context
	return nil
}

// Sets the SessionContext which provides session storage for this request.
func (ac *App) SetSessionContext(context *filters.SessionContext) error {
	if context == nil {
		return errors.New("No SessionContext was supplied to this controller!")
	}

	ac.Session = context

	return nil
}

// Sets the FlashContext which will display a message at the earliest opportunity.
// (Usually on the next controller/action that is FlashContext aware.)
func (ac *App) SetFlashContext(context *filters.FlashContext) error {
	if context == nil {
		return errors.New("No FlashContext was supplied to this controller!")
	}

	ac.Flash = context

	return nil
}

func (ac *App) TestContext(chain []web.ChainableContext) error {
	return testContext(chain)
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
