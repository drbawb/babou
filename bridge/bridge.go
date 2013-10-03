package bridge

import (
	"bytes"
	"fmt"
	"net"

	"github.com/drbawb/babou/lib"
)

// Represents the programs bridge to send messages to the other pack members.
// The default route will discard all messages sent through the bridge.
type Bridge struct {
	transports []Transport // other bridges to deliver messages to

	in chan *Packet // channel of messages to be read from other transports

	subscribers map[string]chan<- *Message

	quit chan bool // send any value to gracefully shutdown the bridge.
}

const (
	BRIDGE_SEND_BUFFER int = 10
	BRIDGE_RECV_BUFFER     = 10
)

// Sets up a bridge w/ a send buffer attached to nothing.
// All messages will be dropped to drain the buffer until transport(s) are available.
func NewBridge(settings *lib.TransportSettings) *Bridge {
	bridge := &Bridge{
		transports:  make([]Transport, 0),
		in:          make(chan *Packet, BRIDGE_RECV_BUFFER),
		quit:        make(chan bool),
		subscribers: make(map[string]chan<- *Message),
	}

	// Implement all transport types for the default bridge.
	switch settings.Transport {
	case lib.UNIX_TRANSPORT:
		go bridge.netListen("unix", settings.Socket)
	case lib.TCP_TRANSPORT:
		go bridge.netListen("tcp", settings.Socket)
	case lib.LOCAL_TRANSPORT:
		bridge.AddTransport(bridge.NewLocalTransport())
	default:
		fmt.Printf("you have selected an unimplemented bridge type. \n")
	}

	go bridge.dispatch()

	return bridge
}

// TODO: Recover from connection failure.
func (b *Bridge) AddTransport(transport Transport) {
	b.transports = append(b.transports, transport)
}

// Wrap message in a packet for transport
func (b *Bridge) broadcast(name string, msg *Message) {
	if len(b.transports) == 0 {
		fmt.Printf("No transports avail. Event dropped: %v \n", msg)
	}

	mpack := &Packet{}
	mpack.SubscriberName = name
	mpack.Payload = msg

	for _, tp := range b.transports {
		tp.Send(mpack)
	}
}

func (b *Bridge) dispatch() {
	for {
		select {
		case msg := <-b.in:
			for name, subscriber := range b.subscribers {
				if name != msg.SubscriberName {
					subscriber <- msg.Payload
				}
			}
		}
	}
}

// currently only listens on unix socket.
func (b *Bridge) netListen(network, addr string) {
	l, err := net.Listen(network, addr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("listening on: %s \n", addr)

	go func(net.Listener) {
		select {
		case _ = <-b.quit:
			_ = l.Close()
			b.quit <- true
		}
	}(l)

	for {
		fd, err := l.Accept()
		if err != nil {
			fmt.Printf("error listening for packet: %s \n", err.Error())
			break
		}

		msgBuf := make([]byte, 1024)
		n, err := fd.Read(msgBuf[:])
		if err != nil {
			fmt.Printf("error reading socket: %s \n", err.Error())
		}

		fmt.Printf("read %d bytes from socket \n", n)

		// gob decode message and stuff it into foreign packet
		packet := &Packet{}

		decodedMessage := decodeMsg(bytes.NewBuffer(msgBuf))
		packet.SubscriberName = "foreign"
		packet.Payload = decodedMessage

		b.in <- packet // send blocked receiver a message
	}
}

// Sends a message on a channel.
// Will block indefinitely if the send-buffer is filled and not being drained.
//
// name: name of a receiver you're listening on [so you will not recv this message]
func (b *Bridge) Publish(name string, msg *Message) {
	// TODO: Basic sanity checks; then forward to bridge for transport.
	if msg != nil {
		b.broadcast(name, msg)
	}
}

// Returns a channel immediately.
//
// When the bridge has sucesfully placed your message
// into the send buffer, a single integer will
// be sent on the returned channel.
func (b *Bridge) APublish(msg *Message) <-chan int {
	// send message to other transports
	// TODO: dummy message in here.
	retChan := make(chan int, 1)
	go func(status chan int) {
		b.Pub("async", msg) // try to send message
		status <- 1         // message sent OK
	}(retChan)

	return retChan
}

// Provide a channel for us to send events too.
// When a new event is published you will receive it.
func (b *Bridge) Subscribe(name string, c chan<- *Message) {
	b.subscribers[name] = c
}

func (b *Bridge) Close() chan bool {
	return b.quit
}
