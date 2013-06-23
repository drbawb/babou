package controllers

import (
	web "babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type HomeController struct {
	actionMap map[string]web.Action
}

// Registers actions for the HomeController and returns it.
// Note: State in the returned controller object is global to
// all requests the controller processes.
func NewHomeController() *HomeController {
	hc := &HomeController{}
	hc.actionMap = make(map[string]web.Action)

	//add your actions here.
	hc.actionMap["index"] = hc.Index

	return hc
}

func (hc *HomeController) HandleRequest(action string,
	params map[string]string) *web.Result {

	if hc.actionMap[action] != nil {
		return hc.actionMap[action](params)
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
}

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
