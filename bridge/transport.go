package bridge

import (
	"fmt"

	"net"
)

type TransportError int

const (
	TRANSPORT_NOT_AVAILABLE TransportError = iota
)

type Transport interface {
	Send(msg *Packet) // Sends a message to the specified socket
}

type UnixTransport struct {
	socketAddr string

	queue chan *Packet // TODO: could repurpose as send buffer in future.
}

type TCPTransport struct {
	socketAddr string

	queue chan *Packet // TODO: could repurpose as send buffer in future.
}

type LocalTransport struct {
	queue chan *Packet
}

// Forwards message to locally available transport.
// Must be used on an existing bridge.
func (b *Bridge) NewLocalTransport() *LocalTransport {
	transport := &LocalTransport{queue: b.inbox} // Use bridge's "receiver" channel to send messages.

	return transport
}

// Loop packet around to bridge's inbox.
func (lt *LocalTransport) Send(msg *Packet) {
	lt.queue <- msg
}

func NewUnixTransport(socketAddr string) *UnixTransport {
	transport := &UnixTransport{socketAddr: socketAddr, queue: make(chan *Packet)}
	go transport.processQueue()

	return transport
}

func (ut *UnixTransport) Send(msg *Packet) {
	ut.queue <- msg
}

func (ut *UnixTransport) processQueue() {
	for {
		select {
		case msg := <-ut.queue:
			c, err := net.Dial("unix", ut.socketAddr)
			if err != nil {
				fmt.Printf("Trouble calling peer[%s]: %s", ut.socketAddr, err.Error())
				continue
			}

			defer c.Close()
			_, err = c.Write(encodeMsg(*msg.Payload))
			if err != nil {
				fmt.Printf("Trouble sending payload to peer[%s]: %s", ut.socketAddr, err.Error())
			}
		}
	}
}

func NewTCPTransport(socketAddr string) *TCPTransport {
	transport := &TCPTransport{socketAddr: socketAddr, queue: make(chan *Packet)}
	go transport.processQueue()

	return transport
}

func (tcp *TCPTransport) Send(msg *Packet) {
	tcp.queue <- msg
}

func (tcp *TCPTransport) processQueue() {
	for {
		select {
		case msg := <-tcp.queue:
			c, err := net.Dial("tcp", tcp.socketAddr)
			if err != nil {
				fmt.Printf("Trouble calling peer[%s]: %s", tcp.socketAddr, err.Error())
				continue
			}

			defer c.Close()
			_, err = c.Write(encodeMsg(*msg.Payload))
			if err != nil {
				fmt.Printf("Trouble sending payload to peer[%s]: %s", tcp.socketAddr, err.Error())
			}
		}
	}
}
