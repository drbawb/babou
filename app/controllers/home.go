//A collection of controllers which can be called from the middleware.
package controllers

import (
	"errors"
	"fmt"

	filters "github.com/drbawb/babou/app/filters"
	web "github.com/drbawb/babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type HomeController struct {
	*App
	auth *filters.AuthContext

	actionMap map[string]web.Action
}

// Creates an instance of this controller with action routes.
// This instance is ready to route requests.
//
// Actions for this controller must be defined here.
func (hc *HomeController) Dispatch(action string) (web.Controller, web.Action) {
	newHc := &HomeController{
		actionMap: make(map[string]web.Action),
		App:       &App{},
	}

	//add your actions here.
	newHc.actionMap["index"] = newHc.Index
	newHc.actionMap["faq"] = newHc.Faq

	return newHc, newHc.actionMap[action]
}

// Controller actions and logic.

// Will display a public welcome page if the user is not logged in
// Otherwise it will redirect the user to the /news page.
func (hc *HomeController) Index() *web.Result {
	if hc.auth.Can("homeIndex") {
		return hc.blog()
	} else {
		return hc.homePage()
	}
}

// Public route - rendered as a public index if the user
// is not logged in or is not authenticated.
func (hc *HomeController) homePage() *web.Result {
	output := &web.Result{Status: 200}
	outData := &struct{}{}

	output.Body = []byte(web.RenderWith(
		"bootstrap",
		"home",
		"index",
		outData, hc.Flash))
	return output
}

// Private route - rendered instead of public index if the user
// is properly authenticated.
func (hc *HomeController) blog() *web.Result {
	output := &web.Result{Status: 200}

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

	output.Body = []byte(web.RenderWith("bootstrap", "home", "news", outData))

	return output
}

// Displays the "about us" page & contact info.
func (hc *HomeController) Faq() *web.Result {
	output := &web.Result{Status: 200}

	// TODO: dump markdown formatted blurb.
	user, err := hc.auth.CurrentUser()
	if err != nil {
		fmt.Printf("error printing user: %s \n", err.Error())
		output.Status = 500
		return output
	}

	outData := &struct {
		Username string
	}{
		Username: user.Username,
	}

	output.Body = []byte(web.RenderWith("bootstrap", "home", "faq", outData))

	return output
}

// Routing methods
// These are required to satisfy the ContextChain interfaces.

// Returns a HomeController instance that is not safe across requests.
// This is useful for routing as well as context-testing.
func NewHomeController() *HomeController {
	return &HomeController{}
}

// Implement Controller interface

// Tests that the current context-chain is suitable for this request.
// For the HomeController: this tests the presence of the default chain
// in addition to the presence of the Authorizaiton Chain.

// Setup contexts

// Sets the AuthContext which allows this controller to test permissions
// for the currently logged in user.
func (hc *HomeController) SetAuthContext(context *filters.AuthContext) error {
	hc.auth = context
	return nil
}

// Calls default context test and then checks for the auth chain.
func (hc *HomeController) TestContext(chain []web.ChainableContext) error {
	if err := hc.App.TestContext(chain); err != nil {
		return err
	}

	for i := 0; i < len(chain); i++ {
		if _, ok := chain[i].(filters.AuthChainLink); ok {
			return nil
		}
	}

	return errors.New("Could not build HomeController, no auth chain for route.")
}
