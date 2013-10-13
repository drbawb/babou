package app

import (
	admin "github.com/drbawb/babou/app/admin"

	controllers "github.com/drbawb/babou/app/controllers"
	filters "github.com/drbawb/babou/app/filters"
	web "github.com/drbawb/babou/lib/web"

	mux "github.com/gorilla/mux"
	log "log"
	http "net/http"
)

func LoadRoutes(s *Server) *mux.Router {
	r := mux.NewRouter()
	web.Router = r

	// Shorthand for controllers

	home := controllers.NewHomeController()
	login := controllers.NewLoginController()
	torrent := controllers.NewTorrentController()

	eventChain := filters.EventChain(s.AppBridge)

	// Handle admin routes
	adminPanel := r.PathPrefix("/admin").Subrouter()
	adminPanel, err := admin.LoadRoutes(adminPanel)
	if err != nil {
		log.Fatalf("Error loading sub-application: /admin, because: %s \n", err.Error())
	}

	// Shows public homepage, redirects to private site if valid session can be found.
	r.HandleFunc("/",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(home, "index")).
		Name("homeIndex")

	r.HandleFunc("/faq",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(home, "faq")).
		Name("aboutUs")

	// Displays a login form.
	r.HandleFunc("/login",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(login, "index")).
		Methods("GET").
		Name("loginIndex")

	r.HandleFunc("/login",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(login, "session")).
		Methods("POST").
		Name("loginSession")

	r.HandleFunc("/logout",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(login, "logout")).
		Methods("GET").
		Name("loginDelete")

	// Displays a registration form
	r.HandleFunc("/register",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(login, "new")).
		Methods("GET").
		Name("loginNew")

	// Handles a new user's registration request.
	r.HandleFunc("/register",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(login, "create")).
		Methods("POST").
		Name("loginCreate")

	// Handle torrent routes:
	r.HandleFunc("/torrents",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Chain(eventChain).
			Resolve(torrent, "index")).
		Methods("GET").
		Name("torrentIndex")

	// (Redirects to /torrents/tv/episodes)
	r.HandleFunc("/torrents/tv",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Chain(eventChain).
			Resolve(torrent, "latestEpisodes")).
		Methods("GET").
		Name("tvLatest")

	// Display recently modified episodes of television.
	r.HandleFunc("/torrents/tv/episodes",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Chain(eventChain).
			Resolve(torrent, "latestEpisodes")).
		Methods("GET").
		Name("tvEpisodes")

	// Display recently modified series of television
	r.HandleFunc("/torrents/tv/series",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Chain(eventChain).
			Resolve(torrent, "latestSeries")).
		Methods("GET").
		Name("tvSeries")

	r.HandleFunc("/torrents/new",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(torrent, "new")).
		Methods("GET").
		Name("torrentNew")

	r.HandleFunc("/torrents/create",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(torrent, "create")).
		Methods("POST").
		Name("torrentCreate")

	r.HandleFunc("/torrents/download/{torrentId}",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Resolve(torrent, "download")).
		Methods("GET").
		Name("torrentDownload")

	r.HandleFunc("/torrents/delete/{torrentId}",
		filters.BuildDefaultChain().
			Chain(filters.AuthChain(false)).
			Chain(eventChain).
			Resolve(torrent, "delete")).
		Methods("GET").
		Name("torrentDelete")

	// Catch-All: Displays all public assets.
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		web.DisableDirectoryListing(http.FileServer(http.Dir("assets/")))))

	return r
}
