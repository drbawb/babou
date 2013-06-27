package filters

import (
	"errors"
	"fmt"
	"net/http"

	web "babou/lib/web"

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

type FlashableController interface {
	web.Controller
	SetFlashContext(*FlashContext) error
}

func FlashChain() web.ChainableContext {
	return &FlashContext{isInit: false}
}

func (fc *FlashContext) GetFlashes() []interface{} {
	fc.lazyLoadSession()

	allFlashes := fc.session.Flashes()
	if allFlashes != nil && len(allFlashes) > 0 {
		fc.sessionContext.SaveAll()
		return allFlashes
	} else {
		return make([]interface{}, 0)
	}

	return nil
}

func (fc *FlashContext) AddFlash(message string) {
	fc.lazyLoadSession()

	fc.session.AddFlash(message)
	fc.sessionContext.SaveAll()
}

// Returns a single helper function which a view could use to render a flash when the page is rendered.
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

// FlashContext requires a chain with a SessionContext and a FlashableController route.
func (fc *FlashContext) TestContext(route web.Route, chain []web.ChainableContext) error {
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

func (fc *FlashContext) NewInstance() web.ChainableContext {
	newFc := &FlashContext{isInit: true}

	return newFc
}

// Applies this context to a controller instance.
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

func (fc *FlashContext) lazyLoadSession() {
	// TestContext ensures that if this context is applied our controller is also setup w/ a session
	if fc.session == nil {
		fc.session, _ = fc.sessionContext.GetSession()
	}
}
