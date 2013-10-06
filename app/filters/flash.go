package filters

import (
	"errors"
	"fmt"
	"net/http"

	web "github.com/drbawb/babou/lib/web"

	sessions "github.com/gorilla/sessions"
)

// Wraps the gorilla muxer's "flash" messages into a `babou/lib/web.Context`
type FlashContext struct {
	isInit bool

	params map[string]string

	controller web.Controller
	request    *http.Request
	response   http.ResponseWriter

	sessionContext SessionChainLink
	session        *sessions.Session
}

type FlashChainLink interface {
	web.ViewableContext
	AddFlash(string)
	GetFlashes() []interface{}
}

// A controller must accept a FlashContext object in order to support
// the FlashContext chain-link
type FlashableController interface {
	web.Controller
	SetFlashContext(*FlashContext) error
}

// Returns an uninitialized FlashContext which is used by the default
// context chainer to test type-requirements at runtime as well as
// route requests to a request-specific instance.
func FlashChain() web.ChainableContext {
	return &FlashContext{isInit: false}
}

func (fc *FlashContext) CloseContext() {}

func (fc *FlashContext) GetFlashes() []interface{} {
	fc.lazyLoadSession()

	allFlashes := fc.session.Flashes()
	if allFlashes != nil && len(allFlashes) > 0 {
		//fc.sessionContext.SaveAll()
		return allFlashes
	} else {
		return make([]interface{}, 0)
	}

	return nil
}

// Add a message to the session's flash context.
// The message will persist until GetFlashes() is called by
// the controller or view layer.
func (fc *FlashContext) AddFlash(message string) {
	fc.lazyLoadSession()

	fc.session.AddFlash(message)
	//fc.sessionContext.SaveAll()
}

// Implements a ViewableContext.
//
// When passed to a rendering context it adds the following helpers to the template
// layer:
//   {{#Flash}}
//		This content is shown when a flash message is present
//		{{Message}}
//		The Message variable holds the first flash message as a string.
//   {{/Flash}}
func (fc *FlashContext) GetViewHelpers() []interface{} {
	out := make([]interface{}, 1)

	flashObj := &struct {
		Flash   bool
		Message string
	}{
		Flash:   false,
		Message: "",
	}

	allFlashes := fc.GetFlashes()
	if len(allFlashes) > 0 {
		flashObj.Flash = true
		flashObj.Message = allFlashes[0].(string)
	} else {
		flashObj.Flash = false
	}

	out[0] = flashObj

	return out
}

// FlashContext requires a chain with a SessionChainLink as well as a FlashableController route.
//
// This method can be called to ensure that both those requirements are satisfied for a given route and context chain.
func (fc *FlashContext) TestContext(route web.Controller, chain []web.ChainableContext) error {
	hasSession := false

	_, ok := route.(FlashableController)
	if !ok {
		return errors.New(fmt.Sprintf("The route :: %T :: does not support the FlashContext.", route))
	}

	for i := 0; i < len(chain); i++ {
		_, ok := chain[i].(SessionChainLink) //cannot use b/c this instance is not bound to a request yet.
		if ok {
			hasSession = true
			break
		}
	}

	if ok && hasSession {
		return nil
	} else {
		return errors.New(fmt.Sprintf("The route :: %T :: does not have a SessionContext that the FlashContext can use.", route))
	}
}

// Returns a clean instance of FlashContext that can be used to service a request.
func (fc *FlashContext) NewInstance() web.ChainableContext {
	newFc := &FlashContext{isInit: true}

	return newFc
}

// Applies this context to an instance of FlashableController
func (fc *FlashContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request, chain []web.ChainableContext) {
	fc.controller = controller
	fc.isInit = true

	v, ok := fc.controller.(FlashableController)
	if ok {
		if err := v.SetFlashContext(fc); err != nil {
			fmt.Printf("Error setting flash context: %s \n", err.Error())
		}
	} else {
		fmt.Printf("Tried to wrap a controller that is not FlashContext aware \n")
	}

	for i := 0; i < len(chain); i++ {
		v, ok := chain[i].(SessionChainLink)
		if ok {
			fc.sessionContext = v
			break
		}
	}

	if fc.sessionContext == nil {
		fmt.Printf("FlashContext could not find a SessionContext in current context-chain.")
	}
}

// Loads a session from the SessionContext chain if one is not already present for this context.
func (fc *FlashContext) lazyLoadSession() {
	// TestContext ensures that if this context is applied our controller is also setup w/ a session
	if fc.session == nil {
		fc.session, _ = fc.sessionContext.GetSession()
	}
}
