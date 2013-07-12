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

	in  chan *Message // channel of messages to be read from other transports
	out chan *Message // channel of messages to be shared with other transports
}

func NewBridge() *Bridge {
	bridge := &Bridge{
		transports: make([]Transport, 0),
		in:         make(chan *Message),
		out:        make(chan *Message),
	}

	go bridge.broadcast() // wait for messages to send to other transports
	go bridge.listen()    // read

	return bridge
}

// currently only listens on unix socket.
func (b *Bridge) listen() {
	fmt.Printf("listening on /tmp/babou.8081.sock \n")
	l, err := net.Listen("unix", "/tmp/babou.8081.sock")
	if err != nil {
		panic(err)
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			fmt.Printf("error listening for pack: %s \n", err.Error())
		}

		msgBuf := make([]byte, 0)
		_, err = fd.Read(msgBuf)
		if err != nil {
			fmt.Printf("error reading file: %s \n", err.Error())
		}

		// gob decode it into a bridge-message
		b.in <- decodeMsg(bytes.NewBuffer(msgBuf)) // send blocked receiver a message
	}
}

// Starts routing requests over transports.
func (b *Bridge) broadcast() {
	for {
		select {
		case msg := <-b.out:
			c, err := net.Dial("unix", "/tmp/babou.8080.sock")
			if err != nil {
				panic(err)
			}

			defer c.Close()

			testMsg := &Message{}
			testMsg.Type = DELETE_USER
			_, err = c.Write(encodeMsg(msg))
			if err != nil {
				fmt.Printf("error sending message to 8080: %s", err.Error())
			}
		}
	}
}

func (b *Bridge) Send(msg *Message) {
	// send message to other transports
	b.out <- &Message{}
}

func (b *Bridge) Recv() <-chan *Message {
	return b.in
}
