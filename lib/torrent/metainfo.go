package torrent

import (
	//bencode "code.google.com/p/bencode-go"
	io "io"

	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"strconv"

	fmt "fmt"
	net "net"
	os "os"

	bencode "github.com/zeebo/bencode"
)

// Represents the torrent's metainfo structure (*.torrent file)
type Torrent struct {
	Info  *TorrentFile
	peers map[string]*Peer
}

type TorrentFile struct {
	Announce     string                 `bencode:"announce"`
	Comment      string                 `bencode:"comment"`
	CreatedBy    string                 `bencode:"created by"`
	CreationDate int64                  `bencode:"creation date"`
	Encoding     string                 `bencode:"encoding"`
	Info         map[string]interface{} `bencode:"info"`
}

type Peer struct {
	ID   string `bencode:"peer id"`
	IP   string `bencode:"ip"`
	Port uint16 `bencode:"port"`
}

// tests reading a torrent file.
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

// Adds a peer or seed for this torrent
func (t *Torrent) AddPeer(peerId, ipAddr, port string) {
	host, _, _ := net.SplitHostPort(ipAddr)
	fmt.Printf("host added: %s, peerId: %s \n", host, peerId)

	portNum, _ := strconv.Atoi(port)

	newPeer := &Peer{ID: peerId, IP: host, Port: uint16(portNum)}
	if t.peers[peerId] == nil && host != "" {
		t.peers[peerId] = newPeer
	}

	fmt.Printf("len peers map: %s", len(t.peers))
}

func (t *Torrent) NumLeech() int {
	return len(t.peers)
}

func (t *Torrent) GetPeerList() string {
	//tempList := make([]*Peer, 0, len(t.peers))

	outBuf := bytes.NewBuffer(make([]byte, 0))

	for _, val := range t.peers {

		if !(val.Port == 0) && !(val.IP == "") && !(val.ID == "") {
			ip := net.ParseIP(val.IP)
			ip = ip.To4()
			binary.Write(outBuf, binary.BigEndian, ip)
			binary.Write(outBuf, binary.BigEndian, val.Port)
		}

	}

	fmt.Printf("peers field: %v \n", outBuf.Bytes())

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
