package filters

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	bridge "github.com/drbawb/babou/bridge"
	web "github.com/drbawb/babou/lib/web"
)

const (
	EVENT_TIMEOUT int = 5
)

type EventChainLink interface {
	web.ChainableContext
}

// An authorizable controller must be willing to accept an AuthContext.
//
// This AuthContext will be tied to a session for a single request.
type EventController interface {
	web.Controller
	SetEventContext(*EventContext) error
}

// An implementation of SessionContext that uses it to provide helper methods for authorizing a user.
type EventContext struct {
	isInit bool

	bridge *bridge.Bridge
}

// Sends a properly typed message over the bridge.
//
// This may block indefinitely if the send buffer is full.
// This should be used for callers that are not time-sensitive OR for
// callers that must take responsibility for delivery of a message.
func (ec *EventContext) SendMessage(msg *bridge.Message) {
	ec.bridge.Send(msg)
}

// Sends a properly typed message over the bridge.
//
// This will send a message w/o blocking the caller for more than
// EVENT_TIMEOUT seconds.
func (ec *EventContext) ASendMessage(msg *bridge.Message) {
	// Start timeout.
	timeout := time.After(time.Duration(EVENT_TIMEOUT) * time.Second)

	select {
	case _ = <-ec.bridge.ASend(nil):
		fmt.Printf("message sent in EC filter \n")
	case _ = <-timeout:
		fmt.Printf("timeout reached sending EC; WARNING: send buffer may be full! \n")
	}
}

// Returns an uninitialized AuthContext suitable for use in a context chain
func EventChain(serverBridge *bridge.Bridge) *EventContext {
	context := &EventContext{isInit: false, bridge: serverBridge}

	return context
}

// Just need a supported controller and a link to the event bridge
func (ec *EventContext) TestContext(route web.Route, chain []web.ChainableContext) error {
	_, ok := route.(EventController)
	if !ok {
		return errors.New(fmt.Sprintf("The route %v does not support the event context!", route))
	}

	return nil
}

// Returns the globally routable instance of the event context
// This context can be safely shared because it does not read or write the request/response
//
// Instead it simply acts a message-broker between controllers and the web server's
// event bridge.
func (ec *EventContext) NewInstance() web.ChainableContext {
	return ec
}

// Applies the context to an authorizable controller.
func (ec *EventContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request, chain []web.ChainableContext) {
	v, ok := controller.(EventController)
	if ok {
		if err := v.SetEventContext(ec); err != nil {
			fmt.Printf("Error setting event context: %s \n", err.Error())
		}
	}
}

// No-op
func (ec *EventContext) CloseContext() {}
