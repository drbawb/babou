package tracker

import (
	"github.com/drbawb/babou/lib/torrent"

	"testing"

	"fmt"
	"time"

	"crypto/rand"
)

// Creates a fake torrent we can use for testing purposes.
func MockTorrent() *torrent.Torrent {
	file := &torrent.TorrentFile{}
	t := torrent.NewTorrent(file)

	file.Announce = "http://localhost:4200/null/null/announce"
	file.Comment = "created by babou test suite"
	file.CreatedBy = "babou v0.0"
	file.CreationDate = time.Now().Unix()
	file.Encoding = "UTF-8"

	file.Info = make(map[string]interface{})

	file.Info["name"] = "fff.mkv"
	file.Info["length"] = 1024

	file.Info["piece length"] = 1024
	file.Info["pieces"] = "01234567890123456789"

	return t
}

// Checks the number of peers returned for a torrent w/ peers > DEFAULT_NUMWANT
func TestNumWantLarge(test *testing.T) {
	torrentFile := MockTorrent()

	// create fifty "peers"
	randomPeerIds := make([][]byte, torrent.DEFAULT_NUMWANT+10)
	for i := 0; i < len(randomPeerIds); i++ {
		randomPeerIds[i] = make([]byte, 20)
		rand.Read(randomPeerIds[i])

		torrentFile.AddPeer(string(randomPeerIds[i]), "[127.0.0.1]:1337", "1337", "abcadefgawalthgrathorp")
	}

	// want no peers
	noPeers := torrentFile.GetPeerList(-1)
	if len(noPeers) > 0 {
		test.Errorf("Requested no peers, and received: %s \n", noPeers)
		test.FailNow()
	}

	// want some subset of peers
	specific := torrentFile.GetPeerList(torrent.DEFAULT_NUMWANT / 2)
	if (len(specific) / 6) != torrent.DEFAULT_NUMWANT/2 {
		test.Errorf("Request %d peers, and received %d \n", (torrent.DEFAULT_NUMWANT / 2),
			len(specific))
		test.FailNow()
	}

	// want default num of peers
	defaultPeers := torrentFile.GetPeerList(0)
	if (len(defaultPeers) / 6) != torrent.DEFAULT_NUMWANT {
		test.Errorf("Request default peers, received: %d \n",
			(len(defaultPeers) / 6))
	}

}

// Checks the number of peers returned for a torrent w/ peers < DEFAULT_NUMWANT
func TestNumWantSmall(test *testing.T) {
	torrentFile := MockTorrent()

	// create fifty "peers"
	randomPeerIds := make([][]byte, torrent.DEFAULT_NUMWANT/2)
	for i := 0; i < len(randomPeerIds); i++ {
		randomPeerIds[i] = make([]byte, 20)
		rand.Read(randomPeerIds[i])

		torrentFile.AddPeer(string(randomPeerIds[i]), "[127.0.0.1]:1337", "1337", "abcadefgawalthgrathorp")
	}

	// want no peers
	noPeers := torrentFile.GetPeerList(-1)
	if len(noPeers) > 0 {
		test.Errorf("Requested no peers, and received: %s \n", noPeers)
		test.FailNow()
	}

	// want some subset of peers [should return requested number]
	specific := torrentFile.GetPeerList(torrent.DEFAULT_NUMWANT / 3)
	if (len(specific) / 6) != torrent.DEFAULT_NUMWANT/3 {
		test.Errorf("Request %d peers, and received %d \n", (torrent.DEFAULT_NUMWANT / 3),
			len(specific))
		test.FailNow()
	}

	// want default num of peers [should return all available peers]
	defaultPeers := torrentFile.GetPeerList(0)
	if (len(defaultPeers) / 6) != torrent.DEFAULT_NUMWANT/2 {
		test.Errorf("Request default peers, received: %d \n",
			(len(defaultPeers) / 6))
	}

}

func TestPeerEnumeration(test *testing.T) {
	torrent := MockTorrent()

	// create four "peers"
	randomPeerIds := make([][]byte, 4)
	for i := 0; i < len(randomPeerIds); i++ {
		randomPeerIds[i] = make([]byte, 20)
		rand.Read(randomPeerIds[i])

		torrent.AddPeer(string(randomPeerIds[i]), "127.0.0.1", "1337", "abcadefgawalthgrathorp")
	}

	// enumerate leechers [4]
	seeders, leechers := torrent.EnumeratePeers()
	if seeders != 0 || leechers != 4 {
		test.Error("Torrent should have four [just added] leechers!")
		test.FailNow()
	}

	// promote two peers
	torrent.UpdateStatsFor(string(randomPeerIds[0]), "0", "1024", "0")
	torrent.UpdateStatsFor(string(randomPeerIds[1]), "0", "1024", "0")

	seeders, leechers = torrent.EnumeratePeers()
	if seeders != 2 || leechers != 2 {
		test.Error("Torrent should have two [just promoted] seeders and two leechers!")
		test.FailNow()
	}
}

// Various unit benchmarks

// Benchmarks how quickly the tracker's `bencoder` can serialize a map.
func BenchmarkResponseMapBencoder(bench *testing.B) {
	bench.StopTimer()
	torrent := MockTorrent()
	responseMap := make(map[string]interface{})

	responseMap["interval"] = 300
	responseMap["min interval"] = 10

	seeding, leeching := torrent.EnumeratePeers()
	responseMap["complete"] = seeding
	responseMap["incomplete"] = leeching

	bench.StartTimer()
	for i := 0; i < bench.N; i++ {
		encodeResponseMap(responseMap)
	}
}
