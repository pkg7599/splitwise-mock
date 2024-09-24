/*
Splitwise Mock API
Description: This Application is designed to split expenses between users.
Version: 1.0.0
Author: Pranjal Kumar Gupta
Email: pranjalkgupta99@gmail.com
*/

package main

import (
	"os"
	"os/signal"
	"splitwise-api/app"
	"syscall"
	"time"
)

func main() {
	api, err := app.CreateApp()
	exitChn := make(chan os.Signal, 1)
	if err != nil {
		panic(err)
	}

	if err := api.Init(); err != nil {
		panic(err)
	}

	if err := api.Start(); err != nil {
		panic(err)
	}
	// Wait for termination signal
	signal.Notify(exitChn, os.Interrupt, syscall.SIGTERM)

	<-exitChn
	// Gracefull Shutdown
	time.Sleep(5 * time.Second)
	api.Stop()
}
