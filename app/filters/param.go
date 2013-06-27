package filters

import (
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

func (dc *DevContext) TestContext(web.Route, []ChainableContext) error {
	//use the chained instance to test if the controller supports it
	//DevContext has no external dependencies

	return nil
}

func (dc *DevContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request) {
	newDc := &DevContext{}
	newDc.SetParams(web.RetrieveAllParams(request))
	newDc.isInit = true

	v, ok := controller.(ParameterizedController)
	if ok {
		if err := v.SetContext(newDc); err != nil {
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
