// `babou` allows you to spawn an instance of the babou private tracker.
package main

import (
	flag "flag"
	fmt "fmt"

	libBabou "babou/lib"
	web "babou/lib/web"
	os "os"
	signal "os/signal"
	syscall "syscall"
)

func main() {
	//Output welcome message:
	fmt.Println("babou fast like veyron.")

	//Parse command line flags
	appSettings := parseFlags()

	if *appSettings.Debug {
		fmt.Println("LOGGING DEBUG MESSAGES TO CONSOLE.")
	}

	//Trap signals from the parent OS
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go trapSignals(c)

	server := &web.Server{}

	server.Start(appSettings)
}

func trapSignals(c chan os.Signal) {
	for sig := range c {
		fmt.Printf("Caught: %s \n", sig.String())
		if sig == syscall.SIGINT || sig == syscall.SIGQUIT || sig == syscall.SIGTERM {
			// Shutdown gracefully.

			// Probably block on webserver shutdown [instant]
			// 	as well as a concurrent block on app shutdown.
			// Exit when they're both finished.
			os.Exit(0)
		} else if sig == syscall.SIGKILL {
			// Get out get out get out!!!
			os.Exit(2)
		}
	}
}

func parseFlags() *libBabou.AppSettings {
	appSettings := &libBabou.AppSettings{}

	appSettings.Debug = flag.Bool("debug", false,
		"Logs debug information to console.")

	appSettings.WebStack = flag.Bool("web-stack", false,
		"Enables the web application server.")
	appSettings.TrackerStack = flag.Bool("track-stack", false,
		"Enables the torrent tracker.")
	appSettings.FullStack = flag.Bool("full-stack", true,
		"Enables the full application stack. - Disabled if track-stack or web-stack are set.")

	appSettings.WebPort = flag.Int("web-port", 8080,
		"Sets the web application's port number.")
	appSettings.TrackerPort = flag.Int("track-port", 4200,
		"Sets the tracker's listening port number.")

	flag.Parse()
	return appSettings
}
