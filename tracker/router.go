package tracker

import (
	"github.com/gorilla/mux"
)

func LoadRoutes(s *Server) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/{secret}/{hash}/announce", wrapAnnounceHandle(s))

	return r
}
