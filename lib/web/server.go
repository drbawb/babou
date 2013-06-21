package web

import (
	libBabou "babou/lib"
	log "log"
)

type Server struct {
	Port int
}

func (s *Server) Start(appSettings *libBabou.AppSettings, serverIO chan int) {
	log.Println("hello, world.")
	serverIO <- libBabou.WEB_SERVER_START
}
