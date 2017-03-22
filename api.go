package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cadenzr/cadenzr/config"
	"github.com/cadenzr/cadenzr/controllers"
	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"

	"github.com/cadenzr/cadenzr/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func startAPI() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	e.Static("/", "app/dist")
	e.Static("/images", "images")

	// Login route
	e.POST("/api/login", controllers.AuthController.Login)

	// Restricted group
	r := e.Group("/api")

	r.GET("/albums", controllers.AlbumController.Index)
	r.GET("/albums/:id", controllers.AlbumController.Show)

	e.GET("/api/songs/:id/stream", func(c echo.Context) error {
		id := controllers.StrToUint(c.Param("id"))
		song := &models.Song{}
		gormDB := db.DB.First(song, "id = ?", id)
		if gormDB.RecordNotFound() {
			log.Debugf("Could not start streaming because song '%d' not in database.", id)
			return c.NoContent(http.StatusNotFound)
		} else if gormDB.Error != nil {
			log.Errorf("Could not start streaming database failed: %v", gormDB.Error)
			return c.NoContent(http.StatusInternalServerError)
		}

		streamer, err := NewFileStreamer(song.Path)
		if err != nil {
			log.Errorf("Could not create streamer: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		defer streamer.Close()

		// TODO: Should be tested
		if c.FormValue("from") == "m3u8" {
			song.Played++
			db.DB.Table("songs").UpdateColumn("played", "played+1").Where("id", song.ID)
		}

		// TODO: set the correct time so browser can cache.
		http.ServeContent(c.Response(), c.Request(), song.Name, time.Time{}, streamer)
		return nil
	})

	jwtConf := middleware.JWTConfig{
		Claims:     &controllers.UserLoginClaim{},
		SigningKey: controllers.Secret,
	}
	r.Use(middleware.JWTWithConfig(jwtConf))

	e.Logger.Fatal(e.Start(config.Config.Hostname + ":" + strconv.Itoa(int(config.Config.Port))))
}
