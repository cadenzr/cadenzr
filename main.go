package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cadenzr/cadenzr/config"
	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/probers"

	"github.com/cadenzr/cadenzr/log"

	_ "github.com/mattn/go-sqlite3"
)

func handleInterrupt(stopProgram chan struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stopProgram <- struct{}{}
	}()
}

func main() {
	if conf, err := config.NewConfigFromFile("./config.json"); err != nil {
		log.Fatalf("Failed read configuration file: %v", err)
	} else {
		config.Config = conf
	}

	logLevel := log.InfoLevel
	switch config.Config.LogLevel {
	case "debug":
		logLevel = log.DebugLevel
	case "info":
		logLevel = log.InfoLevel
	case "warn":
		logLevel = log.WarnLevel
	case "error":
		logLevel = log.ErrorLevel
	default:
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	if err := db.SetupConnection(db.SQLITE, config.Config.Database); err != nil {
		log.Fatalf("Failed to create connection to database: %v", err)
	}

	if err := db.SetupSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	probers.Initialize()

	stopProgram := make(chan struct{})
	handleInterrupt(stopProgram)

	scanCh := make(chan (chan struct{}))
	go scanHandler(scanCh)

	doneCh := make(chan struct{})

	scanCh <- doneCh
	<-doneCh

	<-stopProgram
	log.Info("Stopping cadenzr...")
	db.Shutdown()
}
