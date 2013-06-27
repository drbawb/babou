package controllers

import (
	filters "babou/app/filters"
	web "babou/lib/web"
	errors "errors"
	fmt "fmt"
)

type SessionController struct {
	actionMap    map[string]web.Action
	safeInstance bool //`true` if this instance can service HTTP requests, false otherwise.
	context      *filters.DevContext

	authContext  *filters.AuthContext
	flashContext *filters.FlashContext
}

func (sc *SessionController) Create(params map[string]string) *web.Result {
	resultString := fmt.Sprintf("default response for: %s", params["name"])

	if params["name"] == "get" {
		flashes := sc.flashContext.GetFlashes()
		if len(flashes) > 0 {
			resultString = fmt.Sprintf("first flash is: %v", flashes[0])
		}
	} else if params["name"] == "save" {
		sc.flashContext.AddFlash("hello from drbawbland")
	}

	return &web.Result{Status: 200, Body: []byte(fmt.Sprintf(resultString))}
}

// Returns a routable instance of the Session Controller.
// This instance is not equipped to handle requests.
func NewSessionController() *SessionController {
	sc := &SessionController{}

	//this instance is for routing only; it will not handle requests.
	sc.safeInstance = false
	return sc
}

// Implementations for babou/lib/web.Route and babou/lib/web.DevController
// A controller must be able to accept a context and handle a request.

// Returns a controller capable of handling requests.
// Do not share the returned controller among requests.
// Returns an error if a controller suitable for dispatch is not properly initialized.
func (sc *SessionController) Process(action string) (web.Controller, error) {
	return process(sc, action)
}

// Returns true if this is a controller suitable for servicing requests.
func (sc *SessionController) IsSafeInstance() bool {
	return sc.safeInstance
}

// Returns an instance of SessionController that is equipped to deal with a single request/response
func (sc *SessionController) NewInstance() web.Controller {
	newSc := &SessionController{safeInstance: true, actionMap: make(map[string]web.Action)}

	// Add your actions here.
	newSc.actionMap["create"] = newSc.Create

	return newSc
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

// Sets the parameter-context of a request-facing controller. This includes GET/POST vars.
// More specific contexts with addt'l helpers may be provided by filters.
func (sc *SessionController) SetParams(context *filters.DevContext) error {
	if sc.safeInstance {
		sc.context = context

		return nil
	}
	return errors.New("This instance of SessionController is not equipped to handle request contexts.")
}

// Accepts an authorization context if available for the request.
func (sc *SessionController) SetAuthContext(context *filters.AuthContext) error {
	if sc.safeInstance {
		sc.authContext = context

		return nil
	}

	return errors.New("This instance of SessionController is not equipped to handle request contexts.")
}

func (sc *SessionController) SetFlashContext(context *filters.FlashContext) error {
	if sc.safeInstance {
		sc.flashContext = context
		return nil
	}

	return errors.New("This instance of SessionController is not equipped to handle request contexts.")
}
