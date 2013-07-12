package tracker

import (
	lib "github.com/drbawb/babou/lib"
	libTorrent "github.com/drbawb/babou/lib/torrent"
	libWeb "github.com/drbawb/babou/lib/web"

	models "github.com/drbawb/babou/app/models"

	bencode "github.com/zeebo/bencode"

	"encoding/hex"

	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Handles an announce from a client
// Some TODOs:
//  * Bail out early if secret/hash or request is obviously malformed. (Not from a well-behaved torrent client.)
//  * Cache users and their secrets. (Presumably if they have started one torrent they will start many more.)
//  * Intelligent peer-list generation.
func announceHandle(w http.ResponseWriter, r *http.Request, s *Server) {
	w.Header().Set("Content-Type", "text/plain")

	params := libWeb.RetrieveAllParams(r)
	responseMap := make(map[string]interface{})

	hexHash := hex.EncodeToString([]byte(params.All["info_hash"]))
	fmt.Printf("params.All encoded hash: %s \n", hexHash)
	torrent, ok := s.torrentExists(hexHash)
	fmt.Printf("\n---\nparams: %v \n---\n", params.All)

	user := &models.User{}
	if err := user.SelectSecret(params.All["secret"]); err != nil {
		fmt.Printf("Tracker could not find user by secret\n Error: %s\n Secret: %s \n", err.Error(), params.All["secret"])

		responseMap["failure reason"] = "user could not be found. please redownload the torrent."
		io.Copy(w, encodeResponseMap(responseMap))

		return
	}

	if !libTorrent.CheckHmac(params.All["secret"], params.All["hash"]) {
		fmt.Printf("Tracker did not issue this user-secret, or the shared-secret has rotated since this profile was created.")

		responseMap["failure reason"] = "your user secret is out-of-date, please update it from your profile and redownload the torrent."
		io.Copy(w, encodeResponseMap(responseMap))

		return
	}

	// TODO: tracker request log.
	fmt.Printf("Incoming request (user OK) \n---\n %v \n---\n", params)

	if !ok {
		responseMap["failure reason"] = "tracker could not find requested torrent"
		io.Copy(w, encodeResponseMap(responseMap))

		return
	}

	torrent.AddPeer(params.All["peer_id"], r.RemoteAddr, params.All["port"], params.All["secret"])
	torrent.UpdateStatsFor(params.All["peer_id"], "0", "0", params.All["left"])

	responseMap["interval"] = lib.TRACKER_ANNOUNCE_INTERVAL // intentionally short for debugging purposes.
	responseMap["min interval"] = 10

	seeding, leeching := torrent.EnumeratePeers()
	responseMap["complete"] = seeding
	responseMap["incomplete"] = leeching

	responseMap["peers"] = torrent.GetPeerList(0) //naive peer ranker.
	io.Copy(w, encodeResponseMap(responseMap))
}

// Bencodes a dictionary as a tracker response.
func encodeResponseMap(responseMap map[string]interface{}) io.Reader {
	responseBuf := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(responseBuf)

	if err := encoder.Encode(responseMap); err != nil {
		fmt.Printf("Error encoding response: %s \n", err.Error())
	}

	return responseBuf
}

// Checks if the torrent exists in cache.
// Otherwise looks it up from database.
// TODO: this cache will need to be protected by sync primitives in the future.
// (that can probably wait until the distributed tracker cache / site<->tracker pipeline is finished.)
func (s *Server) torrentExists(infoHash string) (*libTorrent.Torrent, bool) {
	torrent := s.torrentCache[infoHash]

	if torrent == nil {
		//cache miss
		dbTorrent := &models.Torrent{}
		if err := dbTorrent.SelectHash(infoHash); err != nil {
			fmt.Printf("error retrieving torrent by hash: %s \n", infoHash)
			return nil, false
		} else {
			fmt.Printf("found torrent by hash")
			trackerTorrent := libTorrent.NewTorrent(dbTorrent.LoadTorrent())
			s.torrentCache[infoHash] = trackerTorrent
			return s.torrentCache[infoHash], true
		}

		fmt.Printf("missed and didn't find torrent in db.")
		return nil, false
	} else {
		//cache hit.
		fmt.Printf("cache hit for hash \n")
		return torrent, true
	}
}
