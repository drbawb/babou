// `babou` allows you to spawn an instance of the babou private tracker.
package main

import (
	fmt "fmt"
	os "os"
	signal "os/signal"
	syscall "syscall"

	web "github.com/drbawb/babou/app" // The babou application: composed of a server and muxer.
	tracker "github.com/drbawb/babou/tracker"

	bridge "github.com/drbawb/babou/bridge"
	config "github.com/drbawb/babou/lib/config"

	libBabou "github.com/drbawb/babou/lib" // Core babou libraries
	libDb "github.com/drbawb/babou/lib/db"
)

var bridgeIO chan bool = make(chan bool)

func main() {
	//Output welcome message:
	fmt.Println("babou fast like veyron.")

	//Parse command line flags
	appSettings := config.ReadFlags()

	//Trap signals from the parent OS
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go trapSignals(c)

	// Start event bridge
	var appBridge *bridge.Bridge
	webServerIO := make(chan int, 0)
	trackerIO := make(chan int, 0)

	switch appSettings.Bridge.Transport {
	case libBabou.LOCAL_TRANSPORT:
		fmt.Printf("Starting event-bridge \n")
		appBridge = bridge.NewBridge(appSettings.Bridge)
	case libBabou.TCP_TRANSPORT:
		fmt.Printf("Event-bridge listening on TCP \n")
		appBridge = bridge.NewBridge(appSettings.Bridge)
	default:
		panic("Bridge type not impl. yet...")
	}

	for _, peer := range appSettings.BridgePeers {
		switch peer.Transport {
		case libBabou.TCP_TRANSPORT:
			fmt.Printf("Event-bridge listening on TCP \n")
			tcpPeer := bridge.NewTCPTransport(
				fmt.Sprintf("%s:%d", peer.Socket, peer.Port))

			appBridge.AddTransport(tcpPeer)
		default:
			panic("Bridge type not impl. yet...")
		}
	}

	// Connect to the database.
	fmt.Printf("Opening database connection ... \n")
	_, err := libDb.Open(appSettings)
	if err != nil {
		panic("database could not be opened: " + err.Error())
	}

	// Start instance of web-application [if applicable]
	if appSettings.FullStack == true || appSettings.WebStack == true {
		fmt.Printf("Starting web-server \n")
		server := web.NewServer(appSettings, appBridge, webServerIO)

		go server.Start()
	}

	// Start instance of tracker [if applicable]
	if appSettings.FullStack == true || appSettings.TrackerStack == true {
		fmt.Printf("Starting tracker \n")
		server := tracker.NewServer(appSettings, appBridge, trackerIO)

		go server.Start()
	}

	// Catch useless configurations.
	if appSettings.FullStack == false &&
		appSettings.TrackerStack == false &&
		appSettings.WebStack == false {
		fmt.Printf("babou has nothing to do!")
		os.Exit(2)
	}

	// Poll server events indefinitely
	// TODO: Potential deadlock if either coroutine were to terminate.
	for {
		select {
		case webMessage := <-webServerIO:
			if webMessage == libBabou.WEB_SERVER_STARTED {
				fmt.Println("Server has started sucessfully")
			}
		case trackerMessage := <-trackerIO:
			if trackerMessage == libBabou.TRACKER_SERVER_START {
				fmt.Println("Tracker has started successfully")
			}
		}
	}
}

func trapSignals(c chan os.Signal) {
	for sig := range c {
		if sig == syscall.SIGINT || sig == syscall.SIGQUIT || sig == syscall.SIGTERM {
			// Shutdown gracefully.
			fmt.Println("\nbabou is packing up his things ...")

			//TODO: Probably block on webserver shutdown [instant]
			fmt.Println("\nwaiting for webserver to shutdown...")
			fmt.Println("\nwaiting for tracker to shutdown...")
			fmt.Println("\nwaiting for event-bridge to close sockets...")

			os.Exit(0)
		} else if sig == syscall.SIGKILL {
			// Get out get out get out!!!
			os.Exit(2)
		}
	}
}
