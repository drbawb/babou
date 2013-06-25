package controllers

import (
	filters "babou/app/filters"
	lib "babou/lib"
	web "babou/lib/web"
	errors "errors"
	fmt "fmt"
)

type SessionController struct {
	actionMap    map[string]web.Action
	safeInstance bool //`true` if this instance can service HTTP requests, false otherwise.
	context      web.Context
	authContext  filters.SessionContext
}

// Registers actions for session controller and returns it.
// Returns a routable instance of the Session Controller.
func NewSessionController() *SessionController {
	sc := &SessionController{}
	sc.actionMap = make(map[string]web.Action)

	//add your actions here.
	sc.actionMap["create"] = sc.Create

	//this instance is for routing only; it will not handle requests.
	sc.safeInstance = false
	return sc
}

func (sc *SessionController) Create(params map[string]string) *web.Result {
	return &web.Result{Status: 200, Body: []byte(fmt.Sprintf("hello %s :: from session controller test", params["name"]))}
}

// Implementations for babou/lib/web.Route and babou/lib/web.DevController
// A controller must be able to accept a context and handle a request.

// Returns a controller capable of handling requests.
// Do not share the returned controller among requests.
// Returns an error if a controller suitable for dispatch is not properly initialized.
func (sc *SessionController) Process(action string, context web.Context) (web.DevController, error) {
	return process(sc, sc.actionMap, action, context)
}

// Returns true if this is a controller suitable for servicing requests.
func (sc *SessionController) IsSafeInstance() bool {
	return sc.safeInstance
}

// Returns an instance of SessionController that is equipped to deal with a single request/response
func (sc *SessionController) NewInstance() web.DevController {
	return &SessionController{safeInstance: true, actionMap: sc.actionMap}
}

// Will use a request-facing instance of session controller to handle a request.
func (sc *SessionController) HandleRequest(action string) *web.Result {
	// Bail out if this is a shared publicly routable instance.
	if !sc.safeInstance {
		return &web.Result{Status: 500, Body: []byte("Server could not route your request.")}
	}

	// Call the route function from the action map.
	if sc.actionMap[action] != nil {
		return sc.actionMap[action](sc.context.GetParams())
	} else {
		return &web.Result{Status: 404, Body: []byte("")}
	}
}

// Sets the context of a request-facing controller. This includes GET/POST vars.
// More specific contexts with addt'l helpers may be provided by filters.
// Returns an error if this is not a request-facing controller.
func (sc *SessionController) SetContext(context web.Context) error {
	if sc.safeInstance {
		sc.context = context

		return nil
	}
	return errors.New("This instance of SessionController is not equipped to handle request contexts.")
}

func (sc *SessionController) SetAuthContext(context filters.SessionContext) error {
	lib.Println("authContext called from SessionController")

	if sc.safeInstance {
		sc.authContext = context

		return nil
	}

	return errors.New("This instance of SessionController is not equipped to handle request contexts.")
}
