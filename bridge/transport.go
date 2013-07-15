package bridge

import (
	"fmt"
	"net"
)

type Transport interface {
	Send(msg *Message) // Sends a message to the specified socket
}

type UnixTransport struct {
	socketAddr string

	queue chan *Message // TODO: could repurpose as send buffer in future.
}

type TCPTransport struct {
	socketAddr string

	queue chan *Message // TODO: could repurpose as send buffer in future.
}

type LocalTransport struct {
	queue chan *Message
}

// Forwards message to locally available transport.
// Must be used on an existing bridge.
func (b *Bridge) NewLocalTransport() *LocalTransport {
	transport := &LocalTransport{queue: b.in} // Use bridge's "receiver" channel to send messages.

	return transport
}

func (lt *LocalTransport) Send(msg *Message) {
	lt.queue <- msg
}

func NewUnixTransport(socketAddr string) *UnixTransport {
	transport := &UnixTransport{socketAddr: socketAddr, queue: make(chan *Message)}
	go transport.processQueue()

	return transport
}

func (ut *UnixTransport) Send(msg *Message) {
	ut.queue <- msg
}

func (ut *UnixTransport) processQueue() {
	for {
		select {
		case msg := <-ut.queue:
			c, err := net.Dial("unix", ut.socketAddr)
			if err != nil {
				panic(err)
			}

			defer c.Close()

			n, err := c.Write(encodeMsg(msg))
			if err != nil {
				fmt.Printf("error sending message to %s because: %s", ut.socketAddr, err.Error())
			} else {
				fmt.Printf("%d bytes written to socket \n", n)
			}
		}
	}
}

func NewTCPTransport(socketAddr string) *TCPTransport {
	transport := &TCPTransport{socketAddr: socketAddr, queue: make(chan *Message)}
	go transport.processQueue()

	return transport
}

func (tcp *TCPTransport) Send(msg *Message) {
	tcp.queue <- msg
}

func (tcp *TCPTransport) processQueue() {
	for {
		select {
		case msg := <-tcp.queue:
			c, err := net.Dial("tcp", tcp.socketAddr)
			if err != nil {
				panic(err)
			}

			defer c.Close()

			n, err := c.Write(encodeMsg(msg))
			if err != nil {
				fmt.Printf("error sending message to %s because: %s", tcp.socketAddr, err.Error())
			} else {
				fmt.Printf("%d bytes written to socket \n", n)
			}
		}
	}
}
