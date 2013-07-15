Development Notes
===

The following are Babou design documents and development notes.
They are periodically cleaned up and submitted to the Git repository.

Repository Updates
===
Had to perform a filter-branch and nuke some history. My apologies to the 0 contributors that this
likely affects.
(Seriously though: please clone the repository if you have cloned it on or before 28 June 2013.)

Requirements
===
While `babou` aims to be a single-binary solution, we recognize that database changes are a pain-point
for many system administrators.

As such we have bundled `goose` (https://bitbucket.org/liamstask/goose) with all binary releases of `babou`.
If you are installing `babou` from scratch, it is recommended that you get this tool separately.

`goose` is a database migration tool that will allow you to update your schema between arbitrary versions.

`goose` is only used for updating/downgrading `babou`. -- Initial installation can still be performed
by applying the production image manually _or_ running babou's setup scripts.

All migrations and the corresponding database configuration are contained within `db`
`db` is not used as part of the babou runtime (with the exception of the configuration file).


Type Checking
===

Go is a statically typed language: though it does offers an incredibly powerful runtime.

In an effort to create a better programming environment for creating the `babou` web application
we have leveraged the Go runtime when it seems prudent to do so.

Whenever we side-step the static type checker: we add functions which enable the 
`babou` runtime to ensure that all of its required types are satisfied.

These functions are _expected_ to panic across package boundaries: as their failed 
dependencies are considered extraordinary circumstances.

To that end: if `babou` crashes because of incompatible types: it is considered a fault
of `babou`, and will be treated as a high-priority bug.

Some areas to exercise _extreme caution_ when programming include:
- Context Chains
- Views

Context Chains and Routes require thoughtfulness when implementing their `TestContext` methods to ensure
that the runtime will not add a potentially invalid route. It helps to think of implementing 
a context chain-link as test driven development.

Views rely on a templating library that uses lots of reflection to achieve its flexibility.
As such we have wrapped the underlying templating library with some considerably
safer helper methods that aim to add some additional static (& runtime) type checks to the 
rendering calls.

In summary: `babou` is designed to fail fast and fail hard when a programmer makes an error that
can only be detected at runtime.

Rather than using reflection: `babou` codifies its dependencies using interfaces and methods.
We find that this results in more idiomatic code; albeit it does bring with it additional
verbosity.



Scaffolding
===
To speed up development: the application controller contains several suitable default methods
which will handle processing a route as well as adequately testing the default context chain.

Using thsse methods allows you to create a basic implementation of a controller relatively easy.

Customization can be achieved simply by replacing the body of the default methods with
more appropriate logic as you see fit.


Stack
===

Babou has two stacks, as well as a combination of the two (the full stack):

Web Stack:
- Provides a website that responds to HTTP requests.
- Runs periodically scheduled task(s) to populate "best of" lists, etc.

Tracker Stack:
- Provides a BitTorrent tracker that responds to well-formed HTTP
requests from BitTorrent clients.
- Updates database w/ current peer/seed information, ratio information, etc.
- Uses secret-keys to associate users w/ their accounts.
- Logs IP addresses to secret-keys to find abusive users.

Utilities:
- Scripts which can migrate our schema changes.
- Scripts which can migrate data from Gazelle to Babou.


The webstack uses the database for session storage: so if two
webstacks are identically configured they can service requests from
any users currently logged into the site.

This makes adding more nodes an effective strategy for scaling the web-
portion of your Babou tracker.

Event Bridge
===

`babou` uses an event bridge as a secured pipeline between multiple nodes of `babou`.

First, you must decide how you'd like to configure your pipeline.

For a single-instance running the full stack, you should use the default `loopback` bridge.
In this mode: `babou` is using no additional sockets or ports, and he does not waste time serializing data
into a wire format.

In a multi-instance, single-server setup (for e.g: running -web-stack and -track-stack in two separate processes),
you should use UNIX sockets if supported by your OS. This will use an additional file-descriptor, and requires
serialization.

In a multi-instance, multi-server setup, you should probably use TCP which requires binding an IP address
and an additional port to your `babou` instance. This also requires data serialization, but is workable over
complex network setups.

In full-stack mode, `babou` will _always_ reserve an inter-process event pipeline 
that incurs no additional overhead, regardless of the selected transport.

The purpose of specifying `lo` is that babou should not listen for OUTSIDE connections.

---

The relevant configuration
	events:
		transport: lo # Must be one of: unix, tcp, lo
		#socket: /file
		#address: 0.0.0.0
		#port: 3000
		peers:
		  - transport: unix
		    socket: /tmp/babou.8081.sock
		  - transport: tcp
		    address: 127.0.0.1
		    port: 3001

The `events` key contains the information for the receive socket.
The `events.peers` key contains the sockets of OTHER babou instances.

To enable you to share this confiruation file across instances: `babou` will 
never attempt to  connect to his own instance if it is found in the 
`events.peers` dictionary.

If `transport: lo` is selected, the `events.peers` dicitonary is ignored entirely.

---

The event bridge allows [n] babou web-servers to work with [m] babou trackers in concert.
If `babou` is configured to work with other trackers (listening on the `unix` or 
`tcp` tranpsort) then it will attempt to maintain a connection to each `pack member` listed
in the `events.peers` dictionary.


The event bridge maintains two buffers, a send buffer and receive buffer.

If a worker is non-responsive these buffers will exhaust their resources until eventually the
application will reject incoming requests and refuse to update its cache.

(TODO: Failure strategies to deal with unresponsive workers.)


The send-buffer will forward messages to all available pack members. The receive buffer will
receive a message from other pack member(s); if the message is authentic it will be added to the receive
buffer.

The web-server and tracker use the send-buffer to issue messages to other trackers.
Currently the trackers have exclusive holds on the receive buffer. -- In the future we may add
a message router to direct messages to different receivers.

---

The event bridge uses a binary encoding to serialize messages back and forth; these messages serve
to update tracker's caches as website events occur. (For example: announce keys are changed,
user accounts are disabled by ratio-watchers or moderator actions, etc.)


