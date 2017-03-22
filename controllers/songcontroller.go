package controllers

import (
	"net/http"
	"time"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	"github.com/cadenzr/cadenzr/streamers"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

type songController struct {
}

func (c *songController) FileStream(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))
	song := &models.Song{}
	gormDB := db.DB.First(song, "id = ?", id)
	if gormDB.RecordNotFound() {
		log.Debugf("Could not start streaming because song '%d' is not in database.", id)
		return ctx.NoContent(http.StatusNotFound)
	} else if gormDB.Error != nil {
		log.Errorf("Could not start streaming. Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	streamer, err := streamers.NewFileStreamer(song.Path)
	if err != nil {
		log.Errorf("Could not create streamer: %v", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	defer streamer.Close()

	// Since we don't know when songs have been played from m3u8. We just update it at the start.
	if ctx.FormValue("from") == "m3u8" {
		db.DB.Table("songs").Where("id = ?", song.ID).Update("played", gorm.Expr("played+1"))
	}

	// TODO: set the correct time so browser can cache.
	http.ServeContent(ctx.Response(), ctx.Request(), song.Name, time.Time{}, streamer)
	return nil
}

func (c *songController) Played(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))

	if gormDB := db.DB.Table("songs").Where("id = ?", id).Update("played", gorm.Expr("played+1")); gormDB.Error != nil {
		log.Errorf("Failed to increment played cound for song '%d': %v", id, gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	} else if gormDB.RowsAffected == 0 {
		log.Debugf("Could not update song played count. Song '%d' was not found.", id)
		return ctx.NoContent(http.StatusNotFound)
	}

	return ctx.NoContent(http.StatusOK)
}

// SongController Contains the actions for the 'songs' endpoint.
var SongController songController
