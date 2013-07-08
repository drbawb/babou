// The `babou` high-performance BitTorrent tracker.
//
// Implements the core tracker-request router
package tracker

import (
	libBabou "github.com/drbawb/babou/lib"
	libTorrent "github.com/drbawb/babou/lib/torrent"
	libWeb "github.com/drbawb/babou/lib/web"

	bencode "github.com/zeebo/bencode"

	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Parameters for babou's web server
type Server struct {
	Port         int
	serverIO     chan int
	torrentCache map[string]*libTorrent.TorrentFile
}

// Initializes a server using babou/lib settings and a communication channel.
func NewServer(appSettings *libBabou.AppSettings, serverIO chan int) *Server {
	newServer := &Server{torrentCache: make(map[string]*libTorrent.TorrentFile)}

	newServer.Port = *appSettings.TrackerPort
	newServer.serverIO = serverIO

	return newServer
}

func (s *Server) Start() {
	router := LoadRoutes(s)

	go func() {
		// start with custom muxer.
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), router))
	}()
}

func wrapAnnounceHandle(s *Server) http.HandlerFunc {
	//s.torrentCache[string(torrent.EncodeInfo())] = torrent

	fn := func(w http.ResponseWriter, r *http.Request) {
		announceHandle(w, r, s)
	}

	return fn
}

func announceHandle(w http.ResponseWriter, r *http.Request, s *Server) {
	params := libWeb.RetrieveAllParams(r)
	responseMap := make(map[string]interface{})

	fmt.Printf("request from (ip): %s:%s \n, has left(bytes): %s, compact: %s \n --- \n",
		params["ip"], params["port"], params["left"], params["compact"])

	torrent, ok := s.torrentExists(params["info_hash"])
	if ok {
		torrent.AddPeer(params["peer_id"], r.RemoteAddr, params["port"])

		//responseMap["failure reason"] = "tracker found torrent but doesnt know what to do now!"
		responseMap["interval"] = 300
		responseMap["min interval"] = 10

		responseMap["tracker id"] = 12345

		responseMap["complete"] = 1
		responseMap["incomplete"] = 1

		responseMap["peers"] = torrent.GetPeerList()
	} else {
		responseMap["failure reason"] = "tracker could not find requested torrent."
	}

	responseBuf := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(responseBuf)

	err := encoder.Encode(responseMap)
	if err != nil {
		w.Write([]byte("server error"))
	}

	io.Copy(w, responseBuf)
	w.Header().Set("Content-Type", "text/plain")
}

func (s *Server) torrentExists(infoHash string) (*libTorrent.TorrentFile, bool) {
	torrent := s.torrentCache[infoHash]

	if torrent == nil {
		return nil, false
	} else {
		return torrent, true
	}
}

// Test method for loading torrents.
func ReadFile(filename string) *libTorrent.TorrentFile {
	return libTorrent.ReadFile(filename)
}
