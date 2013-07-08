package tracker

import (
	"github.com/gorilla/mux"
)

func LoadRoutes(s *Server) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/announce", wrapAnnounceHandle(s))

	return r
}
