// The `babou` web application server.
//
// The application package includes both a server implementation and a router.
// The web application itself is included in various sub-packages, and is loaded
// by the router once the server has started.
package app

import (
	"github.com/drbawb/babou/bridge"
	libBabou "github.com/drbawb/babou/lib"

	fmt "fmt"
	log "log"

	http "net/http"
)

// Parameters for babou's web server
type Server struct {
	Port int

	serverIO    chan int              // Output for process monitor
	AppBridge   *bridge.Bridge        // Event bridge this server can use to comm. with other trackers.
	AppSettings *libBabou.AppSettings // Settings this server was started with
}

// Initializes a server using babou/lib settings, an event bridge, and a communications channel
// with the monitoring process.
func NewServer(appSettings *libBabou.AppSettings, bridge *bridge.Bridge, serverIO chan int) *Server {
	newServer := &Server{}

	newServer.AppSettings = appSettings
	newServer.Port = appSettings.WebPort
	newServer.serverIO = serverIO

	newServer.AppBridge = bridge

	return newServer
}

// Initializes the babou web application framework
// Starts listening for requests on specified port and passing them
// through the stack.
func (s *Server) Start() {
	log.Printf("Babou is starting his web server on port: %d", s.Port)

	go func() {
		s.loadRoutes()
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil))
	}()

	s.serverIO <- libBabou.WEB_SERVER_STARTED

	// Handle signals
}

// Loads muxer from router.go from `app` package.
func (s *Server) loadRoutes() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Error loading routes: \n %s", r)

		}
	}()

	http.Handle("/", LoadRoutes(s))
}
