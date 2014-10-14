Getting Started with Babou
===

Babou is a private tracker project under heavy development. This project is written entirely in Go, which 
helps us deliver several key advantages.

* Babou itself is statically compiled into a single binary. -- It has _no other_ runtime dependencies
except a database server [PostgreSQL]. -- This means Babou can be _fast_ and extremely easy to deploy.

* Babou can handle requests concurrently. This means that many users can be serviced
simultaneously, with the exception of brief periods of time that help ensure consistency and accuracy.

* Babou is easy to work on and extend. The project has been open source from day zero, 
and the project is currently undergoing a bit of [README Driven Development](http://tom.preston-werner.com/2010/08/23/readme-driven-development.html). This means a
_complete history_ of the project is available to developers. We hope that this reduces the
barrier to entry for developing babou.

---

At this time we do not offer binary installations of `babou`. Please keep in mind that we are
not even in a "public alpha" stage yet. We will use `git tags` to mark different versions of babou
when it is considered ready for extensive testing [and later: production use].

If you are still interested in test-driving `babou` you will need to compile it from source.
This document will outline the tools you will need, and how to use them to build and install
babou.

---

Requirements:

* A PostgreSQL database server, with a database set aside for Babou. Babou uses the public schema and
does not prefix any of his tables. For this reason: we _highly_ recommend that you create a _separate database_ (and user) specifically for the babou application.

* (In the future: we anticipate that the postgresql `hstore` extension may be required. We do not currently use it, but please keep this potential requirement in mind if configuring/compiling postgres from scratch.)

* The Go toolkit. -- You _MUST_ use Go 1.1 [or later] to compile `babou`, as we use some features not available
in Go 1.0. To my knowledge: GCC Go will not work to compile `babou`, this is because it currently lacks support
for some of the features introduced in Go 1.1.

* Git and Mercurial. -- You will need these to use the `go get` tool: which will automate the process of downloading
`babou` and its dependencies. (You do need both: several of our dependencies are hosted on mercurial repositories.)

Recommended for developers:

* An editor that has decent support for Go. -- Currently several configuration values only exist in Go source files.
Several good editors are Sublime Text 2 with the GoSublime plugin, vim with gocode, go-mode.el and go-eldoc for emacs.

* Please read DEVELOP.md in the root of this repository. -- The DEVELOP document explains the ideas behind
babou's core architecture, project goals, how to extend babou, among other things.

* Before submitting any patches, please run them through the `go fmt` tool using its default settings.
This helps ensure usable diffs, keeps the code readable, and enforces a consistent style throughout
the codebase.

Supported Operating Systems:

* Babou has only been tested on Mac OS X (10.7) and Linux (Arch). We do intend to support Windows,
so testing, bug reports, and patches are welcomed to that end.

* Some features of Babou [planned or otherwise] will always work best on Unix-like systems, however.
For e.g: PostgreSQL and the planned communication backend for `babou` can take advantage of Unix
sockets. These sockets can greatly improve speed and security, as well as simplifying small deployments
of `babou`.

---

First you will need to install the Go toolkit for your operating system. This is available from [http://golang.org/doc/install](http://golang.org/doc/install).

You can optionally compile and install Go itself from source, which is described here: [http://golang.org/doc/install/source](http://golang.org/doc/install/source).

To work with `babou` we highly recommend you use the `go get` tool. This will require you to install and
configure mercurial, this process is described here: [http://golang.org/doc/install/source#mercurial](http://golang.org/doc/install/source#mercurial).

You will need to configure your GOPATH environment variable, we recommend that you set it to a directory
inside your home folder (or "Documents" folder on Windows).

An example configuration is this:

export GOPATH=$HOME/projects/go

Where $HOME/projects/go is a folder with the following structure

	projects/go
 		|-- src [dir]
 		|-- pkg
 		|-- bin

For more information, reference this document: [http://golang.org/doc/code.html](http://golang.org/doc/code.html)

After you have sucesfully setup Go (and optionally compiled a small sample program, such as the Hello, World example
from the page linked above) you are ready to download `babou`.

Instead of a standard `git clone` you can actually use a tool included with the Go toolchain called `go get`

`go get github.com/drbawb/babou` will perform the following:

* Create a directory for the babou project under $GOPATH/src/github.com/drbawb/babou

* Clone the `master` branch of that repository [or a specially tagged branch] into that folder.

* Repeat this process for any dependencies of `babou` that you are missing.

This process is described here: [http://golang.org/cmd/go/](http://golang.org/cmd/go/) under the "Download and install packages and dependencies" section.

If this was sucessful, you should be able to go into the $MYGO/src/github.com/drbawb/babou directory and
run the `go build` command. -- Generally speaking the `master` branch should compile except under rare
circumstances.

This directory is a copy of the github repository, so you can branch and work with it just like
any other repository.

---

This project is documented using `go doc`, the tool is somewhat similar to JavaDoc or other automated
documentation tools.

You can use the godoc tool to browse the documentation locally if you do not have an internet connection.
This is briefly described here: [http://golang.org/cmd/go/](http://golang.org/cmd/go/)

Otherwise you can view the documentation online, though it is not always up to date, at: [http://godoc.org/github.com/drbawb/babou](http://godoc.org/github.com/drbawb/babou)

Note that this website will fetch and display the documentation for _any package_ that you can install with the
`go get` command!

This is an easy way to familiarize yourself with the codebase and architecture of the `babou` project.

---

To configure your database connection, please edit the file `lib/db/connect.go`, around [roughly] line
41 in the Open() method you will see a line that looks like this:

`dbConn, err := sql.Open("postgres", "user=rstraw host=localhost dbname=babou sslmode=disable")`

Please change the second string to your PostgreSQL database connection parameters.
The format of this string is described in several places, the `lib/pq` [documentation](http://godoc.org/github.com/lib/pq), and [the readme](https://github.com/lib/pq) of the `lib/pq` database driver.

You will need to put _the same connection string_ in the goose migration tool's configuration file,
which is located at: `db/dbconf.yml`

(Don't worry: we intend to move these parameters to a unified configuration file at a later date.)


---

To setup your database, you will need to `go get` another tool. This tool will [eventually] be shipped
with the binary distribution of `babou`.

Run: `go get bitbucket.org/liamstask/goose/cmd/goose`
This will install `goose` into your `$GOPATH/bin` directory. You can add this to your $PATH variable
or simply execute it directly from that folder.

Execute `goose up` while you are in the _root directory_ of the babou project (`$GOPATH/src/github.com/drbawb/babou`).

This should run several migrations which will add tables and fields to your database.

You will end up with an empty database, currently this is OK because `babou` has open registration and
does not have any sort of permission/rank system yet.

---

At this point you should be able to start babou, to start with the default development settings simply run
`./babou` from the root directory of the babou project. If you'd like to see those settings, and the
available configuration flags, simply run `./babou -help` from the command line.

(Seriously... that's all there is to it.)

By default, babou runs the website on port 8080, and the tracker on port 4200.

`babou` will "panic" and fail to start under the following conditions:

* the selected ports are not available
* the built-in route tests have failed for some reason. (almost always a bug.)

`babou` will "panic" under other conditions, but it will not exit if:
* a template cannot be found on disk. (usually means you are running `babou` from the wrong directory.)
* a template cannot be rendered. (could be a bug in the controller, view, or template.)
* an unexpected value is passed to the webapp. (these will leave long stacktraces on STDOUT. please report
these, as they are ALL considered bugs. typically they will simply return a server error [500] for the request
that caused the problem; if the server itself crashes as a result of one of these stack traces:
the  bug will be considered of a highest priority for `babou` developers.)

It is IMPERATIVE that you run this from the root of the babou project. -- babou uses relative paths
to access templates and assets at runtime.

In the future we will allow you to specify separate directories for assets and app/* templates.

---

You can browse to `babou` at `http://localhost:8080/`

Some routes to try out [even though templates may not have links to them]

`/` (welcome page OR "site news" placeholder.)

`/register` (create an account)

`/login`
`/logout`

`/torrents` (browse torrents)
`/torrents/upload` (upload a .torrent file to the tracker; also displays your personal announce URL)
`/torrents/download/{id}` (where {id} is replaced with the ID number displayed on `/torrents`)

The tracker should be running on `http://localhost:4200` and it currently only listens for a single route:
`/{secret_key}/{secret_hash}/announce`

The secret_key is a key generated for a user when they register an account. This key is pulled from a
random number generator included with the `Go standard library` and it is considered "cryptographically secure."

The hash is simply an HMAC hash of the user's secret key using a site-wide encryption key.
This hash ensures that the user's secret_key was signed by _your instance_ of `babou.`

This site-wide key is currently hard-coded but will soon be a configuration parameter. Furthermore
we will REQUIRE that operators use something other than the default for security purposes. This is useful
if, for example, your database was compromised but the attacker could not gain access to your source or binary
version of `babou`.

I plan to allow for the disabling of the HMAC hash as well, as it does add a small overhead to every request
sent to the tracker. -- When disabling the HMAC the site will function like a more traditional private tracker.

---

IMPORTANT DEVELOPMENT NOTES
===

### Branches and Versioning ###

`master` is generally considered to be the authoritative release branch. Any commit on master will
usually compile on Mac and Linux using the standard Go1.1 compiler and setup.

Ocasionally you will see a develop/ or test/ branch show up on the github repository.
These branches are almost always deleted once their feature is complete and they are merged with master. (Or abandoned.)

These branches are considered temporary, and it is unwise to base any work off of them.

The `master` branch may be tagged with version numbers or explanatory tags. Our current [planned] versioning
scheme is as follows: `major.minor.rev`

Tagged releases typically undergo more rigorous testing and have their documentation brought up-to-date.

At the time of writing: there are no tagged releases of `babou`, and it is considered _pre-alpha_ software.
Contributions in the form of code, assets, issue reports, milk and cookies, et al. are all greatly appreciated.
You may get in touch with me on github.com or at drbawb \[at\] fatalsyntax \[dot\] com

Many features noted in the README are incomplete or non-existent; please see the TODO for a more realistic
look at the state of the project.

No milestones have dates, but progress is marching along rapidly and we hope to tag a 0.1.0 release by the 
end of august.


### Potential Contributors ###

Please submit any code contributions as a github `pull request` to github.com/drbawb/babou
You may add your name [or a pseudonym] and a small tagline to the CONTRIBUTORS.md file if you choose.

Any submitted pull requests submitted will be licensed under babou's BSD license if they are included.
