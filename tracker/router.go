package tracker

import (
	"github.com/gorilla/mux"
)

func LoadRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/announce", announceHandle)

	return r
}
