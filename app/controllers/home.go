//A collection of controllers which can be called from the middleware.
package controllers

import (
	web "babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type HomeController struct {
	actionMap map[string]web.Action
}

// Creates a controller instance w/ an action mapping
// The instance issued is not safe for use across multiple requests.
func NewHomeController() *HomeController {
	hc := &HomeController{}
	hc.actionMap = make(map[string]web.Action)

	//add your actions here.
	hc.actionMap["index"] = hc.Index

	return hc
}

// Will create a request-specific controller instance and
// dispatch a request to the appropriate action mapping.
func (hc *HomeController) HandleRequest(action string,
	params map[string]string) *web.Result {

	if hc.actionMap[action] != nil {
		return hc.actionMap[action](params)
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
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
