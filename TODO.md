babou project TODO list
===

This is a general list of features that are actively being worked on. 

This list _does not_ include bug-fixes, see the github.com/drbawb/babou issue tracker
for a list of issues and their status.  

This list is not in any sort of order and only serves as a very rough measure of progress.

---

General

---

* Multi-node stack. [COMPLETE: 50%. Site currently uses an in-db session cache. Slow, but safe across
multiple nodes. Site could be load balanced as is. -- Tracker uses in-memory single-process cache; cannot be load
balanced.] (General roadmap: store sessions entirely in-cookie [if feasible], and create a distributed cache and/or communication medium that the tracker and site can use to synchronize data across instances.)

* General task scheduler

* Site<->Tracker event pipeline. (Users updated, banned, deleted. Torrents deleted. etc.)
	* This is needed to ensure integrity of the tracker cache.

* Database migration tool. [COMPLETE: 100%. We are using Goose which is a migration tool written in Go.]
(As a sidebar: I'd love to reuse the Goose configuration file for babou's database connectivity.)

* `Getting Started` document. [COMPLETE: 0%]
(A document that briefly describes Go, links to tutorials on installing a working Go toolchain, and describes
how to use that toolchain to retrieve and compile `babou`.)


Event Bridge

---

* Implement event bridge w/ multiple senders and single receiver [DONE]

* Implement TCP, UNIX socket transports, w/ serializable messages. [DONE]

* Implement loopback socket transport [DONE]

* Connection level timeout for TCP/UNIX socket transports so senders don't block indefinitely.


Web Server

---

* Create a light-weight web framework that can serve static files [COMPLETE: 100%]

* Add a basic MVC structure to the web-framework. [COMPLETE: 80%; Model layer needs some work. 
View layer and controller layer have progressed quite nicely.]

* Add a router which can take RESTful URLs and map them to a controller / action pairing. [COMPLETE: 100%]

* Create a renderer for views that can render contexts inside of mustache templates. [COMPLETE: 100%]

* Create helper functions for the view layer. [COMPLETE: 50%; we have some basic form and url helpers.
This list will grow as the project moves forward.]

* Add database connectivity for the model layer. [COMPLETE: 50%; needs a lot of testing, benchmarking, and
general improvement. -- However a working connection to PostgreSQL can be established.]
	* I have change this so that the database connection is opened once by the web server. This is a
	  problem because it means that only a full stack node will have a properly setup DB connection.
  	* However this did improve performance, since we are now using connection pooling. 
  	(500 concurrent requests reading and writing the DB  went from about 800ms to 500ms 
  	on very slow, encrypted, commodity disks.)



* Allow user's to browse the torrent catalog. [COMPLETE: 10%; only a listing of the first 100 torrents 
  is displayed.]
	* (Timeline is roughly: pagination, categories, tags, fulltext search)

* Site authorization. [COMPLETE: 50%; we have reasonably secure authentication, but we still need a permissions system and administration console.]

* Ratio Watcher. Use ratio statistics and various strategies to help promote healthy torrent swarms.
[COMPLETE: 0%; blocked on tracker collecting stats]
(Some sample strategies are: seeding-to-leeching ratio, uploaded-to-downloaded ratio, seeding-to-leeching-over-time ratio, dont-care ratio [global freeleech], etc.)


Tracker

---

* Create a tracker that listens on a specified port and can sucesfully parse a GET requset for /announce.
[COMPLETE: 100%]

* Add basic GET /scrape support to the tracker. [COMPLETE: 0%]

* Add per-user tokens to /announce URL that implement stats-tracking for private torrents. [COMPLETE: 100%]

* Store torrents in memory. [COMPLETE: 50%; a single-process cache is working reasonably well. -- There are
plans to expand this to a distributed cache so that you can load-balance trackers as well.]

* Store torrents in database. [COMPLETE: 60%; the metainfo (.torrent) file is saved to disk. -- This will expand
to include file listings, tags, and other features that benefit the website's catalog.]

* Attach active peers to torrents. [COMPLETE: 80%; peers are added, but never removed. only supports IPv4 at the
moment.]
	* Needs IPv6 support. IPv4 support should be pretty much functional, though.

* Create background jobs to maintain tracker health. [COMPLETE: 65%; Working on the peer reaper which removes
peers that have not announced recently]
	* We still need a job to remove deleted torrents and inactive torrents from the cache.
	* We have a working peer reaper now that runs every 10 minutes through the whole torrent cache.
	  There are a few problem items that need to be addressed.
	* First: peer reaper has no rate-limit; so if you had 100s of thousands of torrents its going to
	  create that many coroutine workers. On the plus side, torrents are locked/unlocked individually.
	  The disadvantage would be high CPU usage of the server in general.
		* I aim to fix this w/ a buffered channel as a work queue. This will be part of a general task scheduler.
	* Second: the peer reaper needs to subscribe to my generalized task scheduler when its created.
	* Third: when we move to distributed trackers there will be a lot of work to ensure that individual
	  nodes do not step on each other's toes.

* Store ratio and bandwidth statistics for each user. [COMPLETE: 0%]

Far Futures

---

The following features are planned but not (currently) under active development:

* Site Blog and possibly staff blogs / user blogs

* Wiki

* Forums

* Static Pages that are easy to update

* Swappable site themes

* IRC integration

---

Development Specific TODOs:

PeerReaper / Announce interaction:
Announce needs to obtain a writelock on the peer map because it updates their "last seen" as well
as their bandwidth statistics.

If the peer reaper is running tracker requests are going to block until the reaper exits.
As such I'm looking to defer peer writes in a separate goroutine so that the goroutines can
safely block while the rest of the request [read-only] continues unblocked.

This would allow for high availability of the tracker even if a torrent is blocked by a peer reaper.

(All this being said, this is a severe case of premature optimization. -- The block is torrent specific, and
each torrent has, at most, probably several thousand peers. -- A blocking linear scan of that list should barely be noticeable to the end-users.)
