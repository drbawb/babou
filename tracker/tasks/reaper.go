package tasks

import (
	"github.com/drbawb/babou/lib"
	"github.com/drbawb/babou/lib/torrent"

	"time"
)

type PeerReaper struct{}

// Loops over a torrent's peers looking for peers which have not announced
// since 2 * TRACKER_ANNOUNCE_INTERVAL.
//
// The underlying task is called asynchronously and will block
// access to the underlying list of peers while it runs.
// TODO: Task timeout.
func (pr *PeerReaper) ReapTorrent(target *torrent.Torrent) {
	reapSince := 2 * lib.TRACKER_ANNOUNCE_INTERVAL

	//TODO: limit num workers MAXPROCS or some other value?
	go pr.doWork(target, reapSince)
}

// Coroutine for reaping a single torrent.
func (pr *PeerReaper) doWork(target *torrent.Torrent, reapSince int) {
	peersToRemove := make([]string, 0)

	// Reap peers that were last seen (2 * ANN_INTERVAL) seconds before the reaper started.
	reapSinceSeconds := time.Duration(reapSince) * time.Second
	reapBefore := time.Now().Add(-reapSinceSeconds)

	// Linear scan of peer map; checks timestamps of peers.
	target.ReadPeers(func(peerMap map[string]*torrent.Peer) {
		for peerId, peer := range peerMap {
			if peer.LastSeen.Before(reapBefore) {
				peersToRemove = append(peersToRemove, peerId)
			}
		}
	})

	// Delete the peers that were marked inactive.
	target.WritePeers(func(peerMap map[string]*torrent.Peer) {
		for _, peerId := range peersToRemove {
			delete(peerMap, peerId)
		}
	})
}
