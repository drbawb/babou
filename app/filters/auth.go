package filters

import (
	"errors"
	"fmt"
	"net/http"

	web "babou/lib/web"

	dbStore "babou/lib/session"
	mux "github.com/gorilla/mux"
	sessions "github.com/gorilla/sessions"
)

// A context that is aware of gorilla-sessions
type SessionContext interface {
	SetRequestPair(http.ResponseWriter, *http.Request)
	SetStore(sessions.Store)
	GetSession(string) (*sessions.Session, error)
}

// An authorizable route must be AuthContext aware.
// That way it gets access to all helpers defined by the context.
type AuthorizableRoute interface {
	Process(string, web.Context) (AuthorizableController, error)
	NewInstance() AuthorizableController
}

type AuthorizableController interface {
	web.Controller
	SetAuthContext(*AuthContext) error
}

// An impl. of SessionContext that uses it to provide helper methods for auth'ing a user.
type AuthContext struct {
	store    sessions.Store
	params   map[string]string
	request  *http.Request
	response http.ResponseWriter
	isInit   bool
}

// Stores GET/POST params for the context.
func (ac *AuthContext) SetParams(params map[string]string) {
	ac.params = params
}

// Retrieves GET/POST variables.
func (ac *AuthContext) GetParams() map[string]string {
	return ac.params
}

// Gives the authorization a request w/ complete headers; used in retrieving/storing the user's session key.
func (ac *AuthContext) SetRequestPair(w http.ResponseWriter, r *http.Request) {
	ac.request = r
	ac.response = w
}

func (ac *AuthContext) SetStore(store sessions.Store) {
	if store == nil {
		ac.store = dbStore.NewDatabaseStore([]byte("3d1fd34f389d799a2539ff554d922683"))
	} else {
		ac.store = store
	}
}

// Retrieves or creates a session key and stores it in the backend.
func (ac *AuthContext) GetSession(name string) (*sessions.Session, error) {
	if ac.request == nil || ac.response == nil {
		return nil, errors.New("Runtime tried to access a session before this AuthContext was fully initialized.")
	}

	session, _ := ac.store.Get(ac.request, name)

	return session, nil
}

// Returns an uninitialized AuthContext suitable for use in a context-chain
func AuthChain() *AuthContext {
	context := &AuthContext{isInit: false}

	return context
}

// Tests if the route implements AuthorizableController interface
func (ac *AuthContext) TestContext(route web.Route) error {
	_, ok := route.(AuthorizableController)
	if ok {
		return nil
	} else {
		return errors.New(fmt.Sprintf("The route :: %T :: does not support the AuthContext.", route))
	}
}

// Implements ChainableContext
func (ac *AuthContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request) {
	//ac.params = retrieveAllParams(request)
	ac.SetRequestPair(response, request)
	ac.SetStore(nil)
	ac.isInit = true

	v, ok := controller.(AuthorizableController)
	if ok {
		if err := v.SetAuthContext(ac); err != nil {
			fmt.Printf("Error setting authorization context: %s \n", err.Error())
		}
	} else {
		fmt.Printf("Tried to wrap a controller that is not AuthContext aware \n")
	}
}

// Wraps a route into a session-aware context [cannot be chained]
func AuthWrap(route web.Route, action string) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		context := &AuthContext{isInit: false, params: retrieveAllParams(request)}
		context.SetRequestPair(response, request)
		context.SetStore(nil)
		context.isInit = true // has valid response,request and session store.

		controller, err := route.Process(action, context)

		v, ok := controller.(AuthorizableController)
		if ok {
			if err := v.SetAuthContext(context); err != nil {
				fmt.Printf("Error setting authorization context: %s \n", err.Error())
			}
		} else {
			fmt.Printf("Tried to wrap a controller that is not AuthContext aware \n")
		}

		if err != nil {
			fmt.Printf("error from authWrap, getting request-instance: %s \n", err.Error())
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

		session, err := context.GetSession("user")
		session.Values["foo"] = "baz"
		sessions.Save(request, response)
	}
}

// Returns true if this auth-context is fully initialized _and_ trustworthy
// (TODO: could handle things like CSRF here, etc.)
func (ac *AuthContext) IsValid() bool {
	return ac.isInit
}

func retrieveAllParams(request *http.Request) map[string]string {
	vars := mux.Vars(request)
	if err := request.ParseForm(); err != nil {
		return vars // could not parse form
	}

	var postVars map[string][]string
	postVars = map[string][]string(request.Form)
	for k, v := range postVars {
		// Ignore duplicate arguments taking the first.
		// POST will supersede any GET data in the event of collisions.
		vars[k] = v[0]
	}

	return vars
}
