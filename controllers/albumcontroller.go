package controllers

import (
	"net/http"
	"strconv"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/models"
	"github.com/labstack/echo"
	"github.com/trtstm/budgetr/log"
)

func StrToUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

type albumController struct {
}

func (c *albumController) Index(ctx echo.Context) error {
	albums := []models.Album{}
	if gormDB := db.DB.Find(&albums); gormDB.Error != nil {
		log.Errorf("AlbumController::Index Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"data": albums,
	})
}

func (c *albumController) Show(ctx echo.Context) error {
	id := StrToUint(ctx.QueryParam("id"))

	album := models.Album{}
	gormDB := db.DB.First(&album, "id = ?", id)
	if gormDB.RecordNotFound() {
		log.Debugf("AlbumController::Show Album '%d' not found.", id)
		return ctx.NoContent(http.StatusNotFound)
	} else if gormDB.Error != nil {
		log.Errorf("AlbumController::Show Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, album)
}

// AlbumController Contains the actions for the 'albums' endpoint.
var AlbumController albumController
