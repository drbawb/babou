package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"errors"
	"flag"

	"fmt"

	libBabou "github.com/drbawb/babou/lib" // Core babou libraries
)

type DatabaseConfig struct {
	ConnectionParams string `json:"open"`
}

type ServerConfig struct {
	DomainName string `json:"domain"`
	ListenAddr string `json:"listen"`
	Port       int    `json:"port"`
}

type BridgePeer struct {
	Transport     string `json:"transport"` //  Socket Type. //TODO: TRANSPORT_TYPE
	SocketAddress string `json:"listen"`    // Address for the socket to send or receive.
	Port          int    `json:"port"`      // Port or suffix [PID,PORT,ETC.] of the remote socket.
}

type BridgeConfig struct {
	LocalBridge BridgePeer    `json:"self"`
	Peers       []*BridgePeer `json:"peers"`
}

// The JSON configuration for the components of the babou stack.
type Config struct {
	Database  *DatabaseConfig `json:"db"`
	WebServer *ServerConfig   `json:"site"`
	Tracker   *ServerConfig   `json:"tracker"`
	Events    *BridgeConfig   `json:"events"`
}

/*
	Parses a configuration file from `config/config.json` OR
	the path passed on the command line.

	Settings on the command-line take precedent over settings
	in the config-file wherever they conflict.

	A config file MUST be specified, even if it is not used
	for actual configuration of the server. For example, hosts
	and ports are used to define announce URLs in .torrent files.
	They may also be used for mailer purposes, etc.
*/

func parseConfig(settings *libBabou.AppSettings) error {
	// Open file and read to byte array
	jsonConfig, err := os.Open(settings.ConfigPath)
	if err != nil {
		return errors.New(fmt.Sprintf("JSON Configuration not found at: %s",
			settings.ConfigPath))
	}
	defer jsonConfig.Close()

	jsonConfigBytes, err := ioutil.ReadAll(jsonConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading [%s]: %s",
			settings.ConfigPath,
			err.Error()))
	}

	// Parse file into configuration struct.
	var parsedConfig Config

	err = json.Unmarshal(jsonConfigBytes, &parsedConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("Error decoding JSON [%s]: %s",
			settings.ConfigPath,
			err.Error()))
	}

	// Start the web-server if it is configured.
	if parsedConfig.WebServer != nil {
		settings.WebHost = parsedConfig.WebServer.DomainName
		settings.WebPort = parsedConfig.WebServer.Port

		settings.WebStack = true
	}

	// Start the tracker if it is configured.
	if parsedConfig.Tracker != nil {
		settings.TrackerHost = parsedConfig.Tracker.DomainName
		settings.TrackerPort = parsedConfig.Tracker.Port

		settings.TrackerStack = true
	}

	// Open a connection pool for the database.
	if parsedConfig.Database != nil {
		settings.DbOpen = parsedConfig.Database.ConnectionParams
	} else {
		return errors.New("Could not find a database configuration in your JSON configuration.")
	}

	settings.FullStack = (settings.WebStack && settings.TrackerStack)

	//TODO: Setup bridge from config file.
	// Setup loopback event bridge and begin discovery process
	// for configured neighbors.
	settings.Bridge = &libBabou.TransportSettings{}
	settings.Bridge.Transport = libBabou.TCP_TRANSPORT
	settings.Bridge.Socket = parsedConfig.Events.LocalBridge.SocketAddress
	settings.Bridge.Port = parsedConfig.Events.LocalBridge.Port

	settings.BridgePeers = make(
		[]*libBabou.TransportSettings, 0, len(parsedConfig.Events.Peers))

	for _, peer := range parsedConfig.Events.Peers {
		peerTransport := &libBabou.TransportSettings{
			Socket:    peer.SocketAddress,
			Port:      peer.Port,
			Transport: libBabou.TCP_TRANSPORT,
		}

		settings.BridgePeers = append(
			settings.BridgePeers, peerTransport)
	}

	return nil
}

func ReadFlags() *libBabou.AppSettings {
	appSettings := &libBabou.AppSettings{}
	var help, debug, webStack, trackStack, fullStack *bool
	var webPort, trackPort *int
	var configPath *string

	help = flag.Bool("help", false, "Prints usage instructions for `babou`.")
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

	configPath = flag.String("config-path", "config/dev.json", "Pathname to your JSON configuration file.")
	flag.Parse()

	appSettings.Debug = *debug

	appSettings.WebStack = *webStack
	appSettings.TrackerStack = *trackStack
	appSettings.FullStack = *fullStack

	appSettings.WebPort = *webPort
	appSettings.TrackerPort = *trackPort

	appSettings.ConfigPath = *configPath

	if *help {
		flag.Usage()
		os.Exit(0)
	}

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
