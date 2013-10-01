package torrent

import (
	"crypto/hmac"
	"crypto/sha256"

	"encoding/hex"
	"errors"
	"net"
	"strconv"

	"sync"
	"time"
)

type PeerStatus int64

const (
	_                   = iota
	SEEDING  PeerStatus = iota // Announced a complete copy of the file.
	LEECHING            = iota // Announced an incomplete copy of the file.
	QUASI               = iota // Announced an incomplete copy; but does not appear to be downloading.
)

// A protected mapping of peerIds to peers.
//
// This map is safe for multiple readers; but read
// and write calls will block while a writer holds the lock.
type PeerMap struct {
	peerMap map[string]*Peer
	rwLock  *sync.RWMutex
}

// A peer being tracked on a torrent
type Peer struct {
	ID     string
	IPAddr net.IP
	Port   uint16

	Status          PeerStatus // what is the peer doing?
	DownloadedBytes int64      // curently completed bytes
	UploadedBytes   int64      // currently uploaded bytes
	LeftBytes       int64      // number of bytes remaining

	LastSeen          time.Time // last announce at:
	LastCompleteBytes int64     // last completed bytes

	Secret string // uniquely identifies a peer
}

// Simpler data structured used for non-compact peer lists.
type BenPeer struct {
	ID   string `bencode:"peer id"`
	IP   string `bencode:"ip"`
	Port int64  `bencode:"port"`
}

// Initalizes a new peer map for a torrent
func NewPeerMap() *PeerMap {
	outMap := &PeerMap{rwLock: &sync.RWMutex{}, peerMap: make(map[string]*Peer)}

	return outMap
}

// Returns the map.
func (pm *PeerMap) Map() map[string]*Peer {
	return pm.peerMap
}

// Returns the syncrhonization primitive that can be used to
// protect concurrent access to this map.
//
// This should be used wherever contention of the map is likely.
func (pm *PeerMap) Sync() *sync.RWMutex {
	return pm.rwLock
}

// Creates a new peer object.
func NewPeer(peerId, ipAddr, port, secret string) *Peer {
	newPeer := &Peer{}

	newPeer.ID = peerId
	newPeer.Secret = secret

	host, _, _ := net.SplitHostPort(ipAddr)
	ip := net.ParseIP(host)
	newPeer.IPAddr = ip

	portNum, _ := strconv.Atoi(port)
	newPeer.Port = uint16(portNum)

	newPeer.LastSeen = time.Now()

	return newPeer
}

// Returns a bencodable version of this peer.
func (p *Peer) EncodePeer() *BenPeer {
	outPeer := &BenPeer{}

	outPeer.ID = p.ID
	outPeer.IP = p.IPAddr.String()
	outPeer.Port = int64(p.Port)

	return outPeer
}

// Updates timestamp to current server time.
func (p *Peer) UpdateLastSeen() {
	p.LastSeen = time.Now()
}

// Updates download statistics and promotes the peer if necessary.
// Strings should be ASCII encoded base 10 numbers. (Per bep-003)
func (p *Peer) UpdateStats(uploaded, downloaded, left string) error {
	//uses int64 and checks for obvious [negative] overflow.
	//overflowing an int64 indicates _incredibly_ large torrents; on the order of 8*10e5 TiB!!!
	uploadedInt, err := strconv.ParseInt(uploaded, 10, 64)
	downloadedInt, err := strconv.ParseInt(downloaded, 10, 64)
	leftInt, err := strconv.ParseInt(left, 10, 64)

	if err != nil {
		return err
	}

	if uploadedInt < 0 || downloadedInt < 0 || leftInt < 0 {
		return errors.New("Statistics failed sanity check. They have been ignored.")
	}

	p.UploadedBytes = uploadedInt
	p.LeftBytes = leftInt

	if p.LastCompleteBytes == 0 {
		// store
		p.DownloadedBytes = downloadedInt
		p.LastCompleteBytes = downloadedInt
	} else {
		//swap and store TODO: (double check this works in order i think it does!)
		p.LastCompleteBytes, p.DownloadedBytes = p.DownloadedBytes, downloadedInt
	}

	if p.LeftBytes == 0 {
		p.Status = SEEDING
	} else {
		p.Status = LEECHING
	}

	return nil
}

// Compares a message (the secret) and an HMAC of that message
// against a secret-key.
//
// This ensures that the .torrent file originated from this tracker and
// is linked to an active user account.
//
//  The secret and hash are []byte arrays and stored as such.
//  They must be deserialized
func CheckHmac(secret, hash string) bool {
	//TODO: taken from users model. should be config param.
	sharedKey := []byte("f75778f7425be4db0369d09af37a6c2b9ab3dea0e53e7bd57412e4b060e607f7")
	secretBytes, err := hex.DecodeString(secret)
	hashBytes, err := hex.DecodeString(hash)

	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, sharedKey)
	mac.Write(secretBytes)

	outVal := hmac.Equal(hashBytes, mac.Sum(nil))

	return outVal
}
