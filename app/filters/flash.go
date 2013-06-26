package filters

import (
	"errors"
	"fmt"
	"net/http"

	web "babou/lib/web"

	dbStore "babou/lib/session"
	sessions "github.com/gorilla/sessions"
)

// Wraps the gorilla muxer's "flash" messages into a `babou/lib/web.Context`
type FlashContext struct {
	isInit bool

	params map[string]string

	request  *http.Request
	response http.ResponseWriter

	store sessions.Store
}

type FlashableController interface {
	web.Controller
	SetFlashContext(*FlashContext) error
}

func FlashChain() ChainableContext {
	return &FlashContext{isInit: false}
}

func (fc *FlashContext) GetFlashes() []interface{} {
	session, _ := fc.GetSession("user")

	allFlashes := session.Flashes()
	if allFlashes != nil && len(allFlashes) > 0 {
		sessions.Save(fc.request, fc.response) // clear flashes
		return allFlashes
	} else {
		return make([]interface{}, 0)
	}
}

func (fc *FlashContext) AddFlash(message string) {
	session, _ := fc.GetSession("user")

	if fc.isInit {
		session.AddFlash(message)
		sessions.Save(fc.request, fc.response)
	}
}

// Implement SessionContext
func (fc *FlashContext) SetRequestPair(response http.ResponseWriter, request *http.Request) {
	fc.request = request
	fc.response = response
}

func (fc *FlashContext) SetStore(store sessions.Store) {
	if fc.store != nil {
		//reuse last store who cares right now.
	} else if store == nil {
		fc.store = dbStore.NewDatabaseStore([]byte("3d1fd34f389d799a2539ff554d922683"))
	} else {
		fc.store = store
	}
}

func (fc *FlashContext) GetSession(name string) (*sessions.Session, error) {
	if fc.request == nil || fc.response == nil {
		return nil, errors.New("Runtime tried to access a session before this FlashContext was fully initialized.")
	}

	if fc.store == nil {
		return nil, errors.New("Runtime did not have a session store available")
	}

	session, err := fc.store.Get(fc.request, name)

	return session, err
}

//Implement Context & ChainableContext

// Sets the GET/POST parameters for this context.
func (fc *FlashContext) SetParams(params map[string]string) {
	fc.params = params
}

// Retrieves GET/POST parameters passed to this context.
func (fc *FlashContext) GetParams() map[string]string {
	return fc.params
}

// Tests if the route implements FlashableController interface
func (fc *FlashContext) TestContext(route web.Route) error {
	_, ok := route.(FlashableController)
	if ok {
		return nil
	} else {
		return errors.New(fmt.Sprintf("The route :: %T :: does not support the FlashContext.", route))
	}
}

// Applies this context to a controller instance.
func (fc *FlashContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request) {
	fc.SetRequestPair(response, request)
	fc.SetStore(nil)
	fc.isInit = true

	v, ok := controller.(FlashableController)
	if ok {
		if err := v.SetFlashContext(fc); err != nil {
			fmt.Printf("Error setting flash context: %s \n", err.Error())
		}
	} else {
		fmt.Printf("Tried to wrap a controller that is not FlashContext aware \n")
	}
}
