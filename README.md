Babou
==

Babou, the rogue ocelot.

![babou logo][logo]
![travis build][travis]

A fully open-source, BSD licensed torrent tracker written in the 
Go programming language.

Babou is inspired by popular private tracker software.
This means we focus on building communities which grow organically,
encourage quality uploads, and whose members contribute bandwidth
to the swarms.

The main goal of the `babou` project is to provide a _full stack_ solution
in a _single binary with minimal dependencies._ This means that with a single 
command: you will be able to bring up both the frontend [the web-server] and 
the backend [the tracker itself].

Deployment of `babou` merely requires access to a PostgreSQL database.
That's. It.

---

Development of babou does not require anything which cannot be obtained
using a standard Go development environment. 
(With the `go` command in your $PATH.)


Requirements
==

We recommend you use the current Go release installation.
All tests and benchmarks published for this project use Go's latest
release tip unless otherwise noted.

- Go `release` (known minimum: release 1.1)
- PostgreSQL 9.3 or higher

(We do not currently offer binary packages. In the future, however, you will
not need `Go` to get started with `babou`. -- The deployment environment itself
does need the `Go` toolkit. These tools are only required for development.)


Features [On the roadmap]
==

* Wiki
	* Support for `markdown` formatting
	* Central repository for site documentation.
	* Collection of user-written articles about uploads on the site.

* Discussions
	* An environment where your community can stay in touch with one-another.
	* Will incl. some form of private messaging as well as group discussions.
	* Will incl. some searching, tagging, and some form of categorization.

* Feedback Tools
	* A simple collection of survey tools (polls, etc.) to collect feedback
	from your community.

* Torrent Searching: and it's going to be [so awesome!](http://www.youtube.com/watch?v=l8JCX9E0bEI)
	* Torrents will be [full-text] searchable.
	* Torrents will also be organized via tagging.


* Ratio Strategies: 
	* We realize that not all private trackers watch ratios the same way. 
	In fact, some trackers may not actively monitor on ratios at all. 
	As such we aim to support several different ratio-watch strategies.

	* For e.g: global freelech, scoped freelech, seeding vs leeching,
	snatches vs grabs, uploaded [bytes] vs downloaded [bytes], etc.


* Torrent "Collections": this includes things like collages, albums, discographies, seasons, series [of books], etc.
	* We want individual torrents to be intelligently grouped together for a 
	myriad of different media types.

* Permissions will be baked into all these features.
	
(All in due time. If you'd like to get a sense of the roadmap please check `TODO.md`)

---

`babou` is designed to be "fast like a veyron" out of the box.

`babou` has a completely distributed architecture.
Your "pack" of trackers and frontends can be connected together using what we
call the "event bridge."

This bridge allows babou's to securely communicate with one-another about
active users, torrents, and peers.

This means you can easily put `babou` behind a reverse proxy, such as `nginx`,
to help distribute your site's load across several machines.


How do I run it?
===

Run `./babou -help` or view `INSTALL.md` to get an idea of how babou should be
configured and deployed.

Many options are configurable through `config/example.json`
You will need to pass the path to this configuration file using the 
`-config-path=` command line flag.


!!! IMPORTANT !!!

Please note that `babou` presently _requires_ some manual configuration of
the source to work in your environment.
`babou` uses many hard-coded encryption secrets which are openly distributed.

Using the default secrets is _dangerous_ and _completely insecure._
`babou` is NOT FIT FOR PRODUCTION ENVIRONMENTS without configuring the secrets.

If you do not understand the implications of this: I ask that you only use
babou in controlled environments. (Such as behind a network firewall.)

---


Directory Structure
===

AKA `What do I need to deploy?`

`babou` must be started from a working directory which contains:

	working dir
 		|--/assets 				(static resources [css, images, js])
 		|--/app/views 			(`mustache` templates, partials)
 		|--/config/dev.json 	(unless specified otherwise by `-config-path`)


`babou` should be installed from a working directory containing the above AND:
	
	working dir
		|--/db 	(`goose` databse config + migrations for dev + install)

All other directories contain source-code and SHOULD NOT be part of your deployment.

If you're serving from a development directory: please make sure 
`app/*`, `config/*`, and `db/*` are not accessible from the outside!

_Only_ the `/assets` directory should be exposed to the outside world.

`/assets` should ideally be served by a CDN in production.
If a CDN is not feasible: you may want to consider using a server that is
very good with static assets. (I prefer `nginx`...)


[logo]: http://fatalsyntax.com/babou_gh.png "babou logo"
[travis]: https://travis-ci.org/drbawb/babou.png "travis build"
