package lib

// web server constants
const (
	WEB_SERVER_STARTED = iota // Receive from web server's io when it is fully initalized and ready to receive requests.
	WEB_SERVER_STOP           // Send on web server's io chan to shut it down.
	WEB_SERVER_STOPPED        // Receive from web server's io when shutdown is complete.
	WEB_SERVER_ERR
	TRACKER_SERVER_START
	TRACKER_SERVER_ERR
)
