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
	Auth    *filters.AuthContext

	Out web.Renderer
}

// Default dispatcher
// This will be called unless your controller provides its
// own impl. of a dispatcher.
//
// Other controllers have no obligation to call or delegate
// this dispatcher; it simply provides a default route.
func (ac *App) Dispatch(action string) (web.Controller, web.Action) {
	newApp := &App{}

	return newApp, func() *web.Result {
		res := &web.Result{}
		res.Body = []byte("your controller needs to provide it's own `Dispatch()` method!")
		res.Status = 200

		return res
	}
}

// Sets the ParameterContext which contains GET/POST data.
// Also sets up the template rendering for this application.
func (ac *App) SetContext(context *filters.DevContext) error {
	if context == nil {
		return errors.New("No context was supplied to this controller!")
	}

	ac.Out = web.NewMustacheRenderer("app/admin/views")
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

func (ac *App) SetAuthContext(context *filters.AuthContext) error {
	if context == nil {
		return errors.New("No AuthContext was supplied to this controller!")
	}

	ac.Auth = context

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