---

A beautiful ASCII diagram of the stack

	?: [external babou web server] --------------\
												 |
										[tcp or unix socket]
												 |
	[babou: process monitor]					 V
			|--- launches event bridge --> 	[event bridge] - - - -|
			|									^				  |
			|									|				  |
			|							[loopback transport]	  |
			|									|				  |
			|									|				  |
			|-?-- launches web server ---> 	[ web server ]		  |
			|													  |
			|-?-- launches tracker ------>	[  tracker   ]<-------/



The event bridge [will eventually] use a shared secret to authenticate messages
from external transports for security.


Directory Layout
===

Babou uses RESTful routing and an MVC-like design.

babou.go
-D-app 
---router.go 	(Web Server router)  
---server.go 	(Webserver)  
--D-views 		(View implementations)  
--D-models 		(Model implementations)  
--D-controllers (Controller implementations)  
--D-filters 	(babou/lib/web.Context implementations)
-D-bridge		(Event Bridge server)
-D-tracker		(Tracker)
--D-tasks		(Scheduled tasks that the tracker runs periodically)
-D-lib  		(Core libraries that are useful across stack)
--D-web 		(Libraries that are commonly included by web-server related code.)  
--D-db  		(Libraries that are commonly included by model related code.)  
--D-session  	(Database backend for HTTP session storage)
--D-torrent  	(Common library methods for reading/writing `metainfo` files)
-D-config	 	(Site settings read when `babou` starts up)
-D-db		 	(Database migrations for upgrades/downgrades)
-D-tests  		(Tests that I promise to write ... eventually.)

router.go contains a list of all routes in the application.
A route is a pattern that is matched against a submitted URL and its corresponding
HTTP method. (GET/POST/PUT/DELETE).

The router opens a response, and forwards the request to an appropriate 
controller/method pairing.

The controller's method will then call upon the model to perform any business-
logic, and finally it will render a response using a view & template.


Views
===
We use a modified version of mustache.go, which supports `lambda` sections.
By default the `render` method will lookup a global application template (application.template)

