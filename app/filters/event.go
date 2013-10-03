package filters

import (
	"errors"
	"fmt"
	"net/http"

	bridge "github.com/drbawb/babou/bridge"
	web "github.com/drbawb/babou/lib/web"
)

const (
	EVENT_TIMEOUT  int    = 5
	EVENT_CTX_NAME string = "web-event-ctx"
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

	bridge   *bridge.Bridge
	memStats map[string]*bridge.TorrentStatMessage
}

// Sends a properly typed message over the bridge.
//
// This may block indefinitely if the send buffer is full.
// This should be used for callers that are not time-sensitive OR for
// callers that must take responsibility for delivery of a message.
func (ec *EventContext) SendMessage(msg *bridge.Message) {
	ec.bridge.Publish(EVENT_CTX_NAME, msg)
}

func (ec *EventContext) ReadStats(infoHash string) *bridge.TorrentStatMessage {
	fmt.Printf("[ec] fetching stats for: %s \n", infoHash)
	return ec.memStats[infoHash]
}

// Returns an uninitialized AuthContext suitable for use in a context chain
// TODO: Synchronized so long as this is the only subscriber writing
// to the event context's internal structures.
func EventChain(serverBridge *bridge.Bridge) *EventContext {
	context := &EventContext{
		isInit:   false,
		bridge:   serverBridge,
		memStats: make(map[string]*bridge.TorrentStatMessage),
	}

	//TODO: factor out
	// Listens for messages from the bridge and dispatches
	// them to the model layer for persistence.
	go func() {
		messages := make(chan *bridge.Message)
		context.bridge.Subscribe(EVENT_CTX_NAME, messages)

		for {
			select {
			case msg := <-messages:
				switch msg.Type {
				case bridge.TORRENT_STAT_TUPLE:
					stats := msg.Payload.(*bridge.TorrentStatMessage)
					fmt.Printf("[ec] Writing stats for %v \n", *stats)
					context.memStats[stats.InfoHash] = stats
				default:
					fmt.Printf(
						"Event bridge has no handler for messages of type: %v \n",
						msg.Type,
					)
				}
			}
		}

	}()

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
