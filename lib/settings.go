package lib

// Application settings that will determine the initial configuration
// of Babou's web server and tracker.

const (
	TRACKER_ANNOUNCE_INTERVAL int = 300
)

type AppSettings struct {
	Debug *bool // Print debug messages

	WebStack     *bool // Enable the web-stack
	TrackerStack *bool // Enable the tracker-stack
	FullStack    *bool // Enable all stacks. (Single binary mode.)

	WebPort     *int // Port the web-stack will listen on
	TrackerPort *int // Port the track-stack will listen on

	WebHost     string // Hostname of the web-server, used for generating URLs
	TrackerHost string //Hostname of tracker, used for generating URLs.
}

// Database settings that will be populated from a flat-file by
// the babou runtime.
//
// DbSettings can be used to open a database connection through
// any lib/db libraries.
type DbSettings struct {
	ServerAddr *string
	ServerPort *int

	SchemaName   *string
	DatabaseName *string
}
