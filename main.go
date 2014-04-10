package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdlayher/wavepipe/core"
)

// testFlag invokes wavepipe in "test" mode, where it will start and exit shortly after.  Used for testing.
var testFlag = flag.Bool("test", false, "Starts " + core.App + " in test mode, causing it to exit shortly after starting.")

func main() {
	// Set up logging, parse flags
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	// Application entry point
	log.Println(core.App, ": starting...")

	// Invoke the manager, with graceful termination and core.Application exit code channels
	killChan := make(chan struct{})
	exitChan := make(chan int)
	go core.Manager(killChan, exitChan)

	// Gracefully handle termination via UNIX signal
	sigChan := make(chan os.Signal, 1)

	// In test mode, wait for a short time, then invoke a signal shutdown
	if *testFlag {
		go func() {
			// Wait 5 seconds, to allow reasonable startup time
			seconds := 5
			log.Println(core.App, ": started in test mode, stopping in", seconds, "seconds.")
			<-time.After(time.Duration(seconds) * time.Second)

			// Send interrupt
			sigChan <- os.Interrupt
		}()
	}

	// Trigger a shutdown if SIGINT or SIGTERM received
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	for sig := range sigChan {
		log.Println(core.App, ": caught signal:", sig)
		killChan <- struct{}{}
		break
	}

	// Force terminate if signaled twice
	go func() {
		for sig := range sigChan {
			log.Println(core.App, ": caught signal:", sig, ", force halting now!")
			os.Exit(1)
		}
	}()

	// Graceful exit
	code := <-exitChan
	log.Println(core.App, ": graceful shutdown complete")
	os.Exit(code)
}