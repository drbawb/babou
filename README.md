Babou
==

Babou, the rogue ocelot.

![babou logo][logo]

An open source, easy to deploy, lightning fast torrent tracker.
In the style of TBDev, Gazelle (what.cd,waffles.fm), etc.

Babou is inspired by popular _private_ tracker software. 
His primary focus is speed, security, and sharing.

Unlike other tracker software, Babou has minimal external dependencies.
Aside from some (included) migration tools: Babou is a single binary application.
Our only external dependency is PostgreSQL. (That was Babou's choice: he says it's
really fast.)

The tracker and website can be run from a single executable.
If more nodes are required, additional instances can be spun up that just
serve the dynamic web content.

Babou offers a migration utility to move data from a MySQL (Gazelle) stack
to our PostgreSQL database. The two applications are NOT database-compatible,
however.

Our stack aims to be simpler than that of Gazelle: all caching is done
by the server executable -- which means we have no external dependencies on 
software such as memcached, redis, or other key-value stores.

In addition we use PostgreSQL and leverage it's fulltext search capabilities.
As opposed to relying on fulltext search daemons.


Requirements
==
We recommend you use the current Go release installation.
All tests and benchmarks published for this project use Go's latest
release tip unless otherwise noted.

- Go 1.1 (or other release version). [For buliding from source.]
- PostgreSQL 9.2.x (latest update).
- That's it . . . optionally you can add a web-server in front of Go 
to proxy requests to the backend.

(You _do not need_ to download and install `Go 1.1` if you have a binary version of 
`babou` -- we are not distributing binaries of `babou` at this time.)


Features [On the roadmap]
==

* Wiki: but not your grandma's wiki. -- Our wiki is lightweight, easy to use, and supports `markdown.`

* Forums: Do you miss the days of bulletin boards and 2400 baud modems? We're going to bring threaded 
  discussions to the table.

* Voting: (Unless you're not into the whole democratically elected sysadmin movement...)

* Torrent Searching: and it's going to be [so awesome!](http://www.youtube.com/watch?v=l8JCX9E0bEI)

* Classes / Permissions: all the features listed above will be access controlled by configurable permissions.

* Ratio Strategies: We realize that not all private trackers watch ratios the same way. -- In fact, some trackers 
  may not act on ratios at all. As such we aim to support several different ratio-watch strategies out of the box
  [which includes global freeleech] -- we also want to make this system fairly easy to extend.

* Torrent "Collections": this includes things like collages, albums, discographies, seasons, series [of books], etc.
  We want individual torrents to be intelligently grouped together for a myriad of multimedia types.

(All in due time. If you'd like to get a sense of the roadmap please check `TODO.md`)

You think ocelot is fast like cheetah?
WELL BABOU IS FAST LIKE A BUGATTI VEYRON. THAT IS NEARLY 4 TIMES AS FAST AS CHEETAH.

Babou is being designed to run behind reverse proxies (load balancers) which will help sites
scale quickly and affordably. We're also making sure that all static assets can be easily
served from a separate server or a CDN.

In addition: babou leverages the powerful concurrency features of the Go runtime to help ensure
high availability of the torrent tracker, even when scheduled tasks [such as statistics collecting, peer pruning]
are consuming the server's time.


How do I run it?
===

We've added `INSTALL.md` to the root of this repository. That document should help you get started
with installing the Go tools and `babou` itself.

Please note that `babou` presently needs some manual configuration of the source to work in your environment.

We are actively working towards creating dedicated configuration files.

---


Directory Structure
===

AKA `What do I need to deploy?`

`babou` must be started from a working directory which contains:

	working dir
 		|--/assets 			(static resources [css, images, js])
 		|--/app/views 			(`mustache` templates, partials)
 		|--/config/dev.json 		(unless specified otherwise by `-config-path`)


`babou` should be installed from a working directory containing the above AND:
	
	working dir
		|--/db (`goose` databse config + migrations for dev + install)

The rest of the directories contain source-code, and there is no reason to deploy
them to production. `babou` is compiled into a statically-linked executable.
It does not interpret or JIT any code at runtime with the exception of templates and
the [not yet implemented] asset pipeline.

For security purposes: I'd recommend that your `/app/* (/app/views/*)` and `/config/*` 
directories not be publicly accessible.

`/assets` should be served up by a server that is good with static assets. (for e.g: nginx) -- (note: babou doesn't need this if you have another web-server rewriting requests for you.)

`/app` and `/config` only needs to be accessible by the `babou` process, `babou` will
not service any requests to that route and your reverse proxies should return
403 or 404 for that route.


[logo]: http://fatalsyntax.com/babou_gh.png "babou logo"