Inside of that application.template it will render (view/&lt;controller&gt;/&lt;action&gt;.template)
whenever it encounters the {{#Yield}}{{/Yield}} lambda.

The contents between two Yield blocks is considered to be a "default" string. -- It is not passed
to the view being rendered; if the view fails to render for any reason the site will instead display
the default string to your user.

If you would not like to yield to the application template, you have other options at
your disposal:

You may call `SetContentType('mime/type')` to set the renderers content type.
You may then use `RenderTo(template, data)` and specify a template (as you would specify
a partial path).

Calling RenderTo() will abort the normal view-chain, meaning the normal `return` flow will
be ignored. If RenderTo() has been called you could safely return an empty struct to save
memory.

We also offer RenderFor(map[string]func() &interface{})
The view layer will pull the extension from the route and look it up in the map.
If no entry is found in the map it will return a 404. The default extension if none
is provided is `.html`

Lastly there is RenderStatus('4xx', data)
(Also RenderStatus('4xx', template, data))

It is idential to RenderTo() except it also returns an HTTP status code.
The first form requires your error pages to be at the root of the view hierarchy
and named after their corresponding status codes.
The latter form allows you to render ANY page while sending an HTTP status code.

---

The Render() methods will prepend a series of "view helpers" to your code.

{{#UrlFor [routeName] [params]}}{{/UrlFor}}
Will generate an HTML-safe URL to a route.
[params] will be split on spaces and passed as arguments to the router.
len(params) must match the expected number of arguments for the named route.

{{#LinkFor [routeName] [params]}}[Display]{{/LinkFor}}
Same as UrlFor except it puts the path inside an <a> tag with
an HTML-escaped display name.

{{#FormFor [id] [http-method]}}
[inner template]
{{/FormFor}}

Creates an HTML form; the inner template inherits the calling context as
well as more pointed "Form" helpers.

{{LabelFor [fieldName]}}[Label]{{/LabelFor}}
Creates a label for a form field identified by `fieldName`

{{TextFieldFor [fieldName] [opt:type]}}{{/TextFieldFor}}
Creates a textfield, optional parameter will be used as the "type" of
text field. (Must be valid HTML, for e.g: hidden, password, etc.)


Router & Controllers
===

Diagram:

	[Request] --> [Router]<-------\  
			 		|	  	 	  |  
			 		|	   	      |  
			 		\------->[Chain]<----------------\  
							   |	  			     |  
							   \----------->[Controller]<---->[View]  
												|  A  
												|  |  	
								[Model]<--------/  |
				  				  |		  		   |
				  				  |		  		   |
				  				  \----------------/

An http Request/Response is given to the router which uses a muxer to
match the request URI to an appropriate route.

A "route" is described by the `babou/lib/web.Route` interface, which implies:
- A route can initialize a controller that is ready to service an HTTP request.
- A route can determine if it is a shared instance or an instance private to the request lifcycle.
- A route can create an instance that accepts a basic context (GET/POST request variables)

A "route" is just a controller instance that knows it cannot service a request. (This is because
routes are expected to be shared across request/response lifecycles. Thus storing context in such an
instance would cause data-races as well as possible leaks of sensitive information.)

A "controller", described by `babou/lib/web.DevController` is just a route that has emitted 
a clean instance capable of accepting some basic context as well as an "action" and emitting an HTTP response.

(Since these are two different interfaces: they could be implement independently of one another -- but
standard practice is to make an objet that is both a route and a controller.)

An "action" is just a string [often part of the URL] which tells the controller what type of request
it needs to respond to. (For example: passing the `index` action to a `login` controller may show a 
login form. While passing the `new` action to the same controller with some POST variables would
create a new session for that user and redirect them to a homepage.)

For a controller#action pairing to be usable by the router: it must first be wrapped by a "filter"
that will transform that pairing into an net/http.HttpHandler function.
This filter will provide the context that will be passed in to the controller.

A default filter is provided by the name of `#wrap()` in the `babou/app` package.
This filter is designed to work on any controller that conforms to the aforementioned Route & Controller
interfaces.

It will simply pass in all GET/POST variables to the controller through the vanilla interface and
then call its standard Process() method.

---

Controllers by default are exposed to an `application controller`, which is simply a package-level collection
of helper functions designed to make implementing a controller easier.

One such method is the `babou/app/controllers#process()` method which will, again, work on any vanilla controller.

More complicated controllers can simply override their own Process() method to implement their own handling logic.
For example they may validate that additional context (such as authentication / session information) is available
before proceeding. This way they could globally check permissions and issue a 403 or a redirect header, etc.


Filters,Contexts, and Sessions [an example]
===

Session state is provided by the babou/app/filters package.
`auth.go` describes the SessionContext interface: being session-aware implies:
- Having access to an http.ResponseWriter and an http.Request object.
- Having access to a session.Store as defined by the github.com/gorilla/sessions package.
- Is able to retrieve a specific session from that store's session registry by name.

An example session-aware filter is the AuthContext.
The AuthContext implements the SessionContext interface which means it can look up
sessions by name for a given request.

The AuthContext also provides a "wrapper" or "filter" method: which allows you to wrap
a Route (controller) around a session object.

The AuthContext asserts that the Route, in addition to fulfilling the vanilla Route/Controller interfaces,
also fulfills the "AuthorizableController" interface.

An AuthorizableController must simply accept an AuthContext through it's SetAuthContext() method.

By implementing this method you now have full access to the AuthContext, which in turn is capable
of reading & writing information from a session.

Chaining Contexts
===

Contexts are a useful and powerful concept: but true power comes with composability.
Contexts usually provide their own "Wrapper" method which returns an http.HandlerFunc()

If a context supports context-chaining [by implementing `ChainableContext` interface] then it can instead be wrapped
up in an executable ContextChain.

ContextChain [struct]

- Controller :: web.DevController created from route.
- (Request,Response) :: http Request & Response
- ChainList :: doubly linked list of ChainableContext items.
- Execute() :: iterates over ChainList to modify the controller's context(s); finally instructing the controller to issue a response.

---

A ContextChain is a list of ChainableContext items:
For e.g: [head] <-> [AuthContext] <-> [WalletContext] <-> [tail]

A ContextChain automatically wraps a route/action with the default context (POST/GET vars).
Adding further links via Chain() will have those contexts _applied_ to the controller instance
that is responding to the request.

Calling Execute() on the chain will apply these contexts until the end of the list is reached.

ContextChains are setup at the route level. Thus the controller cannot 100% rely on a context being
available for a given request. Your controller should plan for context(s) to be unavailable and/or 
fail to apply.

For example if your authentication context is not available: you could abort the request, 
or issue a 403 or 500 response, etc.

---

Type Safety
===

To help maintain type safety at runtime: the context chainer will do a dependency check on the
entire chain when the route is created. (When #Execute() is called on a chain.)

This will iterate through the list, passing in the list and a route as a parameter to TestContext() method's.
Each chain-link can use this TestContext() method to abort the routing if a dependency is not satisfied.

Inside TestContext() you can check two things:
- That any chain links you need are available in your current context chain.
- That any controller(s) you need are available from the route.

These tests are usually done by asserting that the route and chain-link's cast to 
an interface in the `filters` package.

---

ApplyContext() executes similarly; but at this point it is _assumed_ by the runtime that the types
have been checked by each context.

ApplyContext(), however, passes in an _instance chain_ as well as an _instance controller._
The difference is that these instances are short-lived (until the request is rendered) and they
do not share state between requests.

--

CloseContext() is executed on a request-safe instance by iterating over the chain a second-time,
_after the request has been served by the controller._

You can use this as a signal to close file or database handles, write out any stateful information, etc.



Contextual Views
===

Contexts can optionally implement babou/lib/web.ViewableContext.

If they do, they can be passed to the RenderWith() methods, which will automatically add any
helpers that the context has defined to the rendering context.

For example: babou/app/filters.FlashContext offers a boolean view helper which is defined as follows:
{{Flash}}
	Everything in here will be rendered: {{Message}}
{{/Flash}}

Where {{Message}} will be the first message from the session's flash messages.
{{Flash}} is a boolean; so the entire section will not be rendered if no flashes are available.

Again: these will only show up if they're _explicitly passed_ to the RenderWith() methods.
As a benefit, RenderWith() methods will ONLY ACCEPT ViewableContext's as their variadic arguments.
This means that the Go compiler can statically check that you have passed an appropriate context
to the view renderer.


Models & Database
===

The model layer consists of two packages:
babou/app/models
AND
babou/lib/db

babou/lib/db contains connection code which offers several different
connection strategies.

The primary strategy is closure-processing: ExecuteFn() and ExecuteAsync()
open connections to the database. These connections are then passed to a 
closure, which may use the database however it wants.

ExecuteFn() will close the session after the closure has executed. Thus it
expects that the closure will not try to close the session and/or pass the
connection handle to a separate coroutine.

ExecuteAsync() will not defer the closing of the database. The closure
MUST free resources when it is finished executing. Use ExecuteAsync()
with extreme caution.



---

Users Sessions:

The user login flow executes as follows:
The login controller will take a username and password [preferably over HTTPS] and have it
compared against the database salt, hash [and pepper].

On sucessful authentication:
- A session will be created for the user which includes various data about
  the user's presence on the website.
- A userId and sessionId will be stored in a user's session cookie.
- The user's session data will be stored in the http_sessions table.

If a user attempts to login and create a session when one already exists: we assume
they are using a new browser or have removed their login cookies.

We will update the session; but instead of simply updating the `last_seen_` fields
we will instead create a fresh session and store it with the old ID.

IDs that are not marked as `remember me` will be pruned from our database
every 2 hours after they are `last seen.`

IDs that are marked as remember me will only be pruned when they logout
or if they login from another computer w/o checking `remember me.`

---

Cache
===

The babou tracker uses an in-memory cache to store details about active users and torrents. 
This allows quick generation of peer lists, and once a user is marked active they can start 
other torrents quickly without round-trips to the database. In addition the throughput of many
scheduled tasks is greatly improved my keeping torrents in memory.

Once a user is cached, however, the tracker is unaware of further updates to that users' permissions.
This means that users which become disabled or banned while a download is in progress will be able to
use the tracker until the cached record is updated.

To help keep this record consistent with the state of the website, the website has a communications
pipeline which informs known trackers of updates to cached records.

The website will send messages to the tracker under the following circumstances:

* A user's personal announce key is updated.

* A user's account status [disabled, banned, etc.] is updated by an admin OR scheduled task.

* A torrent is removed. (Torrent's cannot be "modified" in any meaningful way to the tracker, since it
relies on a consistent "info" dictionary in the torrent. -- Thus tracker does not care about modifications
to the site's metadata.)

---

Cache consistency:

A tracker handles an announce in three distinct phases:

1. READ AND RESPOND
	* The tracker will read the request parameters; read data [from the cache or disk], and respond
	to the client's request ASAP.
2. DEFER WRITE & LOG
	* The tracker will first issue a cache-write. This is simply a message broadcast to other trackers
	on the event pipeline, instructing them to update their cache [if possible].
	* Then the tracker will write to the database.
	* Failing the database write, the tracker will store the failed transaction in an error log.

Since concurrent reads of the torrent are allowable, we try to prioritize reads to help improve throughput.
Writes are deferred and buffered to help lower database load, as well as incresaing overall response speed.


---

Statistics consistency:

Statistics are updated on every announce but the write is deferred to happen outside the request itself.
This allows requests to completely quickly without blocking on writes that don't affect the outcome
of the tracker's response.

Ratio watch (or rather: how it pertains to disabling accounts) is comptued as part of a scheduled task:
rather than being computed on each announce.

When the ratio watch task is run, it will broadcast an event to active trackers informing them of recently
disabled user accounts.


Tracker
===

The tracker currently reads torrents from the database by looking up the requested info_hash.
Once the torrent is loaded it is stored in an in memory cache, and can be quickly retrieved by
the info_hash for all future requests.

The babou tracker responds to announce requests in the form of
GET /{secret}/{hash}/announce

Which is routed as follows:
- The secret is used to identify the user in the users table.
- The hash is used to verify that babou [at some point] generated this user's secret. -- In the future this
  value could be unique per-torrent.
- The secret is randomly generated; the hash however is a SHA256-based HMAC of the secret appended to
  a shared key known to the web application and torrent server.
  This shared key is configurable, and should be changed for each distinct installation of babou.
  (The shared key MUST BE SHARED among any babou's which are cooperating behind a load-balancer, or otherwise share
  a database.)

If this suceeds: the torrent is looked up in the cache by its info_hash.
If it cannot be found babou will defer to distributed caches; and lastly a round-trip to the database.

If none of these hit: the tracker considers the torrent non-existent. 
(If it is deleted through the site: it will be soft-deleted and the reason will be sent to the torrent client.)

In the event of a hit: the peer is added to the swarm and is given a peer list.
The peer list will include the first `num_want` peers from the torrent's active swarm.
Currently `num_want` is defined as the const `DEFAULT_NUMWANT` (30). -- In the future this
will be set to the default only if the client does not specifically request a number of peers.


More intelligent peer-assignment strategies are planned, ideally they will take into account:
* geographical distribution
* seeder:leecher ratios of individual peers.
* peers will be ranked by completed bytes. (Preferring those with more complete downloads.)
* peer estimated bandwidth will be computed between announce-intervals.

---




