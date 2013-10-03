package torrent

import (
	//bencode "code.google.com/p/bencode-go"
	io "io"

	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"

	"errors"
	fmt "fmt"
	"mime/multipart"

	bencode "github.com/zeebo/bencode"
)

// Performance tuning constants.
const (
	MIN_PEER_THRESHOLD = 30
	DEFAULT_NUMWANT    = 30
)

// Represents a torrent being actively used by the tracker.
// A torrent includes a [decoded] copy of the metainfo file
// as well as a list of active peers that is periodically culled.
type Torrent struct {
	Info  *TorrentFile
	peers *PeerMap
}

// Represents a `babou` torrent.
// We don't care about other fields so they will be discarded from uploaded torrents.
type TorrentFile struct {
	Announce     string                 `bencode:"announce"`
	Comment      string                 `bencode:"comment"`
	CreatedBy    string                 `bencode:"created by"`
	CreationDate int64                  `bencode:"creation date"`
	Encoding     string                 `bencode:"encoding"`
	Info         map[string]interface{} `bencode:"info"`
}

// Writes a new torrent to be used by the tracker for maintaining peer lists.
func NewTorrent(file *TorrentFile) *Torrent {
	out := &Torrent{peers: NewPeerMap()}
	out.Info = file

	return out
}

// Reads a torrent-file from the filesystem.
// TODO: Model will create torrent-file; obsoleting this.
func ReadFile(file multipart.File) *Torrent {
	fmt.Printf("reading file...")

	torrent := &Torrent{Info: &TorrentFile{}, peers: NewPeerMap()}

	decoder := bencode.NewDecoder(file)
	err := decoder.Decode(torrent.Info)
	if err != nil {
		fmt.Printf("error decoding torrent file: %s", err.Error())
	}

	metainfo := torrent.Info
	torrent.Info.Info["private"] = 1

	fmt.Printf("info[] hash: %x \n", metainfo.EncodeInfo())
	fmt.Printf("# of pieces (hashes): %d \n", len(metainfo.Info["pieces"].(string))/20)

	return torrent
}

// Converts torrent to SUPRA-PRIVATE torrent
//
// Sets the private flag to 1 and embeds the supplied secret and hash
// for authentication purposes.
//
// This torrent file SHOULD NOT be shared between users or statistics collection
// and anti-abuse mechanisms will be skewed for that user.
func (t *TorrentFile) WriteFile(secret, hash []byte) ([]byte, error) {
	fmt.Printf("writing file...")

	t.Announce = fmt.Sprintf("http://tracker.fatalsyntax.com:4200/%s/%s/announce", hex.EncodeToString(secret), hex.EncodeToString(hash))
	t.Encoding = "UTF-8"
	infoBuffer := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(infoBuffer)

	err := encoder.Encode(t)

	if err != nil {
		return nil, err
	}

	return infoBuffer.Bytes(), nil

}

// Takes a closure which will obtain a readlock
// for this torrent's peer-map.
//
// Reading from the torrent's peer map is safe from within the closure.
func (t *Torrent) ReadPeers(closure func(map[string]*Peer)) {
	t.peers.Sync().RLock()
	defer t.peers.Sync().RUnlock()

	closure(t.peers.Map())
}

// Takes a closure which will obtain a writelock for this
// torrents' peer-map.
//
// Updates to the map are safe within the context of the closure.
func (t *Torrent) WritePeers(closure func(map[string]*Peer)) {
	t.peers.Sync().Lock()
	defer t.peers.Sync().Unlock()

	closure(t.peers.Map())
}

// Updates the peer-list from an announce requeset.
func (t *Torrent) AddPeer(peerId, ipAddr, port, secret string) {
	// Will either add or update a peer; obtain write lock.
	fn := func(peerList map[string]*Peer) {
		if peerList[peerId] == nil {
			// new peer
			peer := NewPeer(peerId, ipAddr, port, secret)
			peerList[peerId] = peer
		} else {
			// we have seen this peer before.
			peerList[peerId].UpdateLastSeen()
		}
	}

	t.WritePeers(fn)
}

// Updates the in-memory statistics for a peer being tracked for this torrent.
// Returns an error if the peer is not found or the request cannot be fulfilled.
// The stats-collector job will ensure they get written to disk.
func (t *Torrent) UpdateStatsFor(peerId string, uploaded, downloaded, left string) error {
	var status error

	// Will update contents of map so long as peer is found
	fn := func(peerList map[string]*Peer) {
		if peerList[peerId] == nil {
			status = errors.New("Peer w/ ID[%s] not found on this torrent.")
		}

		if err := peerList[peerId].UpdateStats(uploaded, downloaded, left); err != nil {
			status = err
		}
	}

	t.WritePeers(fn)

	return status
}

// Returns the seeders followed by the leechers for this torrent.
func (t *Torrent) EnumeratePeers() (int, int) {
	// Reads number of peers from the map.

	seeding := 0
	leeching := 0

	fn := func(peerList map[string]*Peer) {
		for _, val := range peerList {
			switch {
			case val.Status == 0 || val.Status == LEECHING:
				leeching += 1
			case val.Status == SEEDING:
				seeding += 1
			}
		}
	}

	t.ReadPeers(fn)

	return seeding, leeching
}

