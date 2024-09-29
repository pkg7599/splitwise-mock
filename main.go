/*
Splitwise Mock API
Description: This Application is designed to split expenses between users.
Version: 1.0.0
Author: Pranjal Kumar Gupta
Email: pranjalkgupta99@gmail.com
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
	"splitwise-api/app"
	"splitwise-api/internal"
	"syscall"
	"time"
)

var log internal.CustomLogger = internal.Log

func main() {
	api, err := app.CreateApp()
	exitChn := make(chan os.Signal, 1)
	if err != nil {
		panic(err)
	}

	if err := api.Init(); err != nil {
		log.Error(fmt.Sprintf("error occurred in app initialization: %s", err))
		panic(err)
	}

	if err := api.Start(); err != nil {
		log.Error(fmt.Sprintf("error occurred in app start: %s", err))
		panic(err)
	}
	log.Info("application started")
	// Wait for termination signal
	signal.Notify(exitChn, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	<-exitChn
	// Gracefull Shutdown
	log.Info("stopping application...")
	api.Stop(5 * time.Second)
	log.Info("application stopped")
}
