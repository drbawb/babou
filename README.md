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

IMPORT NOTICE: To build `babou` you will need the mustache templating library.We use a modified version available at `github.com/drbawb/mustache`.
Since the modifications are not merged into master yet, you will manually
have to go to $GOPATH/src/github.com/drbawb/mustache and checkout the
appropriate development branch.

This _only_ applies if you're _building_ babou from source.

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

(All in due time.)

Anything Ocelot can do, we can do quicker.

MULTICORE, DO YOU USE IT?!?!?
YOU THINK OCELOT IS FAST LIKE CHEETAH?
WELL BABOU IS FAST LIKE A BUGATTI VEYRON. THAT IS NEARLY 4 TIMES AS FAST AS CHEETAH.

We even allow you to configure the number of processors you'd like Babou to use for
his INCREDIBLY OVERPOWERED WARP DRIVE.

How do I run it?
===

REALLY EASILY.

Step 1) Compile or obtain the binary & associated assets... throw it in a folder.  
Step 2) `./babou`  


Want more nodes?
Step 3) `babou --web-stack --web-port=8081`  
Step 4) `babou --web-stack --web-port=8082`  
Step 5) . . .  

(`./babou -help` is also available.)

Note: `babou` requires the `assets/` directory and the `app/` directory to be present in the current working directory.

`assets/` include static assets and can safely be served from a load-balancer or reverse proxy.

`app/` contains runtime assets such as templates and partials.

Now what?
===
Time to share some linux distributions!

