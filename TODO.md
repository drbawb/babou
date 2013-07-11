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

* Database migration tool. [COMPLETE: 100%. We are using Goose which is a migration tool written in Go.]
(As a sidebar: I'd love to reuse the Goose configuration file for babou's database connectivity.)

* `Getting Started` document. [COMPLETE: 0%]
(A document that briefly describes Go, links to tutorials on installing a working Go toolchain, and describes
how to use that toolchain to retrieve and compile `babou`.)

Website

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


* Allow user's to browse the torrent catalog. [COMPLETE: 10%; only a listing of the first 100 torrents is displayed.]
(Timeline is roughly: pagination, categories, tags, fulltext search)

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

* Create background jobs to maintain tracker health. [COMPLETE: 5%; Working on the peer reaper which removes
peers that have not announced recently]

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




