Development Notes
===

The following are Babou design documents and development notes.
They are periodically cleaned up and submitted to the Git repository.

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

Directory Layout
===

Babou uses RESTful routing and an MVC-like design.

babou.go
-D-tracker
-D-app 
---router.go (Server router)  
---server.go (Webserver)  
--D-views (View implementations)  
--D-models (Model implementations)  
--D-controllers (Controller implementations)  
--D-filters (babou/lib/web.Context implementations)
-D-lib  (Core libraries that are useful across stack)
--D-web (Libraries that are commonly included by web-server related code.)  
--D-db  (Libraries that are commonly included by model related code.)  
-D-tests  

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

	[Request] --> [Router]<-----\  
			 		|	  		|  
			 		|	   	    |  
			 		\------->[Wrapper]<---------\  
								|				|  
								\----------->[Controller]<---->[View]  
												| 	A  
												| 	|  	
								[Model]<--------/   |
				  				  |		  			|
				  				  |		  			|
				  				  \-----------------/

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

A ContextChain is a doubly-linked list of ChainableContext items:
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

--


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
