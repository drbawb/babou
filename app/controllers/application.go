package controllers

import (
	web "babou/lib/web"
	errors "errors"
)

// A generic routine that will implement `Process` for any `Route` interface
func process(route web.Route, action string) (web.Controller, error) {
	if !route.IsSafeInstance() {
		controller := route.NewInstance() // get a controller

		return controller, nil
	}

	return nil, errors.New("This controller is not equipped to service public facing requests")
}
