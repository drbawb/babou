package torrent

import (
	//bencode "code.google.com/p/bencode-go"
	io "io"

	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"strconv"

	fmt "fmt"
	net "net"
	os "os"

	bencode "github.com/zeebo/bencode"
)

// Represents the torrent's metainfo structure (*.torrent file)
type TorrentFile struct {
	Announce     string                 "announce"
	AnnounceList []interface{}          "announce-list"
	Comment      string                 "comment"
	CreatedBy    string                 "created by"
	CreationDate int64                  "creation date"
	Encoding     string                 "encoding"
	Info         map[string]interface{} "info"
	Private      int64                  "private"

	peers []*Peer
}

type Peer struct {
	ID   string `bencode:"peer id"`
	IP   string `bencode:"ip"`
	Port string `bencode:"port"`
}

// tests reading a torrent file.
func ReadFile(filename string) *TorrentFile {
	fmt.Printf("reading file...")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("cannot open file: %s", err.Error())
		return nil
	}

	torrent := &TorrentFile{peers: make([]*Peer, 0)}

	decoder := bencode.NewDecoder(file)
	err = decoder.Decode(torrent)
	if err != nil {
		fmt.Printf("error decoding torrent file: %s", err.Error())
	}

	fmt.Printf("info[] hash: %x \n", torrent.EncodeInfo())
	fmt.Printf("# of pieces (hashes): %d \n", len(torrent.Info["pieces"].(string))/20)
	if torrent.Info["files"] != nil {
		fmt.Printf("--- \n multi-file mode \n---\n")
		fileList := torrent.Info["files"].([]interface{})
		for _, file := range fileList {
			fileDict := file.(map[string]interface{})
			fmt.Printf("file name: %s \n", fileDict["path"])
			fmt.Printf("file length: %d (KiB) \n", fileDict["length"].(int64)/1024)
			fmt.Printf("   ---   \n")
		}

	} else if torrent.Info["name"] != nil {
		fmt.Printf("--- \n single-file mode \n---\n")
		fmt.Printf("file name: %s \n", torrent.Info["name"])
		fmt.Printf("file length: %d MiB", torrent.Info["length"].(int64)/(1024*1024))
	} else {
		fmt.Printf("malformed torrent? \n")
	}

	return torrent

}

// Adds a peer or seed for this torrent
func (t *TorrentFile) AddPeer(peerId, ipAddr, port string) {
	host, _, _ := net.SplitHostPort(ipAddr)
	fmt.Printf("host added: %s \n", host)

	newPeer := &Peer{ID: peerId, IP: host, Port: port}
	t.peers = append(t.peers, newPeer)
}

func (t *TorrentFile) GetPeerList() string {
	outStrBytes := make([]byte, 6)
	outStr := bytes.NewBuffer(outStrBytes)

	for _, val := range t.peers {
		ip := net.ParseIP(val.IP)
		n, _ := outStr.Write(ip.To4())
		fmt.Printf("wrote %d bytes for ip \n", n)

		portNum, _ := strconv.Atoi(val.Port)

		n = binary.PutVarint(outStrBytes, int64(portNum))
		fmt.Printf("wrote %d bytes for port \n", n)

		fmt.Printf("len peer: %d \n", len(outStrBytes))
	}

	/*peerBuffer := bytes.NewBuffer(make([]byte, 0))

	encoder := bencode.NewEncoder(peerBuffer)
	err := encoder.Encode(t.peers)

	if err != nil {
		fmt.Printf("err: %s encoding peer list \n", err.Error())
	}*/

	fmt.Printf("peer list: %v \n --- \n", string(outStr.Bytes()))

	return string(outStrBytes)
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
