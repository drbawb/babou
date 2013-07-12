package bridge

import (
	"bytes"
	"fmt"
	"net"
)

// Represents the programs bridge to send messages to the other pack members.
// The default route will discard all messages sent through the bridge.
type Bridge struct {
	transports []Transport // other bridges to deliver messages to

	in  chan Message // channel of messages to be read from other transports
	out chan Message // channel of messages to be shared with other transports
}

func NewBridge(listenOn TransportType, socketAddress string) *Bridge {
	bridge := &Bridge{
		transports: make([]Transport, 0),
		in:         make(chan Message),
		out:        make(chan Message),
	}

	// Implement all transport types for the default bridge.
	switch listenOn {
	case UNIX_TRANSPORT:
		go bridge.netListen("unix", socketAddress)

		unix := NewUnixTransport(socketAddress) // TODO: Want to make an unserialized loopback; this works for now.
		bridge.AddTransport(unix)
	case TCP_TRANSPORT:
		go bridge.netListen("tcp", socketAddress)

		tcp := NewTCPTransport(socketAddress) // TODO: Want to make an unserialized loopback; this works for now.
		bridge.AddTransport(tcp)
	default:
		fmt.Printf("you have selected an unimplemented bridge type. \n")
	}

	go bridge.broadcast()

	// TODO: should be from config file.

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
				fmt.Printf("No transports available for send. Message dropped: %v", msg)
			} else {
				for _, tp := range b.transports {
					tp.Send(msg)
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

	for {
		fmt.Printf("top of listen loop \n")
		fd, err := l.Accept()
		if err != nil {
			fmt.Printf("error listening for pack: %s \n", err.Error())
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

func (b *Bridge) Send(msg Message) {
	// send message to other transports
	msg = Message{}
	msg.Type = DELETE_USER

	dmm := DeleteUserMessage{}
	dmm.UserId = 9001

	msg.Payload = dmm

	b.out <- msg
}

func (b *Bridge) Recv() <-chan Message {
	return b.in
}
