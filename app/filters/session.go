package filters

import (
	"errors"
	"fmt"
	"net/http"

	web "babou/lib/web"

	dbStore "babou/lib/session"
	sessions "github.com/gorilla/sessions"
)

// Describes an element in a context chain which is capable of providing
// session-state to all elements in the chain.
//  This interfsce allows a contextChain to determine if a chain requires
//  a SessionChain to be processed. This provides additional type-safety
//  at runtime.
type SessionChainLink interface {
	SetRequestPair(http.ResponseWriter, *http.Request)
	SetStore(sessions.Store)
	GetSession() (*sessions.Session, error)
	SaveAll()
}

type SessionAware interface {
	SetSessionContext(*SessionContext)
}

// A context which provides server-backed session storage to a controller.
type SessionContext struct {
	store    sessions.Store
	request  *http.Request
	response http.ResponseWriter

	params map[string]string

	isInit bool
}

func SessionChain() *SessionContext {
	context := &SessionContext{isInit: false}

	return context
}

func (sc *SessionContext) SetParams(params map[string]string) {
	sc.params = params
}

func (sc *SessionContext) GetParams() map[string]string {
	return sc.params
}

func (sc *SessionContext) TestContext(route web.Route, chain []ChainableContext) error {
	// Controller must be session-aware.
	_, ok := route.(SessionAware)
	if !ok {
		return errors.New(fmt.Sprintf("The route :: %T :: is not SessionAware", route))
	}

	return nil
}

func (sc *SessionContext) NewInstance() ChainableContext {
	newSc := &SessionContext{isInit: false}

	return newSc
}

func (sc *SessionContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request, chain []ChainableContext) {
	sc.SetRequestPair(response, request)
	sc.SetStore(nil)
	sc.isInit = true

	v, ok := controller.(SessionAware)
	if ok {
		v.SetSessionContext(sc)
	} else {
		fmt.Printf("Tried to wrap a controller that is not AuthContext aware \n")
	}
}

// Gives the authorization a request w/ complete headers; used in retrieving/storing the user's session key.
func (sc *SessionContext) SetRequestPair(w http.ResponseWriter, r *http.Request) {
	sc.request = r
	sc.response = w
}

func (sc *SessionContext) SetStore(store sessions.Store) {
	if store == nil {
		sc.store = dbStore.NewDatabaseStore([]byte("3d1fd34f389d799a2539ff554d922683"))
	} else {
		sc.store = store
	}
}

// Retrieves or creates a session key and stores it in the bsckend.
func (sc *SessionContext) GetSession() (*sessions.Session, error) {
	if sc.request == nil || sc.response == nil {
		return nil, errors.New("Runtime tried to sccess a session before this SessionContext was fully initialized.")
	}

	session, _ := sc.store.Get(sc.request, "user")

	return session, nil
}

func (sc *SessionContext) SaveAll() {
	if sc.isInit {
		sessions.Save(sc.request, sc.response)
		return
	}
}
