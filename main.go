package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/signal"
	"syscall"

	"github.com/cadenzr/cadenzr/config"
	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/models"
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

var scanCh chan (chan struct{})
var configFile string = "./config.json"

func main() {
	if conf, err := config.NewConfigFromFile(configFile); err != nil {
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

	// Create admin user.
	shaSum := sha256.Sum256([]byte(config.Config.Password))
	hash := hex.EncodeToString(shaSum[:])
	user := &models.User{
		Username: config.Config.Username,
		Password: hash,
	}
	gormDB := db.DB.FirstOrCreate(user).Where("username = ?", user.Username)
	if gormDB.Error != nil {
		log.Fatalf("Failed to check if user already in database: %v", gormDB.Error)
	}
	if user.Password != hash {
		user.Password = hash
		gormDB = db.DB.Table("users").Where("username = ?", user.Username).UpdateColumn("password", user.Password)
		if gormDB.Error != nil {
			log.Fatalf("Failed to update user '%s' password.: %v", user.Username, gormDB.Error)
		}
	}

	probers.Initialize()

	stopProgram := make(chan struct{})
	handleInterrupt(stopProgram)

	scanCh = make(chan (chan struct{}))
	go scanHandler(scanCh)

	go startAPI()

	<-stopProgram
	log.Info("Stopping cadenzr...")
	db.Shutdown()
}
