package controllers

import (
	libBabou "babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type HomeController struct {
	actionMap map[string]libBabou.Action
}

// Registers actions for the HomeController and returns it.
// Note: State in the returned controller object is global to 
// all requests the controller processes.
func NewHomeController() *HomeController {
	hc := &HomeController{}
	hc.actionMap = make(map[string]libBabou.Action)

	//add your actions here.
	hc.actionMap["index"] = hc.Index

	return hc
}

func (hc *HomeController) HandleRequest(action string, 
	params map[string]string) *libBabou.Result {
	
	if hc.actionMap[action] != nil {
		return hc.actionMap[action](params)
	} else {
		return &libBabou.Result{Status: 404, Body: []byte("Not found")}
	}
}

func (hc *HomeController) Index(params map[string]string) *libBabou.Result {
	output := &libBabou.Result{}
	
	output.Status = 200
	output.Body = []byte("hello from HomeController#index")

	return output
}
