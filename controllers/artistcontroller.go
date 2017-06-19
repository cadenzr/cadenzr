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
	Albums []*albumResponse `json:"albums"`
}

func TransformArtists(artists ...*models.Artist) []*artistResponse {
	r := []*artistResponse{}

	for _, artist := range artists {
		r = append(r, TransformArtist(artist))
	}

	return r
}

func TransformSongsToAlbums(songs ...*models.Song) []*albumResponse {
	return nil
}


func TransformArtist(artist *models.Artist) *artistResponse {
	r := &artistResponse{}
	r.ID = artist.ID
	r.Name = artist.Name

	if artist.Albums != nil {
		r.Albums = TransformAlbums(artist.Albums...)
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
	
	/*
	for _,artist := range artists {
		db.DB.Preload("Songs").Preload("Songs.Artist").Where("Artist = ?", artist.Name).Find(&artist.Albums)
		log.Println(&artist.Albums)
	}*/
	/*
	if gormDB := db.DB.Joins("JOIN songs ON songs.artist_id = artists.id").Joins("JOIN albums ON albums.id = songs.album_id").Find(&artists); gormDB.Error != nil {
		log.Errorf("ArtistController::Index Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	*/
	
	
	for _,artist := range artists {
		if gormDB := db.DB.Preload("Songs").Preload("Songs.Artist").Joins("JOIN songs ON songs.album_id = albums.id").Where("songs.artist_id = ?", artist.ID).Find(&artist.Albums); gormDB.Error != nil {
			log.Errorf("ArtistController::Index Database failed: couldn't get Albums of Artist: %v", gormDB.Error)
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	log.Println(artists)

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
