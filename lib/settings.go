package lib

const (
	TRACKER_ANNOUNCE_INTERVAL int = 300
)

// Available bridge transports.
type TransportType uint8

// Currently we support UNIX and TCP sockets as well as a special loopback socket.
const (
	_               TransportType = iota
	UNIX_TRANSPORT  TransportType = iota
	TCP_TRANSPORT                 = iota
	LOCAL_TRANSPORT               = iota
)

type AppSettings struct {
	Debug bool // Print debug messages

	WebStack     bool // Enable the web-stack
	TrackerStack bool // Enable the tracker-stack
	FullStack    bool // Enable all stacks. (Single binary mode.)

	WebPort     int // Port the web-stack will listen on
	TrackerPort int // Port the track-stack will listen on

	WebHost     string // Hostname of the web-server, used for generating URLs
	TrackerHost string //Hostname of tracker, used for generating URLs.

	Bridge      *TransportSettings   // Local bridge
	BridgePeers []*TransportSettings // Remote bridges
}

type TransportSettings struct {
	Transport TransportType

	Socket string // if applicable
	Port   int    // if applicable
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
