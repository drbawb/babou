package bridge

import (
	"bytes"
	"testing"
)

// Note: if you're testing over a local bridge
// it doesn't matter if `gob` has registered types or not.
//
// Make sure `init()` for `Messages` is called!
func TestEncodeDecodeAreEquivalent(test *testing.T) {
	// Test cases
	messages := []*Message{
		&Message{
			Type:    DELETE_TORRENT,
			Payload: DeleteTorrentMessage{}},
		&Message{
			Type:    DELETE_USER,
			Payload: DeleteUserMessage{}},
		&Message{
			Type:    TORRENT_STAT_TUPLE,
			Payload: TorrentStatMessage{}},
	}

	bytesBuf := bytes.NewBuffer(make([]byte, 0, 1024))
	for _, testCase := range messages {
		bytesBuf.Reset()
		bytesBuf.Write(encodeMsg(*testCase)) // Store test-case encoded form in buffer.

		received := decodeMsg(bytesBuf) // Read test-case from buffer
		if received.Type != testCase.Type {
			test.Logf(
				"Received type[%v] does not match original type[%v]",
				received.Type,
				testCase.Type,
			)

			test.Fail()
		} else if received.Payload != testCase.Payload {
			test.Logf("Received msg[%v] does not match original msg[%v]",
				received.Payload,
				testCase.Payload)
			test.Fail()
		}
	}

}

// Time how long it takes to create & send a message.
func BenchmarkEncoder(bench *testing.B) {
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		dmu := DeleteUserMessage{UserId: 42}

		message := &Message{}
		message.Type = DELETE_USER
		message.Payload = dmu

		encodeMsg(*message)
	}
}

// Time how long it takes to receive & read a message
func BenchmarkDecoder(bench *testing.B) {
	bench.ResetTimer()

	dmu := DeleteUserMessage{UserId: 42}

	message := &Message{}
	message.Type = DELETE_USER
	message.Payload = dmu

	msgBytes := encodeMsg(*message)

	for i := 0; i < bench.N; i++ {
		decodedMsg := decodeMsg(bytes.NewBuffer(msgBytes))
		_ = decodedMsg.Payload.(DeleteUserMessage).UserId
	}
}
