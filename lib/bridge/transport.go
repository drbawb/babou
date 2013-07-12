package bridge

type Transport interface {
	Send() chan<- *Message
	Recv() <-chan *Message
}
