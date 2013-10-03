// This event bridge allows different instances of babou
// to securely exchange messages which help keep the website
// and tracker in sync.
package bridge

import (
	"encoding/gob"
)

type MessageType uint8

const (
	_                           = iota
	UPDATE_USER_KEY MessageType = iota

	CHANGE_USER_TOKEN

	WATCH_USERS

	DELETE_USER
	DISABLE_USER

	DELETE_TORRENT
	DISABLE_TORRENT
)

type Packet struct {
	SubscriberName string
	Payload        *Message
}

// message wrapper for quick decoding on other end.
type Message struct {
	Type    MessageType
	Payload interface{}
}

type DeleteUserMessage struct {
	UserId int
}

type DeleteTorrentMessage struct {
	InfoHash string
	Reason   string
}

func init() {
	gob.Register(&Message{})
	gob.Register(&DeleteUserMessage{})
}

// Instructs trackers to remove a user from their cache ASAP
func DeleteUser(userId int) {
	wrapper := Message{Type: DELETE_USER}
	payload := DeleteUserMessage{UserId: userId}

	wrapper.Payload = payload
}

// Instructs trackers to remove a torrent from their cache ASAP
func DeleteTorrent(torrentHash string) {
	wrapper := Message{Type: DELETE_TORRENT}
	payload := DeleteTorrentMessage{InfoHash: torrentHash}

	wrapper.Payload = payload
}
