package tracker

import (
	libBridge "github.com/drbawb/babou/bridge"
	lib "github.com/drbawb/babou/lib"
	libTorrent "github.com/drbawb/babou/lib/torrent"
	libWeb "github.com/drbawb/babou/lib/web"

	models "github.com/drbawb/babou/app/models"

	bencode "github.com/zeebo/bencode"

	"encoding/hex"

	"bytes"
	//"fmt"
	"io"
	"net/http"
)

// This block defines several preset responses for common failures.
var failureResponses map[PredefinedResponse]([]byte)

type PredefinedResponse int

const (
	RESP_USER_NOT_FOUND PredefinedResponse = iota
	RESP_TORRENT_NOT_FOUND
)

func init() {
	failureResponses = make(map[PredefinedResponse]([]byte))

	var bytesBuf *bytes.Buffer

	bytesBuf = bytes.NewBuffer(make([]byte, 0))
	unf := map[string]interface{}{"failure reason": "user could not be found."}
	io.Copy(bytesBuf, encodeResponseMap(unf))
	failureResponses[RESP_USER_NOT_FOUND] = bytesBuf.Bytes()

	bytesBuf = bytes.NewBuffer(make([]byte, 0))
	tnf := map[string]interface{}{"failure reason": "torrent could not be found."}
	io.Copy(bytesBuf, encodeResponseMap(tnf))
	failureResponses[RESP_TORRENT_NOT_FOUND] = bytesBuf.Bytes()

}

// Handles announce from a client.
// Some TODOs:
//  * Bail out early if secret/hash or request is obviously malformed. (Not from a well-behaved torrent client.)
//  * Cache users and their secrets. (Presumably if they have started one torrent they will start many more.)
//  * Intelligent peer-list generation.
//  * Set `Content-Length` ?
func announceHandle(w http.ResponseWriter, r *http.Request, s *Server) {
	w.Header().Set("Content-Type", "text/plain")

	params := libWeb.RetrieveAllParams(r)
	responseMap := make(map[string]interface{})

	hexHash := hex.EncodeToString([]byte(params.All["info_hash"]))

	torrent, ok := s.torrentExists(hexHash)

	user := &models.User{}
	if err := user.SelectSecret(params.All["secret"]); err != nil {
		w.Write(failureResponses[RESP_USER_NOT_FOUND])

		return
	}

	if !libTorrent.CheckHmac(params.All["secret"], params.All["hash"]) {
		w.Write(failureResponses[RESP_USER_NOT_FOUND])

		return
	}

	// TODO: tracker request log.

	if !ok {
		w.Write(failureResponses[RESP_TORRENT_NOT_FOUND])
		return
	}

	responseMap["interval"] = lib.TRACKER_ANNOUNCE_INTERVAL // intentionally short for debugging purposes.
	responseMap["min interval"] = 10

	seeding, leeching := torrent.EnumeratePeers()
	responseMap["complete"] = seeding
	responseMap["incomplete"] = leeching

	responseMap["peers"], responseMap["peers6"] = torrent.GetPeerList(0) //naive peer ranker.
	io.Copy(w, encodeResponseMap(responseMap))

	// Defer writes outside of response
	// (Just in case we block on DB access or have to contend for the peer list's mutex)
	go func() {
		if params.All["event"] == "stopped" {
			// TODO: remove peer method
			torrent.WritePeers(func(peerMap map[string]*libTorrent.Peer) {
				delete(peerMap, params.All["peer_id"])
			})
		} else {
			torrent.AddPeer(
				params.All["peer_id"],
				r.RemoteAddr,
				params.All["port"],
				params.All["secret"],
			)

			torrent.UpdateStatsFor(params.All["peer_id"], "0", "0", params.All["left"])
		}

		// Send stats over event bridge.
		stats := libBridge.TorrentStatMessage{}
		stats.InfoHash = torrent.InfoHash
		stats.Seeding, stats.Leeching = torrent.EnumeratePeers()

		message := &libBridge.Message{}
		message.Type = libBridge.TORRENT_STAT_TUPLE
		message.Payload = stats

		// TODO: Reaper needs to send this event
		// when a peer is removed.

		s.eventBridge.Publish("tracker", message)

	}()
}

// Bencodes an arbitrary dictionary as a tracker response.
func encodeResponseMap(responseMap map[string]interface{}) io.Reader {
	responseBuf := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(responseBuf)
	_ = encoder.Encode(responseMap)

	return responseBuf
}

// Checks if the torrent exists in cache. Otherwise attempts to fill
// cache from the database.
//
// TODO:
// * Cache-filler should probably have some sort of timeout per torrent.
// * Cache-filler should check that `info_hash` is not obviously malformed.
// * Distributed/coordinated cache fills? [ref: groupcache]
func (s *Server) torrentExists(infoHash string) (*libTorrent.Torrent, bool) {
	torrent := s.torrentCache[infoHash]

	if torrent == nil {
		// Cache miss

		dbTorrent := &models.Torrent{}
		if err := dbTorrent.SelectHash(infoHash); err != nil {
			// Check DB for torrent
			return nil, false
		} else {
			// Attempt to lib.Torrent{} from database model failed.
			prepareTorrent, err := dbTorrent.LoadTorrent()
			if err != nil {
				return nil, false
			}

			trackerTorrent := libTorrent.NewTorrent(prepareTorrent)
			trackerTorrent.InfoHash = dbTorrent.InfoHash
			s.torrentCache[infoHash] = trackerTorrent

			return s.torrentCache[infoHash], true
		}

		// All torrent cache-filling strategies have failed.
		return nil, false
	} else {
		// Cache hit

		return torrent, true
	}
}
