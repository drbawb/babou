package controllers

import (
	web "babou/lib/web"
	errors "errors"
)

// A generic routine that will implement `Process` for any `Route` interface
func process(route web.Route, action string, context web.Context) (web.Controller, error) {
	if !route.IsSafeInstance() {
		controller := route.NewInstance() // get a controller

		if err := controller.SetContext(context); err != nil {
			return nil, err
		}

		return controller, nil
	}

	return nil, errors.New("This controller is not equipped to service public facing requests")
}
