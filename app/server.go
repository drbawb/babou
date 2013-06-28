// The `babou` web application server.
//
// The application package includes both a server implementation and a router.
// The web application itself is included in various sub-packages, and is loaded
// by the router once the server has started.
package app

import (
	libBabou "babou/lib"

	fmt "fmt"
	log "log"

	http "net/http"
)

// Parameters for babou's web server
type Server struct {
	Port     int
	serverIO chan int
}

// Initializes a server using babou/lib settings and a communication channel.
func NewServer(appSettings *libBabou.AppSettings, serverIO chan int) *Server {
	newServer := &Server{}

	newServer.Port = *appSettings.WebPort
	newServer.serverIO = serverIO

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

	s.serverIO <- libBabou.WEB_SERVER_START
}

// Loads muxer from router.go from `app` package.
func (s *Server) loadRoutes() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Error loading routes: \n %s", r)

		}
	}()

	http.Handle("/", LoadRoutes())
}
