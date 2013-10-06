// The `babou` high-performance BitTorrent tracker.
//
// Implements the core tracker-request router
package tracker

import (
	bridge "github.com/drbawb/babou/bridge"
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
	peerReaper   *tasks.PeerReaper

	eventBridge *bridge.Bridge
}

// Initializes a server using babou/lib settings and a communication channel.
func NewServer(appSettings *libBabou.AppSettings, eventBridge *bridge.Bridge, serverIO chan int) *Server {
	newServer := &Server{
		torrentCache: make(map[string]*libTorrent.Torrent),
	}

	newServer.Port = appSettings.TrackerPort
	newServer.serverIO = serverIO
	newServer.peerReaper = &tasks.PeerReaper{} //TODO: constructor.
	newServer.eventBridge = eventBridge

	return newServer
}

func (s *Server) Start() {
	router := LoadRoutes(s)

	go func() {
		// start with custom muxer.
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), router))
	}()

	//TODO: task scheduler of some kind.
	go func() {
		tenMinutes := time.Duration(10) * time.Minute
		timer := time.NewTicker(tenMinutes)

		for {
			select {
			case _ = <-timer.C:
				//TODO: rate limit ...
				fmt.Printf("\n reaping peers . . . \n")
				for _, v := range s.torrentCache {
					s.peerReaper.ReapTorrent(v)
				}
			}
		}

	}()
}

func handleWebEvent(message *bridge.Message) {
	fmt.Printf("Received [%v] from bridge \n", message)
	switch message.Type {
	case bridge.DELETE_TORRENT:
		v := message.Payload.(*bridge.DeleteTorrentMessage)
		fmt.Printf("Removing torrent: %s from cache; deleted because %s \n", v.InfoHash, v.Reason)
	default:
		fmt.Printf("Message dropped; unknown message type \n")
	}
}

func wrapAnnounceHandle(s *Server) http.HandlerFunc {

	fn := func(w http.ResponseWriter, r *http.Request) {
		announceHandle(w, r, s)
	}

	return fn
}
