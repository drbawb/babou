Development Notes
===

The following are Babou design documents and development notes.
They are periodically cleaned up and submitted to the Git repository.


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
--tracker
---TBD
--app
---router.go (Server router)
---server.go (Webserver)
---views (View implementations)
---models (Model implementations)
---controllers (Controller implementations)
--lib
---web (Web Framework core libraries)
--config
---config.json
---database.json
--tests

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


Controllers & Router
===

Diagram:

[Request] --> [Router]<---\
		 |	  |
		 | 	  |
		 \--->[Wrapper]	<-------\
			|		|
			\----------->[Controller]<---->[View]
					| A
					| |
			[Model]<--------/ |
			  |		  |
			  |		  |
			  \----------------

The request is multiplexed by the router to an appropriate controller.
Controllers consist of an Action map which is a map of named routes
to functions capable of servicing an HTTP request.

These functions are given GET/POST parameters. In the case of a
POST request, preferential treatment will be given to POST parameters
if there is a name conflict between parameters.

The controller may call on a model to do business logic. The controller
should block on DB work unless it is a background task, or a task
that is out-of-band with the request being serviced. (For e.g: an
asyncrhonous update that can be pushed later onto a websocket.)

Lastly the controller will call upon the view-layer to render a response.
The view-layer assists binding data to {{mustache}} templates with some
extra helper methods.

After a response has been rendered, the controller must return the response
as a byte-stream to the router.

The router will then return the response to the client.

---

The router can offer additional filters contained in the babou/app package.
These filters are often used for session management, authorization, etc.

A single Controller instance is shared across requests by the router. It is
highly recommended that your controller be stateless: using only the supplied
parameters to construct a response.

The notable exception is accessing the router's session API.

Model Layer
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

Session Management
===

TBA.
