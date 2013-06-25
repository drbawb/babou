package controllers

import (
	web "babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type SessionController struct {
	actionMap map[string]web.Action
}

// Registers actions for session controller and returns it.
// Returns a routable instance of the Session Controller.
func NewSessionController() *SessionController {
	sc := &SessionController{}
	sc.actionMap = make(map[string]web.Action)

	//add your actions here.
	sc.actionMap["create"] = sc.Create

	return sc
}

func (sc *SessionController) HandleRequest(action string,
	params map[string]string) *web.Result {

	if sc.actionMap[action] != nil {
		return sc.actionMap[action](params)
	} else {
		return &web.Result{Status: 404, Body: []byte("")}
	}
}

func (sc *SessionController) Create(params map[string]string) *web.Result {
	return &web.Result{Status: 200, Body: []byte("hello from session controller test")}
}
