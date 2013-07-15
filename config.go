package main

import (
	"errors"
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
	if *settings.WebPort == -1 {
		*settings.WebPort = port
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
	if *settings.TrackerPort == -1 {
		*settings.TrackerPort = port
	}

	settings.TrackerHost = fmt.Sprintf("http://%s:%d", hostname, port)

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
		fmt.Printf("loopback socket \n")
	default:
		fmt.Printf("unknown socket type: %s \n", transport)
		return errors.New("Could not configure event bridge.")
	}

	peers, err := cfg.List("development.events.peers")
	if err != nil {
		return err
	}

	// attempt conversion
	for _, peer := range peers {
		v, ok := peer.(map[string]interface{})
		if !ok {
			fmt.Printf("bad format for peer \n")
			continue
		}

		fmt.Printf("peer transport: %s \n", v["transport"])
	}

	fmt.Printf("\n --- \n")
	return nil
}
