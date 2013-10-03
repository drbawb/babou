package bridge

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func encodeMsg(msg Message) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	encoder := gob.NewEncoder(buf)

	encoder.Encode(msg)

	return buf.Bytes()
}

func decodeMsg(encodedMessage *bytes.Buffer) Message {
	decoder := gob.NewDecoder(encodedMessage)
	msg := Message{}
	err := decoder.Decode(&msg)
	if err != nil {
		fmt.Printf("error decoding message... %s", err.Error())
	}

	return msg
}
