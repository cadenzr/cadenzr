package main

import (
	"bytes"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/cadenzr/cadenzr/config"
	"github.com/cadenzr/cadenzr/controllers"
	"github.com/cadenzr/cadenzr/db"

	static "github.com/Code-Hex/echo-static"
	"github.com/cadenzr/cadenzr/models"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func assets() *assetfs.AssetFS {
	assetInfo := func(path string) (os.FileInfo, error) {
		return os.Stat(path)
	}
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: assetInfo, Prefix: k}
	}
	panic("unreachable")
}

func startAPI() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	e.Use(static.ServeRoot("/", assets()))
	e.Static("/images", "images")

	// Login route
	e.POST("/api/login", controllers.AuthController.Login)

	// Restricted group
	r := e.Group("/api")
	jwtConf := middleware.JWTConfig{
		Claims:     &controllers.UserLoginClaim{},
		SigningKey: controllers.Secret,
	}
	r.Use(middleware.JWTWithConfig(jwtConf))

	jwtConfQuery := jwtConf
	jwtConfQuery.TokenLookup = "query:token"
	rQuery := e.Group("/api")
	rQuery.Use(middleware.JWTWithConfig(jwtConfQuery))

	r.GET("/albums", controllers.AlbumController.Index)
	r.GET("/albums/:id", controllers.AlbumController.Show)
	rQuery.GET("/albums/:id/download", controllers.AlbumController.Download)
	r.GET("/playlists", controllers.PlaylistController.Index)
	r.POST("/playlists", controllers.PlaylistController.Create)
	r.DELETE("/playlists/:id/songs/:sid", controllers.PlaylistController.DeleteSong)
	r.POST("/playlists/:id/songs", controllers.PlaylistController.AddSongs)
	r.GET("/playlists/:id", controllers.PlaylistController.Show)
	r.DELETE("/playlists/:id", controllers.PlaylistController.Delete)

	e.GET("/api/songs/:id/stream", controllers.SongController.FileStream)
	e.POST("/api/songs/:id/played", controllers.SongController.Played)

	r.POST("/upload", upload)

	r.POST("/scan", func(c echo.Context) error {
		done := make(chan struct{})
		scanCh <- done

		<-done
		return c.NoContent(http.StatusOK)
	})

	rQuery.GET("/albums/:id/playlist.m3u8", func(c echo.Context) error {
		id := controllers.StrToUint(c.Param("id"))

		songs := []*models.Song{}
		db.DB.Find(&songs, "album_id = ?", id)

		endpoint := "http://" + config.Config.Hostname
		if config.Config.Hostname == "" {
			endpoint = "http://localhost"
		}
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

		c.Response().Header().Set("Content-Disposition", "attachment; filename=playlist.m3u8")
		return c.Stream(http.StatusOK, "application/x-mpegurl", response)
	})

	e.Logger.Fatal(e.Start(config.Config.Hostname + ":" + strconv.Itoa(int(config.Config.Port))))
}
