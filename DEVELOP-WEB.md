scratchpad
===

Middleware:

The middleware is exposed to the developer through the `BuildChain()` 
or `DefaultChain()` methods.

These methods expose a context chain, which you can then use or expand on
wherever a context chain is required.


A chain has the following life-cycle:

(Boot-up)
* Resolve: `Resolve(route, action)` called; (type checking here.)
* Hooks may be injected here; they receive raw reqs & responses.
	* Hooks op at the req/response level.
	* Hooks can be executed at any point in the `Request` lifecycle.
	* Hooks act as additional middleware extensions that apply for a specific route.
* Returns a HandlerFunc().
	* Prepares a request-facing (Controller,Action) pairing for the route.
	* Which routes the request through the (BeforeAttach) hooks.
	* Attachs context chains to the (Controller)
	* Routes the request through the (AfterAttach) hooks.
	* Executes `Controller.Action()`
	* Routes the request through the AfterExecution hooks.


(Request)
* BeforeAttach :: action is readied, chains are unattached
* AfterAttach :: chains attached, action pending invocation.
* AfterExecution :: controller has executed its action(s)

Default Hooks:
* Resolve :: <babou context tester>

* BeforeAttach 		:: request logger, panic logger
* AfterAttach 		:: nothing
* AfterExecution 	:: responder

Perf Hooks:
* BeforeAttach		:: perf_start
* AfterAttach		:: perf_step
* AfterExecution 	:: perf_stop

Debug Hooks:
* BeforeAttach	:: panic printer, enable `lib/web` debug flags.

Test Hooks:
* ? ? ?

