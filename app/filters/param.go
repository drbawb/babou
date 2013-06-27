package filters

import (
	"errors"
	"fmt"
	"net/http"

	web "babou/lib/web"
)

// A chain which can store and retrieve GET/POST request variables.
type ParamterChainLink interface {
	SetParams(map[string]string)
	GetParams() map[string]string
}

// A controller which accepts GET/POST request variables.
type ParameterizedController interface {
	SetContext(*DevContext) error
}

// Test impl. of Context interface.
type DevContext struct {
	Params map[string]string
	isInit bool
}

func ParameterChain() *DevContext {
	context := &DevContext{isInit: false}

	return context
}

func (dc *DevContext) TestContext(route web.Route, chain []web.ChainableContext) error {
	_, ok := route.(ParameterizedController)
	if !ok {
		return errors.New(fmt.Sprintf("Route :: %T :: does not support the paramter context", route))
	}

	return nil
}

func (dc *DevContext) NewInstance() web.ChainableContext {
	newDc := &DevContext{isInit: false}

	return newDc
}

func (dc *DevContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request, chain []web.ChainableContext) {

	dc.SetParams(web.RetrieveAllParams(request))
	dc.isInit = true

	v, ok := controller.(ParameterizedController)
	if ok {
		if err := v.SetContext(dc); err != nil {
			fmt.Printf("Error setting paramter context: %s \n", err.Error())
		}
	} else {
		fmt.Printf("Tried to wrap a controller that is not request parameter aware \n")
	}
}

// Implements ParameterChain; used as a basis for all contextChains
func (dc *DevContext) SetParams(params map[string]string) {
	dc.Params = params
}

// Implements ParameterChain; used as a basis for all contextChains
func (dc *DevContext) GetParams() map[string]string {
	return dc.Params
}
