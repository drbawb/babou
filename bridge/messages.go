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

	TORRENT_STAT_TUPLE
)

type Packet struct {
	SubscriberName string
	Payload        *Message
}

func init() {
	gob.Register(Message{})
	gob.Register(DeleteUserMessage{})
	gob.Register(DeleteTorrentMessage{})
	gob.Register(TorrentStatMessage{})

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

type TorrentStatMessage struct {
	InfoHash string
	Seeding  int
	Leeching int
}

// Creates a torrent-stat tuple
func TorrentStats(
	infoHash string,
	seeding,
	leeching int) *Message {

	payload := &TorrentStatMessage{
		InfoHash: infoHash,
		Seeding:  seeding,
		Leeching: leeching,
	}

	wrapper := &Message{Type: TORRENT_STAT_TUPLE, Payload: payload}

	return wrapper
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
