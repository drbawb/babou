//A collection of controllers which can be called from the middleware.
package controllers

import (
	errors "errors"
	"fmt"

	filters "github.com/drbawb/babou/app/filters"
	web "github.com/drbawb/babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type HomeController struct {
	safeInstance bool

	context *filters.DevContext
	auth    *filters.AuthContext
	session *filters.SessionContext
	flash   *filters.FlashContext

	actionMap map[string]web.Action
}

// Creates an instance of this controller with action routes.
// This instance is ready to route requests.
//
// Actions for this controller must be defined here.
func (hc *HomeController) NewInstance() web.Controller {
	newHc := &HomeController{safeInstance: true, actionMap: make(map[string]web.Action)}

	//add your actions here.
	newHc.actionMap["index"] = newHc.Index

	return newHc
}

// Controller actions and logic.

// Will display a public welcome page if the user is not logged in
// Otherwise it will redirect the user to the /news page.
func (hc *HomeController) Index(params map[string]string) *web.Result {
	if hc.auth.Can("homeIndex") {
		return hc.blog(params)
	} else {
		return hc.homePage(params)
	}
}

// Public route - rendered as a public index if the user
// is not logged in or is not authenticated.
func (hc *HomeController) homePage(params map[string]string) *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &struct{}{}

	output.Body = []byte(web.RenderWith("public", "home", "index", outData, hc.flash))
	return output
}

// Private route - rendered instead of public index if the user
// is properly authenticated.
func (hc *HomeController) blog(params map[string]string) *web.Result {
	output := &web.Result{}

	output.Status = 200

	testArticles := make([]*struct{ Text string }, 0)
	testArticles = append(testArticles, &struct{ Text string }{Text: "what up bro?"})
	testArticles = append(testArticles, &struct{ Text string }{Text: "JUST WHO THE HELL DO YOU THINK I AM??"})

	user, err := hc.auth.CurrentUser()
	if err != nil {
		fmt.Printf("error printing user: %s \n", err.Error())
		output.Status = 500
		return output
	}

	outData := &struct {
		Username string
		Articles []*struct{ Text string }
	}{
		Username: user.Username,
		Articles: testArticles,
	}

	output.Body = []byte(web.RenderWith("application", "home", "news", outData))

	return output
}

// Routing methods
// These are required to satisfy the ContextChain interfaces.

// Returns a HomeController instance that is not safe across requests.
// This is useful for routing as well as context-testing.
func NewHomeController() *HomeController {
	hc := &HomeController{safeInstance: false}

	return hc
}

// Performs one of the actions mapped to this controller.
// Returns 404 if the action is not found in the map.
func (hc *HomeController) HandleRequest(action string) *web.Result {
	if !hc.safeInstance {
		return &web.Result{Status: 500, Body: []byte("The HomeController cannot service requests from users.")}
	}

	if hc.actionMap[action] != nil {
		return hc.actionMap[action](hc.context.GetParams().All)
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
}

// Process calls the application controller's default route processor.
// This will return a fresh instance of this controller that will be used
// for a single request-response lifecycle.
func (hc *HomeController) Process(action string) (web.Controller, error) {
	return process(hc, action)
}

// Tests that the current context-chain is suitable for this request.
// For the HomeController: this tests the presence of the default chain
// in addition to the presence of the Authorizaiton Chain.
func (hc *HomeController) TestContext(chain []web.ChainableContext) error {
	outFlag := false
	for i := 0; i < len(chain); i++ {
		_, ok := chain[i].(filters.AuthChainLink)
		if ok {
			outFlag = true
			break
		}
	}

	if err := testContext(chain); err != nil {
		return errors.New("Default chain missing from login route")
	}

	if !outFlag {
		return errors.New("Auth chain missing from login route.")
	}

	return nil
}

// Set contexts

// Sets the ParameterContext which contains GET/POST data.
func (hc *HomeController) SetContext(context *filters.DevContext) error {
	if context == nil {
		return errors.New("No context was supplied to this controller!")
	}

	hc.context = context
	return nil
}

// Sets the SessionContext which provides session storage for this request.
func (hc *HomeController) SetSessionContext(context *filters.SessionContext) error {
	if context == nil {
		return errors.New("No SessionContext was supplied to this controller!")
	}

	hc.session = context

	return nil
}

// Sets the FlashContext which will display a message at the earliest opportunity.
// (Usually on the next controller/action that is FlashContext aware.)
func (hc *HomeController) SetFlashContext(context *filters.FlashContext) error {
	if context == nil {
		return errors.New("No FlashContext was supplied to this controller!")
	}

	hc.flash = context

	return nil
}

// Sets the AuthContext which allows this controller to test permissions
// for the currently logged in user.
func (hc *HomeController) SetAuthContext(context *filters.AuthContext) error {
	hc.auth = context
	return nil
}

// Retruns true if this controller is ready to process a request.
func (hc *HomeController) IsSafeInstance() bool {
	return hc.safeInstance
}