// Send numWant -1 for "no peers requested", 0 for default, and n if client wants more peers.
// Returns a ranked peerlist that attempts to maintain a balanced ratio of seeders:leechers.
// First return val is compact-form `peers` dict, second is `peers6` dict.
func (t *Torrent) GetPeerList(numWant int) (string, string) {
	if numWant == -1 {
		return "", "" //peer _specifically requested_ we do not send more peers via numwant => 0
	} else if numWant == 0 {
		numWant = DEFAULT_NUMWANT
	}

	outBuf := bytes.NewBuffer(make([]byte, 0))  //peers buffer
	outBuf6 := bytes.NewBuffer(make([]byte, 0)) //peers6 buffer

	fn := func(peerList map[string]*Peer) {
		// send them everything we got; torrent is just starting off.
		mapLength := len(peerList)

		if mapLength < MIN_PEER_THRESHOLD && mapLength < numWant {
			for _, val := range peerList {

				// Do not add the peer to the IPv4 list if we could
				// not parse a valid IP address for them.
				if ip := val.IPAddr.To4(); ip != nil {
					binary.Write(outBuf, binary.BigEndian, ip)
					binary.Write(outBuf, binary.BigEndian, val.Port)
				} else if ip := val.IPAddr.To16(); ip != nil {
					binary.Write(outBuf6, binary.BigEndian, ip)
					binary.Write(outBuf6, binary.BigEndian, val.Port)
				}
			}
		} else if mapLength < MIN_PEER_THRESHOLD && mapLength > numWant {
			i := 0
			for _, val := range peerList {
				if i >= numWant {
					break
				}

				if ip := val.IPAddr.To4(); ip != nil {
					binary.Write(outBuf, binary.BigEndian, ip)
					binary.Write(outBuf, binary.BigEndian, val.Port)
				} else if ip := val.IPAddr.To16(); ip != nil {
					binary.Write(outBuf6, binary.BigEndian, ip)
					binary.Write(outBuf6, binary.BigEndian, val.Port)
				}

				i++
			}
		} else if mapLength > numWant {
			i := 0
			for _, val := range peerList {
				if i >= numWant {
					break
				}

				if ip := val.IPAddr.To4(); ip != nil {
					binary.Write(outBuf, binary.BigEndian, ip)
					binary.Write(outBuf, binary.BigEndian, val.Port)
				} else if ip := val.IPAddr.To16(); ip != nil {
					binary.Write(outBuf6, binary.BigEndian, ip)
					binary.Write(outBuf6, binary.BigEndian, val.Port)
				}

				i++
			}
		} else {
			for _, val := range peerList {
				if ip := val.IPAddr.To4(); ip != nil {
					binary.Write(outBuf, binary.BigEndian, ip)
					binary.Write(outBuf, binary.BigEndian, val.Port)
				} else if ip := val.IPAddr.To16(); ip != nil {
					binary.Write(outBuf6, binary.BigEndian, ip)
					binary.Write(outBuf6, binary.BigEndian, val.Port)
				}
			}
		}
	}

	t.ReadPeers(fn)

	return string(outBuf.Bytes()), string(outBuf6.Bytes())
}

// Encode's the `info` dictionary into a SHA1 hash; used to uniquely identify a torrent.
func (t *TorrentFile) EncodeInfo() []byte {
	//torrentDict := torrentMetainfo.(map[string]interface{})
	infoBytes := make([]byte, 0) //TODO: intelligenty size buffer based on info
	infoBuffer := bytes.NewBuffer(infoBytes)

	encoder := bencode.NewEncoder(infoBuffer)

	err := encoder.Encode(t.Info)
	if err != nil {
		fmt.Printf("error encoding torrent file: %s", err.Error())
	}

	hash := sha1.New()
	io.Copy(hash, infoBuffer)

	return hash.Sum(nil)
}

// Encode's the `info` dictionary into a SHA1 hash; used to uniquely identify a torrent.
// Converts hash to a string [two chars per byte]
func (t *TorrentFile) EncodeInfoToString() string {
	//torrentDict := torrentMetainfo.(map[string]interface{})
	infoBytes := make([]byte, 0) //TODO: intelligenty size buffer based on info

	encoderBuffer := bytes.NewBuffer(infoBytes)
	encoder := bencode.NewEncoder(encoderBuffer)

	err := encoder.Encode(t.Info)
	if err != nil {
		fmt.Printf("error encoding torrent file: %s", err.Error())
	}

	hash := sha1.New()
	io.Copy(hash, encoderBuffer)

	b := hash.Sum(nil)

	return hex.EncodeToString(b)
}

// Returns a bencoded version of the torrent's info dict.
func (t *TorrentFile) BencodeInfoDict() ([]byte, error) {
	infoBuffer := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(infoBuffer)

	err := encoder.Encode(t.Info)
	if err != nil {
		fmt.Printf("error encoding torrent file: %s", err.Error())
	}

	return infoBuffer.Bytes(), err
}

func DecodeInfoDict(bencodedInfo []byte) (map[string]interface{}, error) {
	decodedMap := make(map[string]interface{})

	bencodedBuffer := bytes.NewBuffer(bencodedInfo)
	dec := bencode.NewDecoder(bencodedBuffer)

	if err := dec.Decode(&decodedMap); err != nil {
		return nil, err
	}

	return decodedMap, nil
}
