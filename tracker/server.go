// The `babou` high-performance BitTorrent tracker.
//
// Implements the core tracker-request router
package tracker

import (
	libBabou "github.com/drbawb/babou/lib"
	libTorrent "github.com/drbawb/babou/lib/torrent"

	"fmt"
	"log"
	"net/http"
)

// Parameters for babou's web server
type Server struct {
	Port         int
	serverIO     chan int
	torrentCache map[string]*libTorrent.Torrent
}

// Initializes a server using babou/lib settings and a communication channel.
func NewServer(appSettings *libBabou.AppSettings, serverIO chan int) *Server {
	newServer := &Server{torrentCache: make(map[string]*libTorrent.Torrent)}

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
	torrentData := ReadFile("/test.torrent")
	if torrentData != nil {
		s.torrentCache[string(torrentData.Info.EncodeInfo())] = torrentData
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		announceHandle(w, r, s)
	}

	return fn
}

// Test method for loading torrents.
func ReadFile(filename string) *libTorrent.Torrent {
	return libTorrent.ReadFile(filename)
}
