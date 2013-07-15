// The `babou` web application server.
//
// The application package includes both a server implementation and a router.
// The web application itself is included in various sub-packages, and is loaded
// by the router once the server has started.
package app

import (
	libBabou "github.com/drbawb/babou/lib"
	libDb "github.com/drbawb/babou/lib/db"

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

	newServer.Port = appSettings.WebPort
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
		s.openDb()
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

	http.Handle("/", LoadRoutes())
}

// Instructs the babou library to open a database connection.
// This DB connection will be closed when babou is gracefully shutdown.
func (s *Server) openDb() {
	_, err := libDb.Open()
	if err != nil {
		panic("database could not be opened: " + err.Error())
	}
}
