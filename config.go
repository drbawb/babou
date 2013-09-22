package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"errors"
	"flag"

	"fmt"

	libBabou "github.com/drbawb/babou/lib" // Core babou libraries
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
	JSON_CONFIG_PATH := "config/config.json"

	// Open file and read to byte array
	jsonConfig, err := os.Open(JSON_CONFIG_PATH)
	if err != nil {
		return errors.New(fmt.Sprintf("JSON Configuration not found at: %s",
			JSON_CONFIG_PATH))
	}

	jsonConfigBytes, err := ioutil.ReadAll(jsonConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading [%s]: %s",
			JSON_CONFIG_PATH,
			err.Error()))
	}

	// Parse byte-array as JSON
	var f interface{}

	err = json.Unmarshal(jsonConfigBytes, &f)
	if err != nil {
		return errors.New(fmt.Sprintf("Error decoding JSON [%s]: %s",
			JSON_CONFIG_PATH,
			err.Error()))
	}

	fmt.Printf("json %v", f)

	return errors.New("JSON CONFIG PARSER NOT IMPL.")
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
