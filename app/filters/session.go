package filters

import (
	"errors"
	"fmt"
	"net/http"

	dbStore "github.com/drbawb/babou/lib/session"
	web "github.com/drbawb/babou/lib/web"

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

// A SessionAware controller must be willing to accept a SessionContext.
// Use of the session context
type SessionAware interface {
	SetSessionContext(*SessionContext) error
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

func (sc *SessionContext) CloseContext() {
	sc.SaveAll()
}

// Tests that the context is being applied to a route which is SessionAware.
//
// This method can be used to ensure that the type-dependencies are satisfied at runtime.
func (sc *SessionContext) TestContext(route web.Controller, chain []web.ChainableContext) error {
	// Controller must be session-aware.
	_, ok := route.(SessionAware)
	if !ok {
		return errors.New(fmt.Sprintf("The route :: %T :: is not SessionAware", route))
	}

	return nil
}

// Returns an instance of this context that can safely be used to process
// a single response-request pairing.
func (sc *SessionContext) NewInstance() web.ChainableContext {
	newSc := &SessionContext{isInit: false}

	return newSc
}

// Applies this context to a SessionAware controller. Exposing all session variables to the underlying route.
func (sc *SessionContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request, chain []web.ChainableContext) {
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

// Sets the session store to the specified store; or fetches one from the
// default database-backed session store.
//
// TODO: Should be configurable; must be shared among cooperative frontends.
func (sc *SessionContext) SetStore(store sessions.Store) {
	if store == nil {
		sc.store = dbStore.NewDatabaseStore([]byte("3d1fd34f389d799a2539ff554d922683"))
	} else {
		sc.store = store
	}
}

// Retrieves or creates a session key and stores it in the backend.
func (sc *SessionContext) GetSession() (*sessions.Session, error) {
	if sc.request == nil || sc.response == nil {
		return nil, errors.New("Runtime tried to sccess a session before this SessionContext was fully initialized.")
	}

	//TODO: Lazy load session here.
	session, _ := sc.store.Get(sc.request, "user")

	return session, nil
}

// Saves all currently open sessions for this request to disk.
func (sc *SessionContext) SaveAll() {
	//TODO: Can I somehow defer this save until the request is finished processing?
	// Maybe a callback from the context-chain.

	// TODO TODO: Is that even a good idea though? Is there an instance where you expect the session
	// perhaps I'd need some kind of "ForceSave()" method in case a controller actually wnats to persist
	// state _now._

	//tl;dr: premature optimization.
	if sc.isInit {
		sessions.Save(sc.request, sc.response)
		return
	}
}
