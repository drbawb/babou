package web

import (
	libBabou "babou/lib"
	log "log"
)

type Server struct {
	Port int
}

func (s *Server) Start(appSettings *libBabou.AppSettings) {
	log.Println("hello, world.")
}
