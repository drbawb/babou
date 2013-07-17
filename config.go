package main

import (
	"errors"
	"flag"
	"fmt"

	libBabou "github.com/drbawb/babou/lib" // Core babou libraries
	"github.com/moraes/config"
)

/*
	Parses a configuration file from `config/config.yml` OR
	the path passed on the command line.

	Settings on the command-line take precedent over settings
	in the config-file wherever they conflict.

	A config file MUST be specified, even if it is not used
	for actual configuration of the server. For example, hosts
	and ports are used to define announce URLs in .torrent files.
	They may also be used for mailer purposes, etc.
*/
func parseConfig(settings *libBabou.AppSettings) error {
	cfg, err := config.ParseYamlFile("config/config.yml")
	if err != nil {
		return err
	}

	fmt.Printf("Reading web-application configuration . . .\n")

	/*hostname, err := cfg.String("development.site.domain")
	if err != nil {
		return errors.New("Could not read hostname (`domain`) from configuraiton file.")
	}*/

	port, err := cfg.Int("development.site.port")
	if err != nil {
		return errors.New("Could not read port from configuration file.")
	}

	// Default, use config file port.
	if settings.WebPort == -1 {
		settings.WebPort = port
	}

	fmt.Printf("Reading tracker configuration . . .\n")

	hostname, err := cfg.String("development.tracker.domain")
	if err != nil {
		return errors.New("Could not read tracker hostname (`domain`) from configuraiton file.")
	}

	port, err = cfg.Int("development.tracker.port")
	if err != nil {
		return errors.New("Could not read tracker port from configuration file.")
	}

	// Default, use config file port.
	if settings.TrackerPort == -1 {
		settings.TrackerPort = port
	}

	settings.TrackerHost = fmt.Sprintf("http://%s:%d", hostname, port)

	dbOpen, err := cfg.String("development.db.open")
	if err != nil {
		return errors.New("development.db.open is not present or not in the expected format.")
	}

	// TODO: Sanity check before database panics
	settings.DbOpen = dbOpen

	err = parseEvent(cfg, settings)
	if err != nil {
		return errors.New("Error parsing event-bridge configuration.")
	}
	return nil
}

func parseEvent(cfg *config.Config, settings *libBabou.AppSettings) error {
	fmt.Printf("\n --- event bridge --- \n")
	transport, err := cfg.String("development.events.transport")
	if err != nil {
		return err
	}
	switch transport {
	case "unix":
		fmt.Printf("unix socket \n")
	case "tcp":
		fmt.Printf("tcp socket \n")
	case "lo":
		fmt.Printf("setting up local socket, ignoring peers \n")
		settings.Bridge = &libBabou.TransportSettings{}

		settings.Bridge.Transport = libBabou.LOCAL_TRANSPORT
		settings.BridgePeers = make([]*libBabou.TransportSettings, 0) // TODO: add peers with config reload.
	default:
		fmt.Printf("unknown socket type: %s \n", transport)
		return errors.New("Could not configure event bridge.")
	}

	fmt.Printf("\n --- \n")
	return nil
}

func parseBridgePeers(settings *libBabou.AppSettings, peerList []interface{}) error {
	for _, peer := range peerList {
		v, ok := peer.(map[string]interface{})
		if !ok {
			return errors.New("bad format for peer")
		}

		fmt.Printf("peer transport: %s \n", v["transport"])
	}

	return nil
}

func parseFlags() *libBabou.AppSettings {
	appSettings := &libBabou.AppSettings{}
	var debug, webStack, trackStack, fullStack *bool
	var webPort, trackPort *int

	debug = flag.Bool("debug", false,
		"Logs debug information to console.")

	webStack = flag.Bool("web-stack", false,
		"Enables the web application server.")
	trackStack = flag.Bool("track-stack", false,
		"Enables the torrent tracker.")
	fullStack = flag.Bool("full-stack", true,
		"Enables the full application stack. - Disabled if track-stack or web-stack are set.")

	webPort = flag.Int("web-port", -1,
		"Sets the web application's port number. -1 to use configuration file's port.")
	trackPort = flag.Int("track-port", -1,
		"Sets the tracker's listening port number. -1 to use configuration file's port.")

	flag.Parse()

	appSettings.Debug = *debug

	appSettings.WebStack = *webStack
	appSettings.TrackerStack = *trackStack
	appSettings.FullStack = *fullStack

	appSettings.WebPort = *webPort
	appSettings.TrackerPort = *trackPort

	// If the user has configured their own stack options, do not use the full stack.
	if appSettings.WebStack == true || appSettings.TrackerStack == true {
		appSettings.FullStack = false
	}

	// Panic if configuration fails.
	err := parseConfig(appSettings)
	if err != nil {
		panic(err.Error())
	}

	return appSettings
}
