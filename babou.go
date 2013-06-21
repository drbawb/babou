package main

import (
	fmt "fmt"
	flag "flag"
)

type AppSettings struct {
	Debug *bool

	WebStack *bool
	TrackerStack *bool
	FullStack *bool

	WebPort *int
	TrackerPort *int
}

func main() {
	//Output welcome message:
	fmt.Println("babou fast like veyron.")

	//Parse command line flags
	appSettings := parseFlags()

	if *appSettings.Debug {
		fmt.Println("LOGGING DEBUG MESSAGES TO CONSOLE.")
	}
}

func parseFlags() *AppSettings {
	appSettings := &AppSettings{}

	appSettings.Debug = flag.Bool("debug", false, "Logs debug information to console.")

	appSettings.WebStack = flag.Bool("web-stack", false, "Enables the web application server.")
	appSettings.TrackerStack = flag.Bool("track-stack", false, "Enables the torrent tracker.")
	appSettings.FullStack = flag.Bool("full-stack", true, "Enables the full application stack. - Disabled if track-stack or web-stack are set.")

	appSettings.WebPort = flag.Int("web-port", 8080, "Sets the web application's port number.")
	appSettings.TrackerPort = flag.Int("track-port", 4200, "Sets the tracker's listening port number.")
	


	flag.Parse()
	return appSettings
}
