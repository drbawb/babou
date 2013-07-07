package torrent

import (
	//bencode "code.google.com/p/bencode-go"
	io "io"

	"bytes"
	"crypto/sha1"

	fmt "fmt"
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
}

// tests reading a torrent file.
func ReadFile(filename string) {
	fmt.Printf("reading file...")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("cannot open file: %s", err.Error())
		return
	}

	torrent := &TorrentFile{}

	decoder := bencode.NewDecoder(file)
	err = decoder.Decode(torrent)
	if err != nil {
		fmt.Printf("error decoding torrent file: %s", err.Error())
	}

	fmt.Printf("info[] hash: %x \n", encodeInfo(torrent.Info))
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

}

// Encode's the `info` dictionary into a SHA1 hash; used to uniquely identify a torrent.
func encodeInfo(infoMap map[string]interface{}) []byte {
	//torrentDict := torrentMetainfo.(map[string]interface{})
	infoBytes := make([]byte, 0) //TODO: intelligenty size buffer based on info
	infoBuffer := bytes.NewBuffer(infoBytes)

	encoder := bencode.NewEncoder(infoBuffer)

	err := encoder.Encode(infoMap)
	if err != nil {
		fmt.Printf("error encoding torrent file: %s", err.Error())
	}

	hash := sha1.New()
	io.Copy(hash, infoBuffer)

	return hash.Sum(nil)
}
