package main

import (
	"bytes"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/cadenzr/cadenzr/config"
	"github.com/cadenzr/cadenzr/controllers"
	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/jinzhu/gorm"

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
			db.DB.Table("songs").Where("id = ?", song.ID).Update("played", gorm.Expr("played+1"))
		}

		// TODO: set the correct time so browser can cache.
		http.ServeContent(c.Response(), c.Request(), song.Name, time.Time{}, streamer)
		return nil
	})

	r.POST("/scan", func(c echo.Context) error {
		done := make(chan struct{})
		scanCh <- done

		<-done
		return c.NoContent(http.StatusOK)
	})

	e.GET("/api/albums/:id/playlist.m3u8", func(c echo.Context) error {
		id := controllers.StrToUint(c.Param("id"))

		songs := []*models.Song{}
		db.DB.Find(&songs, "album_id = ?", id)

		endpoint := "http://" + config.Config.Hostname
		if config.Config.Port != 0 {
			endpoint = endpoint + ":" + strconv.Itoa(int(config.Config.Port))
		}
		endpoint = endpoint + "/api/songs/"

		response := bytes.NewBuffer([]byte{})
		response.WriteString("#EXTM3U\n")
		for _, song := range songs {
			artist := ""
			if song.Artist != nil {
				artist = song.Artist.Name
			}
			response.WriteString("#EXTINF:" + strconv.Itoa(int(math.Ceil(song.Duration.Float64))) + ", " + artist + " - " + song.Name + "\n")
			response.WriteString(endpoint + strconv.Itoa(int(song.ID)) + "/stream?from=m3u8\n")
		}

		return c.Stream(http.StatusOK, "text/plain", response)
	})

	jwtConf := middleware.JWTConfig{
		Claims:     &controllers.UserLoginClaim{},
		SigningKey: controllers.Secret,
	}
	r.Use(middleware.JWTWithConfig(jwtConf))

	e.Logger.Fatal(e.Start(config.Config.Hostname + ":" + strconv.Itoa(int(config.Config.Port))))
}
