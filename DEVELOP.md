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

While you can add additional tracker-stacks, two trackers cannot share the same port.
Logistically: if you were to add additional trackers, they must _always_ be running.
Otherwise stale torrent-files will contain bad connection information, resulting in
a poor user experience.

Babou will prevent you from running additional tracker-stacks, but we do intend to
add a flag which will allow multiple trackers to be run.

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


Controllers
===

Models
===


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


Partials may be rendered with {{#Partial}}&lt;partial_name&gt;{{/Partial}}
The partial name will be prepended with an underscore, it will be called from the current
context's directory. In the case of the appilcation template: this is the `views/` directory.
In the case of a controller/action template, this is the `views/controller/` directory.

If your partial name begins with a `/`, it will be considered an absolute path to the partial
beginning in the `views/` directory.

Examples:

{{#Partial}}nav_bar{{/Partial}}
Called from: views/application.template
Loads: views/_nav_bar.template

{{#Partial}}notifications{{/Partial}}
Called from: views/users/inbox.template
Loads: views/users/_notifications.template

{{#Partial}}/footer{{/Partial}}
Called from: views/torrents/show.template
Loads: views/_footer.template

(Note: The partial renderer will ignore whitespace.)
(In all cases: the partial inherits the calling context of the outer view.)

