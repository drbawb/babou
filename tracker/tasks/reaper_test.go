package tasks

import (
	"testing"
	"time"

	"github.com/drbawb/babou/lib/torrent"
)

const (
	STALE_TIME_IN_SECONDS = 10
)

// TODO: DRY. (Repeated in announce_test.go)
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

// Creates a mock torrent and adds three peers.
// The peers are named mock-1 through mock-3
//
// mock-1 is stale according to `STALE_TIME_IN_SECONDS`
// mock-2 and mock-3 remain fresh for at least `STALE_TIME_IN_SECONDS`
func setupReaperTest() *torrent.Torrent {
	t := MockTorrent()
	t.AddPeer("mock-1", "[::1]:1337", "1337", "abcadefgawalthgrathorp")
	t.AddPeer("mock-2", "[::1]:1337", "1337", "abcadefgawalthgrathorp")
	t.AddPeer("mock-3", "[::1]:1337", "1337", "abcadefgawalthgrathorp")

	// Establish our test clocks
	staleDuration := time.Duration(STALE_TIME_IN_SECONDS) * time.Second
	t.WritePeers(func(peerMap map[string]*torrent.Peer) {
		for _, peer := range peerMap {
			peer.UpdateLastSeen() // current clock
		}
		// ensure that one peer is twice as stale as it needs to be.
		peerMap["mock-1"].LastSeen = time.Now().Add(-(staleDuration * 2))
	})

	return t
}

// Tests that the reaper's underlying task does not remove peers
// that are currently younger than `STALE_TIME_IN_SECONDS`.
func TestReaperWorkerLeavesFresh(test *testing.T) {
	// Create a torrent and add some peers.
	t := setupReaperTest()

	// Reap in specified interval.
	pr := &PeerReaper{}
	pr.doWork(t, STALE_TIME_IN_SECONDS)

	// There should be no write contention on the peer map
	// Read should immediately be false for "mock-1"
	t.ReadPeers(func(peerMap map[string]*torrent.Peer) {
		if peerMap["mock-2"] == nil || peerMap["mock-3"] == nil {
			test.Fail()
		}
	})
}

// Tests that the reaper's underlying task removes stale
// peers appropriately.
func TestReaperWorkerRemovesStale(test *testing.T) {
	// Create a torrent and add some peers.
	t := setupReaperTest()

	// Reap in specified interval.
	pr := &PeerReaper{}
	pr.doWork(t, STALE_TIME_IN_SECONDS)

	// There should be no write contention on the peer map
	// Read should immediately be false for "mock-1"
	t.ReadPeers(func(peerMap map[string]*torrent.Peer) {
		if peerMap["mock-1"] != nil {
			test.Fail()
		}
	})
}
