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

	in  chan *Message // channel of messages to be read from other transports
	out chan *Message // channel of messages to be shared with other transports

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
		transports: make([]Transport, 0),
		in:         make(chan *Message, BRIDGE_RECV_BUFFER),
		out:        make(chan *Message, BRIDGE_SEND_BUFFER),
		quit:       make(chan bool),
	}

	// Implement all transport types for the default bridge.
	switch settings.Transport {
	case lib.UNIX_TRANSPORT:
		go bridge.netListen("unix", settings.Socket)
	case lib.TCP_TRANSPORT:
		go bridge.netListen("tcp", settings.Socket)
	default:
		fmt.Printf("you have selected an unimplemented bridge type. \n")
	}

	go bridge.broadcast()

	return bridge
}

// TODO: Recover from connection failure.
func (b *Bridge) AddTransport(transport Transport) {
	b.transports = append(b.transports, transport)
}

func (b *Bridge) broadcast() {
	for {
		select {
		case msg := <-b.out:
			if len(b.transports) == 0 {
				fmt.Printf("No transports avail. Event dropped: %v \n", msg)
			}

			for _, tp := range b.transports {
				tp.Send(msg)
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
			fmt.Printf("error listening for pack: %s \n", err.Error())
			break
		}

		msgBuf := make([]byte, 1024)
		n, err := fd.Read(msgBuf[:])
		if err != nil {
			fmt.Printf("error reading file: %s \n", err.Error())
		}

		fmt.Printf("read %d bytes from socket \n", n)

		// gob decode it into a bridge-message
		b.in <- decodeMsg(bytes.NewBuffer(msgBuf)) // send blocked receiver a message
	}
}

// Sends a message on a channel.
// Will block indefinitely if the send-buffer is filled and not being drained.
func (b *Bridge) Send(msg *Message) {
	// send message to other transports
	// TODO: dummy message in here.
	msg = &Message{}
	msg.Type = DELETE_USER

	dmm := &DeleteUserMessage{}
	dmm.UserId = 9001

	msg.Payload = dmm

	b.out <- msg
}

// Returns a channel immediately.
//
// When the bridge has sucesfully placed your message
// into the send buffer, a single integer will
// be sent on the returned channel.
func (b *Bridge) ASend(msg *Message) <-chan int {
	// send message to other transports
	// TODO: dummy message in here.
	retChan := make(chan int, 1)
	go func(status chan int) {
		b.Send(msg) // try to send message
		status <- 1 // message sent OK
	}(retChan)

	return retChan
}

func (b *Bridge) Recv() <-chan *Message {
	return b.in
}

func (b *Bridge) Close() chan bool {
	return b.quit
}
