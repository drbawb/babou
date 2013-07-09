package tracker

import (
	libTorrent "github.com/drbawb/babou/lib/torrent"
	libWeb "github.com/drbawb/babou/lib/web"

	"github.com/drbawb/babou/app/models"

	mux "github.com/gorilla/mux"
	bencode "github.com/zeebo/bencode"

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
	routeVars := mux.Vars(r)

	torrent, ok := s.torrentExists(params["info_hash"])

	user := &models.User{}
	if err := user.SelectSecret(routeVars["secret"]); err != nil {
		fmt.Printf("Tracker could not find user by secret\n Error: %s\n Secret: %s \n", err.Error(), routeVars["secret"])

		responseMap["failure reason"] = "user could not be found. please redownload the torrent."
		io.Copy(w, encodeResponseMap(responseMap))

		return
	}

	if !libTorrent.CheckHmac(routeVars["secret"], routeVars["hash"]) {
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

	torrent.AddPeer(params["peer_id"], r.RemoteAddr, params["port"], routeVars["secret"])

	responseMap["interval"] = 300 // intentionally short for debugging purposes.
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
func (s *Server) torrentExists(infoHash string) (*libTorrent.Torrent, bool) {
	torrent := s.torrentCache[infoHash]

	if torrent == nil {
		//cache miss
		return nil, false
	} else {
		//cache hit.
		return torrent, true
	}
}
