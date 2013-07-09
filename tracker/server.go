// The `babou` high-performance BitTorrent tracker.
//
// Implements the core tracker-request router
package tracker

import (
	libBabou "github.com/drbawb/babou/lib"
	libTorrent "github.com/drbawb/babou/lib/torrent"
	libWeb "github.com/drbawb/babou/lib/web"

	"github.com/drbawb/babou/app/models"

	mux "github.com/gorilla/mux"
	bencode "github.com/zeebo/bencode"

	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// Parameters for babou's web server
type Server struct {
	Port         int
	serverIO     chan int
	torrentCache map[string]*libTorrent.Torrent
}

// Initializes a server using babou/lib settings and a communication channel.
func NewServer(appSettings *libBabou.AppSettings, serverIO chan int) *Server {
	newServer := &Server{torrentCache: make(map[string]*libTorrent.Torrent)}

	newServer.Port = *appSettings.TrackerPort
	newServer.serverIO = serverIO

	return newServer
}

func (s *Server) Start() {
	router := LoadRoutes(s)

	go func() {
		// start with custom muxer.
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), router))
	}()
}

func wrapAnnounceHandle(s *Server) http.HandlerFunc {
	torrentData := ReadFile("/home/drbawb/Downloads/[FFF] Hataraku Maou-sama! - 13 [5467C06D].mkv.torrent")
	s.torrentCache[string(torrentData.Info.EncodeInfo())] = torrentData

	fn := func(w http.ResponseWriter, r *http.Request) {
		announceHandle(w, r, s)
	}

	return fn
}

func announceHandle(w http.ResponseWriter, r *http.Request, s *Server) {
	params := libWeb.RetrieveAllParams(r)
	responseMap := make(map[string]interface{})

	fmt.Printf("request from (ip): %s:%s \n, has left(bytes): %s, compact: %s \n --- \n",
		params["ip"], params["port"], params["left"], params["compact"])

	routeVars := mux.Vars(r)

	torrent, ok := s.torrentExists(params["info_hash"])
	userOk := false

	user := &models.User{}
	secretHex, err := hex.DecodeString(routeVars["secret"])
	hashHex, err := hex.DecodeString(routeVars["hash"])

	if err != nil {
		fmt.Printf("error selecting user with secret: %s \n", err.Error())
		userOk = false
	} else {
		err = user.SelectSecret(secretHex)
	}

	if err != nil {
		userOk = false
	} else {
		userOk = checkHmac(user.Secret, hashHex)
	}

	fmt.Printf("request from tracker-id: %s OR %s \n", params["trackerid"], params["tracker id"])

	if ok && userOk {
		//responseMap["failure reason"] = "tracker found torrent but doesnt know what to do now!"
		torrent.AddPeer(params["peer_id"], r.RemoteAddr, params["port"])

		responseMap["interval"] = 300
		responseMap["min interval"] = 10

		responseMap["tracker id"] = randomId()

		responseMap["complete"] = torrent.NumLeech()
		responseMap["incomplete"] = torrent.NumLeech()

		responseMap["peers"] = torrent.GetPeerList()

	} else if !ok {
		responseMap["failure reason"] = "tracker could not find requested torrent"
	} else {
		responseMap["failure reason"] = "invalid user key!"
	}

	responseBuf := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(responseBuf)

	err = encoder.Encode(responseMap)
	if err != nil {
		w.Write([]byte("server error"))
	}

	io.Copy(w, responseBuf)
	w.Header().Set("Content-Type", "text/plain")
}

func checkHmac(secret []byte, hash []byte) bool {
	//TODO: taken from users model. should be config param.
	sharedKey := []byte("f75778f7425be4db0369d09af37a6c2b9ab3dea0e53e7bd57412e4b060e607f7")

	mac := hmac.New(sha256.New, sharedKey)
	mac.Write(secret)

	outVal := hmac.Equal(hash, mac.Sum(nil))
	fmt.Printf("checked hmac, return: %v\n", outVal)
	fmt.Printf("secret: %x \nhash: %x\n", secret, hash)

	return outVal
}

func (s *Server) torrentExists(infoHash string) (*libTorrent.Torrent, bool) {
	torrent := s.torrentCache[infoHash]

	if torrent == nil {
		return nil, false
	} else {
		return torrent, true
	}
}

func randomId() string {
	randBytes := make([]byte, 32)
	_, _ = rand.Read(randBytes)

	hex := hex.EncodeToString(randBytes)

	return hex
}

// Test method for loading torrents.
func ReadFile(filename string) *libTorrent.Torrent {
	return libTorrent.ReadFile(filename)
}
