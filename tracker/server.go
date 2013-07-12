// The `babou` high-performance BitTorrent tracker.
//
// Implements the core tracker-request router
package tracker

import (
	libBabou "github.com/drbawb/babou/lib"
	libTorrent "github.com/drbawb/babou/lib/torrent"
	tasks "github.com/drbawb/babou/tracker/tasks"

	"fmt"
	"log"
	"net/http"
	"time"
)

// Parameters for babou's web server
type Server struct {
	Port int

	serverIO     chan int
	torrentCache map[string]*libTorrent.Torrent

	peerReaper *tasks.PeerReaper
}

// Initializes a server using babou/lib settings and a communication channel.
func NewServer(appSettings *libBabou.AppSettings, serverIO chan int) *Server {
	newServer := &Server{torrentCache: make(map[string]*libTorrent.Torrent)}

	newServer.Port = *appSettings.TrackerPort
	newServer.serverIO = serverIO
	newServer.peerReaper = &tasks.PeerReaper{} //TODO: constructor.

	return newServer
}

func (s *Server) Start() {
	router := LoadRoutes(s)

	go func() {
		// start with custom muxer.
		log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.Port), router))
	}()

	//TODO: task scheduler of some kind.
	go func() {
		tenMinutes := time.Duration(10) * time.Minute
		timer := time.NewTicker(tenMinutes)

		for {
			select {
			case _ = <-timer.C:
				fmt.Printf("\n reaping peers . . . \n")
				for _, v := range s.torrentCache {
					s.peerReaper.ReapTorrent(v)
				}
			}
		}

	}()
}

func wrapAnnounceHandle(s *Server) http.HandlerFunc {

	fn := func(w http.ResponseWriter, r *http.Request) {
		announceHandle(w, r, s)
	}

	return fn
}
