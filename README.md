Babou
==

Babou, the rogue ocelot.

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
to proxy requests to the backend. A sample nginx configuration is included.

(This configuration is highly recommended because then you can use a battle-hardened
webserver on trusted ports (80,443), perform SSL, serve static assets, etc.) 

Features
==
Anything Gazelle can do, we can do ... also.

Wiki: but not your grandma's wiki, this thing supports markdown and is made of awesome.
Forums: but not your grandma's BBS, this thing supports permissions out the wazoo.
Polling: unless you're not into the whole democratically elected sysadmin movement.
Torrent Searching: and it's _awesome._
Permissions!!!: seriously there's like 50 ways to ban a user.
Collages: because we &lt;3 music.

But not yet... kindasortafocusing on making the site, yknow, work.

Anything Ocelot can do, we can do quicker.

MULTICORE, DO YOU USE IT?!?!?
YOU THINK OCELOT IS FAST LIKE CHEETAH?
WELL BABOU IS FAST LIKE A BUGATTI VEYRON. THAT IS NEARLY 4 TIMES AS FAST AS CHEETAH.

We even allow you to configure the number of processors you'd like Babou to use for
his INCREDIBLY OVERPOWERED WARP DRIVE.

How do I run it?
===

REALLY EASILY.

Step 1) Compile or obtain the binary & associated webstuffs... throw it in a folder.
Step 2) `babou --set-it-up`

(Babou doesn't like being told what to do, SO HE WILL CONFIRM THAT YOU REALLY
WANT TO DO A CLEAN INSTALL. Then he will seed the database, randomize encryption 
parameters, and finally spit out an admin username and password. DON'T LOSE IT.
Babou will not let you do this more than once without an additional --yes-im-sure.)

Step 3) `babou --full-stack --web-port=8080 --tracker-port=34200`

Want more nodes?
Step 4) `babou --web-stack --web-port=8081`
Step 5) `babou --web-stack --web-port=8082`
Step 6) . . .

Now what?
===
Time to share some linux distributions, yo.

How do I know that Babou is safe?
===

First of all: it's not illegal to share linux distributions.

Secondly: BABOU BELIEVES IN FREEDOM. He has also had all of his shots.
He is BSD Licensed, top to bottom. Check the code for yourself.

Still don't trust me? Rip out the bits you don't trust, fork it on GitHub,
publicly shame me, submit a pull request, whatever floats your boat. Such is
the miracle of open source development!

Running `babou --license` will spit out the licenses of all frameworks involved.
