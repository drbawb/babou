package controllers

import (
	web "babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type LoginController struct {
	actionMap map[string]web.Action
}

// Registers actions for the HomeController and returns it.
// Note: State in the returned controller object is global to
// all requests the controller processes.
func NewLoginController() *LoginController {
	lc := &LoginController{}
	lc.actionMap = make(map[string]web.Action)

	//add your actions here.
	lc.actionMap["index"] = lc.Index

	return lc
}

func (lc *LoginController) HandleRequest(action string,
	params map[string]string) *web.Result {

	if lc.actionMap[action] != nil {
		return lc.actionMap[action](params)
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
}

func (lc *LoginController) Index(params map[string]string) *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &web.ViewData{Context: &struct{ Name string }{Name: params["name"]}}

	output.Body = []byte(web.Render("home", "index", outData))

	return output
}
