// The `babou` high-performance BitTorrent tracker.
//
// Implements the core tracker-request router
package tracker

import (
	libBabou "github.com/drbawb/babou/lib"
	libTorrent "github.com/drbawb/babou/lib/torrent"
	//libWeb "github.com/drbawb/babou/lib/web"
	bencode "github.com/zeebo/bencode"

	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Parameters for babou's web server
type Server struct {
	Port     int
	serverIO chan int
}

// Initializes a server using babou/lib settings and a communication channel.
func NewServer(appSettings *libBabou.AppSettings, serverIO chan int) *Server {
	newServer := &Server{}

	newServer.Port = *appSettings.TrackerPort
	newServer.serverIO = serverIO

	return newServer
}

func (s *Server) Start() {
	router := LoadRoutes()

	go func() {
		// start with custom muxer.
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), router))
	}()
}

func announceHandle(w http.ResponseWriter, r *http.Request) {
	//params := libWeb.RetrieveAllParams(r)
	responseMap := make(map[string]interface{})
	responseMap["failure reason"] = "tracker could not find requested torrent."

	responseBuf := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(responseBuf)

	err := encoder.Encode(responseMap)
	if err != nil {
		w.Write([]byte("server error"))
	}

	io.Copy(w, responseBuf)
	w.Header().Set("Content-Type", "text/plain")
}

func ReadFile(filename string) {
	libTorrent.ReadFile(filename)
}
