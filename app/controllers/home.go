//A collection of controllers which can be called from the middleware.
package controllers

import (
	errors "errors"

	filters "github.com/drbawb/babou/app/filters"
	web "github.com/drbawb/babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type HomeController struct {
	safeInstance bool

	context *filters.DevContext
	session *filters.SessionContext
	flash   *filters.FlashContext

	actionMap map[string]web.Action
}

// Will display a public welcome page if the user is not logged in
// Otherwise it will redirect the user to the /news page.
func (hc *HomeController) Index(params map[string]string) *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &web.ViewData{Context: &struct {
		Name  string
		Yield func(string, string) string
	}{Name: "Test"}}

	output.Body = []byte(web.RenderIn("public", "home", "index", outData))

	return output
}

// Returns a HomeController instance that is not safe across requests.
func NewHomeController() *HomeController {
	hc := &HomeController{safeInstance: false}

	return hc
}

// Will create a request-specific controller instance and
// dispatch a request to the appropriate action mapping.
func (hc *HomeController) HandleRequest(action string) *web.Result {
	if !hc.safeInstance {
		return &web.Result{Status: 500, Body: []byte("The HomeController cannot service requests from users.")}
	}

	if hc.actionMap[action] != nil {
		return hc.actionMap[action](hc.context.GetParams())
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
}

func (hc *HomeController) SetContext(context *filters.DevContext) error {
	if context == nil {
		return errors.New("No context was supplied to this controller!")
	}

	hc.context = context
	return nil
}

func (hc *HomeController) SetSessionContext(context *filters.SessionContext) error {
	if context == nil {
		return errors.New("No SessionContext was supplied to this controller!")
	}

	hc.session = context

	return nil
}
func (hc *HomeController) SetFlashContext(context *filters.FlashContext) error {
	if context == nil {
		return errors.New("No FlashContext was supplied to this controller!")
	}

	hc.flash = context

	return nil
}

func (hc *HomeController) Process(action string) (web.Controller, error) {
	return process(hc, action)
}

// Tests that the current context-chain is suitable for this request.
func (hc *HomeController) TestContext(chain []web.ChainableContext) error {
	return testContext(chain)
}

func (hc *HomeController) NewInstance() web.Controller {
	newHc := &HomeController{safeInstance: true, actionMap: make(map[string]web.Action)}

	//add your actions here.
	newHc.actionMap["index"] = newHc.Index

	return newHc
}

func (hc *HomeController) IsSafeInstance() bool {
	return hc.safeInstance
}
