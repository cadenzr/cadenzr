package db

import (
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	"github.com/jinzhu/gorm"
	// Load sqlite plugin for gorm.
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Dialect is the database used. E.g sqlite3,mysql,...
// Currently only sqlite3 is supported.
type Dialect string

const (
	// SQLITE dialect name.
	SQLITE Dialect = "sqlite3"
)

// DB is the instance that the other code can use to access the database.
var DB *gorm.DB

// SetupConnection initializes the DB variable in this package.
func SetupConnection(dialect Dialect, args ...interface{}) (err error) {
	log.Infof("Trying to connect to database.")
	DB, err = gorm.Open(string(dialect), args...)
	if err != nil {
		log.Infof("Could not connect to database: %v", err)
	} else {
		log.Infof("Connected to database.")
	}

	return
}

// Shutdown closes the database connection.
func Shutdown() (err error) {
	if DB == nil {
		return nil
	}
	err = DB.Close()
	DB = nil

	return err
}

// SetupSchema Creates the database tables.
func SetupSchema() (err error) {
	log.Info("Updating database schema.")

	db := DB.AutoMigrate(
		&models.Artist{},
		&models.User{},
		&models.Image{},
		&models.Album{},
		&models.Song{},
		&models.Playlist{},
	)
	if db.Error != nil {
		log.Errorf("Failed to update database schema: %v", err)
		return
	}

	log.Info("Database schema updated.")

	return nil
}
