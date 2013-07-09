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
	os "os"

	bencode "github.com/zeebo/bencode"
)

// Performance tuning constants.
const (
	MIN_PEER_THRESHOLD = 30
	DEFAULT_NUMWANT    = 30
)

// Represents the torrent's metainfo structure (*.torrent file)
type Torrent struct {
	Info  *TorrentFile
	peers map[string]*Peer // List of peers for this torrent; as a map of their user-secrets.
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

// Reads a torrent-file from the filesystem.
// TODO: Model will create torrent-file; obsoleting this.
func ReadFile(filename string) *Torrent {
	fmt.Printf("reading file...")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("cannot open file: %s", err.Error())
		return nil
	}

	torrent := &Torrent{Info: &TorrentFile{}, peers: make(map[string]*Peer)}

	decoder := bencode.NewDecoder(file)
	err = decoder.Decode(torrent.Info)
	if err != nil {
		fmt.Printf("error decoding torrent file: %s", err.Error())
	}

	metainfo := torrent.Info
	torrent.Info.Info["private"] = 1

	fmt.Printf("info[] hash: %x \n", metainfo.EncodeInfo())
	fmt.Printf("# of pieces (hashes): %d \n", len(metainfo.Info["pieces"].(string))/20)
	if metainfo.Info["files"] != nil {
		fmt.Printf("--- \n multi-file mode \n---\n")
		fileList := metainfo.Info["files"].([]interface{})
		for _, file := range fileList {
			fileDict := file.(map[string]interface{})
			fmt.Printf("file name: %s \n", fileDict["path"])
			fmt.Printf("file length: %d (KiB) \n", fileDict["length"].(int64)/1024)
			fmt.Printf("   ---   \n")
		}

	} else if metainfo.Info["name"] != nil {
		fmt.Printf("--- \n single-file mode \n---\n")
		fmt.Printf("file name: %s \n", metainfo.Info["name"])
		fmt.Printf("file length: %d MiB", metainfo.Info["length"].(int64)/(1024*1024))
	} else {
		fmt.Printf("malformed torrent? \n")
	}

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

	infoBuffer := bytes.NewBuffer(make([]byte, 0))
	encoder := bencode.NewEncoder(infoBuffer)

	err := encoder.Encode(t)

	if err != nil {
		return nil, err
	}

	return infoBuffer.Bytes(), nil

}

// Updates the peer-list from an announce requeset.
func (t *Torrent) AddPeer(peerId, ipAddr, port, secret string) {
	peer := NewPeer(peerId, ipAddr, port, secret)

	if t.peers[peerId] == nil {
		// new peer
		t.peers[peerId] = peer
	} else {
		// we have seen this peer before.
		t.peers[peerId].UpdateLastSeen()
	}

	fmt.Printf("len peers map: %s", len(t.peers))
}

// Updates the in-memory statistics for a peer being tracked for this torrent.
// Returns an error if the peer is not found or the request cannot be fulfilled.
// The stats-collector job will ensure they get written to disk.
func (t *Torrent) UpdateStatsFor(peerId string, uploaded, downloaded, left string) error {
	if t.peers[peerId] == nil {
		return errors.New("Peer w/ ID[%s] not found on this torrent.")
	}

	if err := t.peers[peerId].UpdateStats(uploaded, downloaded, left); err != nil {
		return err
	}

	return nil
}

// Returns the seeders followed by the leechers for this torrent.
func (t *Torrent) EnumeratePeers() (int, int) {
	seeding := 0
	leeching := 0

	for _, val := range t.peers {
		switch {
		case val.Status == 0 || val.Status == LEECHING:
			leeching += 1
		case val.Status == SEEDING:
			seeding += 1
		}
	}

	return seeding, leeching
}

// Send numWant -1 for "no peers requested", 0 for default, and n if client wants more peers.
// Returns a ranked peerlist that attempts to maintain a balanced ratio of seeders:leechers.
func (t *Torrent) GetPeerList(numWant int) string {
	//tempList := make([]*Peer, 0, len(t.peers))

	if numWant == -1 {
		return "" //peer _specifically requested_ we do not send more peers via numwant => 0
	} else if numWant == 0 {
		numWant = DEFAULT_NUMWANT
	}

	outBuf := bytes.NewBuffer(make([]byte, 0))
	// send them everything we got; torrent is just starting off.
	if len(t.peers) < MIN_PEER_THRESHOLD && len(t.peers) < numWant {
		for _, val := range t.peers {
			ip := val.IPAddr.To4()

			binary.Write(outBuf, binary.BigEndian, ip)
			binary.Write(outBuf, binary.BigEndian, val.Port)
		}
	} else if len(t.peers) < MIN_PEER_THRESHOLD && len(t.peers) > numWant {
		i := 0
		for _, val := range t.peers {
			if i > numWant {
				break
			}

			ip := val.IPAddr
			binary.Write(outBuf, binary.BigEndian, ip)
			binary.Write(outBuf, binary.BigEndian, val.Port)

			i++
		}
	}

	return string(outBuf.Bytes())
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
