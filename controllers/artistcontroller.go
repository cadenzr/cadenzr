package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	"github.com/labstack/echo"
)

type artistResponse struct {
	ID    uint            `json:"id"`
	Name  string          `json:"name"`
	Songs []*songResponse `json:"songs"`
}

func TransformArtists(artists ...*models.Artist) []*artistResponse {
	r := []*artistResponse{}

	for _, artist := range artists {
		r = append(r, TransformArtist(artist))
	}

	return r
}

func TransformArtist(artist *models.Artist) *artistResponse {
	r := &artistResponse{}
	r.ID = artist.ID
	r.Name = artist.Name

	if artist.Songs != nil {
		r.Songs = TransformSongs(artist.Songs...)
	}

	return r
}

type artistController struct {
}

func (c *artistController) Index(ctx echo.Context) error {
	artists := []*models.Artist{}
	if gormDB := db.DB.Preload("Songs").Preload("Songs.Album").Preload("Songs.Artist").Preload("Songs.Cover").Find(&artists); gormDB.Error != nil {
		log.Errorf("ArtistController::Index Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	data, err := json.MarshalIndent(artists, "", "\t")
	log.Println(err)
	log.Println("Artists: ", string(data))

	return ctx.JSON(http.StatusOK, echo.Map{
		"data": TransformArtists(artists...),
	})
}

func (c *artistController) Show(echo.Context) error {
	panic("Not implemented")
}

func (c *artistController) Create(echo.Context) error {
	panic("Not implemented")
}

func (c *artistController) Update(echo.Context) error {
	panic("Not implemented")
}

func (c *artistController) Delete(echo.Context) error {
	panic("Not implemented")
}

// ArtistController Contains the actions for the 'artists' endpoint.
var ArtistController artistController
