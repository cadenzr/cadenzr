package controllers

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	"github.com/labstack/echo"
)

type songResponse struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	Artist   models.NullString  `json:"artist"`
	Album    models.NullString  `json:"album"`
	Year     models.NullInt64   `json:"year"`
	Genre    models.NullString  `json:"genre"`
	Duration models.NullFloat64 `json:"duration"`
	Mime     string             `json:"mime"`
	Cover    models.NullString  `json:"cover"`
	Played   uint               `json:"played"`
}

type imageResponse struct {
	ID   uint   `json:"id"`
	Path string `json:"path"`
	Link string `json:"link"`
	Mime string `json:"mime"`
	Hash string `json:"hash"`
}
type albumResponse struct {
	ID    uint              `json:"id"`
	Name  string            `json:"name"`
	Year  models.NullInt64  `json:"year"`
	Cover models.NullString `json:"cover"`
	Songs []*songResponse   `json:"songs"`
}

func TransformImage(image *models.Image) *imageResponse {
	return &imageResponse{
		ID:   image.ID,
		Path: image.Path,
		Link: image.Link,
		Mime: image.Mime,
		Hash: image.Hash,
	}
}

func TransformAlbum(album *models.Album) *albumResponse {
	r := &albumResponse{}
	r.ID = album.ID
	r.Name = album.Name
	r.Year = album.Year

	if album.Cover != nil {
		r.Cover.Set(album.Cover.Link)
	}

	if album.Songs != nil {
		r.Songs = TransformSongs(album.Songs...)
	}

	return r
}

func TransformAlbums(albums ...*models.Album) []*albumResponse {
	r := []*albumResponse{}

	for _, album := range albums {
		r = append(r, TransformAlbum(album))
	}

	return r
}

func TransFormSong(song *models.Song) *songResponse {
	r := &songResponse{}
	r.ID = song.ID
	r.Name = song.Name
	r.Year = song.Year
	r.Genre = song.Genre
	r.Duration = song.Duration
	r.Mime = song.Mime
	r.Played = song.Played

	if song.Artist != nil {
		r.Artist.Set(song.Artist.Name)
	}

	if song.Album != nil {
		r.Album.Set(song.Album.Name)
	}

	if song.Cover != nil {
		r.Cover.Set(song.Cover.Link)
	}

	return r
}

func TransformSongs(songs ...*models.Song) []*songResponse {
	r := []*songResponse{}

	for _, song := range songs {
		r = append(r, TransFormSong(song))
	}

	return r
}

// StrToUint does what the name implies.
func StrToUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

type albumController struct {
}

func (c *albumController) Index(ctx echo.Context) error {
	albums := []*models.Album{}
	if gormDB := db.DB.Preload("Songs").Preload("Cover").Preload("Songs.Album").Preload("Songs.Artist").Preload("Songs.Cover").Find(&albums); gormDB.Error != nil {
		log.Errorf("AlbumController::Index Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"data": TransformAlbums(albums...),
	})
}

func (c *albumController) Show(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))

	album := &models.Album{}
	gormDB := db.DB.Preload("Songs").Preload("Cover").Preload("Songs.Album").Preload("Songs.Artist").Preload("Songs.Cover").First(&album, "id = ?", id)
	if gormDB.RecordNotFound() {
		log.Debugf("AlbumController::Show Album '%d' not found.", id)
		return ctx.NoContent(http.StatusNotFound)
	} else if gormDB.Error != nil {
		log.Errorf("AlbumController::Show Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, TransformAlbum(album))
}

func (c *albumController) Download(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))

	album := &models.Album{}
	gormDB := db.DB.Preload("Songs").Preload("Cover").Preload("Songs.Album").Preload("Songs.Artist").Preload("Songs.Cover").First(&album, "id = ?", id)
	if gormDB.RecordNotFound() {
		log.Debugf("AlbumController::Download Album '%d' not found.", id)
		return ctx.NoContent(http.StatusNotFound)
	} else if gormDB.Error != nil {
		log.Errorf("AlbumController::Download Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// Create the zip archive containing songs
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	for _, song := range album.Songs {
		f, err := w.Create(filepath.Base(song.Path))
		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadFile(song.Path)
		if err != nil {
			log.Errorf("AlbumController::Download Couldn't read song path: %v", song.Path)
			return ctx.NoContent(http.StatusInternalServerError)
		}

		_, err = f.Write(body)
		if err != nil {
			log.Errorf("AlbumController::Download Couldn't write song to zip: %v", song.Path)
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	// Make sure to check the error on Close.
	err := w.Close()
	if err != nil {
		log.Errorf("AlbumController::Download Couldn't create archive for download: %v", err.Error())
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.Stream(http.StatusOK, "application/zip", buf)
}

// AlbumController Contains the actions for the 'albums' endpoint.
var AlbumController albumController
